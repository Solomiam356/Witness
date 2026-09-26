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
		return fmt.Errorf("не вдалося перевірити підключення: %w", err)
	}

	log.Println("Успішно підключено до PostgreSQL!")
	return nil
}

func InitSchema() error {
	if DB == nil {
		return fmt.Errorf("database pool is not initialized")
	}

	schema := `
	-- Очищення застарілих таблиць для перестворення схеми під час розробки
	-- УВАГА: На продакшені замість DROP використовуються SQL-міграції!
	DROP TABLE IF EXISTS reports CASCADE;
	DROP TABLE IF EXISTS password_reset_tokens CASCADE;
	DROP TABLE IF EXISTS email_verification_tokens CASCADE;
	DROP TABLE IF EXISTS daily_reflections CASCADE;
	DROP TABLE IF EXISTS daily_verses CASCADE;
	DROP TABLE IF EXISTS saved_testimonies CASCADE;
	DROP TABLE IF EXISTS testimonies CASCADE;
	DROP TABLE IF EXISTS tasks CASCADE;
	DROP TABLE IF EXISTS sessions CASCADE;
	DROP TABLE IF EXISTS users CASCADE;

	CREATE TABLE IF NOT EXISTS users (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		email VARCHAR(255) UNIQUE NOT NULL,
		password_hash VARCHAR(255) NOT NULL,
		display_name VARCHAR(100) NOT NULL,
		role VARCHAR(50) DEFAULT 'user',
		email_verified BOOLEAN DEFAULT FALSE,
		current_streak INT DEFAULT 0,
		last_active_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		deleted_at TIMESTAMP WITH TIME ZONE,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS sessions (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		refresh_token_hash VARCHAR(255) NOT NULL,
		device_info TEXT,
		expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
		revoked_at TIMESTAMP WITH TIME ZONE,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS tasks (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		title VARCHAR(255) NOT NULL,
		description TEXT,
		status VARCHAR(50) DEFAULT 'pending',
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS testimonies (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		title VARCHAR(255),
		content TEXT NOT NULL,
		summary TEXT,
		tags TEXT[],
		category VARCHAR(50) DEFAULT 'general',
		language VARCHAR(10) DEFAULT 'uk',
		title_en VARCHAR(255),
		content_en TEXT,
		summary_en TEXT,
		prayer_count INT DEFAULT 0,
		is_published BOOLEAN DEFAULT false,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_testimonies_published_created ON testimonies (is_published, created_at DESC, id);

	CREATE TABLE IF NOT EXISTS saved_testimonies (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		testimony_id UUID NOT NULL REFERENCES testimonies(id) ON DELETE CASCADE,
		saved_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		UNIQUE(user_id, testimony_id)
	);

	CREATE TABLE IF NOT EXISTS daily_verses (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		verse_ref VARCHAR(100) NOT NULL,
	    verse_text TEXT NOT NULL,
		date DATE UNIQUE NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);


	CREATE TABLE IF NOT EXISTS daily_reflections (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		verse_ref VARCHAR(100) NOT NULL,
		verse_text TEXT,
		morning_note TEXT,
		evening_note TEXT,
		date DATE NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		UNIQUE (user_id, date)
	);

	CREATE TABLE IF NOT EXISTS email_verification_tokens (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		token_hash VARCHAR(255) NOT NULL,
		expires_at TIMESTAMPTZ NOT NULL,
		created_at TIMESTAMPTZ DEFAULT NOW(),
		used_at TIMESTAMPTZ
	);

	CREATE INDEX IF NOT EXISTS idx_verification_token_hash ON email_verification_tokens(token_hash);

	CREATE TABLE IF NOT EXISTS password_reset_tokens (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		token_hash TEXT NOT NULL UNIQUE,
		expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS reports (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		testimony_id UUID NOT NULL REFERENCES testimonies(id) ON DELETE CASCADE,
		user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		reason TEXT NOT NULL,
		status VARCHAR(50) DEFAULT 'pending',
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_testimonies_user_id ON testimonies (user_id);
	CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions (user_id);
	CREATE INDEX IF NOT EXISTS idx_sessions_refresh_hash ON sessions (refresh_token_hash);
	CREATE INDEX IF NOT EXISTS idx_saved_testimonies_user_id ON saved_testimonies (user_id);
	CREATE INDEX IF NOT EXISTS idx_reports_testimony_id ON reports (testimony_id);
	CREATE INDEX IF NOT EXISTS idx_reports_user_id ON reports (user_id);
	CREATE INDEX IF NOT EXISTS idx_daily_verses_date ON daily_verses (date);
	CREATE INDEX IF NOT EXISTS idx_daily_reflections_user_id ON daily_reflections (user_id, date);
	`

	_, err := DB.Exec(context.Background(), schema)
	if err != nil {
		return fmt.Errorf("не вдалося створити схему даних Witness: %w", err)
	}
	log.Println("Фінальна схема Witness V2.1 успішно розгорнута!")
	return nil
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}