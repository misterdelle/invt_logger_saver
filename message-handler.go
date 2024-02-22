package main

import (
	// "fmt"
	"bytes"
	"encoding/json"
	"fmt"
	"invt_logger_saver/pkg/data"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func onMessageReceived(client mqtt.Client, message mqtt.Message) {
	log.Printf("Received message: %s from topic: %s\n", message.Payload(), message.Topic())

	parseStationData(message.Payload(), message.Topic())
}

func parseStationData(msgPayload []byte, msgTopic string) {

	if msgTopic == app.MQTTTopicName+"/station/pvDayEnergy" {
		stationData.TotalProduction = fromByteArrayToFloat32(msgPayload)
		stationData.TotalProductionRead = true
	}

	if msgTopic == app.MQTTTopicName+"/station/gridDayEnergy" {
		stationData.FeedIn = fromByteArrayToFloat32(msgPayload)
		stationData.FeedInRead = true
	}

	if msgTopic == app.MQTTTopicName+"/station/batteryChargeDayEnergy" {
		stationData.BatteryCharge = fromByteArrayToFloat32(msgPayload)
		stationData.BatteryChargeRead = true
	}

	if msgTopic == app.MQTTTopicName+"/station/loadDayEnergy" {
		stationData.TotalConsumption = fromByteArrayToFloat32(msgPayload)
		stationData.TotalConsumptionRead = true
	}

	if msgTopic == app.MQTTTopicName+"/station/purchasingDayEnergy" {
		stationData.PowerPurchased = fromByteArrayToFloat32(msgPayload)
		stationData.PowerPurchasedRead = true
	}

	if msgTopic == app.MQTTTopicName+"/station/batteryDischargeDayEnergy" {
		stationData.BatteryDischarge = fromByteArrayToFloat32(msgPayload)
		stationData.BatteryDischargeRead = true
	}

	if msgTopic == app.MQTTTopicName+"/station/batterySOC" {
		stationData.BatterySOC = fromByteArrayToFloat32(msgPayload)
		stationData.BatterySOCRead = true
	}

	if msgTopic == app.MQTTTopicName+"/station/lastUpdateTime" {
		// 2024-01-19 10:28:48
		lastUpdate := strings.TrimSpace(string(msgPayload))
		lastUpdateTS, err := time.Parse("2006-01-02 15:04:05", lastUpdate)
		if err != nil {
			log.Println(fmt.Sprintf("Error parsing lastUpdateTime: %s", err))
		}
		stationData.LastUpdateTime = lastUpdateTS
		stationData.LastUpdateTimeRead = true
	}

	if stationData.TotalProductionRead && stationData.FeedInRead && stationData.BatteryChargeRead {
		stationData.SelfUsed = stationData.TotalProduction - stationData.FeedIn - stationData.BatteryCharge
		stationData.SelfUsedRead = true
	}

	if stationData.TotalConsumptionRead && stationData.PowerPurchasedRead && stationData.BatteryDischargeRead {
		stationData.Production = stationData.TotalConsumption - stationData.PowerPurchased - stationData.BatteryDischarge
		stationData.ProductionRead = true
	}

	if stationData.SelfUsedRead && stationData.ProductionRead && stationData.BatterySOCRead && stationData.LastUpdateTimeRead {
		// Scrivi nel db
		_, err := RetryWithBackoff(app.DB.InsertStationData, 5, 2*time.Second, stationData)
		if err != nil {
			log.Println(fmt.Sprintf("Error %s", err))
		}

		if stationData.LastUpdateTime.Minute() == 0 || stationData.LastUpdateTime.Minute() == 30 {
			go sendTelegramMessage(stationData.String())
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

func sendTelegramMessage(text string) {

	request_url := "https://api.telegram.org/bot" + app.TelegramToken + "/sendMessage"
	log.Println(fmt.Sprintf("Request URL %s", request_url))

	client := &http.Client{}

	values := map[string]string{"text": text, "chat_id": app.TelegramChatID}
	json_paramaters, _ := json.Marshal(values)
	log.Println(fmt.Sprintf("json_paramaters %s", json_paramaters))

	req, _ := http.NewRequest("POST", request_url, bytes.NewBuffer(json_paramaters))
	req.Header.Set("Content-Type", "application/json")

	_, err := client.Do(req)

	if err != nil {
		fmt.Println(err)
	}
}
