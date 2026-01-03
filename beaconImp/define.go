package beaconImp

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type Beacon struct {
	state        *GameState
	r            *gin.Engine
	stateLock    sync.RWMutex // 全局游戏状态读写锁
	lastTickTime time.Time    // 上次tick时间（不持久化）
}

// ========== User ==========

type User struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	CityIDs  []uint `json:"city_ids"` // 玩家所有城池的ID列表
}

// ========== City（树型结构：包含建筑、部队、队列） ==========

type City struct {
	ID     uint   `json:"id"`
	UserID uint   `json:"user_id"`
	Name   string `json:"name"`
	PosX   int    `json:"pos_x"`
	PosY   int    `json:"pos_y"`

	// 资源
	Wood  int `json:"wood"`
	Stone int `json:"stone"`
	Iron  int `json:"iron"`
	Food  int `json:"food"`
	Gold  int `json:"gold"`

	// 资源累积器（浮点数，不足1时累积）
	WoodAcc  float64 `json:"wood_acc"`
	StoneAcc float64 `json:"stone_acc"`
	IronAcc  float64 `json:"iron_acc"`
	FoodAcc  float64 `json:"food_acc"`

	// 建筑（每个建筑独立成员）
	Government *BaseBuilding `json:"government"`
	Lumberyard *BaseBuilding `json:"lumberyard"`
	Quarry     *BaseBuilding `json:"quarry"`
	IronMine   *BaseBuilding `json:"iron_mine"`
	Farm       *BaseBuilding `json:"farm"`
	Warehouse  *BaseBuilding `json:"warehouse"`
	Barracks   *BaseBuilding `json:"barracks"`

	// 部队
	Troops []*Troop `json:"troops"`

	// 队列（允许多个任务排队，但同一时间只执行第一个）
	BuildingUpgradeQueue []*BuildingUpgradeQueue `json:"building_upgrade_queue"`
	RecruitQueue         []*RecruitQueue         `json:"recruit_queue"`
}

// ========== Building ==========

type BuildingType string

const (
	BuildingGovernment BuildingType = "government"
	BuildingLumberyard BuildingType = "lumberyard"
	BuildingQuarry     BuildingType = "quarry"
	BuildingIronMine   BuildingType = "iron_mine"
	BuildingFarm       BuildingType = "farm"
	BuildingWarehouse  BuildingType = "warehouse"
	BuildingBarracks   BuildingType = "barracks"
)

// BaseBuilding 建筑基础结构（城池内部，无需ID）
type BaseBuilding struct {
	Type  BuildingType `json:"type"`
	Level int          `json:"level"`
}

// ========== Troop ==========

type TroopType string

const (
	TroopSupplyCart    TroopType = "supply_cart"
	TroopLightInfantry TroopType = "light_infantry"
	TroopSpearman      TroopType = "spearman"
	TroopArcher        TroopType = "archer"
	TroopEliteInfantry TroopType = "elite_infantry"
	TroopCrossbowman   TroopType = "crossbowman"
	TroopLightCavalry  TroopType = "light_cavalry"
	TroopHeavyInfantry TroopType = "heavy_infantry"
	TroopHeavyCavalry  TroopType = "heavy_cavalry"
	TroopMarksman      TroopType = "marksman"
	TroopChariot       TroopType = "chariot"
	TroopCatapult      TroopType = "catapult"
)

// Troop 士兵（城池内部，无需ID）
type Troop struct {
	Type     TroopType `json:"type"`
	Quantity int       `json:"quantity"`
}

// ========== Queue Structs ==========

// BuildingUpgradeQueue 建筑升级队列
// 注意：使用相对剩余时间，避免服务停止期间时间推进
type BuildingUpgradeQueue struct {
	BuildingType   BuildingType `json:"building_type"`    // 要升级的建筑类型
	BuildingNameCN string       `json:"building_name_cn"` // 中文名称（前端显示）
	TargetLevel    int          `json:"target_level"`     // 升到的等级
	RemainingTime  float64      `json:"remaining_time"`   // 剩余时间（秒，浮点）
}

// RecruitQueue 招募队列
// 注意：使用相对剩余时间和剩余数量，逐个完成
type RecruitQueue struct {
	TroopType     TroopType `json:"troop_type"`
	TroopNameCN   string    `json:"troop_name_cn"`  // 中文名称（前端显示）
	TotalQuantity int       `json:"total_quantity"` // 总数量
	RemainingQty  int       `json:"remaining_qty"`  // 剩余数量
	TimePerUnit   float64   `json:"time_per_unit"`  // 单个招募时间（秒）
	RemainingTime float64   `json:"remaining_time"` // 当前单位剩余时间（秒）
}

// ========== City Helper Methods ==========

// GetBuildingByType 根据类型获取建筑指针
func (c *City) GetBuildingByType(buildingType BuildingType) *BaseBuilding {
	switch buildingType {
	case BuildingGovernment:
		return c.Government
	case BuildingLumberyard:
		return c.Lumberyard
	case BuildingQuarry:
		return c.Quarry
	case BuildingIronMine:
		return c.IronMine
	case BuildingFarm:
		return c.Farm
	case BuildingWarehouse:
		return c.Warehouse
	case BuildingBarracks:
		return c.Barracks
	default:
		return nil
	}
}

// GetAllBuildings 获取所有建筑的列表（用于遍历）
func (c *City) GetAllBuildings() []*BaseBuilding {
	return []*BaseBuilding{
		c.Government,
		c.Lumberyard,
		c.Quarry,
		c.IronMine,
		c.Farm,
		c.Warehouse,
		c.Barracks,
	}
}
