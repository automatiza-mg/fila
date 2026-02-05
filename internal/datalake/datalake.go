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
	db.SetMaxIdleConns(3)
	db.SetMaxOpenConns(3)

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	return &DataLake{db: db}, nil
}

func (d *DataLake) Close() error {
	return d.db.Close()
}
