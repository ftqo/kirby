package database

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v4/stdlib"

	"github.com/disgoorg/log"
	"github.com/ftqo/kirby/config"
)

func Open(ctx context.Context, log log.Logger, c config.DBConfig) (*sql.DB, error) {
	log.Info("opening database connection pool")
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.Host, c.Port, c.Username, c.Password, c.Database)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %v", err)
	}
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping database connection: %v", err)
	}
	return db, nil
}
