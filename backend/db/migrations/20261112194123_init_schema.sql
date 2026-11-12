-- +goose Up
CREATE TABLE monitors (
    id               BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name             TEXT        NOT NULL,
    url              TEXT        NOT NULL,
    method           TEXT        NOT NULL DEFAULT 'GET' CHECK (method IN ('GET', 'HEAD')),
    interval_seconds INT         NOT NULL DEFAULT 60 CHECK (interval_seconds BETWEEN 30 AND 3600),
    timeout_seconds  INT         NOT NULL DEFAULT 10 CHECK (timeout_seconds BETWEEN 1 AND 30),
    expected_status  INT         NOT NULL DEFAULT 200,
    is_active        BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (timeout_seconds < interval_seconds)
);

CREATE TABLE check_results (
    id               BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    monitor_id       BIGINT      NOT NULL REFERENCES monitors (id) ON DELETE CASCADE,
    checked_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    is_up            BOOLEAN     NOT NULL,
    status_code      INT,
    response_time_ms INT,
    error_message    TEXT
);

CREATE INDEX check_results_monitor_checked_idx
    ON check_results (monitor_id, checked_at DESC);

-- +goose Down
DROP TABLE check_results;
DROP TABLE monitors;