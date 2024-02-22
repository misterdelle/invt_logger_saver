package data

import (
	"fmt"
	"time"
)

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
	BatterySOC           float32
	BatterySOCRead       bool
}

func NewStation() Station {
	return Station{}
}

func (s *Station) String() string {
	ret := "Station: \n"
	ret += fmt.Sprintf("LastUpdateTime: %s  \n", s.LastUpdateTime.String())
	ret += fmt.Sprintf("TotalProduction: %.2f \n", s.TotalProduction)
	ret += fmt.Sprintf("FeedIn: %.2f \n", s.FeedIn)
	ret += fmt.Sprintf("BatteryCharge: %.2f \n", s.BatteryCharge)
	ret += fmt.Sprintf("SelfUsed: %.2f \n", s.SelfUsed)
	ret += fmt.Sprintf("TotalConsumption: %.2f \n", s.TotalConsumption)
	ret += fmt.Sprintf("PowerPurchased: %.2f \n", s.PowerPurchased)
	ret += fmt.Sprintf("BatteryDischarge: %.2f \n", s.BatteryDischarge)
	ret += fmt.Sprintf("Production: %.2f \n", s.Production)
	ret += fmt.Sprintf("BatterySOC: %.2f \n", s.BatterySOC)

	return ret
}
