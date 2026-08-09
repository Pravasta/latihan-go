package postgres

import (
	"database/sql"
	"fmt"
	"os"
)

// Bertanggung Jawab kepada Open Connection - Ping - Close

type DB struct {
	*sql.DB
}

func NewDB() (*DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("db_host"), os.Getenv("db_port"), os.Getenv("db_user"),
		os.Getenv("db_password"), os.Getenv("db_name"),
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
