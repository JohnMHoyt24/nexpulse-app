-- name: CreateMonitor :one
INSERT INTO monitors (name, url, method, interval_seconds, timeout_seconds, expected_status)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetMonitor :one
SELECT * FROM monitors
WHERE id = $1;

-- name: ListMonitors :many
SELECT * FROM monitors
ORDER BY created_at DESC;

-- name: DeleteMonitor :execrows
DELETE FROM monitors
WHERE id = $1;

-- name: UpdateMonitor :one
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