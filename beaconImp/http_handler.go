package beaconImp

import (
	"beacon/common"
	"beacon/log"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var sessions = map[string]uint{}

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userName, err := c.Cookie("session_id")
		userId, ok := sessions[userName]
		if err != nil || !ok {
			log.Errorf("Auth failed, redirecting to login.")
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}
		log.Infof("Login success: username=%s, userId=%d", userName, userId)
		c.Set("userName", userName)
		c.Set("userId", userId)
		c.Next()
	}
}

func (B *Beacon) RegisterHttpHandler() {
	protected := B.r.Group("/beacon")
	protected.Use(authMiddleware())
	{
		protected.GET("/main", func(c *gin.Context) {
			username, _ := c.Get("userName")
			userIDVal, _ := c.Get("userId")

			var resource Resource
			if userID, ok := userIDVal.(uint); ok {
				result := B.db.FirstOrCreate(&resource, Resource{UserID: userID})
				if result.Error != nil {
					log.Errorf("Failed to load resource: %v", result.Error)
				}
				log.Infof("Loaded resource for user %d: wood=%d", userID, resource.Wood)
			}

			c.HTML(http.StatusOK, "main.html", gin.H{
				"UserName": username,
				"Resource": resource,
			})

			protected.GET("/architecture", func(c *gin.Context) {
				userIDVal, _ := c.Get("userId")
				var userID uint
				if uid, ok := userIDVal.(uint); ok {
					userID = uid
				}

				var resource Resource
				result := B.db.FirstOrCreate(&resource, Resource{UserID: userID})
				if result.Error != nil {
					log.Errorf("Failed to load resource: %v", result.Error)
				}

				var buildings []Building
				B.db.Where("user_id = ?", userID).Find(&buildings)

				c.HTML(http.StatusOK, "architecture.html", gin.H{
					"Resource":  resource,
					"Buildings": buildings,
				})
			})
		})
	}
	B.r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{})
	})
	B.r.GET("/register", func(c *gin.Context) {
		c.HTML(http.StatusOK, "register.html", gin.H{})
	})
	B.r.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", gin.H{})
	})
	B.r.POST("/register", func(c *gin.Context) {
		username := c.PostForm("username")
		password := c.PostForm("password")

		var existingUser User
		result := B.db.Where("username = ?", username).First(&existingUser)

		if result.Error == nil {
			c.HTML(http.StatusBadRequest, "register.html", gin.H{
				"Error": "用户名已存在",
			})
			return
		}

		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.HTML(http.StatusInternalServerError, "register.html", gin.H{
				"Error": "数据库错误",
			})
			return
		}

		hashedPassword, err := common.HashPassword(password)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "register.html", gin.H{
				"Error": "密码哈希失败",
			})
			return
		}

		user := User{Username: username, Password: hashedPassword}
		createResult := B.db.Create(&user)
		if createResult.Error != nil {
			c.HTML(http.StatusInternalServerError, "register.html", gin.H{
				"Error": "创建用户失败",
			})
			return
		}
		resource := Resource{UserID: user.ID}
		B.db.Create(&resource)

		buildingTypes := []BuildingType{
			MainHall, Barracks, Smithy,
			Warehouse, LumberMill, Quarry,
		}
		for _, bt := range buildingTypes {
			building := Building{
				UserID: user.ID,
				Type:   bt,
				Level:  0,
			}
			B.db.Create(&building)
		}

		log.Infof("New user registered: %s", username)
		c.Redirect(http.StatusFound, "/login")
	})

	B.r.POST("/login", func(c *gin.Context) {
		username := c.PostForm("username")
		password := c.PostForm("password")

		var user User
		result := B.db.Where("username = ?", username).First(&user)

		if result.Error != nil {
			log.Infof("Login attempt failed for %s: %s", username, result.Error)
			c.HTML(http.StatusUnauthorized, "login.html", gin.H{
				"Error": "用户名或密码错误",
			})
			return
		}

		if !common.CheckPasswordHash(password, user.Password) {
			log.Infof("Login attempt failed for %s: Invalid password", username)
			c.HTML(http.StatusUnauthorized, "login.html", gin.H{
				"Error": "用户名或密码错误",
			})
			return
		}

		sessions[user.Username] = user.ID
		c.SetCookie("session_id", user.Username, 3600, "/", "", false, true)
		log.Infof("User %s logged in.", username)
		c.Redirect(http.StatusFound, "/beacon/main")
	})
	B.r.GET("/logout", func(c *gin.Context) {
		userName, err := c.Cookie("session_id")
		if err == nil {
			delete(sessions, userName)
			log.Infof("User %s logged out.", userName)
		}
		c.SetCookie("session_id", "", -1, "/", "", false, true)
		c.Redirect(http.StatusFound, "/")
	})
	B.r.GET("/decision", func(c *gin.Context) {
		c.HTML(200, "decision.html", gin.H{})
	})
	B.r.GET("/map", func(c *gin.Context) {
		c.HTML(200, "map.html", gin.H{})
	})

}
