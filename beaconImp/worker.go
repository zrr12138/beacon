package beaconImp

import (
	"beacon/config"
	"beacon/log"
	"time"
)

// Start 启动HTTP服务
func (B *Beacon) Start() {
	err := B.r.Run(":8000")
	if err != nil {
		log.Fatal(err)
	}
	log.Info("Beacon started on port 8000")
}

// StartWorker 启动后台工作线程（动态tick间隔 + 每10秒快照）
func (B *Beacon) StartWorker() {
	// 初始化上次tick时间
	B.lastTickTime = time.Now()

	// 游戏逻辑 tick（每秒，但会根据实际时间差计算）
	gameTicker := time.NewTicker(1 * time.Second)
	go func() {
		for range gameTicker.C {
			B.tick()
		}
	}()

	// 快照持久化（每10秒）
	snapshotTicker := time.NewTicker(10 * time.Second)
	go func() {
		for range snapshotTicker.C {
			B.saveSnapshotTask()
		}
	}()

	log.Info("Background worker started: game tick every 1s, snapshot every 10s")
}

// saveSnapshotTask 定期保存快照（持有读锁）
func (B *Beacon) saveSnapshotTask() {
	B.stateLock.RLock()
	defer B.stateLock.RUnlock()

	if err := B.SaveSnapshot(); err != nil {
		log.Errorf("Failed to save snapshot: %v", err)
	}
}

// tick 每秒执行的游戏逻辑（持有写锁）
func (B *Beacon) tick() {
	B.stateLock.Lock()
	defer B.stateLock.Unlock()

	// 计算实际时间差（秒，浮点）
	now := time.Now()
	deltaSeconds := now.Sub(B.lastTickTime).Seconds()
	B.lastTickTime = now

	// 遍历所有城池，更新状态
	for _, city := range B.state.Cities {
		// 1. 更新资源产出
		B.updateCityResources(city, deltaSeconds)

		// 2. 处理建筑升级队列（只处理第一个）
		B.processCityBuildingUpgrade(city, deltaSeconds)

		// 3. 处理招募队列（只处理第一个）
		B.processCityRecruit(city, deltaSeconds)
	}
}

// updateCityResources 更新城池资源（基于实际时间差，使用浮点累积）
func (B *Beacon) updateCityResources(city *City, deltaSeconds float64) {
	// 直接获取4个资源生产建筑
	lumberyardConf := config.GetBuildingLevel(string(BuildingLumberyard), city.Lumberyard.Level)
	quarryConf := config.GetBuildingLevel(string(BuildingQuarry), city.Quarry.Level)
	ironMineConf := config.GetBuildingLevel(string(BuildingIronMine), city.IronMine.Level)
	farmConf := config.GetBuildingLevel(string(BuildingFarm), city.Farm.Level)

	var woodRate, stoneRate, ironRate, foodRate float64

	// 计算每秒产量 = 每小时产量 / 3600
	if lumberyardConf != nil {
		woodRate = float64(lumberyardConf.ProductionPerHour) / 3600.0
	}
	if quarryConf != nil {
		stoneRate = float64(quarryConf.ProductionPerHour) / 3600.0
	}
	if ironMineConf != nil {
		ironRate = float64(ironMineConf.ProductionPerHour) / 3600.0
	}
	if farmConf != nil {
		foodRate = float64(farmConf.ProductionPerHour) / 3600.0
	}

	// 累积资源（浮点数）
	city.WoodAcc += woodRate * deltaSeconds
	city.StoneAcc += stoneRate * deltaSeconds
	city.IronAcc += ironRate * deltaSeconds
	city.FoodAcc += foodRate * deltaSeconds

	// 转换为整数资源
	if city.WoodAcc >= 1.0 {
		toAdd := int(city.WoodAcc)
		city.Wood += toAdd
		city.WoodAcc -= float64(toAdd)
	}
	if city.StoneAcc >= 1.0 {
		toAdd := int(city.StoneAcc)
		city.Stone += toAdd
		city.StoneAcc -= float64(toAdd)
	}
	if city.IronAcc >= 1.0 {
		toAdd := int(city.IronAcc)
		city.Iron += toAdd
		city.IronAcc -= float64(toAdd)
	}
	if city.FoodAcc >= 1.0 {
		toAdd := int(city.FoodAcc)
		city.Food += toAdd
		city.FoodAcc -= float64(toAdd)
	}

	// 检查仓库容量上限
	B.applyCityResourceCap(city)
}

// applyCityResourceCap 应用仓库容量上限
func (B *Beacon) applyCityResourceCap(city *City) {
	// 获取仓库建筑
	warehouse := city.Warehouse
	if warehouse == nil {
		return
	}

	warehouseConf := config.GetBuildingLevel("warehouse", warehouse.Level)
	if warehouseConf == nil {
		return
	}

	capacity := warehouseConf.Capacity
	if city.Wood > capacity {
		city.Wood = capacity
	}
	if city.Stone > capacity {
		city.Stone = capacity
	}
	if city.Iron > capacity {
		city.Iron = capacity
	}
	if city.Food > capacity {
		city.Food = capacity
	}
}

// processCityBuildingUpgrade 处理城池的建筑升级队列（只处理第一个任务）
func (B *Beacon) processCityBuildingUpgrade(city *City, deltaSeconds float64) {
	if len(city.BuildingUpgradeQueue) == 0 {
		return
	}

	// 只处理队列第一个任务
	queue := city.BuildingUpgradeQueue[0]

	// 递减剩余时间
	queue.RemainingTime -= deltaSeconds

	if queue.RemainingTime <= 0 {
		// 升级完成
		city.FinishCurrentBuildingUpgrade()
		log.Infof("Building upgrade completed: city=%d, type=%s, level=%d",
			city.ID, queue.BuildingNameCN, queue.TargetLevel)
	}
}

// processCityRecruit 处理城池的招募队列（只处理第一个任务，逐个完成士兵）
func (B *Beacon) processCityRecruit(city *City, deltaSeconds float64) {
	if len(city.RecruitQueue) == 0 {
		return
	}

	// 只处理队列第一个任务
	queue := city.RecruitQueue[0]
	if queue.RemainingQty <= 0 {
		return
	}

	// 递减当前单位剩余时间
	queue.RemainingTime -= deltaSeconds

	if queue.RemainingTime <= 0 {
		// 完成一个单位
		city.CompleteCurrentRecruitUnit()
		log.Debugf("Recruit unit completed: city=%d, type=%s, remaining=%d",
			city.ID, queue.TroopNameCN, queue.RemainingQty)

		if queue.RemainingQty <= 0 {
			log.Infof("Recruit queue completed: city=%d, type=%s",
				city.ID, queue.TroopNameCN)
		}
	}
}
