package admin

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

func (r *Repository) ListUsers() ([]domain.User, error) {
	var users []domain.User
	err := r.db.Order("created_at DESC").Find(&users).Error
	return users, err
}

func (r *Repository) RecentUsers(limit int) ([]domain.User, error) {
	var users []domain.User
	err := r.db.Order("created_at DESC").Limit(limit).Find(&users).Error
	return users, err
}

func (r *Repository) GetUser(id uint) (*domain.User, error) {
	var u domain.User
	if err := r.db.First(&u, id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repository) CountUsers() (int64, error) {
	var c int64
	err := r.db.Model(&domain.User{}).Count(&c).Error
	return c, err
}

func (r *Repository) CountAdmins() (int64, error) {
	var c int64
	err := r.db.Model(&domain.User{}).Where("role = ?", domain.RoleAdmin).Count(&c).Error
	return c, err
}

func (r *Repository) CountActiveUsers() (int64, error) {
	var c int64
	err := r.db.Model(&domain.User{}).Where("is_active = ?", true).Count(&c).Error
	return c, err
}

func (r *Repository) CountCreatedSince(t time.Time) (int64, error) {
	var c int64
	err := r.db.Model(&domain.User{}).Where("created_at >= ?", t).Count(&c).Error
	return c, err
}

func (r *Repository) SetRole(id uint, role domain.Role) error {
	return r.db.Model(&domain.User{}).
		Where("id = ?", id).
		Updates(map[string]any{"role": role, "updated_at": time.Now()}).Error
}

func (r *Repository) SetStatus(id uint, active bool) error {
	return r.db.Model(&domain.User{}).
		Where("id = ?", id).
		Updates(map[string]any{"is_active": active, "updated_at": time.Now()}).Error
}

// Delete removes the user together with their sessions and reset tokens.
func (r *Repository) Delete(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", id).Delete(&domain.Session{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", id).Delete(&domain.PasswordResetToken{}).Error; err != nil {
			return err
		}
		return tx.Delete(&domain.User{}, id).Error
	})
}
