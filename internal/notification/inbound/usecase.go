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

type ucStream interface {
	StreamNotifications(ctx context.Context, userID int64) <-chan usecase.StreamEvent
}

type uc interface {
	ucConsumer
	ucStream

	DeviceRegister(ctx context.Context, in usecase.DeviceRegisterInput) error
	DeviceRemove(ctx context.Context, in usecase.DeviceRemoveInput) error
	ListCategories(ctx context.Context) ([]entity.Category, error)
	ListSettings(ctx context.Context) ([]entity.UserSetting, error)
	UpdateSettings(ctx context.Context, in usecase.UpdateSettingsInput) error
	ListInbox(ctx context.Context, in usecase.ListInboxInput) ([]entity.NotificationItem, error)
	MarkInboxRead(ctx context.Context, in usecase.MarkInboxReadInput) error
	MarkAllInboxRead(ctx context.Context) error
	DeleteInbox(ctx context.Context, in usecase.DeleteInboxInput) error
}
