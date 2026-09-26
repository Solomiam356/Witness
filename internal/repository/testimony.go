package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/Solomiam356/witness-backend/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lib/pq"
)

type TestimonyRepository struct {
	db *pgxpool.Pool
}

func NewTestimonyRepository(db *pgxpool.Pool) *TestimonyRepository {
	return &TestimonyRepository{db: db}
}

func (r *TestimonyRepository) Create(ctx context.Context, t *domain.Testimony) error {
	query := `
		INSERT INTO testimonies (id, user_id, title, content, summary, tags, is_published, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, false, NOW(), NOW())
		RETURNING id, is_published, created_at, updated_at
	`
	return r.db.QueryRow(ctx, query, t.UserID, t.Title, t.Content, t.Summary, t.Tags).Scan(&t.ID, &t.IsPublished, &t.CreatedAt, &t.UpdatedAt)
}

func (r *TestimonyRepository) UpdateModerationStatus(
	ctx context.Context,
	id string,
	summary string,
	tags pq.StringArray,
	contentEN string,
	summaryEN string,
	isPublished bool,
) error {
	query := `
		UPDATE testimonies
		SET summary = $1,
		    tags = $2,
		    content_en = $3,
		    summary_en = $4,
		    is_published = $5,
		    updated_at = NOW()
		WHERE id = $6
	`
	_, err := r.db.Exec(ctx, query, summary, tags, contentEN, summaryEN, isPublished, id)
	return err
}

func (r *TestimonyRepository) GetAllByUserID(ctx context.Context, userID string) ([]domain.Testimony, error) {
	query := `SELECT id, user_id, title, content, summary, tags, is_published, created_at
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
		if err := rows.Scan(&t.ID, &t.UserID, &t.Title, &t.Content, &t.Summary, &t.Tags, &t.IsPublished, &t.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, t)
	}

	return list, rows.Err()
}

func (r *TestimonyRepository) DeleteByID(ctx context.Context, id string, userID string) error {
	query := `DELETE FROM testimonies WHERE id = $1 AND user_id = $2`

	result, err := r.db.Exec(ctx, query, id, userID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("свідчення не знайдено або користувач не має прав на видалення")
	}

	return nil
}

func (r *TestimonyRepository) HardDelete(ctx context.Context, id string) error {
	query := `DELETE FROM testimonies WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *TestimonyRepository) GetFeed(ctx context.Context, cursor string, limit int, search string, filterUserID string) ([]domain.Testimony, error) {
	if limit <= 0 {
		limit = 10
	}

	query := `SELECT id, user_id, title, content, summary, tags, is_published, created_at FROM testimonies WHERE is_published = true`
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
		if err := rows.Scan(&t.ID, &t.UserID, &t.Title, &t.Content, &t.Summary, &t.Tags, &t.IsPublished, &t.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, t)
	}

	return list, rows.Err()
}
