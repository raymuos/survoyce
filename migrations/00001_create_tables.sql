-- +goose Up
CREATE TABLE businesses (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE feedback (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    business_id UUID NOT NULL REFERENCES businesses(id),
    audio_key   TEXT NOT NULL,
    transcript  TEXT NOT NULL DEFAULT '',
    sentiment   DOUBLE PRECISION NOT NULL DEFAULT 0,
    themes      TEXT[] NOT NULL DEFAULT '{}',
    status      TEXT NOT NULL DEFAULT 'pending'
                CHECK (status IN ('pending', 'processing', 'done', 'failed')),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_feedback_business_created ON feedback (business_id, created_at DESC);

-- +goose Down
DROP TABLE feedback;
DROP TABLE businesses;
