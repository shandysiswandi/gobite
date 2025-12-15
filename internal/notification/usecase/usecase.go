package usecase

import (
	"context"
	"time"

	"github.com/shandysiswandi/gobite/internal/notification/entity"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgclock"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgconfig"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgjwt"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgmail"
	"github.com/shandysiswandi/gobite/internal/pkg/pkguid"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgvalidator"
)

type repoDB interface {
	NotificationCreate(ctx context.Context, n entity.NotificationCreate) error
	NotificationCreateWithDeliveryLog(ctx context.Context, n entity.NotificationCreate, log entity.DeliveryLogCreate) (logID int64, err error)
	NotificationGetUserNotificationPaginate(ctx context.Context, userID int64, limit, offset int32) ([]entity.UserNotificationItem, error)
	NotificationCountUnread(ctx context.Context, userID int64) (int64, error)
	NotificationMarkRead(ctx context.Context, userID, notificationID int64) error
	NotificationsMarkAllRead(ctx context.Context, userID int64) error
	NotificationSoftDelete(ctx context.Context, userID, notificationID int64) error

	NotificationTemplateGetByTrigger(ctx context.Context, triggerKey string, channel entity.Channel) (*entity.Template, error)

	NotificationDeliveryLogUpdateStatus(ctx context.Context, id int64, status entity.DeliveryStatus, providerResponse entity.JSONMap, nextRetryAt *time.Time) error

	NotificationUserDeviceRegister(ctx context.Context, userID int64, deviceToken, platform string) error
	NotificationUserDeviceRemove(ctx context.Context, deviceToken string) error
}

type Usecase struct {
	repoDB repoDB

	cfg       pkgconfig.Config
	uid       pkguid.NumberID
	clock     pkgclock.Clocker
	validator pkgvalidator.Validator

	jwtTempToken pkgjwt.JWT[map[string]any]
	repoMail     repoMail
}

type Dependency struct {
	RepoDB repoDB

	Config    pkgconfig.Config
	UID       pkguid.NumberID
	Clock     pkgclock.Clocker
	Validator pkgvalidator.Validator

	JWTTempToken pkgjwt.JWT[map[string]any]
	RepoMail     repoMail
}

type repoMail interface {
	Send(ctx context.Context, msg pkgmail.Message) error
}

func NewNotification(dep Dependency) *Usecase {
	return &Usecase{
		repoDB:       dep.RepoDB,
		cfg:          dep.Config,
		uid:          dep.UID,
		clock:        dep.Clock,
		validator:    dep.Validator,
		jwtTempToken: dep.JWTTempToken,
		repoMail:     dep.RepoMail,
	}
}
