package entity

type Channel int16

const (
	ChannelInApp Channel = 0
	ChannelEmail Channel = 1
	ChannelSMS   Channel = 2
	ChannelPush  Channel = 3
)

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
	StatusQueued     DeliveryStatus = 0
	StatusProcessing DeliveryStatus = 1
	StatusSent       DeliveryStatus = 2
	StatusFailed     DeliveryStatus = 3
)

func (s DeliveryStatus) String() string {
	switch s {
	case StatusQueued:
		return "queued"
	case StatusProcessing:
		return "processing"
	case StatusSent:
		return "sent"
	case StatusFailed:
		return "failed"
	default:
		return "unknown"
	}
}
