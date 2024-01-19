package main

import (
	"fmt"
	"invt_logger_saver/pkg/data"
	"strconv"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func onMessageReceived(client mqtt.Client, message mqtt.Message) {
	fmt.Printf("Received message: %s from topic: %s\n", message.Payload(), message.Topic())

	parseStationData(message.Payload(), message.Topic())
}

func parseStationData(msgPayload []byte, msgTopic string) {

	if msgTopic == app.MQTTTopicName+"/station/pvDayEnergy" {
		stationData.TotalProduction = fromByteArrayToFloat32(msgPayload)
	}

	if msgTopic == app.MQTTTopicName+"/station/gridDayEnergy" {
		stationData.FeedIn = fromByteArrayToFloat32(msgPayload)
	}

	if msgTopic == app.MQTTTopicName+"/station/batteryChargeDayEnergy" {
		stationData.BatteryCharge = fromByteArrayToFloat32(msgPayload)
	}

	if msgTopic == app.MQTTTopicName+"/station/loadDayEnergy" {
		stationData.TotalConsumption = fromByteArrayToFloat32(msgPayload)
	}

	if msgTopic == app.MQTTTopicName+"/station/purchasingDayEnergy" {
		stationData.PowerPurchased = fromByteArrayToFloat32(msgPayload)
	}

	if msgTopic == app.MQTTTopicName+"/station/batteryDischargeDayEnergy" {
		stationData.BatteryDischarge = fromByteArrayToFloat32(msgPayload)
	}

	if msgTopic == app.MQTTTopicName+"/station/lastUpdateTime" {
		// 2024-01-19 10:28:48
		lastUpdate := strings.TrimSpace(string(msgPayload))
		lastUpdateTS, err := time.Parse("2006-01-02 15:04:05", lastUpdate)
		if err != nil {
			fmt.Println(err)
		}
		stationData.LastUpdateTime = lastUpdateTS
	}

	if stationData.TotalProduction >= 0 && stationData.FeedIn >= 0 && stationData.BatteryCharge >= 0 && stationData.SelfUsed == -1 {
		stationData.SelfUsed = stationData.TotalProduction - stationData.FeedIn - stationData.BatteryCharge
	}

	if stationData.TotalConsumption >= 0 && stationData.PowerPurchased >= 0 && stationData.BatteryDischarge >= 0 && stationData.Production == -1 {
		stationData.Production = stationData.TotalConsumption - stationData.PowerPurchased - stationData.BatteryDischarge
	}

	if stationData.SelfUsed > 0 && stationData.Production > 0 && !stationData.LastUpdateTime.IsZero() {
		// Scrivi nel db
		_, err := RetryWithBackoff(app.DB.InsertStationData, 5, 2*time.Second, stationData)
		if err != nil {
			fmt.Println(fmt.Sprintf("Error %s", err))
		}

		stationData = data.NewStation()
	}

}

func fromByteArrayToFloat32(b []byte) float32 {
	num, err := strconv.ParseFloat(strings.TrimSpace(string(b)), 64)
	if err != nil {
		fmt.Println("fromByteArrayToFloat32 failed:", err)

		return 0
	}

	return float32(num)
}
