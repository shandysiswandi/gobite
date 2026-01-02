package uid

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"strings"
	"sync/atomic"
	"time"
)

// ErrStableNodeIdentityUnavailable indicates no stable node identity is available.
var ErrStableNodeIdentityUnavailable = errors.New("uid: cannot determine stable node identity (machine-id/hostname unavailable)")

// ObjectIDGenerator generates 32-byte distributed-safe IDs (hex output).
type ObjectIDGenerator struct {
	nodeID  [6]byte
	pid     uint16
	counter uint32
}

// NewObjectIDGenerator creates a generator with stable node identity.
func NewObjectIDGenerator() (*ObjectIDGenerator, error) {
	g := &ObjectIDGenerator{}
	g.pid = uint16(os.Getpid())

	// stable node identity source: /etc/machine-id OR hostname
	src, err := g.machineIDOrHostnameStrict()
	if err != nil {
		return nil, err
	}

	sum := sha256.Sum256([]byte(src))
	copy(g.nodeID[:], sum[:6])

	// Seed counter from crypto/rand
	var b [4]byte
	if _, err := rand.Read(b[:]); err != nil {
		return nil, err
	}
	g.counter = uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])

	return g, nil
}

// machineIDOrHostnameStrict returns a stable identity string or an error.
func (g *ObjectIDGenerator) machineIDOrHostnameStrict() (string, error) {
	// Try /etc/machine-id (Linux)
	if b, err := os.ReadFile("/etc/machine-id"); err == nil {
		s := strings.TrimSpace(string(b))
		if s != "" {
			return s, nil
		}
	}

	// Fallback hostname
	if h, err := os.Hostname(); err == nil {
		h = strings.TrimSpace(h)
		if h != "" {
			return h, nil
		}
	}

	return "", ErrStableNodeIdentityUnavailable
}

// Generate returns a 64-char hex string representing 32 bytes (URL-safe).
func (g *ObjectIDGenerator) Generate() string {
	var raw [32]byte

	// 6-byte timestamp (ms, big-endian)
	ts := uint64(time.Now().UnixMilli())
	raw[0] = byte(ts >> 40)
	raw[1] = byte(ts >> 32)
	raw[2] = byte(ts >> 24)
	raw[3] = byte(ts >> 16)
	raw[4] = byte(ts >> 8)
	raw[5] = byte(ts)

	// 6-byte node id (stable)
	copy(raw[6:12], g.nodeID[:])

	// 2-byte pid (big-endian)
	raw[12] = byte(g.pid >> 8)
	raw[13] = byte(g.pid)

	// 4-byte counter
	c := atomic.AddUint32(&g.counter, 1)
	raw[14] = byte(c >> 24)
	raw[15] = byte(c >> 16)
	raw[16] = byte(c >> 8)
	raw[17] = byte(c)

	// 14 random bytes (best effort). If it fails, deterministic fallback.
	if _, err := rand.Read(raw[18:]); err != nil {
		var seed [18]byte
		copy(seed[0:6], raw[0:6])
		copy(seed[6:12], raw[6:12])
		copy(seed[12:14], raw[12:14])
		copy(seed[14:18], raw[14:18])

		sum := sha256.Sum256(seed[:])
		copy(raw[18:], sum[:14])
	}

	var hexBuf [64]byte
	hex.Encode(hexBuf[:], raw[:])
	return string(hexBuf[:])
}
