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
	"syscall"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/joho/godotenv"
)

type application struct {
	Env            string
	MQTTURL        string
	MQTTUser       string
	MQTTPassword   string
	MQTTTopicName  string
	DSN            string
	DB             repository.DatabaseRepository
	MQTTClient     mqtt.Client
	TelegramToken  string
	TelegramChatID string
}

var (
	app         = application{TelegramChatID: "529623659"}
	stationData = data.NewStation()
)

func init() {
	flag.Parse()

	if app.Env != "" {
		log.Printf("app.Env          : %s \n", app.Env)
		godotenv.Load(".env." + app.Env + ".local")
		godotenv.Load(".env." + app.Env)
	} else {
		log.Println("app.Env NON settato, carico i dati dal file .env")
		godotenv.Load() // The Original .env
		app.Env = os.Getenv("Env")
		log.Printf("app.Env          : %s \n", app.Env)
	}

	app.MQTTURL = os.Getenv("mqtt.url")
	app.MQTTUser = os.Getenv("mqtt.user")
	app.MQTTPassword = os.Getenv("mqtt.password")
	app.MQTTTopicName = os.Getenv("mqtt.prefix")
	app.DSN = os.Getenv("DSN")
	app.TelegramToken = os.Getenv("TELEGRAM_TOKEN")
	app.TelegramChatID = os.Getenv("TELEGRAM_CHAT_ID")

	log.Printf("app.MQTTURL       : %s \n", app.MQTTURL)
	log.Printf("app.MQTTUser      : %s \n", app.MQTTUser)
	log.Printf("app.MQTTPassword  : %s \n", app.MQTTPassword)
	log.Printf("app.MQTTTopicName : %s \n", app.MQTTTopicName)
	log.Printf("app.DSN           : %s \n", app.DSN)
	log.Printf("app.TelegramToken : %s \n", app.TelegramToken)
	log.Printf("app.TelegramChatID: %s \n", app.TelegramChatID)
}

func main() {
	connRDBMS, err := app.connectToDB()
	if err != nil {
		log.Fatalf(fmt.Sprintf("Error %s", err))
	}
	defer connRDBMS.Close()

	app.DB = &db.PostgresDBRepo{DB: connRDBMS}

	var broker = app.MQTTURL
	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID("go_mqtt_client")
	opts.SetUsername(app.MQTTUser)
	opts.SetPassword(app.MQTTPassword)
	opts.SetConnectRetry(true)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	// client := mqtt.NewClient(opts)
	app.MQTTClient = mqtt.NewClient(opts)

	if token := app.MQTTClient.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf(fmt.Sprintf("Error connecting to MQTT broker: %s", token.Error()))
	}

	go app.subscribeTopic()

	// Wait for a signal to exit the program gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	app.MQTTClient.Unsubscribe(app.MQTTTopicName)
	app.MQTTClient.Disconnect(250)
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	log.Println("Connected")
	go app.subscribeTopic()
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	log.Printf("Connect lost: %v", err)
}

func (app *application) subscribeTopic() {
	topicFilter := map[string]byte{
		app.MQTTTopicName + "/station/lastUpdateTime":            0,
		app.MQTTTopicName + "/station/purchasingDayEnergy":       0,
		app.MQTTTopicName + "/station/batteryChargeDayEnergy":    0,
		app.MQTTTopicName + "/station/batterySOC":                0,
		app.MQTTTopicName + "/station/gridDayEnergy":             0,
		app.MQTTTopicName + "/station/pvDayEnergy":               0,
		app.MQTTTopicName + "/station/batteryDischargeDayEnergy": 0,
		app.MQTTTopicName + "/station/loadDayEnergy":             0,
	}

	if token := app.MQTTClient.SubscribeMultiple(topicFilter, onMessageReceived); token.Wait() && token.Error() != nil {
		log.Fatalf(fmt.Sprintf("Error subscribing to topic: %s", token.Error()))
	}
}
