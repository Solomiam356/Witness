package repository

import (
	"context"
	"time"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/Solomiam356/witness-backend/internal/domain"
)

type SessionRepository struct {
	db *pgxpool.Pool
}

func NewSessionRepository(db *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Create(ctx context.Context, s *domain.Session) error {
	query := `
	INSERT INTO sessions (id, user_id, refresh_token_hash, device_info, expires_at, created_at)
	VALUES (gen_random_uuid(), $1, $2, $3, $4, NOW())
	RETURNING id, created_at
	`
	return r.db.QueryRow(ctx, query, s.UserID, s.RefreshTokenHash, s.DeviceInfo, s.ExpiresAt).Scan(&s.ID, &s.CreatedAt)
}

func (r *SessionRepository) GetByHash(ctx context.Context, hash string) (*domain.Session, error) {
	query := `
	SELECT id, user_id, refresh_token_hash, device_info, expires_at, revoked_at, created_at
	FROM sessions
	WHERE refresh_token_hash = $1	
 `
 var s domain.Session
 err := r.db.QueryRow(ctx, query, hash).Scan(
	&s.ID, &s.UserID, &s.RefreshTokenHash, &s.DeviceInfo, &s.ExpiresAt, &s.RevokedAt, &s.CreatedAt,
 )
 if err != nil {
	return nil, err
 }
 return &s, nil
}

func (r *SessionRepository) Revoke(ctx context.Context, id string) error {
	query := `UPDATE sessions SET revoked_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *SessionRepository) RevokeAllByUserID(ctx context.Context, userID string) error {
	query := `UPDATE sessions SET revoked_at = NOW() WHERE user_id = $1 AND revoked_at IS NULL`
	_, err := r.db.Exec(ctx, query, userID)
	return err
}

func (r *SessionRepository) GetSessionsByUserID(ctx context.Context, userID string) ([]domain.Session, error) {
	query := `
		SELECT id, user_id, refresh_token_hash, device_info, expires_at, revoked_at, created_at
		FROM sessions
		WHERE user_id = $1 AND revoked_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []domain.Session
	for rows.Next() {
		var s domain.Session

		err := rows.Scan(&s.ID, &s.UserID, &s.RefreshTokenHash, &s.DeviceInfo, &s.ExpiresAt, &s.RevokedAt, &s.CreatedAt)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return sessions, nil
}

func (r *SessionRepository) SaveVerificationToken(ctx context.Context, userID string, tokenHash string, expiresAt time.Time) error {
	query := `
	INSERT INTO email_verification_tokens (user_id, token_hash, expires_at)
	VALUES ($1, $2, $3)
	`
	_, err := r.db.Exec(ctx, query, userID, tokenHash, expiresAt)
	return err
}