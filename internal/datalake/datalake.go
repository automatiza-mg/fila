package datalake

import (
	"context"
	"database/sql"
	"errors"
	"time"

	_ "github.com/sclgo/impala-go"
)

var (
	ErrNotFound = errors.New("record not found")
)

type DataLake struct {
	db *sql.DB
}

func New(ctx context.Context, cfg *Config) (*DataLake, error) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	db, err := sql.Open("impala", cfg.connString())
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(30 * time.Second)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	var now time.Time
	if err := db.QueryRow("SELECT NOW()").Scan(&now); err != nil {
		return nil, err
	}

	return &DataLake{db: db}, nil
}

func (d *DataLake) Close() error {
	return d.db.Close()
}

func (d *DataLake) Stats() sql.DBStats {
	return d.db.Stats()
}
