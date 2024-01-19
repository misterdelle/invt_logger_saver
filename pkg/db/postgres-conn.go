package db

import (
	"database/sql"
	"time"
)

const dbTimeout = time.Second * 300

type PostgresConn struct {
	DB *sql.DB
}

func (m *PostgresDBRepo) Connection() *sql.DB {
	return m.DB
}
