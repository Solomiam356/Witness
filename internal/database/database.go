package database

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/Solomiam356/witness-backend/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

func Connect(cfg config.Config) error {
	sslMode := os.Getenv("DB_SSLMODE")
	if sslMode == "" {
		sslMode = "require"
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
		sslMode,
	)

	var err error
	DB, err = pgxpool.New(context.Background(), dsn)
	if err != nil {
		return fmt.Errorf("не вдалося підключитися до бази даних: %w", err)
	}

	if err := DB.Ping(context.Background()); err != nil {
		DB.Close()
		DB = nil

		return fmt.Errorf(
			"не вдалося перевірити підключення: %w",
			err,
		)
	}

	log.Println("Успішно підключено до PostgreSQL!")
	return nil
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}
