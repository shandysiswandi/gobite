package entity

type NotificationStatus string

const (
	NotificationStatusAll    NotificationStatus = "all"
	NotificationStatusUnread NotificationStatus = "unread"
	NotificationStatusRead   NotificationStatus = "read"
)
