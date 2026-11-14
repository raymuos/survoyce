// Package domain holds the core types and the interfaces the rest of the app
// depends on. It must not import any other package from this module.
package domain

import (
	"context"
	"errors"
	"io"
	"time"
)

var ErrNotFound = errors.New("not found")

type Status string

const (
	StatusPending    Status = "pending"
	StatusProcessing Status = "processing"
	StatusDone       Status = "done"
	StatusFailed     Status = "failed"
)

type Feedback struct {
	ID         string
	BusinessID string
	AudioKey   string
	Transcript string
	Sentiment  float64 // -1 (negative) to 1 (positive)
	Themes     []string
	Status     Status
	CreatedAt  time.Time
}

// Analysis is the structured result of analyzing a transcript.
type Analysis struct {
	Sentiment float64
	Themes    []string
}

// FeedbackStore persists feedback records.
type FeedbackStore interface {
	Create(ctx context.Context, f *Feedback) error
	Get(ctx context.Context, id string) (*Feedback, error)
	UpdateStatus(ctx context.Context, id string, s Status) error
	SaveResult(ctx context.Context, id string, a Analysis, transcript string) error
}

// AudioStorage stores raw audio files (S3/MinIO).
type AudioStorage interface {
	Put(ctx context.Context, key string, r io.Reader) error
	Get(ctx context.Context, key string) (io.ReadCloser, error)
}

// Transcriber turns audio into English text.
type Transcriber interface {
	Transcribe(ctx context.Context, audio io.Reader) (text string, lang string, err error)
}

// Analyzer extracts sentiment and themes from text.
type Analyzer interface {
	Analyze(ctx context.Context, text string) (Analysis, error)
}
