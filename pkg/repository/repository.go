package repository

import (
	"database/sql"
)

type DatabaseRepository interface {
	Connection() *sql.DB
	InsertStationData(args ...interface{}) (interface{}, error)
}
