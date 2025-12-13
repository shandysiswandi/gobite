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
-- mfa_factors table
-- ----- ----- ----- ----- -----

-- name: MfaFactorGetByUserID :many
SELECT * FROM mfa_factors 
WHERE 
    user_id = @user_id AND 
    is_verified = @is_verified;