-- +goose Up
-- +goose StatementBegin

INSERT INTO identity_users (id, email, full_name, avatar_url, status, created_by, updated_by) 
VALUES 
    (1, 'admin@gobite.com', 'Admin', 'https://ui-avatars.com/api/?name=Admin', 2, 1, 1),
    (2, 'user@gobite.com', 'User', 'https://ui-avatars.com/api/?name=User', 2, 1, 1);

INSERT INTO identity_user_credentials (user_id, password) 
VALUES 
    (1, '$2a$12$CmaDcrhMrG9YMPxW2wWnmO/dAObJfNXHWmM0x45SzjP/nOB8Y1Rli'),
    (2, '$2a$12$CmaDcrhMrG9YMPxW2wWnmO/dAObJfNXHWmM0x45SzjP/nOB8Y1Rli');

-- "odwh-23IX-j5Pl" -- "LZEG-lnN4-w8hQ" -- "96qV-hU4h-I9XV" -- "bVjS-aFmF-VLBD" -- "V7Tr-dzp6-HABd" 
-- "Lfj2-advW-rqBT" -- "JxPf-jbod-QdtW" -- "Ycti-dfZp-jEhh" -- "Tdse-CWnw-eFiG" -- "cH2e-h2ad-tc0o"
INSERT INTO public.identity_mfa_backup_codes (id, user_id, code) 
VALUES
	(1, 2, '$argon2id$v=19$m=32768,t=3,p=2$5nxw7Ax0BCeF79WoUy7Qdg$YfdP3h5pswzFIOS1mwu6jDzZO/V0fKD/7qmz/FOTbas'),
	(2, 2, '$argon2id$v=19$m=32768,t=3,p=2$gHLxuEfI/ZJ1U69ZtXMenQ$vGz/ySVvbDkxdr6fmwHTkudWuBxyS6OpvEIDEF4uwQc'),
	(3, 2, '$argon2id$v=19$m=32768,t=3,p=2$Tz/kkoaY0nefXYLvydz/RQ$u5WQ+k0Wz8Rt+WCjUQsh8qUDq5HmnmkgnJU1Su3vYVE'),
	(4, 2, '$argon2id$v=19$m=32768,t=3,p=2$R1roJzoxVD66qeLJwnrF1g$KHVXTUGXhrDEtb8snhYpdMJhrynPow6RfvdDrIZbrds'),
	(5, 2, '$argon2id$v=19$m=32768,t=3,p=2$8bdxqUw5oY8z8FsV+TNX/Q$I0lBWlnso/lctwmQzhzjUVyyXvPGDMd0hlbYB/fH5tU'),
	(6, 2, '$argon2id$v=19$m=32768,t=3,p=2$I1WV8fRiHRVZp3kO4UgqoQ$ZngclOoGWc1PgslRIsCONlTIWWfy2qDhuCthgu/cwrg'),
	(7, 2, '$argon2id$v=19$m=32768,t=3,p=2$rtxiWdSOATPt0rAc1AGl0Q$HzAxlXBHjiXnD4nh6E2nf/JZVu7yt0sOuKfgix1yW0I'),
	(8, 2, '$argon2id$v=19$m=32768,t=3,p=2$6UyKv1TV9V21+xarSVpnWA$4y8NgaCxQCmJpPt9xtudQ8Eom2zcm0JRQAgnxsvMUS4'),
	(9, 2, '$argon2id$v=19$m=32768,t=3,p=2$4hO5a6rsauH07J0RAkm+Ug$4lcWHoY6FUry1CnbCoH9BafFUivDAcNgTgllwZPgQqQ'),
	(10, 2, '$argon2id$v=19$m=32768,t=3,p=2$J1c62QAYnxLaS/T5dTWnUg$EHzuFZDN0ilm2AvuK9f5VJRWvBl4PEF1a0SlhRENHHU');

-- key: KFP6EBHKHTWE2PHK5GOK7K2ARBZWQDBV
INSERT INTO public.identity_mfa_factors (id, user_id, "type", friendly_name, secret, key_version, is_verified) 
VALUES
	(1, 2, 1, 'user', decode('00013B264043A88BD1B2DAF89CFDA4511848117B551D66398E3F1AFA75ACD5F27F692840D2B96FD43E3AA7D821A3016FCA96938D42B66A311C2F1BDE3942','hex'), 1, true),
	(2, 2, 3, 'user', decode('','hex'), 1, true);


INSERT INTO identity_casbin_rules (ptype, v0, v1, v2) 
VALUES 
    ('p', 'admin', '*', '*'),
    ('p', 'viewer', 'identity:management:users', 'read'),
    ('p', 'user', 'domain:resource', 'action');

-- user-role assignments
INSERT INTO identity_casbin_rules (ptype, v0, v1) 
VALUES
    ('g', '1', 'admin'),
    ('g', '2', 'viewer');

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM identity_users WHERE id IN (1, 2);
DELETE FROM identity_user_credentials WHERE id IN (1, 2);
TRUNCATE TABLE identity_casbin_rules RESTART IDENTITY;
-- +goose StatementEnd
