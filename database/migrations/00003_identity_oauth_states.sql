-- +goose Up
-- +goose StatementBegin
CREATE TABLE identity_oauth_states (
    id BIGINT PRIMARY KEY,
    state VARCHAR NOT NULL,
    provider VARCHAR NOT NULL,
    code_verifier VARCHAR NOT NULL,
    redirect_path VARCHAR NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_identity_oauth_states_state UNIQUE(state)
);

CREATE INDEX idx_identity_oauth_states_expires_at ON identity_oauth_states(expires_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS identity_oauth_states;
-- +goose StatementEnd
