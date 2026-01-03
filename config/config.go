package config

import (
	"os"
	"sync"

	"github.com/pelletier/go-toml/v2"
)

// 全局配置实例
var (
	BuildingConfig *BuildingConf
	TroopConfig    *TroopConf
	once           sync.Once
)

// BuildingLevelConf 建筑等级配置
type BuildingLevelConf struct {
	Level              int `toml:"level" json:"level"`
	ProductionPerHour  int `toml:"production_per_hour" json:"production_per_hour"` // 资源产量/小时
	Capacity           int `toml:"capacity" json:"capacity"`                       // 容量（仓库）
	BuildSpeedBoost    int `toml:"build_speed_boost" json:"build_speed_boost"`     // 建造加速百分比（官府）
	RecruitSpeedBoost  int `toml:"recruit_speed_boost" json:"recruit_speed_boost"` // 招募加速百分比（兵营）
	UpgradeTimeSeconds int `toml:"upgrade_time_seconds" json:"upgrade_time_seconds"`
	UpgradeCostWood    int `toml:"upgrade_cost_wood" json:"upgrade_cost_wood"`
	UpgradeCostStone   int `toml:"upgrade_cost_stone" json:"upgrade_cost_stone"`
	UpgradeCostIron    int `toml:"upgrade_cost_iron" json:"upgrade_cost_iron"`
	UpgradeCostFood    int `toml:"upgrade_cost_food" json:"upgrade_cost_food"`
	UpgradeCostGold    int `toml:"upgrade_cost_gold" json:"upgrade_cost_gold"`
	PopulationCost     int `toml:"population_cost" json:"population_cost"`
}

// BuildingConf 建筑配置
type BuildingConf struct {
	Building map[string]struct {
		InitialLevel int                 `toml:"initial_level"`
		MaxLevel     int                 `toml:"max_level"`
		Levels       []BuildingLevelConf `toml:"levels"`
	} `toml:"building"`
}

// TroopConf 部队配置
type TroopConf struct {
	Troops []struct {
		Type               string `toml:"type"`
		Name               string `toml:"name"`
		MeleeAttack        int    `toml:"melee_attack"`
		RangedAttack       int    `toml:"ranged_attack"`
		MeleeDefense       int    `toml:"melee_defense"`
		RangedDefense      int    `toml:"ranged_defense"`
		Speed              int    `toml:"speed"`
		Capacity           int    `toml:"capacity"`
		FoodConsumption    int    `toml:"food_consumption"`
		RecruitTimeSeconds int    `toml:"recruit_time_seconds"`
		RecruitCostWood    int    `toml:"recruit_cost_wood"`
		RecruitCostIron    int    `toml:"recruit_cost_iron"`
		RecruitCostFood    int    `toml:"recruit_cost_food"`
		RecruitCostStone   int    `toml:"recruit_cost_stone"`
	} `toml:"troop"`
}

// LoadConfig 加载所有配置（启动时调用一次）
func LoadConfig() error {
	var loadErr error
	once.Do(func() {
		// 加载建筑配置
		buildingData, err := os.ReadFile("conf/buildings.toml")
		if err != nil {
			loadErr = err
			return
		}
		BuildingConfig = &BuildingConf{}
		if err := toml.Unmarshal(buildingData, BuildingConfig); err != nil {
			loadErr = err
			return
		}

		// 加载部队配置
		troopData, err := os.ReadFile("conf/troops.toml")
		if err != nil {
			loadErr = err
			return
		}
		TroopConfig = &TroopConf{}
		if err := toml.Unmarshal(troopData, TroopConfig); err != nil {
			loadErr = err
			return
		}
	})
	return loadErr
}

// GetBuildingLevel 获取指定建筑的指定等级配置
func GetBuildingLevel(buildingType string, level int) *BuildingLevelConf {
	if BuildingConfig == nil {
		return nil
	}
	bConf, ok := BuildingConfig.Building[buildingType]
	if !ok {
		return nil
	}
	for _, lv := range bConf.Levels {
		if lv.Level == level {
			return &lv
		}
	}
	return nil
}

// TroopAttr 兵种属性
type TroopAttr struct {
	Type               string `json:"type"`
	Name               string `json:"name"`
	MeleeAttack        int    `json:"melee_attack"`
	RangedAttack       int    `json:"ranged_attack"`
	MeleeDefense       int    `json:"melee_defense"`
	RangedDefense      int    `json:"ranged_defense"`
	Speed              int    `json:"speed"`
	Capacity           int    `json:"capacity"`
	FoodConsumption    int    `json:"food_consumption"`
	RecruitTimeSeconds int    `json:"recruit_time_seconds"`
	RecruitCostWood    int    `json:"recruit_cost_wood"`
	RecruitCostIron    int    `json:"recruit_cost_iron"`
	RecruitCostFood    int    `json:"recruit_cost_food"`
	RecruitCostStone   int    `json:"recruit_cost_stone"`
}

// GetTroopConfig 获取指定兵种配置
func GetTroopConfig(troopType string) *TroopAttr {
	if TroopConfig == nil {
		return nil
	}
	for i := range TroopConfig.Troops {
		if TroopConfig.Troops[i].Type == troopType {
			t := &TroopConfig.Troops[i]
			return &TroopAttr{
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
			}
		}
	}
	return nil
}
