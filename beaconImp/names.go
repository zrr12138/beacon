package beaconImp

// BuildingNameCN 建筑中文名称映射
var BuildingNameCN = map[BuildingType]string{
	BuildingGovernment: "官府",
	BuildingLumberyard: "伐木场",
	BuildingQuarry:     "采石场",
	BuildingIronMine:   "铁矿场",
	BuildingFarm:       "农田",
	BuildingWarehouse:  "仓库",
	BuildingBarracks:   "兵营",
}

// TroopNameCN 兵种中文名称映射
var TroopNameCN = map[TroopType]string{
	TroopSupplyCart:    "粮草兵",
	TroopLightInfantry: "轻步兵",
	TroopSpearman:      "长枪兵",
	TroopArcher:        "弓箭手",
	TroopEliteInfantry: "精锐步兵",
	TroopCrossbowman:   "强弩手",
	TroopLightCavalry:  "轻骑兵",
	TroopHeavyInfantry: "重甲步兵",
	TroopHeavyCavalry:  "重骑兵",
	TroopMarksman:      "神射手",
	TroopChariot:       "战车",
	TroopCatapult:      "投石车",
}

// GetBuildingNameCN 获取建筑中文名
func GetBuildingNameCN(t BuildingType) string {
	if name, ok := BuildingNameCN[t]; ok {
		return name
	}
	return string(t)
}

// GetTroopNameCN 获取兵种中文名
func GetTroopNameCN(t TroopType) string {
	if name, ok := TroopNameCN[t]; ok {
		return name
	}
	return string(t)
}
