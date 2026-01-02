package inbound

import (
	"context"

	"github.com/shandysiswandi/gobite/internal/notification/entity"
	"github.com/shandysiswandi/gobite/internal/notification/usecase"
)

type ucConsumer interface {
	ConsumeUserRegistration(ctx context.Context, in usecase.ConsumeUserRegistrationInput) error
	ConsumeUserForgotPassword(ctx context.Context, msg usecase.ConsumeUserForgotPasswordInput) error
}

type uc interface {
	ucConsumer

	DeviceRegister(ctx context.Context, in usecase.DeviceRegisterInput) error
	DeviceRemove(ctx context.Context, in usecase.DeviceRemoveInput) error
	ListCategories(ctx context.Context) ([]entity.Category, error)
	ListSettings(ctx context.Context) ([]entity.UserSetting, error)
	UpdateSettings(ctx context.Context, in usecase.UpdateSettingsInput) error
	ListNotifications(ctx context.Context, in usecase.ListNotificationsInput) ([]entity.NotificationItem, error)
	MarkNotificationRead(ctx context.Context, in usecase.MarkNotificationReadInput) error
	MarkAllNotificationsRead(ctx context.Context) error
	DeleteNotification(ctx context.Context, in usecase.DeleteNotificationInput) error
}
