package notifications

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// NotificationRecord is the minimal data the framework needs from a stored notification.
type NotificationRecord struct {
	ID          string
	UserID      string
	Title       string
	Body        string
	Level       string
	Icon        string
	ActionURL   string
	ActionLabel string
	Read        bool
	CreatedAt   time.Time
}

// NotificationRepository is the interface to persist notifications.
// Implement it in your project using your own ORM or database layer.
type NotificationRepository interface {
	Create(ctx context.Context, n NotificationRecord) error
	GetAll(ctx context.Context, userID string, limit int) ([]NotificationRecord, error)
	GetUnread(ctx context.Context, userID string, limit int) ([]NotificationRecord, error)
	MarkRead(ctx context.Context, userID, notifID string) error
	MarkAllRead(ctx context.Context, userID string) error
	UnreadCount(ctx context.Context, userID string) (int, error)
}

// DatabaseStore is a persistent notification store backed by a NotificationRepository.
// It keeps the in-memory SSE subscriber map for live streaming while
// persisting every notification to the database.
type DatabaseStore struct {
	repo        NotificationRepository
	mu          sync.RWMutex
	subscribers map[string][]chan *Notification
	maxPerUser  int
}

// NewDatabaseStore creates a DatabaseStore using the provided NotificationRepository.
func NewDatabaseStore(repo NotificationRepository, maxPerUser int) *DatabaseStore {
	if maxPerUser <= 0 {
		maxPerUser = 100
	}
	return &DatabaseStore{
		repo:        repo,
		subscribers: make(map[string][]chan *Notification),
		maxPerUser:  maxPerUser,
	}
}

// Send persists a notification and broadcasts it to SSE subscribers.
func (s *DatabaseStore) Send(userID string, n *Notification) {
	if n.ID == "" {
		n.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	}
	if n.CreatedAt.IsZero() {
		n.CreatedAt = time.Now()
	}
	if n.Level == "" {
		n.Level = LevelInfo
	}
	n.UserID = userID

	ctx := context.Background()
	_ = s.repo.Create(ctx, NotificationRecord{
		ID:          n.ID,
		UserID:      userID,
		Title:       n.Title,
		Body:        n.Body,
		Level:       string(n.Level),
		Icon:        n.Icon,
		ActionURL:   n.ActionURL,
		ActionLabel: n.ActionLabel,
		Read:        false,
		CreatedAt:   n.CreatedAt,
	})

	s.mu.RLock()
	subs := s.subscribers[userID]
	s.mu.RUnlock()

	for _, ch := range subs {
		select {
		case ch <- n:
		default:
		}
	}
}

// GetAll returns all notifications for a user (newest first).
func (s *DatabaseStore) GetAll(userID string) []*Notification {
	ctx := context.Background()
	rows, err := s.repo.GetAll(ctx, userID, s.maxPerUser)
	if err != nil {
		return nil
	}
	return recordsToNotifications(rows)
}

// GetUnread returns unread notifications for a user (newest first).
func (s *DatabaseStore) GetUnread(userID string) []*Notification {
	ctx := context.Background()
	rows, err := s.repo.GetUnread(ctx, userID, s.maxPerUser)
	if err != nil {
		return nil
	}
	return recordsToNotifications(rows)
}

// MarkRead marks a single notification as read.
func (s *DatabaseStore) MarkRead(userID, notifID string) {
	ctx := context.Background()
	_ = s.repo.MarkRead(ctx, userID, notifID)
}

// MarkAllRead marks all notifications as read for a user.
func (s *DatabaseStore) MarkAllRead(userID string) {
	ctx := context.Background()
	_ = s.repo.MarkAllRead(ctx, userID)
}

// UnreadCount returns the number of unread notifications for a user.
func (s *DatabaseStore) UnreadCount(userID string) int {
	ctx := context.Background()
	count, err := s.repo.UnreadCount(ctx, userID)
	if err != nil {
		return 0
	}
	return count
}

// Subscribe returns a channel that receives new notifications for a user.
func (s *DatabaseStore) Subscribe(ctx context.Context, userID string) <-chan *Notification {
	ch := make(chan *Notification, 16)

	s.mu.Lock()
	s.subscribers[userID] = append(s.subscribers[userID], ch)
	s.mu.Unlock()

	go func() {
		<-ctx.Done()
		s.mu.Lock()
		defer s.mu.Unlock()
		subs := s.subscribers[userID]
		for i, sub := range subs {
			if sub == ch {
				s.subscribers[userID] = append(subs[:i], subs[i+1:]...)
				break
			}
		}
		close(ch)
	}()

	return ch
}

func recordsToNotifications(rows []NotificationRecord) []*Notification {
	out := make([]*Notification, len(rows))
	for i, r := range rows {
		out[i] = &Notification{
			ID:          r.ID,
			UserID:      r.UserID,
			Title:       r.Title,
			Body:        r.Body,
			Level:       Level(r.Level),
			Icon:        r.Icon,
			ActionURL:   r.ActionURL,
			ActionLabel: r.ActionLabel,
			Read:        r.Read,
			CreatedAt:   r.CreatedAt,
		}
	}
	return out
}
