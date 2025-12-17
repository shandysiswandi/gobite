package notification

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shandysiswandi/gobite/internal/notification/inbound"
	"github.com/shandysiswandi/gobite/internal/notification/outbound"
	"github.com/shandysiswandi/gobite/internal/notification/usecase"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgclock"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgconfig"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgjwt"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgmail"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgmessaging"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgrouter"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgroutine"
	"github.com/shandysiswandi/gobite/internal/pkg/pkguid"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgvalidator"
)

type Dependency struct {
	Ctx       context.Context
	DBConn    *pgxpool.Pool
	Messaging pkgmessaging.Messaging
	Config    pkgconfig.Config
	UID       pkguid.NumberID
	Clock     pkgclock.Clocker
	Goroutine *pkgroutine.Manager
	Validator pkgvalidator.Validator
	Router    *pkgrouter.Router
	Mail      pkgmail.Mail

	JWTTempToken pkgjwt.JWT[map[string]any]
}

func New(dep Dependency) error {
	dbNotif := outbound.NewSQL(dep.DBConn)
	repoMail := outbound.NewMail(dep.Mail)

	uc := usecase.NewNotification(usecase.Dependency{
		RepoDB:       dbNotif,
		Config:       dep.Config,
		UID:          dep.UID,
		Clock:        dep.Clock,
		Validator:    dep.Validator,
		JWTTempToken: dep.JWTTempToken,
		RepoMail:     repoMail,
	})

	inbound.RegisterHTTPEndpoint(dep.Router, uc)
	if dep.Ctx != nil && dep.Goroutine != nil && dep.Messaging != nil {
		inbound.RegisterMQConsumer(dep.Ctx, dep.Goroutine, dep.Messaging, uc)
	}

	return nil
}
