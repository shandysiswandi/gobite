package notification

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shandysiswandi/gobite/internal/notification/inbound"
	"github.com/shandysiswandi/gobite/internal/notification/outbound"
	"github.com/shandysiswandi/gobite/internal/notification/usecase"
	"github.com/shandysiswandi/gobite/internal/pkg/clock"
	"github.com/shandysiswandi/gobite/internal/pkg/config"
	"github.com/shandysiswandi/gobite/internal/pkg/goroutine"
	"github.com/shandysiswandi/gobite/internal/pkg/jwt"
	"github.com/shandysiswandi/gobite/internal/pkg/mail"
	"github.com/shandysiswandi/gobite/internal/pkg/messaging"
	"github.com/shandysiswandi/gobite/internal/pkg/router"
	"github.com/shandysiswandi/gobite/internal/pkg/uid"
	"github.com/shandysiswandi/gobite/internal/pkg/validator"
)

type Dependency struct {
	Ctx       context.Context
	DBConn    *pgxpool.Pool
	Messaging messaging.Messaging
	Config    config.Config
	UID       uid.NumberID
	Clock     clock.Clocker
	Goroutine *goroutine.Manager
	Validator validator.Validator
	Router    *router.Router
	Mail      mail.Mail
	JWT       jwt.JWT
}

func New(dep Dependency) error {
	dbNotif := outbound.NewSQL(dep.DBConn)
	repoMail := outbound.NewMail(dep.Mail)

	uc := usecase.NewNotification(usecase.Dependency{
		RepoDB:    dbNotif,
		Config:    dep.Config,
		UID:       dep.UID,
		Clock:     dep.Clock,
		Validator: dep.Validator,
		JWT:       dep.JWT,
		RepoMail:  repoMail,
	})

	inbound.RegisterHTTPEndpoint(dep.Router, uc)
	if dep.Ctx != nil && dep.Goroutine != nil && dep.Messaging != nil && dep.Config != nil {
		inbound.RegisterMQConsumer(dep.Ctx, dep.Config, dep.Goroutine, dep.Messaging, uc)
	}

	return nil
}
