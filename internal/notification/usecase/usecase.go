package usecase

import (
	"bytes"
	"context"
	"errors"
	"html/template"
	"log/slog"
	"sync"

	"github.com/shandysiswandi/gobite/internal/notification/entity"
	"github.com/shandysiswandi/gobite/internal/pkg/clock"
	"github.com/shandysiswandi/gobite/internal/pkg/config"
	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
	"github.com/shandysiswandi/gobite/internal/pkg/instrument"
	"github.com/shandysiswandi/gobite/internal/pkg/jwt"
	"github.com/shandysiswandi/gobite/internal/pkg/mail"
	"github.com/shandysiswandi/gobite/internal/pkg/uid"
	"github.com/shandysiswandi/gobite/internal/pkg/validator"
	"go.opentelemetry.io/otel/trace"
)

type repoDB interface {
	RegisterUserDevice(ctx context.Context, userID int64, deviceToken, platform string) error
	RemoveUserDevice(ctx context.Context, deviceToken string) error

	GetTemplateByTriggerChannel(ctx context.Context, tk entity.TriggerKey, ch entity.Channel) (*entity.Template, error)
	CreateNotification(ctx context.Context, data entity.CreateNotification) error
	CreateNotificationWithDeliveryLog(ctx context.Context, n entity.CreateNotification, dl entity.CreateDeliveryLog) (int64, error)
	UpdateDeliveryLogStatus(ctx context.Context, u entity.UpdateDeliveryLog) error

	ListCategories(ctx context.Context) ([]entity.Category, error)
	ListUserSettings(ctx context.Context, userID int64) ([]entity.UserSetting, error)
	UpsertUserSettings(ctx context.Context, userID int64, settings []entity.UserSetting) error
	ListNotifications(ctx context.Context, userID int64, status entity.NotificationStatus, limit, offset int32) ([]entity.NotificationItem, error)
	CountUnreadNotifications(ctx context.Context, userID int64) (int64, error)
	MarkNotificationRead(ctx context.Context, userID, notificationID int64) (bool, error)
	MarkNotificationsReadAll(ctx context.Context, userID int64) (int64, error)
	SoftDeleteNotification(ctx context.Context, userID, notificationID int64) (bool, error)
}

type Usecase struct {
	repoDB    repoDB
	cfg       config.Config
	uid       uid.NumberID
	clock     clock.Clocker
	validator validator.Validator
	jwt       jwt.JWT
	repoMail  repoMail
	ins       instrument.Instrumentation
	streamMu  sync.RWMutex
	streams   map[int64]map[*subscriber]struct{}
}

type Dependency struct {
	RepoDB     repoDB
	Config     config.Config
	UID        uid.NumberID
	Clock      clock.Clocker
	Validator  validator.Validator
	JWT        jwt.JWT
	RepoMail   repoMail
	Instrument instrument.Instrumentation
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
		ins:       dep.Instrument,
		streams:   make(map[int64]map[*subscriber]struct{}),
	}
}

func (s *Usecase) startSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	return s.ins.Tracer("notification.usecase").Start(ctx, name)
}

func (s *Usecase) renderTemplate(name, tpl string, data map[string]any) (string, error) {
	t, err := template.New(name).Option("missingkey=zero").Parse(tpl)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (s *Usecase) baseEmailTemplateData() map[string]any {
	return map[string]any{
		"support_email":   "support@gobite.com",
		"company_name":    "GoBite Inc",
		"company_address": "Kota Jakarta Selatan, Daerah Khusus Ibukota Jakarta 12160",
		"year":            s.clock.Now().Format("2006"),
	}
}

func (s *Usecase) getTemplate(ctx context.Context, tk entity.TriggerKey, ch entity.Channel) *entity.Template {
	tpl, err := s.repoDB.GetTemplateByTriggerChannel(ctx, tk, ch)
	if errors.Is(err, goerror.ErrNotFound) {
		slog.WarnContext(ctx, "notification template not found", "trigger_key", tk, "channel", ch.String())
		return nil
	}
	if err != nil {
		slog.ErrorContext(ctx, "failed to repo get template by trigger channel", "trigger_key", tk, "channel", ch.String(), "error", err)
		return nil
	}

	return tpl
}
