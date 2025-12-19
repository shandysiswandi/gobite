package usecase

import (
	"context"
	"time"

	"github.com/shandysiswandi/gobite/internal/notification/entity"
	"github.com/shandysiswandi/gobite/internal/pkg/clock"
	"github.com/shandysiswandi/gobite/internal/pkg/config"
	"github.com/shandysiswandi/gobite/internal/pkg/jwt"
	"github.com/shandysiswandi/gobite/internal/pkg/mail"
	"github.com/shandysiswandi/gobite/internal/pkg/uid"
	"github.com/shandysiswandi/gobite/internal/pkg/validator"
	"github.com/shandysiswandi/gobite/internal/pkg/valueobject"
)

const (
	keyPayloadUserID = "user_id"
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

	NotificationDeliveryLogUpdateStatus(ctx context.Context, id int64, status entity.DeliveryStatus, providerResponse valueobject.JSONMap, nextRetryAt *time.Time) error

	NotificationUserDeviceRegister(ctx context.Context, userID int64, deviceToken, platform string) error
	NotificationUserDeviceRemove(ctx context.Context, deviceToken string) error
}

type Usecase struct {
	repoDB    repoDB
	cfg       config.Config
	uid       uid.NumberID
	clock     clock.Clocker
	validator validator.Validator
	jwt       jwt.JWT
	repoMail  repoMail
}

type Dependency struct {
	RepoDB    repoDB
	Config    config.Config
	UID       uid.NumberID
	Clock     clock.Clocker
	Validator validator.Validator
	JWT       jwt.JWT
	RepoMail  repoMail
}

type repoMail interface {
	Send(ctx context.Context, msg mail.Message) error
}

func NewNotification(dep Dependency) *Usecase {
	return &Usecase{
		repoDB:    dep.RepoDB,
		cfg:       dep.Config,
		uid:       dep.UID,
		clock:     dep.Clock,
		validator: dep.Validator,
		jwt:       dep.JWT,
		repoMail:  dep.RepoMail,
	}
}
