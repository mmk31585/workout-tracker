package storage

import (
	"database/sql"
	"fmt"

	"github.com/mmk31585/workout-tracker/internal/config"
	"github.com/mmk31585/workout-tracker/internal/db"
	"github.com/pressly/goose/v3"
)

type Storage struct {
	db *sql.DB
}

func New(dbCfg config.DBConfig) (*Storage, error) {
	dataBase, err := db.New(dbCfg.DBAddr, dbCfg.MaxOpenConns, dbCfg.MaxIdleConns, dbCfg.MaxIdleTime)
	if err != nil {
		return nil, fmt.Errorf("storage: cannot connect to db: %w", err)
	}
	if err := goose.Up(dataBase, "migrations"); err != nil {
		dataBase.Close()
		return nil, fmt.Errorf("storage: migrations failed: %w", err)
	}
	return &Storage{db: dataBase}, nil
}
func (s *Storage) DB() *sql.DB {
	return s.db
}

func (s *Storage) Close() error {
	return s.db.Close()
}
