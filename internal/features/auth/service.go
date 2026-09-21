package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log/slog"
	"net/mail"
	"strings"
	"time"

	"boilerplate/internal/domain"
	"boilerplate/internal/session"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailTaken         = errors.New("email already registered")
	ErrAccountDisabled    = errors.New("account is deactivated")
	ErrInvalidToken       = errors.New("invalid or expired reset token")
)

// UserRepository is the storage contract; the GORM implementation lives in
// repository.go.
type UserRepository interface {
	FindByEmail(email string) (*domain.User, error)
	ExistsByEmail(email string) (bool, error)
	Create(user *domain.User) error
	CreateResetToken(token *domain.PasswordResetToken) error
	FindValidResetToken(tokenHash string) (*domain.PasswordResetToken, error)
	MarkResetTokenUsed(token *domain.PasswordResetToken) error
	UpdatePassword(userID uint, hash string) error
}

type Service struct {
	repo     UserRepository
	sessions *session.Manager
	log      *slog.Logger
}

func NewService(repo UserRepository, sessions *session.Manager, log *slog.Logger) *Service {
	return &Service{repo: repo, sessions: sessions, log: log}
}

// dummyHash keeps login timing similar for unknown emails, so responses don't
// reveal whether an account exists.
var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("timing-equalizer"), bcrypt.DefaultCost)

func NormalizeEmail(raw string) string {
	return strings.ToLower(strings.TrimSpace(raw))
}

func ValidEmail(raw string) bool {
	if raw == "" || len(raw) > 190 {
		return false
	}
	if strings.ContainsAny(raw, " <>\"") {
		return false
	}
	_, err := mail.ParseAddress(raw)
	return err == nil
}

func ValidPassword(pw string) bool {
	return len(pw) >= 8
}

func (s *Service) Register(name, email, password string) (*domain.User, error) {
	email = NormalizeEmail(email)
	taken, err := s.repo.ExistsByEmail(email)
	if err != nil {
		return nil, err
	}
	if taken {
		return nil, ErrEmailTaken
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user := &domain.User{
		Name:     name,
		Email:    email,
		Password: string(hash),
		Role:     domain.RoleUser,
		IsActive: true,
	}
	if err := s.repo.Create(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Service) Login(email, password string) (*domain.User, error) {
	user, err := s.repo.FindByEmail(NormalizeEmail(email))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(password))
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
		return nil, ErrInvalidCredentials
	}
	if !user.IsActive {
		return nil, ErrAccountDisabled
	}
	return user, nil
}

// RequestPasswordReset creates a single-use token valid for one hour. It
// returns a token only when the account exists; callers must not reveal the
// difference. There is no mailer in the boilerplate, so the handler logs the
// reset link.
func (s *Service) RequestPasswordReset(email string) (string, error) {
	user, err := s.repo.FindByEmail(NormalizeEmail(email))
	if err != nil {
		return "", nil
	}
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	token := hex.EncodeToString(raw)
	sum := sha256.Sum256([]byte(token))
	err = s.repo.CreateResetToken(&domain.PasswordResetToken{
		UserID:    user.ID,
		TokenHash: hex.EncodeToString(sum[:]),
		ExpiresAt: time.Now().Add(time.Hour),
	})
	if err != nil {
		return "", err
	}
	return token, nil
}

// ValidResetToken reports whether a raw token can still be used, so the
// reset page can show the expired state before any submit.
func (s *Service) ValidResetToken(rawToken string) bool {
	if !isHex64(rawToken) {
		return false
	}
	sum := sha256.Sum256([]byte(rawToken))
	_, err := s.repo.FindValidResetToken(hex.EncodeToString(sum[:]))
	return err == nil
}

func (s *Service) ResetPassword(rawToken, newPassword string) error {
	if !isHex64(rawToken) {
		return ErrInvalidToken
	}
	sum := sha256.Sum256([]byte(rawToken))
	rt, err := s.repo.FindValidResetToken(hex.EncodeToString(sum[:]))
	if err != nil {
		return ErrInvalidToken
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if err := s.repo.UpdatePassword(rt.UserID, string(hash)); err != nil {
		return err
	}
	if err := s.repo.MarkResetTokenUsed(rt); err != nil {
		return err
	}
	// A password change invalidates every existing session.
	s.sessions.DestroyForUser(rt.UserID, "")
	return nil
}

func isHex64(s string) bool {
	if len(s) != 64 {
		return false
	}
	_, err := hex.DecodeString(s)
	return err == nil
}
