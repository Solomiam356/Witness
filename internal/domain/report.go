package domain

import "time"

type Report struct {
	ID          string    `json:"id"`
	ReporterID  string    `json:"reporter_id"`
	TestimonyID string    `json:"testimony_id"`
	Reason      string    `json:"reason"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}