-- +goose Up
-- +goose StatementBegin
CREATE TABLE auth_users (
    id BIGINT PRIMARY KEY,
    email VARCHAR NOT NULL,
    full_name VARCHAR NOT NULL,
    avatar_url VARCHAR NOT NULL,
    status SMALLINT NOT NULL DEFAULT 0, -- (e.g., 0: Unverified, 1: Active, 2: Banned).
    deleted_at TIMESTAMPTZ DEFAULT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create a unique index on the lowercase version of the email.
-- This enforces uniqueness across all users, including those that are soft-deleted.
CREATE UNIQUE INDEX idx_auth_users_lower_case_email ON auth_users (lower(email));

CREATE TRIGGER trg_auth_users_set_updated_at
BEFORE UPDATE ON auth_users
FOR EACH ROW
EXECUTE FUNCTION trigger_set_timestamp();

CREATE TABLE auth_user_credentials (
    user_id BIGINT PRIMARY KEY,
    password VARCHAR NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT fk_auth_user_credentials_user
        FOREIGN KEY(user_id) 
        REFERENCES auth_users(id)
        ON DELETE CASCADE
);

CREATE TRIGGER trg_auth_user_credentials_set_updated_at
BEFORE UPDATE ON auth_user_credentials
FOR EACH ROW
EXECUTE FUNCTION trigger_set_timestamp();

CREATE TABLE auth_user_connections (
    id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    provider VARCHAR NOT NULL,
    provider_user_id VARCHAR NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_auth_user_connections_user
        FOREIGN KEY(user_id) 
        REFERENCES auth_users(id)
        ON DELETE CASCADE,
        
    UNIQUE(provider, provider_user_id)
);

CREATE INDEX idx_auth_user_connections_user_id ON auth_user_connections(user_id);

CREATE TABLE auth_challenges (
    id BIGINT PRIMARY KEY,
    token VARCHAR NOT NULL, -- Store a hash, not the raw token
    user_id BIGINT NOT NULL,
    purpose SMALLINT NOT NULL DEFAULT 0, -- 0: unknown, 1: mfa_login, 2: password_reset_verify, 3: device_verify
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ DEFAULT NULL,
    metadata JSONB DEFAULT '{}'::JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_auth_challenges_user 
        FOREIGN KEY(user_id) 
        REFERENCES auth_users(id) 
        ON DELETE CASCADE
);

CREATE UNIQUE INDEX idx_auth_challenges_token ON auth_challenges(token);
CREATE INDEX idx_auth_challenges_user_id ON auth_challenges(user_id);
CREATE INDEX idx_auth_challenges_expires_at ON auth_challenges(expires_at);

CREATE TABLE auth_refresh_tokens (
    id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    token VARCHAR NOT NULL, -- Store a hash, not the raw token
    expires_at TIMESTAMPTZ NOT NULL,
    revoked BOOLEAN NOT NULL DEFAULT FALSE,
    replaced_by_token_id BIGINT DEFAULT NULL,
    metadata JSONB DEFAULT '{}'::JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT fk_auth_refresh_tokens_user 
        FOREIGN KEY(user_id) 
        REFERENCES auth_users(id) 
        ON DELETE CASCADE
);

CREATE UNIQUE INDEX idx_auth_refresh_tokens_token ON auth_refresh_tokens(token);
CREATE INDEX idx_auth_refresh_tokens_user_id ON auth_refresh_tokens(user_id);


-- ---------------------------------
-- RBAC (Role Based Access Control)
-- ---------------------------------

CREATE TABLE auth_permissions (
    id BIGINT PRIMARY KEY,
    action VARCHAR NOT NULL,
    resource VARCHAR NOT NULL,
    description VARCHAR NOT NULL,
    UNIQUE(action, resource)
);

CREATE TABLE auth_roles (
    id BIGINT PRIMARY KEY,
    name VARCHAR UNIQUE NOT NULL,
    description VARCHAR NOT NULL
);

CREATE TABLE auth_role_permissions (
    role_id BIGINT NOT NULL,
    permission_id BIGINT NOT NULL,
    PRIMARY KEY (role_id, permission_id),
    CONSTRAINT fk_auth_role_permissions_role FOREIGN KEY(role_id) REFERENCES auth_roles(id) ON DELETE CASCADE,
    CONSTRAINT fk_auth_permission FOREIGN KEY(permission_id) REFERENCES auth_permissions(id) ON DELETE CASCADE
);

CREATE INDEX idx_auth_role_permissions_permission_id ON auth_role_permissions(permission_id);

CREATE TABLE auth_user_roles (
    user_id BIGINT NOT NULL,
    role_id BIGINT NOT NULL,
    PRIMARY KEY (user_id, role_id),
    CONSTRAINT fk_auth_user_roles_user FOREIGN KEY(user_id) REFERENCES auth_users(id) ON DELETE CASCADE,
    CONSTRAINT fk_auth_user_roles_role FOREIGN KEY(role_id) REFERENCES auth_roles(id) ON DELETE CASCADE
);

CREATE INDEX idx_auth_user_roles_role_id ON auth_user_roles(role_id);

-- ---------------------------------
-- MFA (Multi-Factor Authentication)
-- ---------------------------------

CREATE TABLE auth_mfa_factors (
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
    CONSTRAINT fk_auth_mfa_factors_user FOREIGN KEY(user_id) REFERENCES auth_users(id) ON DELETE CASCADE
);

CREATE INDEX idx_auth_mfa_factors_user_id ON auth_mfa_factors(user_id);

CREATE TRIGGER trg_auth_mfa_factors_set_updated_at
BEFORE UPDATE ON auth_mfa_factors
FOR EACH ROW
EXECUTE FUNCTION trigger_set_timestamp();

CREATE TABLE auth_mfa_backup_codes (
    id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    code VARCHAR NOT NULL,
    used_at TIMESTAMPTZ DEFAULT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT fk_auth_mfa_backup_codes_user 
        FOREIGN KEY(user_id) 
        REFERENCES auth_users(id) 
        ON DELETE CASCADE
);

CREATE INDEX idx_auth_mfa_backup_codes_user_id ON auth_mfa_backup_codes(user_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trg_auth_mfa_factors_set_updated_at ON auth_mfa_factors;
DROP TRIGGER IF EXISTS trg_auth_user_credentials_set_updated_at ON auth_user_credentials;
DROP TRIGGER IF EXISTS trg_auth_users_set_updated_at ON auth_users;

DROP TABLE IF EXISTS auth_mfa_backup_codes;
DROP TABLE IF EXISTS auth_mfa_factors;

DROP TABLE IF EXISTS auth_user_roles;
DROP TABLE IF EXISTS auth_role_permissions;
DROP TABLE IF EXISTS auth_roles;
DROP TABLE IF EXISTS auth_permissions;

DROP TABLE IF EXISTS auth_refresh_tokens;
DROP TABLE IF EXISTS auth_challenges;
DROP TABLE IF EXISTS auth_user_connections;
DROP TABLE IF EXISTS auth_user_credentials;
DROP TABLE IF EXISTS auth_users;
-- +goose StatementEnd
