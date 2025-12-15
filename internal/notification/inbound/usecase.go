package inbound

import (
	"context"

	"github.com/shandysiswandi/gobite/internal/notification/entity"
	"github.com/shandysiswandi/gobite/internal/notification/usecase"
)

type uc interface {
	ListNotifications(ctx context.Context, in usecase.ListNotificationsInput) (*usecase.ListNotificationsOutput, error)
	CountUnread(ctx context.Context, in usecase.CountUnreadInput) (*usecase.CountUnreadOutput, error)
	MarkRead(ctx context.Context, in usecase.MarkReadInput) error
	MarkAllRead(ctx context.Context, in usecase.MarkAllReadInput) error
	DeleteNotification(ctx context.Context, in usecase.DeleteNotificationInput) error
	RegisterDevice(ctx context.Context, in usecase.RegisterDeviceInput) error
	RemoveDevice(ctx context.Context, in usecase.RemoveDeviceInput) error

	UserRegistrationNotification(ctx context.Context, msg entity.UserRegistrationMessage) error
}
