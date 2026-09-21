package templates

import (
	"time"

	"boilerplate/internal/domain"
)

type OverviewData struct {
	Theme       string
	CSRF        string
	FlashType   string
	FlashMsg    string
	Greeting    string
	SignInCount int64
	LastSignIn  string
	MemberSince time.Time
	Recent      []domain.Session
}
