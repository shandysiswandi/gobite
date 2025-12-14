-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
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
CREATE UNIQUE INDEX idx_users_lower_case_email ON users (lower(email));

CREATE TRIGGER trg_users_set_updated_at
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION trigger_set_timestamp();

CREATE TABLE user_credentials (
    user_id BIGINT PRIMARY KEY,
    password VARCHAR NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT fk_user_credentials_user
        FOREIGN KEY(user_id) 
        REFERENCES users(id)
        ON DELETE CASCADE
);

CREATE TRIGGER trg_user_credentials_set_updated_at
BEFORE UPDATE ON user_credentials
FOR EACH ROW
EXECUTE FUNCTION trigger_set_timestamp();

CREATE TABLE user_password_resets (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    token VARCHAR NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ DEFAULT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_user_password_resets_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

CREATE UNIQUE INDEX idx_user_password_resets_token ON user_password_resets(token);
CREATE INDEX idx_user_password_resets_user_id ON user_password_resets(user_id);

CREATE TABLE user_connections (
    id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    provider VARCHAR NOT NULL,
    provider_user_id VARCHAR NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_user_connections_user
        FOREIGN KEY(user_id) 
        REFERENCES users(id)
        ON DELETE CASCADE,
        
    UNIQUE(provider, provider_user_id)
);

CREATE INDEX idx_user_connections_user_id ON user_connections(user_id);

CREATE TABLE permissions (
    id BIGSERIAL PRIMARY KEY,
    action VARCHAR NOT NULL,
    resource VARCHAR NOT NULL,
    description VARCHAR NOT NULL,
    UNIQUE(action, resource)
);

CREATE TABLE roles (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR UNIQUE NOT NULL,
    description VARCHAR NOT NULL
);

CREATE TABLE role_permissions (
    role_id BIGINT NOT NULL,
    permission_id BIGINT NOT NULL,
    PRIMARY KEY (role_id, permission_id),
    CONSTRAINT fk_role_permissions_role FOREIGN KEY(role_id) REFERENCES roles(id) ON DELETE CASCADE,
    CONSTRAINT fk_permission FOREIGN KEY(permission_id) REFERENCES permissions(id) ON DELETE CASCADE
);

CREATE INDEX idx_role_permissions_permission_id ON role_permissions(permission_id);

CREATE TABLE user_roles (
    user_id BIGINT NOT NULL,
    role_id BIGINT NOT NULL,
    PRIMARY KEY (user_id, role_id),
    CONSTRAINT fk_user_roles_user FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_user_roles_role FOREIGN KEY(role_id) REFERENCES roles(id) ON DELETE CASCADE
);

CREATE INDEX idx_user_roles_role_id ON user_roles(role_id);

-- ---------------------------------
-- MFA (Multi-Factor Authentication)
-- ---------------------------------

CREATE TABLE mfa_factors (
    id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    type SMALLINT NOT NULL, -- (e.g., 0: unknown, 1: totp, 2: sms, 3: backup_code)
    friendly_name VARCHAR NOT NULL,
    secret BYTEA NOT NULL, -- ciphertext of the shared secret, encrypted/peppered in app
    key_version SMALLINT NOT NULL DEFAULT 1, -- key version of rotation 
    is_verified BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_mfa_factors_user FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_mfa_factors_user_id ON mfa_factors(user_id);

CREATE TRIGGER trg_mfa_factors_set_updated_at
BEFORE UPDATE ON mfa_factors
FOR EACH ROW
EXECUTE FUNCTION trigger_set_timestamp();

CREATE TABLE mfa_backup_codes (
    id BIGINT PRIMARY KEY,
    mfa_factor_id BIGINT NOT NULL,
    code VARCHAR NOT NULL, -- hashed
    is_used BOOLEAN NOT NULL DEFAULT FALSE,
    CONSTRAINT fk_mfa_backup_codes_mfa_factor FOREIGN KEY(mfa_factor_id) REFERENCES mfa_factors(id) ON DELETE CASCADE
);

CREATE INDEX idx_mfa_backup_codes_mfa_factor_id ON mfa_backup_codes(mfa_factor_id);

-- ---------------------------------
-- Seeder -- Secret123!
-- ---------------------------------

INSERT INTO users (id, email, full_name, avatar_url, status) VALUES (1, 'admin@admin.com', 'Admin', '', 1);

INSERT INTO user_credentials (user_id, password) 
VALUES (1, '$2a$12$CmaDcrhMrG9YMPxW2wWnmO/dAObJfNXHWmM0x45SzjP/nOB8Y1Rli');

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trg_mfa_factors_set_updated_at ON mfa_factors;
DROP TRIGGER IF EXISTS trg_user_credentials_set_updated_at ON user_credentials;
DROP TRIGGER IF EXISTS trg_users_set_updated_at ON users;
DROP TABLE IF EXISTS mfa_backup_codes;
DROP TABLE IF EXISTS mfa_factors;
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS user_connections;
DROP TABLE IF EXISTS user_credentials;
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
