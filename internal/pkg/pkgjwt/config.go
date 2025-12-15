package pkgjwt

import (
	"time"

	"github.com/shandysiswandi/gobite/internal/pkg/pkgclock"
	"github.com/shandysiswandi/gobite/internal/pkg/pkguid"
)

// Config defines the inputs for building a JWT implementation.
//
// Clock and UUID are required so token generation is testable and produces a
// unique JTI (token ID).
type Config struct {
	Secret   []byte
	Issuer   string
	Audience string
	TTL      time.Duration
	// ---
	Clock pkgclock.Clocker
	UUID  pkguid.StringID
}
