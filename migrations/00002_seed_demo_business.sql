-- +goose Up
-- Dev-only: a fixed demo business so you can upload feedback before building auth.
INSERT INTO businesses (id, name)
VALUES ('00000000-0000-0000-0000-000000000001', 'Demo Business');

-- +goose Down
DELETE FROM businesses WHERE id = '00000000-0000-0000-0000-000000000001';
