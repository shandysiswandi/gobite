package db

import "context"

func (s *DB) RemoveUserDevice(ctx context.Context, deviceToken string) (err error) {
	ctx, span := s.startSpan(ctx, "RemoveUserDevice")
	defer func() { s.endSpan(span, err) }()

	err = s.query.RemoveNotificationUserDevice(ctx, deviceToken)
	return s.mapError(err)
}
