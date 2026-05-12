-- +goose Up
CREATE TABLE IF NOT EXISTS vulnerabilities (
    -- No surrogate id — cve_id is the natural primary key.
    -- VARCHAR(20): CVE-YYYY- (9 chars) + up to 11-digit sequence = 20 chars max.
    cve_id     VARCHAR(20)              NOT NULL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE IF EXISTS vulnerabilities;
