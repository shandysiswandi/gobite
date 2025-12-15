package inbound

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/shandysiswandi/gobite/internal/notification/entity"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgmessaging"
)

type MQHandler struct {
	uc uc
}

func (h *MQHandler) UserRegistrationNotification(ctx context.Context, msg pkgmessaging.Message) error {
	var payload entity.UserRegistrationMessage
	if err := json.Unmarshal(msg.Body(), &payload); err != nil {
		slog.Error("failed to parse message body", "error", err, "body", msg.Body())
		return nil
	}

	return h.uc.UserRegistrationNotification(ctx, payload)
}
