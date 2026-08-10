package postgres

import (
	"database/sql"
	"fmt"
	"taskflow-api/internal/config"
)

// Bertanggung Jawab kepada Open Connection - Ping - Close

type DB struct {
	*sql.DB
}

func ConnectDB(cfg *config.Config) (*DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password, cfg.Database.Name,
	)

	sqlDB, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return &DB{
		DB: sqlDB,
	}, nil
}
