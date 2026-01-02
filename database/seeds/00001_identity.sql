-- +goose Up
-- +goose StatementBegin

INSERT INTO identity_users (id, email, full_name, avatar_url, status) 
VALUES 
    (1, 'admin@gobite.com', 'Admin', 'https://ui-avatars.com/api/?name=Admin', 2),
    (2, 'user@gobite.com', 'User', 'https://ui-avatars.com/api/?name=User', 2),
    -- 
    (3, 'siti@gobite.com', 'Siti Rahma', 'https://ui-avatars.com/api/?name=Siti+Rahma', 3),
    (4, 'bima@gobite.com', 'Bima Pratama', 'https://ui-avatars.com/api/?name=Bima+Pratama', 4),
    (5, 'nabila@gobite.com', 'Nabila Azzahra', 'https://ui-avatars.com/api/?name=Nabila+Azzahra', 1),
    (6, 'rafi@gobite.com', 'Rafi Alfarizi', 'https://ui-avatars.com/api/?name=Rafi+Alfarizi', 2),
    (7, 'nadine@gobite.com', 'Nadine Putri', 'https://ui-avatars.com/api/?name=Nadine+Putri', 3),
    (8, 'arya@gobite.com', 'Arya Saputra', 'https://ui-avatars.com/api/?name=Arya+Saputra', 4),
    (9, 'intan@gobite.com', 'Intan Maharani', 'https://ui-avatars.com/api/?name=Intan+Maharani', 1),
    (10, 'fajar@gobite.com', 'Fajar Nugroho', 'https://ui-avatars.com/api/?name=Fajar+Nugroho', 2),
    (11, 'putra@gobite.com', 'Putra Wicaksono', 'https://ui-avatars.com/api/?name=Putra+Wicaksono', 3),
    (12, 'maya@gobite.com', 'Maya Salsabila', 'https://ui-avatars.com/api/?name=Maya+Salsabila', 4);


INSERT INTO identity_user_credentials (user_id, password) 
VALUES 
    (1, '$2a$12$CmaDcrhMrG9YMPxW2wWnmO/dAObJfNXHWmM0x45SzjP/nOB8Y1Rli'),
    (2, '$2a$12$CmaDcrhMrG9YMPxW2wWnmO/dAObJfNXHWmM0x45SzjP/nOB8Y1Rli'),
    (3, '$2a$12$CmaDcrhMrG9YMPxW2wWnmO/dAObJfNXHWmM0x45SzjP/nOB8Y1Rli'),
    (4, '$2a$12$CmaDcrhMrG9YMPxW2wWnmO/dAObJfNXHWmM0x45SzjP/nOB8Y1Rli'),
    (5, '$2a$12$CmaDcrhMrG9YMPxW2wWnmO/dAObJfNXHWmM0x45SzjP/nOB8Y1Rli'),
    (6, '$2a$12$CmaDcrhMrG9YMPxW2wWnmO/dAObJfNXHWmM0x45SzjP/nOB8Y1Rli'),
    (7, '$2a$12$CmaDcrhMrG9YMPxW2wWnmO/dAObJfNXHWmM0x45SzjP/nOB8Y1Rli'),
    (8, '$2a$12$CmaDcrhMrG9YMPxW2wWnmO/dAObJfNXHWmM0x45SzjP/nOB8Y1Rli'),
    (9, '$2a$12$CmaDcrhMrG9YMPxW2wWnmO/dAObJfNXHWmM0x45SzjP/nOB8Y1Rli'),
    (10, '$2a$12$CmaDcrhMrG9YMPxW2wWnmO/dAObJfNXHWmM0x45SzjP/nOB8Y1Rli'),
    (11, '$2a$12$CmaDcrhMrG9YMPxW2wWnmO/dAObJfNXHWmM0x45SzjP/nOB8Y1Rli'),
    (12, '$2a$12$CmaDcrhMrG9YMPxW2wWnmO/dAObJfNXHWmM0x45SzjP/nOB8Y1Rli');

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM identity_users WHERE id IN (1, 2);
DELETE FROM identity_user_credentials WHERE id IN (1, 2);
-- +goose StatementEnd
