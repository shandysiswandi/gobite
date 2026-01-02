package entity

import (
	"time"

	"github.com/shandysiswandi/gobite/internal/pkg/valueobject"
)

type CreateNotification struct {
	ID         int64
	UserID     int64
	CategoryID int64
	TriggerKey TriggerKey
	Data       valueobject.JSONMap
	Metadata   valueobject.JSONMap
}

type CreateDeliveryLog struct {
	NotificationID int64
	Channel        Channel
	Status         DeliveryStatus
}

type UpdateDeliveryLog struct {
	ID               int64
	Status           DeliveryStatus
	ProviderResponse valueobject.JSONMap
	NextRetryAt      *time.Time
}

type Template struct {
	ID         int64
	TriggerKey TriggerKey
	CategoryID int64
	Channel    Channel
	Subject    string
	Body       string
}

type Category struct {
	ID          int64
	Name        string
	Description string
	IsMandatory bool
}

type UserSetting struct {
	CategoryID int64
	Channel    Channel
	IsEnabled  bool
}

type NotificationItem struct {
	ID         int64
	CategoryID int64
	TriggerKey TriggerKey
	Data       valueobject.JSONMap
	Metadata   valueobject.JSONMap
	ReadAt     *time.Time
	CreatedAt  time.Time
}
