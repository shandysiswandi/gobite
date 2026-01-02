-- +goose Up
-- +goose StatementBegin
-- Create a shared trigger function to automatically update the updated_at timestamp.
CREATE OR REPLACE FUNCTION trigger_set_timestamp()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP FUNCTION IF EXISTS trigger_set_timestamp();
-- +goose StatementEnd
