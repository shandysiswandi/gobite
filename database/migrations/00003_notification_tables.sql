-- +goose Up
-- +goose StatementBegin

-- Group notifications (e.g., 'Security', 'Social', 'Promotions') to manage user settings.
CREATE TABLE notification_categories (
    id BIGINT PRIMARY KEY,
    name VARCHAR NOT NULL UNIQUE, -- e.g. 'security', 'social'
    description VARCHAR NOT NULL, -- e.g. 'Password resets and login alerts'
    is_mandatory BOOLEAN NOT NULL DEFAULT FALSE, -- If true, users cannot disable this category
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TRIGGER trg_notification_categories_set_updated_at
BEFORE UPDATE ON notification_categories
FOR EACH ROW
EXECUTE FUNCTION trigger_set_timestamp();

-- This service owns push tokens because they are strictly for delivery.
CREATE TABLE notification_user_devices (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    device_token VARCHAR NOT NULL, -- FCM/APNS Token
    platform VARCHAR NOT NULL, -- android, ios, web
    last_active_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    UNIQUE(device_token)
);
CREATE INDEX idx_notification_user_devices_user_id ON notification_user_devices(user_id);

-- Stores user preferences. Absence of a record usually implies "Default On" in logic.
CREATE TABLE notification_user_settings (
    user_id BIGINT NOT NULL,
    category_id BIGINT NOT NULL,
    channel SMALLINT NOT NULL, -- (0: unknwon, 1: In-App, 2: Email, 3: SMS, 4: Push)
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (user_id, category_id, channel),
    
    CONSTRAINT fk_notification_user_settings_category 
        FOREIGN KEY(category_id) REFERENCES notification_categories(id) ON DELETE CASCADE
);

CREATE TRIGGER trg_notification_user_settings_set_updated_at
BEFORE UPDATE ON notification_user_settings
FOR EACH ROW
EXECUTE FUNCTION trigger_set_timestamp();

-- Maps specific events (trigger_key) to their text content or external template IDs.
CREATE TABLE notification_templates (
    id BIGINT PRIMARY KEY,
    trigger_key VARCHAR NOT NULL,   -- The unique event name used in code (e.g., 'password_reset', 'new_comment')
    category_id BIGINT NOT NULL,    -- Links event to the broad category
    channel SMALLINT NOT NULL,      -- (0: unknwon, 1: In-App, 2: Email, 3: SMS, 4: Push)
    subject VARCHAR NOT NULL, 
    body TEXT NOT NULL,    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Prevent duplicate templates for the same event+channel
    UNIQUE (trigger_key, channel),

    CONSTRAINT fk_notification_templates_category 
        FOREIGN KEY(category_id) REFERENCES notification_categories(id) ON DELETE CASCADE
);

CREATE TRIGGER trg_notification_templates_set_updated_at
BEFORE UPDATE ON notification_templates
FOR EACH ROW
EXECUTE FUNCTION trigger_set_timestamp();

-- The central history table. This serves as the "In-App Notification Box".
CREATE TABLE notifications (
    id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    category_id BIGINT NOT NULL,
    trigger_key VARCHAR NOT NULL, -- Keeps track of which event caused this
    -- Data: The content payload for rendering the message
    -- Example: { "username": "John", "comment_snippet": "Nice post!" }
    data JSONB NOT NULL DEFAULT '{}'::JSONB, 
    -- Metadata: System references for logic/routing
    -- Example: { "actor_id": 55, "resource_id": "1023", "resource_type": "post", "click_action": "/posts/1023" }
    metadata JSONB NOT NULL DEFAULT '{}'::JSONB,
    
    read_at TIMESTAMPTZ DEFAULT NULL, -- NULL = Unread
    deleted_at TIMESTAMPTZ DEFAULT NULL, -- Soft delete
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_notifications_category 
        FOREIGN KEY(category_id) REFERENCES notification_categories(id) ON DELETE CASCADE
);

-- Index for counting unread items quickly
CREATE INDEX idx_notifications_user_unread ON notifications(user_id) WHERE read_at IS NULL;
-- Index for showing notification list in order
CREATE INDEX idx_notifications_user_created ON notifications(user_id, created_at DESC);

-- Queues external messages (Email/SMS) so the main API response isn't slowed down.
CREATE TABLE notification_delivery_logs (
    id BIGSERIAL PRIMARY KEY,
    notification_id BIGINT NOT NULL,
    channel SMALLINT NOT NULL, -- -- (0: unknwon, 1: In-App, 2: Email, 3: SMS, 4: Push)
    
    -- Processing status
    status SMALLINT NOT NULL DEFAULT 0, -- (e.g., 0: queued, 1: processing, 2: sent, 3: failed)
    attempt_count INT NOT NULL DEFAULT 0,
    next_retry_at TIMESTAMPTZ DEFAULT NULL,
    
    -- Provider feedback
    provider_response JSONB DEFAULT NULL,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_delivery_logs_notification 
        FOREIGN KEY(notification_id) REFERENCES notifications(id) ON DELETE CASCADE
);

CREATE INDEX idx_notification_delivery_logs_status ON notification_delivery_logs(status) WHERE status IN (0, 1);

CREATE TRIGGER trg_notification_delivery_logs_set_updated_at
BEFORE UPDATE ON notification_delivery_logs
FOR EACH ROW
EXECUTE FUNCTION trigger_set_timestamp();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trg_notification_delivery_logs_set_updated_at ON notification_delivery_logs;
DROP TABLE IF EXISTS notification_delivery_logs;

DROP TABLE IF EXISTS notifications;

DROP TRIGGER IF EXISTS trg_notification_templates_set_updated_at ON notification_templates;
DROP TABLE IF EXISTS notification_templates;

DROP TRIGGER IF EXISTS trg_notification_user_settings_set_updated_at ON notification_user_settings;
DROP TABLE IF EXISTS notification_user_settings;

DROP TABLE IF EXISTS notification_user_devices;

DROP TRIGGER IF EXISTS trg_notification_categories_set_updated_at ON notification_categories;
DROP TABLE IF EXISTS notification_categories;
-- +goose StatementEnd
