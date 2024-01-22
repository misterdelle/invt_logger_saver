package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"invt_logger_saver/pkg/data"
)

type PostgresDBRepo struct {
	DB *sql.DB
}

// InsertStationData inserts Station data into the database
func (m *PostgresDBRepo) InsertStationData(args ...interface{}) (interface{}, error) {
	stationData := args[0].(data.Station)

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	// TotalProduction  int
	// FeedIn           int
	// BatteryCharge    int
	// SelfUsed         int
	// TotalConsumption int
	// PowerPurchased   int
	// BatteryDischarge int
	// Production       int

	stmt := `insert into "Station".Station (last_update_ts, total_production, feed_in, battery_charge, self_used, total_consumption, power_purchased, battery_discharge, production)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	row := m.DB.QueryRowContext(ctx, stmt,
		stationData.LastUpdateTime,
		stationData.TotalProduction,
		stationData.FeedIn,
		stationData.BatteryCharge,
		stationData.SelfUsed,
		stationData.TotalConsumption,
		stationData.PowerPurchased,
		stationData.BatteryDischarge,
		stationData.Production,
	)

	if row.Err() != nil {
		return nil, errors.New(fmt.Sprintf("error inserting stationData: %s", row.Err()))
	}

	return nil, nil
}
