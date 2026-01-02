-- +goose Up
-- +goose StatementBegin
CREATE TABLE identity_users (
    id BIGINT PRIMARY KEY,
    email VARCHAR NOT NULL,
    full_name VARCHAR NOT NULL,
    avatar_url VARCHAR NOT NULL,
    status SMALLINT NOT NULL DEFAULT 0, -- (e.g., 0: Unverified, 1: Active, 2: Banned).
    deleted_at TIMESTAMPTZ DEFAULT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by BIGINT NOT NULL,
    updated_by BIGINT NOT NULL,
    deleted_by BIGINT DEFAULT NULL,

    CONSTRAINT fk_identity_users_created_by 
        FOREIGN KEY (created_by) 
        REFERENCES identity_users(id) 
        ON DELETE SET NULL,
    CONSTRAINT fk_identity_users_updated_by 
        FOREIGN KEY (updated_by) 
        REFERENCES identity_users(id) 
        ON DELETE SET NULL,
    CONSTRAINT fk_identity_users_deleted_by 
        FOREIGN KEY (deleted_by) 
        REFERENCES identity_users(id) 
        ON DELETE SET NULL
);

-- Create a unique index on the lowercase version of the email.
-- This enforces uniqueness across all users, including those that are soft-deleted.
CREATE UNIQUE INDEX idx_identity_users_lower_case_email ON identity_users (lower(email));

CREATE TRIGGER trg_identity_users_set_updated_at
BEFORE UPDATE ON identity_users
FOR EACH ROW
EXECUTE FUNCTION trigger_set_timestamp();

CREATE TABLE identity_user_credentials (
    user_id BIGINT PRIMARY KEY,
    password VARCHAR NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT fk_identity_user_credentials_user
        FOREIGN KEY(user_id) 
        REFERENCES identity_users(id)
        ON DELETE CASCADE
);

CREATE TRIGGER trg_identity_user_credentials_set_updated_at
BEFORE UPDATE ON identity_user_credentials
FOR EACH ROW
EXECUTE FUNCTION trigger_set_timestamp();

CREATE TABLE identity_user_connections (
    id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    provider VARCHAR NOT NULL,
    provider_user_id VARCHAR NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_identity_user_connections_provider_provider_user_id UNIQUE(provider, provider_user_id),
    CONSTRAINT fk_identity_user_connections_user
        FOREIGN KEY(user_id) 
        REFERENCES identity_users(id)
        ON DELETE CASCADE
);

CREATE INDEX idx_identity_user_connections_user_id ON identity_user_connections(user_id);

CREATE TABLE identity_challenges (
    id BIGINT PRIMARY KEY,
    token VARCHAR NOT NULL, -- Store a hash, not the raw token
    user_id BIGINT NOT NULL,
    purpose SMALLINT NOT NULL DEFAULT 0, -- 0: unknown, 1: mfa_login, 2: password_reset_verify, 3: device_verify
    expires_at TIMESTAMPTZ NOT NULL,
    metadata JSONB DEFAULT '{}'::JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_identity_challenges_user 
        FOREIGN KEY(user_id) 
        REFERENCES identity_users(id) 
        ON DELETE CASCADE
);

CREATE UNIQUE INDEX idx_identity_challenges_token ON identity_challenges(token);
CREATE INDEX idx_identity_challenges_user_id ON identity_challenges(user_id);
CREATE INDEX idx_identity_challenges_expires_at ON identity_challenges(expires_at);

CREATE TABLE identity_refresh_tokens (
    id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    token VARCHAR NOT NULL, -- Store a hash, not the raw token
    expires_at TIMESTAMPTZ NOT NULL,
    revoked BOOLEAN NOT NULL DEFAULT FALSE,
    replaced_by_token_id BIGINT DEFAULT NULL,
    metadata JSONB DEFAULT '{}'::JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT fk_identity_refresh_tokens_user 
        FOREIGN KEY(user_id) 
        REFERENCES identity_users(id) 
        ON DELETE CASCADE
);

CREATE UNIQUE INDEX idx_identity_refresh_tokens_token ON identity_refresh_tokens(token);
CREATE INDEX idx_identity_refresh_tokens_user_id ON identity_refresh_tokens(user_id);

-- ---------------------------------
-- MFA (Multi-Factor Authentication)
-- ---------------------------------

CREATE TABLE identity_mfa_factors (
    id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    type SMALLINT NOT NULL, -- (e.g., 0: unknown, 1: totp, 2: sms)
    friendly_name VARCHAR NOT NULL,
    secret BYTEA NOT NULL, -- ciphertext of the shared secret, encrypted/peppered in app
    key_version SMALLINT NOT NULL DEFAULT 1, -- key version of rotation 
    is_verified BOOLEAN NOT NULL DEFAULT FALSE,
    last_used_at TIMESTAMPTZ DEFAULT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_identity_mfa_factors_user FOREIGN KEY(user_id) REFERENCES identity_users(id) ON DELETE CASCADE
);

CREATE INDEX idx_identity_mfa_factors_user_id ON identity_mfa_factors(user_id);

CREATE TRIGGER trg_identity_mfa_factors_set_updated_at
BEFORE UPDATE ON identity_mfa_factors
FOR EACH ROW
EXECUTE FUNCTION trigger_set_timestamp();

CREATE TABLE identity_mfa_backup_codes (
    id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    code VARCHAR NOT NULL,
    used_at TIMESTAMPTZ DEFAULT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_identity_mfa_backup_codes_user_code UNIQUE(user_id, code),
    CONSTRAINT fk_identity_mfa_backup_codes_user 
        FOREIGN KEY(user_id) 
        REFERENCES identity_users(id) 
        ON DELETE CASCADE
);

CREATE INDEX idx_identity_mfa_backup_codes_user_id ON identity_mfa_backup_codes(user_id);

-- ---------------------------------
-- Authz Casbin
-- ---------------------------------
CREATE TABLE identity_casbin_rules (
    id BIGSERIAL PRIMARY KEY,
    ptype VARCHAR NOT NULL,
    v0 VARCHAR,
    v1 VARCHAR,
    v2 VARCHAR,
    v3 VARCHAR,
    v4 VARCHAR,
    v5 VARCHAR,

    -- Prevent duplicates (optional but recommended)
    CONSTRAINT uq_identity_casbin_rules UNIQUE (ptype, v0, v1, v2, v3, v4, v5)
);

CREATE INDEX IF NOT EXISTS idx_identity_casbin_rules_ptype ON identity_casbin_rules(ptype);
CREATE INDEX IF NOT EXISTS idx_identity_casbin_rules_v0 ON identity_casbin_rules(v0);
CREATE INDEX IF NOT EXISTS idx_identity_casbin_rules_v1 ON identity_casbin_rules(v1);
CREATE INDEX IF NOT EXISTS idx_identity_casbin_rules_v2 ON identity_casbin_rules(v2);
-- Often useful composite index for RBAC lookups
CREATE INDEX IF NOT EXISTS idx_identity_casbin_rules_ptype_v0 ON identity_casbin_rules(ptype, v0);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trg_identity_mfa_factors_set_updated_at ON identity_mfa_factors;
DROP TRIGGER IF EXISTS trg_identity_user_credentials_set_updated_at ON identity_user_credentials;
DROP TRIGGER IF EXISTS trg_identity_users_set_updated_at ON identity_users;

DROP TABLE IF EXISTS identity_mfa_backup_codes;
DROP TABLE IF EXISTS identity_mfa_factors;

DROP TABLE IF EXISTS identity_refresh_tokens;
DROP TABLE IF EXISTS identity_challenges;
DROP TABLE IF EXISTS identity_user_connections;
DROP TABLE IF EXISTS identity_user_credentials;
DROP TABLE IF EXISTS identity_users;
DROP TABLE IF EXISTS identity_casbin_rules;
-- +goose StatementEnd
