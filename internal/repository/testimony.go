package repository

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/Solomiam356/witness-backend/internal/domain"
)

type TestimonyRepository struct {
	db *pgxpool.Pool
}

func NewTestimonyRepository(db *pgxpool.Pool) *TestimonyRepository {
	return &TestimonyRepository{db: db}
}

func (r *TestimonyRepository) Create(ctx context.Context, t *domain.Testimony) error {
	query := `
		INSERT INTO testimonies (id, user_id, title, content, summary, tags, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(ctx, query, t.UserID, t.Title, t.Content, t.Summary, t.Tags).Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)
}

func (r *TestimonyRepository) GetAllByUserID(ctx context.Context, userID string) ([]domain.Testimony, error) {
	query := `SELECT id, user_id, title, content, summary, tags, created_at
	          FROM testimonies
	          WHERE user_id = $1
	          ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.Testimony
	for rows.Next() {
		var t domain.Testimony
		if err := rows.Scan(&t.ID, &t.UserID, &t.Title, &t.Content, &t.Summary, &t.Tags, &t.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return list, nil
}

func (r *TestimonyRepository) DeleteByID(ctx context.Context, id string, userID string) error {
	query := `DELETE FROM testimonies WHERE id = $1 AND user_id = $2`

	result, err := r.db.Exec(ctx, query, id, userID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return context.DeadlineExceeded
	}

	return nil
}

func (r *TestimonyRepository) GetFeed(ctx context.Context, cursor string, limit int, search string, filterUserID string) ([]domain.Testimony, error) {
	if limit <= 0 {
		limit = 10
	}

	query := `SELECT id, user_id, title, content, summary, tags, created_at FROM testimonies WHERE 1=1`
	var args []interface{}
	argCounter := 1

	if cursor != "" {
		query += fmt.Sprintf(" AND created_at < $%d", argCounter)
		args = append(args, cursor)
		argCounter++
	}

	if filterUserID != "" {
		query += fmt.Sprintf(" AND user_id = $%d", argCounter)
		args = append(args, filterUserID)
		argCounter++
	}

	if search != "" {
		query += fmt.Sprintf(" AND (title ILIKE $%d OR content ILIKE $%d)", argCounter, argCounter)
		args = append(args, "%"+search+"%")
		argCounter++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d", argCounter)
	args = append(args, limit)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.Testimony
	for rows.Next() {
		var t domain.Testimony
		if err := rows.Scan(&t.ID, &t.UserID, &t.Title, &t.Content, &t.Summary, &t.Tags, &t.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return list, nil
}
