package usecase

import "github.com/shandysiswandi/gobite/internal/notification/entity"

type ListNotificationsInput struct {
	Limit  int32 `validate:"gte=1,lte=100"`
	Offset int32 `validate:"gte=0,lte=100000"`
}

type ListNotificationsOutput struct {
	Notifications []entity.UserNotificationItem
	Limit         int32
	Offset        int32
}

type CountUnreadInput struct{}

type CountUnreadOutput struct {
	Count int64
}

type MarkReadInput struct {
	NotificationID int64 `validate:"gt=0"`
}

type MarkAllReadInput struct{}

type DeleteNotificationInput struct {
	NotificationID int64 `validate:"gt=0"`
}

type RegisterDeviceInput struct {
	DeviceToken string `validate:"required"`
	Platform    string `validate:"required,oneof=android ios web"`
}

type RemoveDeviceInput struct {
	DeviceToken string `validate:"required"`
}
