-- +goose Up
-- +goose StatementBegin

-- ---------------------------------
-- Seeder -- pass: Secret123!, code: 123456
-- ---------------------------------

INSERT INTO auth_users (id, email, full_name, avatar_url, status) 
VALUES (1, 'admin@admin.com', 'Admin', '', 2);

INSERT INTO auth_user_credentials (user_id, password) 
VALUES (1, '$2a$12$CmaDcrhMrG9YMPxW2wWnmO/dAObJfNXHWmM0x45SzjP/nOB8Y1Rli');

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM auth_users WHERE id = 1;
-- +goose StatementEnd
