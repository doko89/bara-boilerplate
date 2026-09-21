package settings

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

func (r *Repository) EmailTakenBy(email string, exceptUserID uint) (bool, error) {
	var c int64
	err := r.db.Model(&domain.User{}).
		Where("email = ? AND id <> ?", email, exceptUserID).
		Count(&c).Error
	return c > 0, err
}

func (r *Repository) UpdateProfile(userID uint, name, email string) error {
	return r.db.Model(&domain.User{}).
		Where("id = ?", userID).
		Updates(map[string]any{"name": name, "email": email, "updated_at": time.Now()}).Error
}

func (r *Repository) UpdatePassword(userID uint, hash string) error {
	return r.db.Model(&domain.User{}).
		Where("id = ?", userID).
		Updates(map[string]any{"password": hash, "updated_at": time.Now()}).Error
}
