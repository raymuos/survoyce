package store_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/raymuos/survoyce/internal/domain"
	"github.com/raymuos/survoyce/internal/store"
)

const demoBusinessID = "00000000-0000-0000-0000-000000000001"

// Integration test: needs a migrated database. Run with `make test-integration`.
func TestPostgres_FeedbackLifecycle(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	s := store.NewPostgres(pool)

	fb := &domain.Feedback{
		ID:         uuid.NewString(),
		BusinessID: demoBusinessID,
		AudioKey:   "audio/test.webm",
	}
	if err := s.Create(ctx, fb); err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := s.Get(ctx, fb.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Status != domain.StatusPending {
		t.Errorf("status = %q, want pending", got.Status)
	}

	res := domain.Analysis{Sentiment: 0.8, Themes: []string{"friendly staff"}}
	if err := s.SaveResult(ctx, fb.ID, res, "great service"); err != nil {
		t.Fatalf("save result: %v", err)
	}

	got, _ = s.Get(ctx, fb.ID)
	if got.Status != domain.StatusDone || got.Transcript != "great service" || got.Sentiment != 0.8 {
		t.Errorf("unexpected row after save: %+v", got)
	}

	if _, err := s.Get(ctx, uuid.NewString()); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("missing row: err = %v, want ErrNotFound", err)
	}
}
