package repository

import (
	"database/sql"
	"invt_logger_saver/pkg/data"
)

type DatabaseRepository interface {
	Connection() *sql.DB
	InsertStationData(stationData data.Station) error
}
