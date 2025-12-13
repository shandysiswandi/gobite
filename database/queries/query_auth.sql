-- name: UserGetByEmail :one
SELECT * FROM users 
WHERE 
    email = @email AND 
    deleted_at IS NULL;

-- ----- ----- ----- ----- -----
-- 
-- ----- ----- ----- ----- -----

-- name: UserCredentialGetByUserID :one
SELECT * FROM user_credentials 
WHERE 
    user_id = @user_id;

-- ----- ----- ----- ----- -----
-- 
-- ----- ----- ----- ----- -----

-- name: MfaFactorGetByUserID :many
SELECT * FROM mfa_factors 
WHERE 
    user_id = @user_id AND 
    is_verified = @is_verified;