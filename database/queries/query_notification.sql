-- name: NotificationCreate :exec
INSERT INTO notifications (id, user_id, category_id, trigger_key, data, metadata) 
VALUES (@id, @user_id, @category_id, @trigger_key, @data, @metadata);

-- name: NotificationGetUserNotificationPaginate :many
-- Fetches a paginated list of notifications for a user.
-- Joins with categories to get the display name (e.g., "Security", "Social").
SELECT 
    n.id, 
    n.trigger_key, 
    n.data, 
    n.metadata, 
    n.read_at, 
    n.created_at,
    nc.name as category_name,
    nc.description as category_description
FROM notifications n
JOIN notification_categories nc ON n.category_id = nc.id
WHERE 
    n.user_id = @user_id AND 
    n.deleted_at IS NULL
ORDER BY n.created_at DESC
LIMIT $1 OFFSET $2;

-- name: NotificationCountUnread :one
SELECT COUNT(*) 
FROM notifications 
WHERE 
    user_id = @user_id AND 
    read_at IS NULL AND 
    deleted_at IS NULL;

-- name: NotificationMarkRead :exec
UPDATE notifications 
SET read_at = NOW() 
WHERE 
    id = @id AND 
    user_id = @user_id AND
    read_at IS NULL;

-- name: NotificationsMarkAllRead :exec
UPDATE notifications 
SET read_at = NOW() 
WHERE 
    user_id = @user_id AND
    read_at IS NULL;

-- name: NotificationSoftDelete :exec
UPDATE notifications 
SET deleted_at = NOW() 
WHERE id = @id AND user_id = @user_id;

-- name: NotificationTemplateGetByTrigger :one
SELECT * FROM notification_templates
WHERE 
    trigger_key = @trigger_key AND
    channel = @channel;

-- name: NotificationDeliveryLogCreate :one
INSERT INTO notification_delivery_logs (notification_id, channel, status)
VALUES (@notification_id, @channel, @status) RETURNING id;

-- name: NotificationDeliveryLogUpdateStatus :exec
UPDATE notification_delivery_logs
SET
    status = @status,
    provider_response = @provider_response,
    updated_at = NOW(),
    attempt_count = attempt_count + 1,
    next_retry_at = @next_retry_at
WHERE id = @id;

-- name: NotificationDeliveryLogGetPendings :many
-- Fetches batch of logs to process.
-- Criteria: Status is Queued (0) OR (Failed (3) AND time to retry has passed)
-- Limit is essential for worker batching.
SELECT * FROM notification_delivery_logs
WHERE 
    (status = @status_default::SMALLINT) OR 
    (status = @status_failed::SMALLINT AND attempt_count < @max_attempt_count::int AND next_retry_at <= NOW())
ORDER BY created_at ASC
LIMIT $1;

-- name: NotificationUserDeviceRegister :exec
-- Saves or updates a device token. 
-- If the token exists, we update the user_id (in case of account switch).
INSERT INTO notification_user_devices (user_id, device_token, platform, last_active_at)
VALUES (@user_id, @device_token, @platform, NOW())
ON CONFLICT (device_token) 
DO UPDATE SET 
    user_id = EXCLUDED.user_id,
    platform = EXCLUDED.platform,
    last_active_at = NOW();

-- name: NotificationUserDeviceRemove :exec
-- Called when a user logs out or the provider tells us the token is invalid.
DELETE FROM notification_user_devices 
WHERE device_token = @device_token;

-- name: NotificationUserDeviceGetByUser :many
-- Fetches all active devices for a user to broadcast push notifications.
SELECT device_token, platform 
FROM notification_user_devices 
WHERE user_id = @user_id;