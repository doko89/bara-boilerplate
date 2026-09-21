package domain

import (
	"strings"
	"time"
)

// Role is the access level assigned to a user.
type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

func (r Role) Valid() bool { return r == RoleAdmin || r == RoleUser }

// User is the account entity shared across features.
type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:120;not null" json:"name"`
	Email     string    `gorm:"size:190;uniqueIndex;not null" json:"email"`
	Password  string    `gorm:"size:255;not null" json:"-"`
	Role      Role      `gorm:"size:20;not null;default:user;index" json:"role"`
	IsActive  bool      `gorm:"not null;default:true" json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (u *User) IsAdmin() bool { return u.Role == RoleAdmin }

// Initials renders the first letters of the first and last name for avatars.
func (u *User) Initials() string {
	parts := strings.Fields(u.Name)
	if len(parts) == 0 {
		return "?"
	}
	first := strings.ToUpper(parts[0][:1])
	if len(parts) == 1 {
		return first
	}
	return first + strings.ToUpper(parts[len(parts)-1][:1])
}
