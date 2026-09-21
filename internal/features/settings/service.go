package settings

import (
	"errors"

	"boilerplate/internal/domain"
	"boilerplate/internal/session"

	"golang.org/x/crypto/bcrypt"
)

var ErrEmailTaken = errors.New("email already registered")
var ErrWrongPassword = errors.New("current password is incorrect")

// SettingsRepository is implemented by repository.go (GORM).
type SettingsRepository interface {
	EmailTakenBy(email string, exceptUserID uint) (bool, error)
	UpdateProfile(userID uint, name, email string) error
	UpdatePassword(userID uint, hash string) error
}

type Service struct {
	repo     SettingsRepository
	sessions *session.Manager
}

func NewService(repo SettingsRepository, sessions *session.Manager) *Service {
	return &Service{repo: repo, sessions: sessions}
}

func (s *Service) Sessions(userID uint, limit int) []domain.Session {
	return s.sessions.ListForUser(userID, limit)
}

func (s *Service) UpdateProfile(userID uint, name, email string) error {
	taken, err := s.repo.EmailTakenBy(email, userID)
	if err != nil {
		return err
	}
	if taken {
		return ErrEmailTaken
	}
	return s.repo.UpdateProfile(userID, name, email)
}

// ChangePassword verifies the current password, stores the new hash and signs
// out every other session of the user.
func (s *Service) ChangePassword(u *domain.User, currentSessionID, current, newPassword string) error {
	if bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(current)) != nil {
		return ErrWrongPassword
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if err := s.repo.UpdatePassword(u.ID, string(hash)); err != nil {
		return err
	}
	s.sessions.DestroyForUser(u.ID, currentSessionID)
	return nil
}

func (s *Service) RevokeOtherSessions(userID uint, currentSessionID string) {
	s.sessions.DestroyForUser(userID, currentSessionID)
}
