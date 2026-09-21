package session

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"boilerplate/internal/domain"

	"gorm.io/gorm"
)

// Manager persists login sessions server-side. Only the random 256-bit session
// ID ever leaves the server (inside an HttpOnly cookie).
type Manager struct {
	db  *gorm.DB
	ttl time.Duration
}

func NewManager(db *gorm.DB, ttl time.Duration) *Manager {
	return &Manager{db: db, ttl: ttl}
}

// TTL exposes the configured session lifetime (for cookie Max-Age).
func (m *Manager) TTL() time.Duration { return m.ttl }

func NewID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (m *Manager) Create(userID uint, userAgent, ip string) (*domain.Session, error) {
	// Opportunistic garbage collection of expired rows.
	m.db.Where("expires_at < ?", time.Now()).Delete(&domain.Session{})

	id, err := NewID()
	if err != nil {
		return nil, err
	}
	now := time.Now()
	s := &domain.Session{
		ID:         id,
		UserID:     userID,
		UserAgent:  userAgent,
		IP:         ip,
		CreatedAt:  now,
		LastSeenAt: now,
		ExpiresAt:  now.Add(m.ttl),
	}
	if err := m.db.Create(s).Error; err != nil {
		return nil, err
	}
	return s, nil
}

// Get returns the session with its user preloaded. Expired sessions and
// sessions belonging to deactivated users are destroyed on access.
func (m *Manager) Get(id string) (*domain.Session, error) {
	if id == "" {
		return nil, gorm.ErrRecordNotFound
	}
	var s domain.Session
	if err := m.db.Preload("User").First(&s, "id = ?", id).Error; err != nil {
		return nil, err
	}
	if s.Expired() || !s.User.IsActive {
		m.Destroy(s.ID)
		return nil, gorm.ErrRecordNotFound
	}
	return &s, nil
}

func (m *Manager) Touch(s *domain.Session) {
	m.db.Model(&domain.Session{}).Where("id = ?", s.ID).Update("last_seen_at", time.Now())
}

func (m *Manager) Destroy(id string) {
	m.db.Where("id = ?", id).Delete(&domain.Session{})
}

// DestroyForUser invalidates every session of a user, optionally keeping one
// (used by "sign out other sessions" and password resets).
func (m *Manager) DestroyForUser(userID uint, exceptSessionID string) {
	q := m.db.Where("user_id = ?", userID)
	if exceptSessionID != "" {
		q = q.Where("id <> ?", exceptSessionID)
	}
	q.Delete(&domain.Session{})
}

func (m *Manager) CountActive() int64 {
	var c int64
	m.db.Model(&domain.Session{}).Where("expires_at > ?", time.Now()).Count(&c)
	return c
}

func (m *Manager) CountForUser(userID uint) int64 {
	var c int64
	m.db.Model(&domain.Session{}).Where("user_id = ?", userID).Count(&c)
	return c
}

func (m *Manager) ListForUser(userID uint, limit int) []domain.Session {
	var sessions []domain.Session
	m.db.Where("user_id = ?", userID).
		Order("last_seen_at DESC").
		Limit(limit).
		Find(&sessions)
	return sessions
}
