package beaconImp

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Beacon struct {
	db *gorm.DB
	r  *gin.Engine
}
