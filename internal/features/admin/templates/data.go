package templates

import (
	"boilerplate/internal/domain"
)

type OverviewData struct {
	Theme          string
	CSRF           string
	FlashType      string
	FlashMsg       string
	TotalUsers     int64
	ActiveUsers    int64
	Admins         int64
	NewThisWeek    int64
	ActiveSessions int64
	RecentUsers    []domain.User
}

type UsersData struct {
	Theme         string
	CSRF          string
	FlashType     string
	FlashMsg      string
	Users         []domain.User
	CurrentUserID uint
}
