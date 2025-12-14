-- ----- ----- ----- ----- -----
-- users table
-- ----- ----- ----- ----- -----

-- name: UserGetByEmail :one
SELECT * FROM users 
WHERE 
    email = @email AND 
    deleted_at IS NULL;

-- name: UserGetByID :one
SELECT * FROM users 
WHERE 
    id = @id AND 
    deleted_at IS NULL;

-- name: UserGetValidByIDForUpdate :one
SELECT * FROM users 
WHERE 
    id = @id AND 
    status = 1 AND -- mean: user status is active
    deleted_at IS NULL
FOR UPDATE;

-- name: UserCreate :exec
INSERT INTO users (id, email, full_name, avatar_url, status)
VALUES (@id, @email, @full_name, @avatar_url, @status);

-- ----- ----- ----- ----- -----
-- user_credentials table
-- ----- ----- ----- ----- -----

-- name: UserCredentialGetByUserID :one
SELECT * FROM user_credentials 
WHERE 
    user_id = @user_id;

-- name: UserCredentialCreate :exec
INSERT INTO user_credentials (user_id, password)
VALUES (@user_id, @password);

-- name: UserCredentialUpdate :exec
UPDATE user_credentials 
SET 
    password = @password
WHERE user_id = @user_id;

-- ----- ----- ----- ----- -----
-- user_password_resets table
-- ----- ----- ----- ----- -----

-- name: UserPasswordResetCreate :exec
INSERT INTO user_password_resets (user_id, token, expires_at)
VALUES (@user_id, @token, @expires_at);

-- name: UserPasswordResetGetValidForUpdate :one
SELECT *
FROM user_password_resets 
WHERE
    token = @token AND
    used_at IS NULL AND
    expires_at > @now
FOR UPDATE;

-- name: UserPasswordResetMarkUsed :exec
UPDATE user_password_resets
SET used_at = @used_at
WHERE
    id = @id AND
    used_at IS NULL;

-- ----- ----- ----- ----- -----
-- mfa_factors table
-- ----- ----- ----- ----- -----

-- name: MfaFactorGetByUserID :many
SELECT * FROM mfa_factors 
WHERE 
    user_id = @user_id AND 
    is_verified = @is_verified;
