package admin

import (
	"errors"
	"time"

	"boilerplate/internal/domain"
	"boilerplate/internal/session"
)

var (
	ErrNotFound   = errors.New("user not found")
	ErrSelfAction = errors.New("cannot target your own account")
	ErrLastAdmin  = errors.New("the workspace needs at least one admin")
	ErrBadRole    = errors.New("invalid role")
)

// UserRepository is implemented by repository.go (GORM).
type UserRepository interface {
	ListUsers() ([]domain.User, error)
	RecentUsers(limit int) ([]domain.User, error)
	GetUser(id uint) (*domain.User, error)
	CountUsers() (int64, error)
	CountAdmins() (int64, error)
	CountActiveUsers() (int64, error)
	CountCreatedSince(t time.Time) (int64, error)
	SetRole(id uint, role domain.Role) error
	SetStatus(id uint, active bool) error
	Delete(id uint) error
}

type Service struct {
	repo     UserRepository
	sessions *session.Manager
}

func NewService(repo UserRepository, sessions *session.Manager) *Service {
	return &Service{repo: repo, sessions: sessions}
}

type Stats struct {
	TotalUsers     int64
	ActiveUsers    int64
	Admins         int64
	NewThisWeek    int64
	ActiveSessions int64
	RecentUsers    []domain.User
}

func (s *Service) Overview() Stats {
	weekAgo := time.Now().AddDate(0, 0, -7)
	recent, _ := s.repo.RecentUsers(5)
	stats := Stats{
		RecentUsers:    recent,
		ActiveSessions: s.sessions.CountActive(),
	}
	stats.TotalUsers, _ = s.repo.CountUsers()
	stats.ActiveUsers, _ = s.repo.CountActiveUsers()
	stats.Admins, _ = s.repo.CountAdmins()
	stats.NewThisWeek, _ = s.repo.CountCreatedSince(weekAgo)
	return stats
}

func (s *Service) Users() ([]domain.User, error) {
	return s.repo.ListUsers()
}

func (s *Service) SetRole(actor *domain.User, targetID uint, role domain.Role) error {
	if !role.Valid() {
		return ErrBadRole
	}
	if actor.ID == targetID {
		return ErrSelfAction
	}
	target, err := s.repo.GetUser(targetID)
	if err != nil {
		return ErrNotFound
	}
	if target.IsAdmin() && role == domain.RoleUser {
		admins, err := s.repo.CountAdmins()
		if err != nil {
			return err
		}
		if admins <= 1 {
			return ErrLastAdmin
		}
	}
	return s.repo.SetRole(targetID, role)
}

func (s *Service) SetStatus(actor *domain.User, targetID uint, active bool) error {
	if actor.ID == targetID {
		return ErrSelfAction
	}
	target, err := s.repo.GetUser(targetID)
	if err != nil {
		return ErrNotFound
	}
	if target.IsAdmin() && !active {
		admins, err := s.repo.CountAdmins()
		if err != nil {
			return err
		}
		if admins <= 1 {
			return ErrLastAdmin
		}
	}
	if err := s.repo.SetStatus(targetID, active); err != nil {
		return err
	}
	if !active {
		// Deactivating a user kicks them out immediately.
		s.sessions.DestroyForUser(targetID, "")
	}
	return nil
}

func (s *Service) Delete(actor *domain.User, targetID uint) error {
	if actor.ID == targetID {
		return ErrSelfAction
	}
	target, err := s.repo.GetUser(targetID)
	if err != nil {
		return ErrNotFound
	}
	if target.IsAdmin() {
		admins, err := s.repo.CountAdmins()
		if err != nil {
			return err
		}
		if admins <= 1 {
			return ErrLastAdmin
		}
	}
	return s.repo.Delete(targetID)
}
