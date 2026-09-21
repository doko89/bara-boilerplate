package domain

import "time"

// Session is a server-side login session. The browser only ever holds the
// random session ID inside an HttpOnly cookie.
type Session struct {
	ID         string    `gorm:"size:64;primaryKey"`
	UserID     uint      `gorm:"index;not null"`
	User       User      `gorm:"constraint:OnDelete:CASCADE"`
	UserAgent  string    `gorm:"size:255"`
	IP         string    `gorm:"size:64"`
	CreatedAt  time.Time
	LastSeenAt time.Time
	ExpiresAt  time.Time `gorm:"index"`
}

func (s *Session) Expired() bool { return time.Now().After(s.ExpiresAt) }
