-- Queries for the monitors table. sqlc turns each "-- name:" block into a Go
-- method in internal/db/monitors.sql.go; the suffix sets what it returns
-- (:one = a single row, :many = a slice, :execrows = rows affected).
-- Comments placed right after a "-- name:" line become that method's doc comment.

-- name: CreateMonitor :one
-- CreateMonitor inserts a new monitor and returns it with its generated ID and
-- timestamps. is_active is left to its default (true).
INSERT INTO monitors (name, url, method, interval_seconds, timeout_seconds, expected_status)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetMonitor :one
-- GetMonitor returns the monitor with the given ID, or pgx.ErrNoRows if none exists.
SELECT * FROM monitors
WHERE id = $1;

-- name: ListMonitors :many
-- ListMonitors returns every monitor, newest first.
SELECT * FROM monitors
ORDER BY created_at DESC;

-- name: DeleteMonitor :execrows
-- DeleteMonitor removes a monitor (and, by cascade, its check results) and
-- returns the number of rows deleted; 0 means the ID did not exist.
DELETE FROM monitors
WHERE id = $1;

-- name: UpdateMonitor :one
-- UpdateMonitor replaces every editable field of a monitor, bumps updated_at,
-- and returns the updated row.
UPDATE monitors
SET name = $2,
    url = $3,
    method = $4,
    interval_seconds = $5,
    timeout_seconds = $6,
    expected_status = $7,
    is_active = $8,
    updated_at = now()
WHERE id = $1
RETURNING *;
