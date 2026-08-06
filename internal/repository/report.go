package repository

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/Solomiam356/witness-backend/internal/domain"
)

type ReportRepository struct {
	db *pgxpool.Pool
}

func NewReportRepository(db *pgxpool.Pool) *ReportRepository {
	return &ReportRepository{db: db}
}

func (r *ReportRepository) Create(ctx context.Context, report *domain.Report) error {
	query := `
	INSERT INTO reports (id, reporter_id, testimony_id, reason, status, created_at, updated_at)
	VALUES (gen_random_uuid(), $1, $2, $3, 'pending', NOW(), NOW())
	RETURNING id, status, created_at, updated_at
	`

	return r.db.QueryRow(ctx, query, report.ReporterID, report.TestimonyID, report.Reason).Scan(&report.ID, &report.Status, &report.CreatedAt, &report.UpdatedAt)
}