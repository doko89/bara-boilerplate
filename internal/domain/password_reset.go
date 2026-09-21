package domain

import "time"

// PasswordResetToken stores only the SHA-256 hash of the reset token that was
// emailed (in development: logged) to the user.
type PasswordResetToken struct {
	ID        uint `gorm:"primaryKey"`
	UserID    uint `gorm:"index;not null"`
	User      User `gorm:"constraint:OnDelete:CASCADE"`
	TokenHash string `gorm:"size:64;uniqueIndex"`
	ExpiresAt time.Time
	UsedAt    *time.Time
	CreatedAt time.Time
}

func (t *PasswordResetToken) Expired() bool { return time.Now().After(t.ExpiresAt) }

func (t *PasswordResetToken) Used() bool { return t.UsedAt != nil }
