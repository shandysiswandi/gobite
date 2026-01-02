-- +goose Up
-- +goose StatementBegin

INSERT INTO iam_casbin_rules (ptype, v0, v1, v2) 
VALUES 
    ('p', 'admin', '*', '*'),
    ('p', 'user', 'domain:resource', 'action');

-- user-role assignments
INSERT INTO iam_casbin_rules (ptype, v0, v1) VALUES
('g', 'admin@gobite.com', 'admin'),
('g', 'user@gobite.com', 'user');

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
TRUNCATE TABLE iam_casbin_rules RESTART IDENTITY;
-- +goose StatementEnd
