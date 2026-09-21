package auth

import (
	"time"

	"boilerplate/internal/domain"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindByEmail(email string) (*domain.User, error) {
	var u domain.User
	if err := r.db.Where("email = ?", email).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repository) ExistsByEmail(email string) (bool, error) {
	var c int64
	if err := r.db.Model(&domain.User{}).Where("email = ?", email).Count(&c).Error; err != nil {
		return false, err
	}
	return c > 0, nil
}

func (r *Repository) Create(user *domain.User) error {
	return r.db.Create(user).Error
}

func (r *Repository) CreateResetToken(t *domain.PasswordResetToken) error {
	return r.db.Create(t).Error
}

func (r *Repository) FindValidResetToken(tokenHash string) (*domain.PasswordResetToken, error) {
	var t domain.PasswordResetToken
	err := r.db.
		Where("token_hash = ? AND used_at IS NULL AND expires_at > ?", tokenHash, time.Now()).
		First(&t).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *Repository) MarkResetTokenUsed(t *domain.PasswordResetToken) error {
	return r.db.Model(t).Update("used_at", time.Now()).Error
}

func (r *Repository) UpdatePassword(userID uint, hash string) error {
	return r.db.Model(&domain.User{}).
		Where("id = ?", userID).
		Updates(map[string]any{"password": hash, "updated_at": time.Now()}).Error
}
