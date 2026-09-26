package repository

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/Solomiam356/witness-backend/internal/domain"
)

type ReflectionRepository struct {
	db *pgxpool.Pool
}

func NewReflectionRepository(db *pgxpool.Pool) *ReflectionRepository {
	return &ReflectionRepository{db: db}
}

// GetTodayVerse повертає вірш дня
func (r *ReflectionRepository) GetTodayVerse(ctx context.Context) (string, string, error) {
	query := `SELECT verse_ref, verse_text FROM daily_verses WHERE date = CURRENT_DATE LIMIT 1`
	var ref, text string
	err := r.db.QueryRow(ctx, query).Scan(&ref, &text)
	if err != nil {
		// Якщо о 6:00 вірш ще не згенерувався, повертаємо дефолтний
		return "Івана 3:16", "Бо так Бог полюбив світ...", nil
	}
	return ref, text, nil
}

// SaveMorningNote зберігає ранковий роздум
func (r *ReflectionRepository) SaveMorningNote(ctx context.Context, userID, verseRef, verseText, note string) error {
	query := `
		INSERT INTO daily_reflections (user_id, verse_ref, verse_text, morning_note, date, updated_at)
		VALUES ($1, $2, $3, $4, CURRENT_DATE, NOW())
		ON CONFLICT (user_id, date)
		DO UPDATE SET morning_note = EXCLUDED.morning_note, updated_at = NOW()
	`
	_, err := r.db.Exec(ctx, query, userID, verseRef, verseText, note)
	return err
}

// SaveEveningNote зберігає вечірню нотатку
func (r *ReflectionRepository) SaveEveningNote(ctx context.Context, userID, note string) error {
	query := `
		UPDATE daily_reflections
		SET evening_note = $1, updated_at = NOW()
		WHERE user_id = $2 AND date = CURRENT_DATE
	`
	_, err := r.db.Exec(ctx, query, note, userID)
	return err
}