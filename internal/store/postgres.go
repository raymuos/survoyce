// Package store implements domain.FeedbackStore on top of Postgres.
package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/raymuos/survoyce/internal/domain"
)

type Postgres struct {
	pool *pgxpool.Pool
}

// Compile-time check that Postgres satisfies the interface.
var _ domain.FeedbackStore = (*Postgres)(nil)

func NewPostgres(pool *pgxpool.Pool) *Postgres {
	return &Postgres{pool: pool}
}

func (p *Postgres) Create(ctx context.Context, f *domain.Feedback) error {
	if f.Status == "" {
		f.Status = domain.StatusPending
	}
	err := p.pool.QueryRow(ctx,
		`INSERT INTO feedback (id, business_id, audio_key, status)
		 VALUES ($1, $2, $3, $4)
		 RETURNING created_at`,
		f.ID, f.BusinessID, f.AudioKey, string(f.Status),
	).Scan(&f.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert feedback: %w", err)
	}
	return nil
}

func (p *Postgres) Get(ctx context.Context, id string) (*domain.Feedback, error) {
	var (
		f      domain.Feedback
		status string
	)
	err := p.pool.QueryRow(ctx,
		`SELECT id::text, business_id::text, audio_key, transcript,
		        sentiment, themes, status, created_at
		 FROM feedback WHERE id = $1`, id,
	).Scan(&f.ID, &f.BusinessID, &f.AudioKey, &f.Transcript,
		&f.Sentiment, &f.Themes, &status, &f.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get feedback: %w", err)
	}
	f.Status = domain.Status(status)
	return &f, nil
}

func (p *Postgres) UpdateStatus(ctx context.Context, id string, s domain.Status) error {
	tag, err := p.pool.Exec(ctx,
		`UPDATE feedback SET status = $2 WHERE id = $1`, id, string(s))
	if err != nil {
		return fmt.Errorf("update status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (p *Postgres) SaveResult(ctx context.Context, id string, a domain.Analysis, transcript string) error {
	themes := a.Themes
	if themes == nil {
		themes = []string{}
	}
	tag, err := p.pool.Exec(ctx,
		`UPDATE feedback
		 SET transcript = $2, sentiment = $3, themes = $4, status = 'done'
		 WHERE id = $1`,
		id, transcript, a.Sentiment, themes)
	if err != nil {
		return fmt.Errorf("save result: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
