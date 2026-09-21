package dashboard

import (
	"time"

	"boilerplate/internal/domain"
	"boilerplate/internal/session"
)

type Stats struct {
	SignInCount int64
	HasSignIn   bool
	LastSignIn  time.Time
	MemberSince time.Time
	Recent      []domain.Session
}

type Service struct {
	sessions *session.Manager
}

func NewService(sessions *session.Manager) *Service {
	return &Service{sessions: sessions}
}

func (s *Service) Overview(u *domain.User) Stats {
	recent := s.sessions.ListForUser(u.ID, 6)
	stats := Stats{
		SignInCount: s.sessions.CountForUser(u.ID),
		MemberSince: u.CreatedAt,
		Recent:      recent,
	}
	if len(recent) > 0 {
		stats.HasSignIn = true
		stats.LastSignIn = recent[0].CreatedAt
	}
	return stats
}
