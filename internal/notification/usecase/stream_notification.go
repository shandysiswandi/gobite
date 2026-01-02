package usecase

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/shandysiswandi/gobite/internal/notification/entity"
	"github.com/shandysiswandi/gobite/internal/pkg/valueobject"
)

// StreamEvent represents a notification update sent over SSE.
type StreamEvent struct {
	ID         int64               `json:"id"`
	UserID     int64               `json:"user_id"`
	CategoryID int64               `json:"category_id"`
	TriggerKey entity.TriggerKey   `json:"trigger_key"`
	Data       valueobject.JSONMap `json:"data"`
	Metadata   valueobject.JSONMap `json:"metadata"`
	CreatedAt  time.Time           `json:"created_at"`
}

type subscriber struct {
	ch     chan StreamEvent
	closed atomic.Bool
}

// StreamNotifications registers a stream for a user and closes it when ctx is done.
func (s *Usecase) StreamNotifications(ctx context.Context, userID int64) <-chan StreamEvent {
	sub := &subscriber{ch: make(chan StreamEvent, 10)}

	s.streamMu.Lock()
	if s.streams[userID] == nil {
		s.streams[userID] = make(map[*subscriber]struct{})
	}
	s.streams[userID][sub] = struct{}{}
	s.streamMu.Unlock()

	go func() {
		<-ctx.Done()
		s.streamMu.Lock()
		if subs := s.streams[userID]; subs != nil {
			delete(subs, sub)
			if len(subs) == 0 {
				delete(s.streams, userID)
			}
		}
		s.streamMu.Unlock()
		close(sub.ch)
	}()

	return sub.ch
}

func (s *Usecase) publishNotification(evt StreamEvent) {
	s.streamMu.RLock()
	subs := s.streams[evt.UserID]
	s.streamMu.RUnlock()

	for sub := range subs {
		if sub.closed.Load() {
			continue
		}

		select {
		case sub.ch <- evt:
		default:
		}
	}
}

func (s *Usecase) buildStreamEvent(n entity.CreateNotification) StreamEvent {
	return StreamEvent{
		ID:         n.ID,
		UserID:     n.UserID,
		CategoryID: n.CategoryID,
		TriggerKey: n.TriggerKey,
		Data:       n.Data,
		Metadata:   n.Metadata,
		CreatedAt:  s.clock.Now(),
	}
}
