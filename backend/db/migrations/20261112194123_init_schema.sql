-- Initial schema: the monitors to check and the results of each check.
-- Managed by goose: "Up" applies the migration, "Down" reverts it.
-- sqlc also reads this file to learn the table shapes for code generation.

-- +goose Up

-- monitors: one row per URL that Nexpulse checks on a schedule.
CREATE TABLE monitors (
    id               BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name             TEXT        NOT NULL,                 -- display name
    url              TEXT        NOT NULL,                 -- address to request
    method           TEXT        NOT NULL DEFAULT 'GET' CHECK (method IN ('GET', 'HEAD')),
    interval_seconds INT         NOT NULL DEFAULT 60 CHECK (interval_seconds BETWEEN 30 AND 3600), -- time between checks
    timeout_seconds  INT         NOT NULL DEFAULT 10 CHECK (timeout_seconds BETWEEN 1 AND 30),     -- how long to wait for a response
    expected_status  INT         NOT NULL DEFAULT 200,     -- HTTP status that counts as "up"
    is_active        BOOLEAN     NOT NULL DEFAULT TRUE,    -- false pauses checking without deleting
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    -- A check must finish before the next one is due, so checks never overlap.
    CHECK (timeout_seconds < interval_seconds)
);

-- check_results: one row per check performed against a monitor.
CREATE TABLE check_results (
    id               BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    -- ON DELETE CASCADE: deleting a monitor also deletes its history.
    monitor_id       BIGINT      NOT NULL REFERENCES monitors (id) ON DELETE CASCADE,
    checked_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    is_up            BOOLEAN     NOT NULL,
    -- The columns below are nullable: a request that times out or fails to
    -- connect has no status code or response time, only an error message.
    status_code      INT,
    response_time_ms INT,
    error_message    TEXT
);

-- Speeds up the common query "latest results for monitor X, newest first".
CREATE INDEX check_results_monitor_checked_idx
    ON check_results (monitor_id, checked_at DESC);

-- +goose Down
-- Drop check_results first because it references monitors.
DROP TABLE check_results;
DROP TABLE monitors;
