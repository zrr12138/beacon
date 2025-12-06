package beaconImp

import "gorm.io/gorm"

type User struct {
	gorm.Model        // 包含 ID, CreatedAt, UpdatedAt, DeletedAt
	Username   string `gorm:"unique"`
	Password   string
}

type Resource struct {
	gorm.Model
	UserID uint `gorm:"uniqueIndex"`
	Wood   int  `gorm:"default:1000"`
	Stone  int  `gorm:"default:150"`
	Gold   int  `gorm:"default:300"`
	Food   int  `gorm:"default:100"`
}
type BuildingType string

const (
	MainHall   BuildingType = "main_hall"
	Barracks   BuildingType = "barracks"
	Smithy     BuildingType = "smithy"
	Warehouse  BuildingType = "warehouse"
	LumberMill BuildingType = "lumber_mill"
	Quarry     BuildingType = "quarry"
)

type Building struct {
	gorm.Model
	UserID         uint         `gorm:"index"`
	Type           BuildingType `gorm:"not null"`
	Level          int          `gorm:"default:1"`
	IsUpgrading    bool         `gorm:"default:false"`
	UpgradeEndTime int64        `gorm:"default:0"`
}
