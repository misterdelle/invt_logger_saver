package data

import "time"

type Station struct {
	LastUpdateTime   time.Time
	TotalProduction  float32
	FeedIn           float32
	BatteryCharge    float32
	SelfUsed         float32
	TotalConsumption float32
	PowerPurchased   float32
	BatteryDischarge float32
	Production       float32
}

func NewStation() Station {
	return Station{
		TotalProduction:  -1,
		FeedIn:           -1,
		BatteryCharge:    -1,
		SelfUsed:         -1,
		TotalConsumption: -1,
		PowerPurchased:   -1,
		BatteryDischarge: -1,
		Production:       -1,
	}
}
