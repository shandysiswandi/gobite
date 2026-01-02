-- ***** ***** *****
-- SELECT DATA
-- ***** ***** *****

-- name: GetIdentityUserLoginInfo :one
SELECT u.id, u.email, u.status, c.password, EXISTS (SELECT 1 FROM identity_mfa_factors m WHERE m.user_id = u.id AND m.is_verified = TRUE) AS has_mfa
FROM identity_users AS u
JOIN identity_user_credentials AS c ON u.id = c.user_id
WHERE 
    lower(u.email) = lower(@email) 
    AND u.deleted_at IS NULL;

-- name: GetIdentityUserCredentialInfo :one
SELECT u.id, u.email, u.status, c.password
FROM identity_users AS u
JOIN identity_user_credentials AS c ON u.id = c.user_id
WHERE
    u.id = @id
    AND u.deleted_at IS NULL;

-- name: GetIdentityUserByEmail :one
SELECT * 
FROM identity_users 
WHERE 
    lower(email) = lower(@email)
    AND deleted_at IS NULL;

-- name: GetIdentityUserByEmailIncludeDeleted :one
SELECT * 
FROM identity_users
WHERE 
    lower(email) = lower(@email);

-- name: GetIdentityChallengeUserByTokenPurpose :one
SELECT u.id AS user_id, u.status, u.email, c.id, c.token, c.purpose, c.metadata
FROM identity_challenges c
JOIN identity_users AS u ON u.id = c.user_id
WHERE 
    u.deleted_at IS NULL
    AND c.token = @token 
    AND c.purpose = @purpose 
    AND c.expires_at > NOW();

-- name: GetIdentityUserRefreshToken :one
SELECT rt.id, rt.user_id, rt.token, rt.expires_at, rt.revoked, rt.replaced_by_token_id, u.email, u.status AS user_status
FROM identity_refresh_tokens rt
JOIN identity_users u ON u.id = rt.user_id
WHERE 
    rt.token = @token
    AND u.deleted_at IS NULL;

-- name: GetIdentityMFAFactorByUserID :many
SELECT * 
FROM identity_mfa_factors 
WHERE 
    user_id = @user_id AND 
    is_verified = @is_verified
ORDER BY created_at ASC;

-- name: GetIdentityMFAFactorByID :one
SELECT * 
FROM identity_mfa_factors 
WHERE 
    id = @id AND 
    user_id = @user_id;

-- name: GetIdentityMFABackupCodeByUserID :many
SELECT * 
FROM identity_mfa_backup_codes 
WHERE 
    user_id = @user_id
    AND used_at IS NULL;

-- ***** ***** *****
-- CREATE DATA
-- ***** ***** *****

-- name: CreateIdentityRefreshToken :exec
INSERT INTO identity_refresh_tokens (id, user_id, token, expires_at, metadata) 
VALUES (@id, @user_id, @token, @expires_at, @metadata);

-- name: CreateIdentityChallenge :exec
INSERT INTO identity_challenges (id, user_id, token, purpose, expires_at, metadata) 
VALUES (@id, @user_id, @token, @purpose, @expires_at, @metadata);

-- name: CreateIdentityMFAFactor :exec
INSERT INTO identity_mfa_factors (id, user_id, type, friendly_name, secret, key_version, is_verified)
VALUES (@id, @user_id, @type, @friendly_name, @secret, @key_version, @is_verified);

-- name: CreateIdentityUser :exec
INSERT INTO identity_users (id, email, full_name, avatar_url, status)
VALUES (@id, @email, @full_name, @avatar_url, @status);

-- name: CreateIdentityUserCredential :exec
INSERT INTO identity_user_credentials (user_id, password)
VALUES (@user_id, @password);

-- name: CreateIdentityMFABackupCodes :copyfrom
INSERT INTO identity_mfa_backup_codes (id, user_id, code)
VALUES (@id, @user_id, @code);

-- ***** ***** *****
-- UPDATE DATA
-- ***** ***** *****

-- name: VerifyIdentityMFAFactor :exec
UPDATE identity_mfa_factors
SET 
    is_verified = TRUE
WHERE
    id = @id AND
    user_id = @user_id;

-- name: UpdateIdentityMFALastUsedAt :exec
UPDATE identity_mfa_factors
SET 
    last_used_at = NOW()
WHERE
    id = @id AND
    user_id = @user_id;

-- name: UpdateIdentityUserStatus :exec
UPDATE identity_users
SET 
    status = @new_status
WHERE 
    id = @id 
    AND status = @old_status
    AND deleted_at IS NULL;

-- name: RevokeIdentityRefreshToken :exec
UPDATE identity_refresh_tokens 
SET 
    revoked = TRUE
WHERE 
    token = @token;

-- name: RevokeAllIdentityRefreshToken :exec
UPDATE identity_refresh_tokens 
SET 
    revoked = TRUE
WHERE 
    user_id = @user_id;

-- name: ReplaceIdentityRefreshToken :execrows
UPDATE identity_refresh_tokens 
SET 
    revoked = TRUE, 
    replaced_by_token_id = @new_token_id::BIGINT
WHERE 
    id = @old_token_id;

-- name: UpdateIdentityUserName :exec
UPDATE identity_users
SET full_name = @full_name
WHERE
    id = @id AND
    deleted_at IS NULL;

-- name: UpdateIdentityUserAvatar :exec
UPDATE identity_users
SET avatar_url = @avatar_url
WHERE
    id = @id AND
    deleted_at IS NULL;

-- name: UpdateIdentityUserCredential :exec
UPDATE identity_user_credentials 
SET 
    password = @password
WHERE 
    user_id = @user_id;

-- name: MarkIdentityMFABackupCodeUsed :execrows
UPDATE identity_mfa_backup_codes
SET 
    used_at = NOW()
WHERE 
    user_id = @user_id 
    AND id = @id 
    AND used_at IS NULL;

-- ***** ***** *****
-- DELETE DATA
-- ***** ***** *****

-- name: DeleteIdentityChallengeByID :exec
DELETE FROM identity_challenges WHERE id = @id;

-- name: DeleteIdentityMFABackupCodeByUserID :exec
DELETE FROM identity_mfa_backup_codes WHERE user_id = @user_id;