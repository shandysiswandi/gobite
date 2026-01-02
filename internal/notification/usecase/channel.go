package usecase

import (
	"strings"

	"github.com/shandysiswandi/gobite/internal/notification/entity"
)

func channelFromString(raw string) (entity.Channel, bool) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "in_app":
		return entity.ChannelInApp, true
	case "email":
		return entity.ChannelEmail, true
	case "sms":
		return entity.ChannelSMS, true
	case "push":
		return entity.ChannelPush, true
	default:
		return entity.ChannelUnknown, false
	}
}

func channelToString(ch entity.Channel) string {
	switch ch {
	case entity.ChannelInApp:
		return "in_app"
	case entity.ChannelEmail:
		return "email"
	case entity.ChannelSMS:
		return "sms"
	case entity.ChannelPush:
		return "push"
	case entity.ChannelUnknown:
		return "in_app"
	default:
		return "unknown"
	}
}
