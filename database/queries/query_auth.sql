-- ***** ***** *****
-- SELECT DATA
-- ***** ***** *****

-- name: GetAuthUserLoginInfo :one
SELECT u.id, u.email, u.status, c.password, EXISTS (SELECT 1 FROM auth_mfa_factors m WHERE m.user_id = u.id AND m.is_verified = TRUE) AS has_mfa
FROM auth_users AS u
JOIN auth_user_credentials AS c ON u.id = c.user_id
WHERE 
    lower(u.email) = lower(@email) 
    AND u.deleted_at IS NULL;

-- name: GetAuthUserCredentialInfo :one
SELECT u.id, u.email, u.status, c.password
FROM auth_users AS u
JOIN auth_user_credentials AS c ON u.id = c.user_id
WHERE
    u.id = @id
    AND u.deleted_at IS NULL;

-- name: GetAuthUserByEmail :one
SELECT * 
FROM auth_users 
WHERE 
    lower(email) = lower(@email)
    AND deleted_at IS NULL;

-- name: GetAuthUserByEmailIncludeDeleted :one
SELECT * 
FROM auth_users
WHERE 
    lower(email) = lower(@email);

-- name: GetAuthChallengeUserByTokenPurpose :one
SELECT u.id AS user_id, u.status, u.email, c.id, c.token, c.purpose, c.metadata
FROM auth_challenges c
JOIN auth_users AS u ON u.id = c.user_id
WHERE 
    c.token = @token 
    AND c.purpose = @purpose 
    AND c.expires_at > NOW();

-- name: GetAuthUserRefreshToken :one
SELECT rt.id, rt.user_id, rt.token, rt.expires_at, rt.revoked, rt.replaced_by_token_id, u.email, u.status AS user_status
FROM auth_refresh_tokens rt
JOIN auth_users u ON u.id = rt.user_id
WHERE 
    rt.token = @token
    AND u.deleted_at IS NULL;

-- name: GetAuthMFAFactorByUserID :many
SELECT * FROM auth_mfa_factors 
WHERE 
    user_id = @user_id AND 
    is_verified = @is_verified
ORDER BY created_at ASC;

-- ***** ***** *****
-- CREATE DATA
-- ***** ***** *****

-- name: CreateAuthRefreshToken :exec
INSERT INTO auth_refresh_tokens (id, user_id, token, expires_at, metadata) 
VALUES (@id, @user_id, @token, @expires_at, @metadata);

-- name: CreateAuthChallenge :exec
INSERT INTO auth_challenges (id, user_id, token, purpose, expires_at, metadata) 
VALUES (@id, @user_id, @token, @purpose, @expires_at, @metadata);

-- name: CreateAuthMFAFactor :exec
INSERT INTO auth_mfa_factors (id, user_id, type, friendly_name, secret, key_version, is_verified)
VALUES (@id, @user_id, @type, @friendly_name, @secret, @key_version, @is_verified);

-- name: CreateAuthUser :exec
INSERT INTO auth_users (id, email, full_name, avatar_url, status)
VALUES (@id, @email, @full_name, @avatar_url, @status);

-- name: CreateAuthUserCredential :exec
INSERT INTO auth_user_credentials (user_id, password)
VALUES (@user_id, @password);

-- ***** ***** *****
-- UPDATE DATA
-- ***** ***** *****

-- name: RevokeAuthRefreshToken :exec
UPDATE auth_refresh_tokens 
SET 
    revoked = TRUE
WHERE 
    token = @token;

-- name: RevokeAllAuthRefreshToken :exec
UPDATE auth_refresh_tokens 
SET 
    revoked = TRUE
WHERE 
    user_id = @user_id;

-- name: ReplaceAuthRefreshToken :exec
UPDATE auth_refresh_tokens 
SET 
    revoked = TRUE, 
    replaced_by_token_id = @new_token_id::BIGINT
WHERE 
    id = @old_token_id;

-- name: UpdateAuthUserName :exec
UPDATE auth_users
SET full_name = @full_name
WHERE
    id = @id AND
    deleted_at IS NULL;

-- name: UpdateAuthUserCredential :exec
UPDATE auth_user_credentials 
SET 
    password = @password
WHERE 
    user_id = @user_id;

-- name: MarkUsedAuthChallengeByID :exec
UPDATE auth_challenges 
SET
    used_at = @used_at
WHERE 
    id = @id;

-- ***** ***** *****
-- DELETE DATA
-- ***** ***** *****

-- name: DeleteAuthChallengeByID :exec
DELETE FROM auth_challenges WHERE id = @id;
