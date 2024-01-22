package data

import "time"

type Station struct {
	LastUpdateTime       time.Time
	LastUpdateTimeRead   bool
	TotalProduction      float32
	TotalProductionRead  bool
	FeedIn               float32
	FeedInRead           bool
	BatteryCharge        float32
	BatteryChargeRead    bool
	SelfUsed             float32
	SelfUsedRead         bool
	TotalConsumption     float32
	TotalConsumptionRead bool
	PowerPurchased       float32
	PowerPurchasedRead   bool
	BatteryDischarge     float32
	BatteryDischargeRead bool
	Production           float32
	ProductionRead       bool
}

func NewStation() Station {
	return Station{}
}
