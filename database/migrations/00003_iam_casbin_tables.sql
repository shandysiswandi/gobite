-- +goose Up
-- +goose StatementBegin
CREATE TABLE iam_casbin_rules (
    id BIGSERIAL PRIMARY KEY,
    ptype VARCHAR NOT NULL,
    v0 VARCHAR,
    v1 VARCHAR,
    v2 VARCHAR,
    v3 VARCHAR,
    v4 VARCHAR,
    v5 VARCHAR,

    -- Prevent duplicates (optional but recommended)
    CONSTRAINT uq_iam_casbin_rules UNIQUE (ptype, v0, v1, v2, v3, v4, v5)
);

CREATE INDEX IF NOT EXISTS idx_iam_casbin_rules_ptype ON iam_casbin_rules(ptype);
CREATE INDEX IF NOT EXISTS idx_iam_casbin_rules_v0 ON iam_casbin_rules(v0);
CREATE INDEX IF NOT EXISTS idx_iam_casbin_rules_v1 ON iam_casbin_rules(v1);
CREATE INDEX IF NOT EXISTS idx_iam_casbin_rules_v2 ON iam_casbin_rules(v2);
-- Often useful composite index for RBAC lookups
CREATE INDEX IF NOT EXISTS idx_iam_casbin_rules_ptype_v0 ON iam_casbin_rules(ptype, v0);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS iam_casbin_rules;
-- +goose StatementEnd
