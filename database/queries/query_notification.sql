-- ***** ***** *****
-- SELECT DATA
-- ***** ***** *****

-- name: GetNotificationTemplateByTriggerChannel :one
SELECT id, trigger_key, category_id, channel, subject, body
FROM notification_templates
WHERE 
    trigger_key = @trigger_key AND 
    channel = @channel;

-- name: ListNotificationCategories :many
SELECT id, name, description, is_mandatory
FROM notification_categories
ORDER BY id ASC;

-- name: ListNotificationUserSettings :many
SELECT user_id, category_id, channel, is_enabled
FROM notification_user_settings
WHERE 
    user_id = @user_id;

-- name: ListNotificationsByUserAll :many
SELECT id, user_id, category_id, trigger_key, data, metadata, read_at, created_at
FROM notifications
WHERE 
    user_id = @user_id AND 
    deleted_at IS NULL
ORDER BY 
    created_at DESC, 
    id DESC
LIMIT @page_limit OFFSET @page_offset;

-- name: ListNotificationsByUserUnread :many
SELECT id, user_id, category_id, trigger_key, data, metadata, read_at, created_at
FROM notifications
WHERE 
    user_id = @user_id AND 
    read_at IS NULL AND 
    deleted_at IS NULL
ORDER BY 
    created_at DESC, 
    id DESC
LIMIT @page_limit OFFSET @page_offset;

-- name: ListNotificationsByUserRead :many
SELECT id, user_id, category_id, trigger_key, data, metadata, read_at, created_at
FROM notifications
WHERE 
    user_id = @user_id AND 
    read_at IS NOT NULL AND 
    deleted_at IS NULL
ORDER BY 
    created_at DESC, 
    id DESC
LIMIT @page_limit OFFSET @page_offset;

-- name: CountNotificationsUnread :one
SELECT COUNT(*)::BIGINT
FROM notifications
WHERE 
    user_id = @user_id AND 
    read_at IS NULL AND 
    deleted_at IS NULL;

-- ***** ***** *****
-- CREATE DATA
-- ***** ***** *****

-- name: RegisterNotificationUserDevice :exec
INSERT INTO notification_user_devices (user_id, device_token, platform, last_active_at)
VALUES (@user_id, @device_token, @platform, NOW())
ON CONFLICT (device_token) 
DO UPDATE SET 
    user_id = EXCLUDED.user_id,
    platform = EXCLUDED.platform,
    last_active_at = NOW();

-- name: CreateNotification :exec
INSERT INTO notifications (id, user_id, category_id, trigger_key, data, metadata) 
VALUES (@id, @user_id, @category_id, @trigger_key, @data, @metadata);

-- name: CreateNotificationDeliveryLog :one
INSERT INTO notification_delivery_logs (notification_id, channel, status)
VALUES (@notification_id, @channel, @status) RETURNING id;

-- name: UpdateNotificationDeliveryLogStatus :exec
UPDATE notification_delivery_logs
SET
    status = @status,
    provider_response = @provider_response,
    updated_at = NOW(),
    attempt_count = attempt_count + 1,
    next_retry_at = @next_retry_at
WHERE id = @id;

-- name: UpsertNotificationUserSetting :exec
INSERT INTO notification_user_settings (user_id, category_id, channel, is_enabled)
VALUES (@user_id, @category_id, @channel, @is_enabled)
ON CONFLICT (user_id, category_id, channel)
DO UPDATE SET
    is_enabled = EXCLUDED.is_enabled,
    updated_at = NOW();

-- name: MarkNotificationRead :execrows
UPDATE notifications
SET read_at = NOW()
WHERE 
    id = @id AND 
    user_id = @user_id AND 
    read_at IS NULL AND 
    deleted_at IS NULL;

-- name: MarkNotificationsReadAll :execrows
UPDATE notifications
SET read_at = NOW()
WHERE 
    user_id = @user_id AND 
    read_at IS NULL AND 
    deleted_at IS NULL;

-- ***** ***** *****
-- DELETE DATA
-- ***** ***** *****

-- name: RemoveNotificationUserDevice :exec
DELETE FROM notification_user_devices 
WHERE 
    device_token = @device_token;

-- name: SoftDeleteNotification :execrows
UPDATE notifications
SET deleted_at = NOW()
WHERE 
    id = @id AND 
    user_id = @user_id AND 
    deleted_at IS NULL;
