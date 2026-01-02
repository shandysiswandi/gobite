package inbound

import (
	"time"

	"github.com/shandysiswandi/gobite/internal/pkg/valueobject"
)

type RegisterDeviceRequest struct {
	DeviceToken string `json:"device_token"`
	Platform    string `json:"platform"`
}

type RemoveDeviceRequest struct {
	DeviceToken string `json:"device_token"`
}

type NotificationCategoryResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsMandatory bool   `json:"is_mandatory"`
}

type NotificationCategoriesResponse struct {
	Categories []NotificationCategoryResponse `json:"categories"`
}

type NotificationSettingResponse struct {
	CategoryID int64  `json:"category_id"`
	Channel    string `json:"channel"`
	IsEnabled  bool   `json:"is_enabled"`
}

type NotificationSettingsResponse struct {
	Settings []NotificationSettingResponse `json:"settings"`
}

type NotificationSettingRequest struct {
	CategoryID int64  `json:"category_id"`
	Channel    string `json:"channel"`
	IsEnabled  bool   `json:"is_enabled"`
}

type NotificationSettingsUpdateRequest struct {
	Settings []NotificationSettingRequest `json:"settings"`
}

type NotificationResponse struct {
	ID         int64               `json:"id"`
	CategoryID int64               `json:"category_id"`
	TriggerKey string              `json:"trigger_key"`
	Data       valueobject.JSONMap `json:"data" swaggertype:"object"`
	Metadata   valueobject.JSONMap `json:"metadata" swaggertype:"object"`
	ReadAt     *time.Time          `json:"read_at,omitempty"`
	CreatedAt  time.Time           `json:"created_at"`
}

type NotificationsResponse struct {
	Notifications []NotificationResponse `json:"notifications"`
}
