package beaconImp

import (
	"beacon/common"
	"beacon/config"
	"beacon/log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

var sessions = map[string]uint{}

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userName, err := c.Cookie("session_id")
		userId, ok := sessions[userName]
		if err != nil || !ok {
			log.Errorf("Auth failed, redirecting to login.")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权，请先登录"})
			c.Abort()
			return
		}
		log.Debugf("Login success: username=%s, userId=%d", userName, userId)
		c.Set("userName", userName)
		c.Set("userId", userId)
		c.Next()
	}
}

// parseCityID 从请求中解析 city_id 参数（支持 query 和 form）
func parseCityID(c *gin.Context) (uint, error) {
	cityIDStr := c.Query("city_id")
	if cityIDStr == "" {
		cityIDStr = c.PostForm("city_id")
	}
	if cityIDStr == "" {
		return 0, gin.Error{Err: nil, Type: gin.ErrorTypePublic, Meta: "缺少 city_id 参数"}
	}

	cityID, err := strconv.ParseUint(cityIDStr, 10, 32)
	if err != nil {
		return 0, gin.Error{Err: err, Type: gin.ErrorTypePublic, Meta: "city_id 格式错误"}
	}

	return uint(cityID), nil
}

// validateCityAccess 验证用户是否有权访问指定城市
func (B *Beacon) validateCityAccess(userID, cityID uint) (*City, error) {
	B.stateLock.RLock()
	defer B.stateLock.RUnlock()

	city, err := B.state.GetCity(cityID)
	if err != nil {
		return nil, gin.Error{Err: err, Type: gin.ErrorTypePublic, Meta: "城市不存在"}
	}

	if city.UserID != userID {
		return nil, gin.Error{Err: nil, Type: gin.ErrorTypePublic, Meta: "无权访问该城市"}
	}

	return city, nil
}

func (B *Beacon) RegisterHttpHandler() {
	// ========== API 路由组 ==========
	api := B.r.Group("/api")
	api.Use(authMiddleware())
	{
		// ========== 用户城市列表 ==========
		// GET /api/cities
		api.GET("/cities", func(c *gin.Context) {
			userIDVal, _ := c.Get("userId")
			userID := userIDVal.(uint)

			B.stateLock.RLock()
			cities := B.state.ListCitiesByUser(userID)
			B.stateLock.RUnlock()

			type CityBasicInfo struct {
				ID   uint   `json:"id"`
				Name string `json:"name"`
				PosX int    `json:"pos_x"`
				PosY int    `json:"pos_y"`
			}

			citiesInfo := make([]CityBasicInfo, 0, len(cities))
			for _, city := range cities {
				citiesInfo = append(citiesInfo, CityBasicInfo{
					ID:   city.ID,
					Name: city.Name,
					PosX: city.PosX,
					PosY: city.PosY,
				})
			}

			c.JSON(http.StatusOK, gin.H{
				"cities": citiesInfo,
			})
		})

		// ========== 城市基本信息 ==========
		// G/info?city_id=1
		api.GET("/city/info", func(c *gin.Context) {
			userIDVal, _ := c.Get("userId")
			userID := userIDVal.(uint)

			cityID, err := parseCityID(c)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "缺少或无效的 city_id"})
				return
			}

			city, err := B.validateCityAccess(userID, cityID)
			if err != nil {
				c.JSON(http.StatusForbidden, gin.H{"error": "无权访问该城市"})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"id":    city.ID,
				"name":  city.Name,
				"pos_x": city.PosX,
				"pos_y": city.PosY,
			})
		})

		// ========== 原子化API：资源查询 ==========
		// GET /api/resources?city_id=1
		api.GET("/resources", func(c *gin.Context) {
			userIDVal, _ := c.Get("userId")
			userID := userIDVal.(uint)

			cityID, err := parseCityID(c)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "缺少或无效的 city_id"})
				return
			}

			city, err := B.validateCityAccess(userID, cityID)
			if err != nil {
				c.JSON(http.StatusForbidden, gin.H{"error": "无权访问该城市"})
				return
			}

			// 获取仓库容量
			capacity := 0
			if city.Warehouse != nil {
				conf := config.GetBuildingLevel("warehouse", city.Warehouse.Level)
				if conf != nil {
					capacity = conf.Capacity
				}
			}

			c.JSON(http.StatusOK, gin.H{
				"city_id":  city.ID,
				"wood":     city.Wood,
				"stone":    city.Stone,
				"iron":     city.Iron,
				"food":     city.Food,
				"gold":     city.Gold,
				"capacity": capacity,
			})
		})

		// ========== 原子化API：兵力查询 ==========
		// GET /api/troops?city_id=1
		api.GET("/troops", func(c *gin.Context) {
			userIDVal, _ := c.Get("userId")
			userID := userIDVal.(uint)

			cityID, err := parseCityID(c)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "缺少或无效的 city_id"})
				return
			}

			city, err := B.validateCityAccess(userID, cityID)
			if err != nil {
				c.JSON(http.StatusForbidden, gin.H{"error": "无权访问该城市"})
				return
			}

			type TroopDisplay struct {
				Type     string `json:"type"`
				NameCN   string `json:"name_cn"`
				Quantity int    `json:"quantity"`
			}

			var troops []TroopDisplay
			for _, t := range city.Troops {
				troops = append(troops, TroopDisplay{
					Type:     string(t.Type),
					NameCN:   GetTroopNameCN(t.Type),
					Quantity: t.Quantity,
				})
			}

			c.JSON(http.StatusOK, gin.H{
				"city_id": city.ID,
				"troops":  troops,
			})
		})

		// ========== 原子化API：建筑升级队列查询 ==========
		// GET /api/building-queue?city_id=1
		api.GET("/building-queue", func(c *gin.Context) {
			userIDVal, _ := c.Get("userId")
			userID := userIDVal.(uint)

			cityID, err := parseCityID(c)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "缺少或无效的 city_id"})
				return
			}

			city, err := B.validateCityAccess(userID, cityID)
			if err != nil {
				c.JSON(http.StatusForbidden, gin.H{"error": "无权访问该城市"})
				return
			}

			type BuildingQueueDisplay struct {
				BuildingType   string  `json:"building_type"`
				BuildingNameCN string  `json:"building_name_cn"`
				TargetLevel    int     `json:"target_level"`
				RemainingTime  float64 `json:"remaining_time"`
			}

			var buildingQueue []BuildingQueueDisplay
			for _, queue := range city.BuildingUpgradeQueue {
				buildingQueue = append(buildingQueue, BuildingQueueDisplay{
					BuildingType:   string(queue.BuildingType),
					BuildingNameCN: queue.BuildingNameCN,
					TargetLevel:    queue.TargetLevel,
					RemainingTime:  queue.RemainingTime,
				})
			}

			c.JSON(http.StatusOK, gin.H{
				"city_id":        city.ID,
				"building_queue": buildingQueue,
			})
		})

		// ========== 原子化API：招募队列查询 ==========
		// GET /api/recruit-queue?city_id=1
		api.GET("/recruit-queue", func(c *gin.Context) {
			userIDVal, _ := c.Get("userId")
			userID := userIDVal.(uint)

			cityID, err := parseCityID(c)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "缺少或无效的 city_id"})
				return
			}

			city, err := B.validateCityAccess(userID, cityID)
			if err != nil {
				c.JSON(http.StatusForbidden, gin.H{"error": "无权访问该城市"})
				return
			}

			type RecruitQueueDisplay struct {
				TroopType     string  `json:"troop_type"`
				TroopNameCN   string  `json:"troop_name_cn"`
				TotalQuantity int     `json:"total_quantity"`
				RemainingQty  int     `json:"remaining_qty"`
				TimePerUnit   float64 `json:"time_per_unit"`
				RemainingTime float64 `json:"remaining_time"`
			}

			var recruitQueue []RecruitQueueDisplay
			for _, queue := range city.RecruitQueue {
				recruitQueue = append(recruitQueue, RecruitQueueDisplay{
					TroopType:     string(queue.TroopType),
					TroopNameCN:   queue.TroopNameCN,
					TotalQuantity: queue.TotalQuantity,
					RemainingQty:  queue.RemainingQty,
					TimePerUnit:   queue.TimePerUnit,
					RemainingTime: queue.RemainingTime,
				})
			}

			c.JSON(http.StatusOK, gin.H{
				"city_id":       city.ID,
				"recruit_queue": recruitQueue,
			})
		})

		// ========== 原子化API：单个建筑信息查询 ==========
		// GET /api/building/government?city_id=1
		api.GET("/building/government", func(c *gin.Context) {
			B.getBuildingInfo(c, BuildingGovernment)
		})

		api.GET("/building/lumberyard", func(c *gin.Context) {
			B.getBuildingInfo(c, BuildingLumberyard)
		})

		api.GET("/building/quarry", func(c *gin.Context) {
			B.getBuildingInfo(c, BuildingQuarry)
		})

		api.GET("/building/iron_mine", func(c *gin.Context) {
			B.getBuildingInfo(c, BuildingIronMine)
		})

		api.GET("/building/farm", func(c *gin.Context) {
			B.getBuildingInfo(c, BuildingFarm)
		})

		api.GET("/building/warehouse", func(c *gin.Context) {
			B.getBuildingInfo(c, BuildingWarehouse)
		})

		api.GET("/building/barracks", func(c *gin.Context) {
			B.getBuildingInfo(c, BuildingBarracks)
		})

		// ========== 原子化API：所有建筑列表 ==========
		// GET /api/buildings?city_id=1
		api.GET("/buildings", func(c *gin.Context) {
			userIDVal, _ := c.Get("userId")
			userID := userIDVal.(uint)

			cityID, err := parseCityID(c)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "缺少或无效的 city_id"})
				return
			}

			city, err := B.validateCityAccess(userID, cityID)
			if err != nil {
				c.JSON(http.StatusForbidden, gin.H{"error": "无权访问该城市"})
				return
			}

			type BuildingDisplay struct {
				Type          string                    `json:"type"`
				NameCN        string                    `json:"name_cn"`
				Level         int                       `json:"level"`
				CurrentEffect string                    `json:"current_effect"`
				NextEffect    string                    `json:"next_effect"`
				NextLevelConf *config.BuildingLevelConf `json:"next_level_conf"`
				IsUpgrading   bool                      `json:"is_upgrading"`
			}

			displayBuildings := make([]BuildingDisplay, 0, 7)

			// 检查哪些建筑在升级中
			upgradingBuildings := make(map[BuildingType]bool)
			for _, queue := range city.BuildingUpgradeQueue {
				upgradingBuildings[queue.BuildingType] = true
			}

			// 遍历所有建筑
			for _, b := range city.GetAllBuildings() {
				if b == nil {
					continue
				}
				currentConf := config.GetBuildingLevel(string(b.Type), b.Level)
				nextConf := config.GetBuildingLevel(string(b.Type), b.Level+1)

				display := BuildingDisplay{
					Type:          string(b.Type),
					NameCN:        GetBuildingNameCN(b.Type),
					Level:         b.Level,
					NextLevelConf: nextConf,
					IsUpgrading:   upgradingBuildings[b.Type],
				}

				if currentConf != nil {
					if currentConf.ProductionPerHour > 0 {
						display.CurrentEffect = strconv.Itoa(currentConf.ProductionPerHour) + "/小时"
					} else if currentConf.Capacity > 0 {
						display.CurrentEffect = "容量: " + strconv.Itoa(currentConf.Capacity)
					} else if currentConf.BuildSpeedBoost > 0 {
						display.CurrentEffect = "建造加速: " + strconv.Itoa(currentConf.BuildSpeedBoost) + "%"
					} else if currentConf.RecruitSpeedBoost > 0 {
						display.CurrentEffect = "招募加速: " + strconv.Itoa(currentConf.RecruitSpeedBoost) + "%"
					}
				}

				if nextConf != nil {
					if nextConf.ProductionPerHour > 0 {
						display.NextEffect = strconv.Itoa(nextConf.ProductionPerHour) + "/小时"
					} else if nextConf.Capacity > 0 {
						display.NextEffect = "容量: " + strconv.Itoa(nextConf.Capacity)
					} else if nextConf.BuildSpeedBoost > 0 {
						display.NextEffect = "建造加速: " + strconv.Itoa(nextConf.BuildSpeedBoost) + "%"
					} else if nextConf.RecruitSpeedBoost > 0 {
						display.NextEffect = "招募加速: " + strconv.Itoa(nextConf.RecruitSpeedBoost) + "%"
					}
				}

				displayBuildings = append(displayBuildings, display)
			}

			c.JSON(http.StatusOK, gin.H{
				"city_id":   city.ID,
				"buildings": displayBuildings,
			})
		})

		// ========== 建筑升级 ==========
		// POST /api/building/upgrade
		// Form: city_id, building_type
		api.POST("/building/upgrade", func(c *gin.Context) {
			userIDVal, _ := c.Get("userId")
			userID := userIDVal.(uint)

			cityID, err := parseCityID(c)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "缺少或无效的 city_id"})
				return
			}

			buildingTypeStr := c.PostForm("building_type")
			if buildingTypeStr == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 building_type"})
				return
			}
			buildingType := BuildingType(buildingTypeStr)

			B.stateLock.Lock()
			defer B.stateLock.Unlock()

			city, err := B.state.GetCity(cityID)
			if err != nil || city.UserID != userID {
				c.JSON(http.StatusForbidden, gin.H{"error": "无权访问该城市"})
				return
			}

			// 不再检查队列是否为空，允许多个任务排队

			// 查找指定类型的建筑
			building := city.GetBuildingByType(buildingType)

			if building == nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "建筑不存在"})
				return
			}

			// 获取升级配置
			nextConf := config.GetBuildingLevel(string(building.Type), building.Level+1)
			if nextConf == nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "已达最高等级"})
				return
			}

			// 检查资源
			if city.Wood < nextConf.UpgradeCostWood ||
				city.Stone < nextConf.UpgradeCostStone ||
				city.Iron < nextConf.UpgradeCostIron ||
				city.Food < nextConf.UpgradeCostFood ||
				city.Gold < nextConf.UpgradeCostGold {
				c.JSON(http.StatusBadRequest, gin.H{"error": "资源不足"})
				return
			}

			// 扣除资源
			city.Wood -= nextConf.UpgradeCostWood
			city.Stone -= nextConf.UpgradeCostStone
			city.Iron -= nextConf.UpgradeCostIron
			city.Food -= nextConf.UpgradeCostFood
			city.Gold -= nextConf.UpgradeCostGold

			// 创建升级队列并添加到队列中
			queue := &BuildingUpgradeQueue{
				BuildingType:   building.Type,
				BuildingNameCN: GetBuildingNameCN(building.Type),
				TargetLevel:    building.Level + 1,
				RemainingTime:  float64(nextConf.UpgradeTimeSeconds),
			}
			city.AddBuildingUpgradeToQueue(queue)

			log.Infof("Building upgrade queued: city=%d, building=%s, level=%d->%d, time=%.0fs",
				city.ID, queue.BuildingNameCN, building.Level, queue.TargetLevel, queue.RemainingTime)

			c.JSON(http.StatusOK, gin.H{"success": true})
		})

		// ========== 招募列表 ==========
		// GET /api/recruit/list
		api.GET("/recruit/list", func(c *gin.Context) {
			// 转换为 TroopAttr 数组以获得正确的 JSON 标签
			troops := make([]*config.TroopAttr, 0, len(config.TroopConfig.Troops))
			for i := range config.TroopConfig.Troops {
				t := &config.TroopConfig.Troops[i]
				troops = append(troops, &config.TroopAttr{
					Type:               t.Type,
					Name:               t.Name,
					MeleeAttack:        t.MeleeAttack,
					RangedAttack:       t.RangedAttack,
					MeleeDefense:       t.MeleeDefense,
					RangedDefense:      t.RangedDefense,
					Speed:              t.Speed,
					Capacity:           t.Capacity,
					FoodConsumption:    t.FoodConsumption,
					RecruitTimeSeconds: t.RecruitTimeSeconds,
					RecruitCostWood:    t.RecruitCostWood,
					RecruitCostIron:    t.RecruitCostIron,
					RecruitCostFood:    t.RecruitCostFood,
					RecruitCostStone:   t.RecruitCostStone,
				})
			}
			c.JSON(http.StatusOK, gin.H{
				"troops": troops,
			})
		})

		// ========== 招募确认 ==========
		// POST /api/recruit/confirm
		// Form: city_id, troop_type, quantity
		api.POST("/recruit/confirm", func(c *gin.Context) {
			userIDVal, _ := c.Get("userId")
			userID := userIDVal.(uint)

			cityID, err := parseCityID(c)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "缺少或无效的 city_id"})
				return
			}

			troopType := c.PostForm("troop_type")
			quantityStr := c.PostForm("quantity")
			quantity, _ := strconv.Atoi(quantityStr)

			if quantity <= 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "数量必须大于0"})
				return
			}

			troopConf := config.GetTroopConfig(troopType)
			if troopConf == nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "兵种不存在"})
				return
			}

			B.stateLock.Lock()
			defer B.stateLock.Unlock()

			city, err := B.state.GetCity(cityID)
			if err != nil || city.UserID != userID {
				c.JSON(http.StatusForbidden, gin.H{"error": "无权访问该城市"})
				return
			}

			// 不再检查队列是否为空，允许多个任务排队

			// 根据配置计算资源消耗
			costWood := quantity * troopConf.RecruitCostWood
			costStone := quantity * troopConf.RecruitCostStone
			costIron := quantity * troopConf.RecruitCostIron
			costFood := quantity * troopConf.RecruitCostFood
			//log.Infof("debug: %s %s %+v %d %d %d %d", troopType, GetTroopNameCN(TroopType(troopType)), troopConf, costWood, costStone, costIron, costFood)
			if city.Wood < costWood || city.Stone < costStone ||
				city.Iron < costIron || city.Food < costFood {
				c.JSON(http.StatusBadRequest, gin.H{"error": "资源不足"})
				return
			}
			//log.Infof("debug: %d %d %d %d", city.Wood, city.Stone, city.Iron, city.Food)
			city.Wood -= costWood
			city.Stone -= costStone
			city.Iron -= costIron
			city.Food -= costFood
			//log.Infof("+debug: %d %d %d %d", city.Wood, city.Stone, city.Iron, city.Food)
			// 创建招募队列
			queue := &RecruitQueue{
				TroopType:     TroopType(troopType),
				TroopNameCN:   troopConf.Name,
				TotalQuantity: quantity,
				RemainingQty:  quantity,
				TimePerUnit:   float64(troopConf.RecruitTimeSeconds),
				RemainingTime: float64(troopConf.RecruitTimeSeconds),
			}
			city.AddRecruitToQueue(queue)

			log.Infof("Recruit queued: city=%d, type=%s, qty=%d, time_per_unit=%.0fs",
				city.ID, queue.TroopNameCN, quantity, queue.TimePerUnit)

			c.JSON(http.StatusOK, gin.H{"success": true})
		})
	}

	// ========== 静态页面路由（不需要认证） ==========
	B.r.GET("/", func(c *gin.Context) {
		c.File("./static/index.html")
	})

	B.r.GET("/register", func(c *gin.Context) {
		c.File("./static/register.html")
	})

	B.r.GET("/login", func(c *gin.Context) {
		c.File("./static/login.html")
	})

	// ========== 需要认证的静态页面 ==========
	protected := B.r.Group("")
	protected.Use(func(c *gin.Context) {
		userName, err := c.Cookie("session_id")
		_, ok := sessions[userName]
		if err != nil || !ok {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}
		c.Next()
	})
	{
		protected.GET("/main", func(c *gin.Context) {
			c.File("./static/main.html")
		})

		protected.GET("/buildings", func(c *gin.Context) {
			c.File("./static/buildings.html")
		})

		protected.GET("/recruit", func(c *gin.Context) {
			c.File("./static/recruit.html")
		})

		protected.GET("/map", func(c *gin.Context) {
			c.File("./static/map.html")
		})
	}

	// ========== 认证相关 API ==========
	B.r.POST("/api/register", func(c *gin.Context) {
		username := c.PostForm("username")
		password := c.PostForm("password")

		B.stateLock.Lock()
		defer B.stateLock.Unlock()

		// 检查用户名是否存在
		if _, err := B.state.GetUserByUsername(username); err == nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "用户名已存在",
			})
			return
		}

		// 哈希密码
		hashedPassword, err := common.HashPassword(password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "密码哈希失败",
			})
			return
		}

		// 创建用户
		user := &User{Username: username, Password: hashedPassword}
		if err := B.state.CreateUser(user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "创建用户失败",
			})
			return
		}

		// 创建默认城池
		city := &City{
			UserID: user.ID,
			Name:   "我的城池",
			PosX:   100,
			PosY:   100,
			Wood:   5000,
			Stone:  5000,
			Iron:   3000,
			Food:   5000,
			Gold:   1000,
		}

		// 创建初始建筑（所有建筑初始等级1）
		city.Government = &BaseBuilding{Type: BuildingGovernment, Level: 1}
		city.Lumberyard = &BaseBuilding{Type: BuildingLumberyard, Level: 1}
		city.Quarry = &BaseBuilding{Type: BuildingQuarry, Level: 1}
		city.IronMine = &BaseBuilding{Type: BuildingIronMine, Level: 1}
		city.Farm = &BaseBuilding{Type: BuildingFarm, Level: 1}
		city.Warehouse = &BaseBuilding{Type: BuildingWarehouse, Level: 1}
		city.Barracks = &BaseBuilding{Type: BuildingBarracks, Level: 1}

		B.state.CreateCity(city)

		log.Infof("New user registered: %s (ID=%d)", username, user.ID)
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "注册成功",
		})
	})

	B.r.POST("/api/login", func(c *gin.Context) {
		username := c.PostForm("username")
		password := c.PostForm("password")

		B.stateLock.RLock()
		user, err := B.state.GetUserByUsername(username)
		B.stateLock.RUnlock()

		if err != nil {
			log.Infof("Login attempt failed for %s: %s", username, err)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "用户名或密码错误",
			})
			return
		}

		if !common.CheckPasswordHash(password, user.Password) {
			log.Infof("Login attempt failed for %s: Invalid password", username)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "用户名或密码错误",
			})
			return
		}

		sessions[user.Username] = user.ID
		c.SetCookie("session_id", user.Username, 3600, "/", "", false, true)
		log.Infof("User %s logged in.", username)
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "登录成功",
		})
	})

	B.r.GET("/api/logout", func(c *gin.Context) {
		userName, err := c.Cookie("session_id")
		if err == nil {
			delete(sessions, userName)
			log.Infof("User %s logged out.", userName)
		}
		c.SetCookie("session_id", "", -1, "/", "", false, true)
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "登出成功",
		})
	})

	// ========== 静态资源服务 ==========
	B.r.Static("/static", "./static")
}

func (B *Beacon) getBuildingInfo(c *gin.Context, buildingType BuildingType) {
	userIDVal, _ := c.Get("userId")
	userID := userIDVal.(uint)

	cityID, err := parseCityID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少或无效的 city_id"})
		return
	}

	city, err := B.validateCityAccess(userID, cityID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权访问该城市"})
		return
	}

	// 查找指定类型的建筑
	building := city.GetBuildingByType(buildingType)

	if building == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "建筑不存在"})
		return
	}

	currentConf := config.GetBuildingLevel(string(building.Type), building.Level)
	nextConf := config.GetBuildingLevel(string(building.Type), building.Level+1)

	// 检查是否在升级中
	isUpgrading := false
	for _, queue := range city.BuildingUpgradeQueue {
		if queue.BuildingType == buildingType {
			isUpgrading = true
			break
		}
	}

	type BuildingInfo struct {
		Type        string                    `json:"type"`
		NameCN      string                    `json:"name_cn"`
		Level       int                       `json:"level"`
		CurrentConf *config.BuildingLevelConf `json:"current_conf"`
		NextConf    *config.BuildingLevelConf `json:"next_conf"`
		IsUpgrading bool                      `json:"is_upgrading"`
	}

	c.JSON(http.StatusOK, gin.H{
		"city_id": city.ID,
		"building": BuildingInfo{
			Type:        string(building.Type),
			NameCN:      GetBuildingNameCN(building.Type),
			Level:       building.Level,
			CurrentConf: currentConf,
			NextConf:    nextConf,
			IsUpgrading: isUpgrading,
		},
	})
}
