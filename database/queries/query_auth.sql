-- ----- ----- ----- ----- -----
-- users table
-- ----- ----- ----- ----- -----

-- name: UserGetByEmail :one
SELECT * FROM users 
WHERE 
    email = @email AND 
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

-- ----- ----- ----- ----- -----
-- mfa_factors table
-- ----- ----- ----- ----- -----

-- name: MfaFactorGetByUserID :many
SELECT * FROM mfa_factors 
WHERE 
    user_id = @user_id AND 
    is_verified = @is_verified;