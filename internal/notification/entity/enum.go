package entity

import (
	"strings"
)

type Channel int16

const (
	ChannelUnknown Channel = 0
	ChannelInApp   Channel = 1
	ChannelEmail   Channel = 2
	ChannelSMS     Channel = 3
	ChannelPush    Channel = 4
)

func ChannelFromString(raw string) Channel {
	switch strings.TrimSpace(raw) {
	case "in_app":
		return ChannelInApp
	case "email":
		return ChannelEmail
	case "sms":
		return ChannelSMS
	case "push":
		return ChannelPush
	default:
		return ChannelUnknown
	}
}

func (c Channel) String() string {
	switch c {
	case ChannelInApp:
		return "in_app"
	case ChannelEmail:
		return "email"
	case ChannelSMS:
		return "sms"
	case ChannelPush:
		return "push"
	default:
		return "unknown"
	}
}

type DeliveryStatus int16

const (
	DeliveryStatusUnknown    DeliveryStatus = 0
	DeliveryStatusQueued     DeliveryStatus = 1
	DeliveryStatusProcessing DeliveryStatus = 2
	DeliveryStatusSent       DeliveryStatus = 3
	DeliveryStatusFailed     DeliveryStatus = 4
)

func (s DeliveryStatus) String() string {
	switch s {
	case DeliveryStatusQueued:
		return "queued"
	case DeliveryStatusProcessing:
		return "processing"
	case DeliveryStatusSent:
		return "sent"
	case DeliveryStatusFailed:
		return "failed"
	default:
		return "unknown"
	}
}

type TriggerKey string

const (
	TriggerKeyEmailVerify   TriggerKey = "email_verify"
	TriggerKeyPasswordReset TriggerKey = "password_reset"
	TriggerKeyUserWelcome   TriggerKey = "user_welcome"
)

func (tk TriggerKey) String() string {
	return string(tk)
}
