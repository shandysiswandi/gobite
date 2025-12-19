package entity

import (
	"time"

	"github.com/shandysiswandi/gobite/internal/pkg/valueobject"
)

type NotificationCreate struct {
	ID         int64
	UserID     int64
	CategoryID int64
	TriggerKey string
	Data       valueobject.JSONMap
	Metadata   valueobject.JSONMap
}

type UserNotificationItem struct {
	ID                  int64
	TriggerKey          string
	Data                valueobject.JSONMap
	Metadata            valueobject.JSONMap
	ReadAt              *time.Time
	CreatedAt           time.Time
	CategoryName        string
	CategoryDescription string
}

type Template struct {
	ID         int64
	TriggerKey string
	CategoryID int64
	Channel    Channel
	Subject    string
	Body       string
}

type DeliveryLogCreate struct {
	NotificationID int64
	Channel        Channel
	Status         DeliveryStatus
}

const (
	TriggerEmailVerify = "email_verify"
	TriggerUserWelcome = "user_welcome"
)
