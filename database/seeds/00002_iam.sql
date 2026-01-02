-- +goose Up
-- +goose StatementBegin

INSERT INTO iam_casbin_rules (ptype, v0, v1, v2) 
VALUES 
    ('p', 'admin', '*', '*'),
    ('p', 'viewer', 'identity:management:users', 'read'),
    ('p', 'user', 'domain:resource', 'action');

-- user-role assignments
INSERT INTO iam_casbin_rules (ptype, v0, v1) VALUES
('g', '1', 'admin'),
('g', '2', 'viewer'),
('g', '6', 'user'),
('g', '10', 'user');

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
TRUNCATE TABLE iam_casbin_rules RESTART IDENTITY;
-- +goose StatementEnd
