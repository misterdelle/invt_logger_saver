package main

import (
	"database/sql"
	"log"
	"time"
)

var counts int64

const dbTimeout = time.Second * 3

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (app *application) connectToDB() (*sql.DB, error) {
	//	dsn := os.Getenv("DSN")
	// Locale
	// dsn := "host=localhost port=5500 user=postgres password=postgres dbname=ectm-users sslmode=disable timezone=UTC connect_timeout=5"
	// Casa
	// dsn := "host=homeassistant.local port=5432 user=sql_admin password=sql_password dbname=ECTM sslmode=disable timezone=UTC connect_timeout=5"

	for {
		connection, err := openDB(app.DSN)
		if err != nil {
			log.Println(err)
			log.Println("Postgres not yet ready ...")
			counts++
		} else {
			log.Println("Connected to Postgres!")
			return connection, nil
		}

		if counts > 10 {
			log.Println(err)
			return nil, err
		}

		log.Println("Backing off for two seconds....")
		time.Sleep(2 * time.Second)
		continue
	}
}
