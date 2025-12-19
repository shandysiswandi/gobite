package inbound

import (
	"time"

	"github.com/shandysiswandi/gobite/internal/pkg/valueobject"
)

type NotificationItemResponse struct {
	ID                  int64               `json:"id"`
	TriggerKey          string              `json:"trigger_key"`
	Data                valueobject.JSONMap `json:"data"`
	Metadata            valueobject.JSONMap `json:"metadata"`
	ReadAt              *time.Time          `json:"read_at,omitempty"`
	CreatedAt           time.Time           `json:"created_at"`
	CategoryName        string              `json:"category_name"`
	CategoryDescription string              `json:"category_description"`
}

type ListNotificationsResponse struct {
	Notifications []NotificationItemResponse `json:"notifications"`
	Limit         int32                      `json:"limit"`
	Offset        int32                      `json:"offset"`
}

type CountUnreadResponse struct {
	Count int64 `json:"count"`
}

type RegisterDeviceRequest struct {
	DeviceToken string `json:"device_token"`
	Platform    string `json:"platform"`
}

type RemoveDeviceRequest struct {
	DeviceToken string `json:"device_token"`
}
