package beaconImp

import (
	"beacon/log"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func (B *Beacon) Init() {
	var err error
	B.db, err = gorm.Open(sqlite.Open("beacon.db"), &gorm.Config{Logger: &log.Logger{}})
	if err != nil {
		log.Fatal("Failed to connect database:", err)
	}
	err = B.db.AutoMigrate(&User{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}
	B.r = gin.Default()
	B.r.LoadHTMLGlob("templates/*.html")
	B.RegisterHttpHandler()
}
