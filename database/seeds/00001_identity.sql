-- +goose Up
-- +goose StatementBegin

INSERT INTO identity_users (id, email, full_name, avatar_url, status) 
VALUES 
    (1, 'admin@gobite.com', 'Admin', 'https://ui-avatars.com/api/?name=Admin', 2),
    (2, 'user@gobite.com', 'User', 'https://ui-avatars.com/api/?name=User', 2);

INSERT INTO identity_user_credentials (user_id, password) 
VALUES 
    (1, '$2a$12$CmaDcrhMrG9YMPxW2wWnmO/dAObJfNXHWmM0x45SzjP/nOB8Y1Rli'),
    (2, '$2a$12$CmaDcrhMrG9YMPxW2wWnmO/dAObJfNXHWmM0x45SzjP/nOB8Y1Rli');

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM identity_users WHERE id IN (1, 2);
DELETE FROM identity_user_credentials WHERE id IN (1, 2);
-- +goose StatementEnd
