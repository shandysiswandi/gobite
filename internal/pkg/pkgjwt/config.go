package pkgjwt

import (
	"time"

	"github.com/shandysiswandi/gobite/internal/pkg/pkgclock"
	"github.com/shandysiswandi/gobite/internal/pkg/pkguid"
)

type Config struct {
	Secret   []byte
	Issuer   string
	Audience string
	TTL      time.Duration
	// ---
	Clock pkgclock.Clocker
	UUID  pkguid.StringID
}
