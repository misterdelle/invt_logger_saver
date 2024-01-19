package main

import (
	"flag"
	"fmt"
	"invt_logger_saver/pkg/data"
	"invt_logger_saver/pkg/db"
	"invt_logger_saver/pkg/repository"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/joho/godotenv"
)

type application struct {
	Env           string
	MQTTURL       string
	MQTTUser      string
	MQTTPassword  string
	MQTTTopicName string
	DSN           string
	DB            repository.DatabaseRepository
}

var (
	app         = application{}
	stationData = data.NewStation()
)

func init() {
	flag.Parse()

	if app.Env != "" {
		fmt.Printf("app.Env        : %s \n", app.Env)
		godotenv.Load(".env." + app.Env + ".local")
		godotenv.Load(".env." + app.Env)
	} else {
		fmt.Println("app.Env NON settato, carico i dati dal file .env")
		godotenv.Load() // The Original .env
		app.Env = os.Getenv("Env")
		fmt.Printf("app.Env                 : %s \n", app.Env)
	}

	app.MQTTURL = os.Getenv("mqtt.url")
	app.MQTTUser = os.Getenv("mqtt.user")
	app.MQTTPassword = os.Getenv("mqtt.password")
	app.MQTTTopicName = os.Getenv("mqtt.prefix")
	app.DSN = os.Getenv("DSN")

	fmt.Printf("app.MQTTURL      : %s \n", app.MQTTURL)
	fmt.Printf("app.MQTTUser     : %s \n", app.MQTTUser)
	fmt.Printf("app.MQTTPassword : %s \n", app.MQTTPassword)
	fmt.Printf("app.MQTTTopicName: %s \n", app.MQTTTopicName)
	fmt.Printf("app.DSN          : %s \n", app.MQTTTopicName)
}

func main() {

	connRDBMS, err := app.connectToDB()
	if err != nil {
		log.Fatal(fmt.Sprintf("Error %s", err))
	}
	defer connRDBMS.Close()

	app.DB = &db.PostgresDBRepo{DB: connRDBMS}

	var broker = app.MQTTURL
	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID("go_mqtt_client")
	opts.SetUsername(app.MQTTUser)
	opts.SetPassword(app.MQTTPassword)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	client := mqtt.NewClient(opts)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(fmt.Sprintf("Error connecting to MQTT broker: %s", token.Error()))
	}

	// if token := client.Subscribe(app.MQTTTopicName, 0, onMessageReceived); token.Wait() && token.Error() != nil {
	// 	panic(fmt.Sprintf("Error subscribing to topic: %s", token.Error()))
	// }
	topicFilter := map[string]byte{
		// app.MQTTTopicName + "/inverter":                          0,
		app.MQTTTopicName + "/station/lastUpdateTime":            0,
		app.MQTTTopicName + "/station/purchasingDayEnergy":       0,
		app.MQTTTopicName + "/station/batteryChargeDayEnergy":    0,
		app.MQTTTopicName + "/station/batterySOC":                0,
		app.MQTTTopicName + "/station/gridDayEnergy":             0,
		app.MQTTTopicName + "/station/pvDayEnergy":               0,
		app.MQTTTopicName + "/station/batteryDischargeDayEnergy": 0,
		app.MQTTTopicName + "/station/loadDayEnergy":             0,
	}

	if token := client.SubscribeMultiple(topicFilter, onMessageReceived); token.Wait() && token.Error() != nil {
		panic(fmt.Sprintf("Error subscribing to topic: %s", token.Error()))
	}

	// Wait for a signal to exit the program gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	client.Unsubscribe(app.MQTTTopicName)
	client.Disconnect(250)
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connect lost: %v", err)
}

func onMessageReceived(client mqtt.Client, message mqtt.Message) {
	fmt.Printf("Received message: %s from topic: %s\n", message.Payload(), message.Topic())

	if message.Topic() == app.MQTTTopicName+"/station/pvDayEnergy" {
		stationData.TotalProduction = fromByteArrayToFloat32(message.Payload())
	}

	if message.Topic() == app.MQTTTopicName+"/station/gridDayEnergy" {
		stationData.FeedIn = fromByteArrayToFloat32(message.Payload())
	}

	if message.Topic() == app.MQTTTopicName+"/station/batteryChargeDayEnergy" {
		stationData.BatteryCharge = fromByteArrayToFloat32(message.Payload())
	}

	if message.Topic() == app.MQTTTopicName+"/station/loadDayEnergy" {
		stationData.TotalConsumption = fromByteArrayToFloat32(message.Payload())
	}

	if message.Topic() == app.MQTTTopicName+"/station/purchasingDayEnergy" {
		stationData.PowerPurchased = fromByteArrayToFloat32(message.Payload())
	}

	if message.Topic() == app.MQTTTopicName+"/station/batteryDischargeDayEnergy" {
		stationData.BatteryDischarge = fromByteArrayToFloat32(message.Payload())
	}

	if message.Topic() == app.MQTTTopicName+"/station/lastUpdateTime" {
		// 2024-01-19 10:28:48
		lastUpdate := strings.TrimSpace(string(message.Payload()))
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
		err := app.DB.InsertStationData(stationData)
		if err != nil {
			log.Println(fmt.Sprintf("Error %s", err))
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
