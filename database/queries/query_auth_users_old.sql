-- -- name: UserGetByEmailStatus :one
-- SELECT * FROM auth_users 
-- WHERE 
--     email = @email AND 
--     status = @status AND
--     deleted_at IS NULL;

-- -- name: UserGetByID :one
-- SELECT * FROM auth_users 
-- WHERE 
--     id = @id AND 
--     deleted_at IS NULL;

-- -- name: UserGetByIDStatus :one
-- SELECT * FROM auth_users 
-- WHERE 
--     id = @id AND 
--     status = @status AND
--     deleted_at IS NULL;

-- -- name: UserGetValidByIDForUpdate :one
-- SELECT * FROM auth_users 
-- WHERE 
--     id = @id AND 
--     status = 1 AND -- mean: user status is active
--     deleted_at IS NULL
-- FOR UPDATE;

-- -- name: UserUpdateStatus :exec
-- UPDATE auth_users
-- SET status = @new_status
-- WHERE
--     id = @id AND
--     status = @old_status AND
--     deleted_at IS NULL;

-- -- name: UserCredentialGetByUserID :one
-- SELECT * FROM auth_user_credentials 
-- WHERE 
--     user_id = @user_id;

-- -- name: UserPasswordResetGetValidForUpdate :one
-- SELECT *
-- FROM auth_user_password_resets 
-- WHERE
--     token = @token AND
--     used_at IS NULL AND
--     expires_at > @now
-- FOR UPDATE;

-- -- name: UserPasswordResetMarkUsed :exec
-- UPDATE auth_user_password_resets
-- SET used_at = @used_at
-- WHERE
--     id = @id AND
--     used_at IS NULL;