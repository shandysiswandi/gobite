package usecase

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"strings"

	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
	"github.com/shandysiswandi/gobite/internal/pkg/jwt"
	"github.com/shandysiswandi/gobite/internal/pkg/storage"
)

//nolint:gochecknoglobals // global for fast reuse
var avatarContentTypeExt = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

var errAvatarTooLarge = errors.New("avatar exceeds max size")

type ProfileUpdateAvatarInput struct {
	File        io.Reader
	ContentType string
}

func (s *Usecase) ProfileUpdateAvatar(ctx context.Context, in ProfileUpdateAvatarInput) error {
	ctx, span := s.startSpan(ctx, "ProfileUpdateAvatar")
	defer span.End()

	clm := jwt.GetAuth(ctx)
	if clm == nil {
		return goerror.NewBusiness("authentication required", goerror.CodeUnauthorized)
	}

	if in.File == nil {
		return goerror.NewInvalidInput(nil, "avatar", "avatar file is required")
	}

	contentType := strings.ToLower(strings.TrimSpace(in.ContentType))
	ext, ok := avatarContentTypeExt[contentType]
	if !ok {
		return goerror.NewInvalidInput(nil, "avatar", "unsupported avatar content type")
	}

	user, err := s.repoDB.GetUserByEmail(ctx, clm.UserEmail, false)
	if errors.Is(err, goerror.ErrNotFound) {
		slog.WarnContext(ctx, "user account not found", "email", clm.UserEmail)
		return goerror.NewBusiness("authentication required", goerror.CodeUnauthorized)
	}
	if err != nil {
		slog.ErrorContext(ctx, "failed to repo get user by email", "email", clm.UserEmail, "error", err)
		return goerror.NewServer(err)
	}

	if err := s.ensureUserStatusAllowed(ctx, user.ID, user.Status); err != nil {
		return err
	}

	bucket := strings.TrimSpace(s.cfg.GetString("modules.identity.avatar_bucket"))
	baseURL := strings.TrimSpace(s.cfg.GetString("modules.identity.avatar_base_url"))
	key := fmt.Sprintf("%d/%s%s", user.ID, s.uuid.Generate(), ext)
	maxSize := s.cfg.GetInt64("modules.identity.avatar_max_size_bytes")

	reader := &maxBytesReader{
		r:   in.File,
		max: maxSize,
	}
	_, err = s.storage.PutObject(ctx, bucket, key, reader, storage.PutOptions{
		Size:        -1,
		ContentType: contentType,
		Metadata:    map[string]string{"user_id": strconv.FormatInt(user.ID, 10)},
	})
	if err != nil {
		if errors.Is(err, errAvatarTooLarge) {
			return goerror.NewInvalidInput(errAvatarTooLarge)
		}
		slog.ErrorContext(ctx, "failed to upload user avatar", "user_id", user.ID, "error", err)
		return goerror.NewServer(err)
	}

	avatarURL := baseURL + "/" + key
	if err := s.repoDB.UpdateUserAvatar(ctx, user.ID, avatarURL); err != nil {
		slog.ErrorContext(ctx, "failed to update user avatar", "user_id", user.ID, "error", err)
		return goerror.NewServer(err)
	}

	return nil
}

type maxBytesReader struct {
	r     io.Reader
	max   int64
	read  int64
	buf   [1]byte
	ended bool
}

func (m *maxBytesReader) Read(p []byte) (int, error) {
	if m.read >= m.max {
		if m.ended {
			return 0, errAvatarTooLarge
		}

		n, err := m.r.Read(m.buf[:])
		if n > 0 {
			m.ended = true
			return 0, errAvatarTooLarge
		}
		if err == nil {
			m.ended = true
			return 0, errAvatarTooLarge
		}
		return 0, err
	}

	remaining := m.max - m.read
	if int64(len(p)) > remaining {
		p = p[:remaining]
	}

	n, err := m.r.Read(p)
	m.read += int64(n)
	return n, err
}
