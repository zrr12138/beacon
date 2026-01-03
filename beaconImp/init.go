package beaconImp

import (
	"beacon/config"
	"beacon/log"
	"os"

	"github.com/gin-gonic/gin"
)

func (B *Beacon) Init() {
	// 加载配置
	if err := config.LoadConfig(); err != nil {
		log.Fatal("Failed to load config:", err)
	}
	log.Info("Config loaded successfully")

	// 创建快照目录
	if err := os.MkdirAll(snapshotDir, 0755); err != nil {
		log.Fatal("Failed to create snapshot dir:", err)
	}

	// 加载最新快照到内存
	if err := B.LoadLatestSnapshot(); err != nil {
		log.Fatal("Failed to load snapshot:", err)
	}
	// 统计所有城市的建筑总数
	totalBuildings := 0
	for _, city := range B.state.Cities {
		totalBuildings += len(city.GetAllBuildings())
	}

	log.Infof("Game state loaded: %d users, %d cities, %d buildings",
		len(B.state.Users), len(B.state.Cities), totalBuildings)

	// 初始化 Gin
	B.r = gin.Default()
	B.RegisterHttpHandler()

	// 启动后台工作线程
	B.StartWorker()
}
