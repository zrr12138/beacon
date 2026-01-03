package beaconImp

import "errors"

// ========== GameState - 树型结构 ==========

// GameState 包含所有游戏数据
// 架构：User -> City -> Buildings/Troops/Queues（树型包含）
// 注意：玩家最近10秒内的操作可能在崩溃时丢失（设计权衡）
// 注意：不记录绝对时间戳，避免服务停止期间时间推进
type GameState struct {
	NextUserID uint             `json:"next_user_id"`
	NextCityID uint             `json:"next_city_id"`
	Users      map[string]*User `json:"users"`  // username -> User
	Cities     map[uint]*City   `json:"cities"` // cityID -> City
}

// NewGameState 创建初始空状态
func NewGameState() *GameState {
	return &GameState{
		NextUserID: 1,
		NextCityID: 1,
		Users:      make(map[string]*User),
		Cities:     make(map[uint]*City),
	}
}

// ========== User Methods ==========

// CreateUser 创建用户
func (gs *GameState) CreateUser(u *User) error {
	if _, exists := gs.Users[u.Username]; exists {
		return errors.New("user already exists")
	}
	u.ID = gs.NextUserID
	gs.NextUserID++
	u.CityIDs = []uint{} // 初始化空城池列表
	gs.Users[u.Username] = u
	return nil
}

// GetUserByUsername 查询用户
func (gs *GameState) GetUserByUsername(username string) (*User, error) {
	u, ok := gs.Users[username]
	if !ok {
		return nil, errors.New("user not found")
	}
	return u, nil
}

// ========== City Methods ==========

// CreateCity 创建城池
func (gs *GameState) CreateCity(c *City) error {
	c.ID = gs.NextCityID
	gs.NextCityID++

	// 初始化城池内部结构
	c.Troops = []*Troop{}
	c.BuildingUpgradeQueue = []*BuildingUpgradeQueue{}
	c.RecruitQueue = []*RecruitQueue{}

	// 查找城池所属用户
	var user *User
	for _, u := range gs.Users {
		if u.ID == c.UserID {
			user = u
			break
		}
	}

	gs.Cities[c.ID] = c

	// 添加到用户的城池列表
	if user != nil {
		user.CityIDs = append(user.CityIDs, c.ID)
	}

	return nil
}

// GetCity 获取城池
func (gs *GameState) GetCity(cityID uint) (*City, error) {
	c, ok := gs.Cities[cityID]
	if !ok {
		return nil, errors.New("city not found")
	}
	return c, nil
}

// ListCitiesByUser 获取用户所有城池
func (gs *GameState) ListCitiesByUser(userID uint) []*City {
	cities := make([]*City, 0)
	for _, city := range gs.Cities {
		if city.UserID == userID {
			cities = append(cities, city)
		}
	}
	return cities
}

// ========== Troop Methods (on City) ==========

// AddTroop 向城池添加部队（自动合并同类型）
func (c *City) AddTroop(troopType TroopType, quantity int) {
	for _, t := range c.Troops {
		if t.Type == troopType {
			t.Quantity += quantity
			return
		}
	}
	// 不存在则新建
	c.Troops = append(c.Troops, &Troop{
		Type:     troopType,
		Quantity: quantity,
	})
}

// GetTroop 获取城池的指定类型部队
func (c *City) GetTroop(troopType TroopType) *Troop {
	for _, t := range c.Troops {
		if t.Type == troopType {
			return t
		}
	}
	return nil
}

// ========== Building Queue Methods (on City) ==========

// AddBuildingUpgradeToQueue 添加建筑升级任务到队列
func (c *City) AddBuildingUpgradeToQueue(q *BuildingUpgradeQueue) {
	c.BuildingUpgradeQueue = append(c.BuildingUpgradeQueue, q)
}

// FinishCurrentBuildingUpgrade 完成当前建筑升级（队列第一个）
func (c *City) FinishCurrentBuildingUpgrade() {
	if len(c.BuildingUpgradeQueue) > 0 {
		queue := c.BuildingUpgradeQueue[0]
		// 升级建筑等级
		building := c.GetBuildingByType(queue.BuildingType)
		if building != nil {
			building.Level = queue.TargetLevel
		}
		// 移除队列第一个元素
		c.BuildingUpgradeQueue = c.BuildingUpgradeQueue[1:]
	}
}

// ========== Recruit Queue Methods (on City) ==========

// AddRecruitToQueue 添加招募任务到队列
func (c *City) AddRecruitToQueue(q *RecruitQueue) {
	c.RecruitQueue = append(c.RecruitQueue, q)
}

// CompleteCurrentRecruitUnit 完成当前招募的一个单位（队列第一个）
func (c *City) CompleteCurrentRecruitUnit() {
	if len(c.RecruitQueue) > 0 {
		queue := c.RecruitQueue[0]
		if queue.RemainingQty > 0 {
			c.AddTroop(queue.TroopType, 1)
			queue.RemainingQty--

			// 所有单位完成，移除队列第一个元素
			if queue.RemainingQty <= 0 {
				c.RecruitQueue = c.RecruitQueue[1:]
			} else {
				// 重置下一个单位的时间
				queue.RemainingTime = queue.TimePerUnit
			}
		}
	}
}
