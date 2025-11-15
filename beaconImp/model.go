package beaconImp

import "gorm.io/gorm"

type User struct {
	gorm.Model        // 包含 ID, CreatedAt, UpdatedAt, DeletedAt
	Username   string `gorm:"unique"`
	Password   string
}
