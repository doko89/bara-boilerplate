package templates

import (
	"boilerplate/internal/domain"
)

type SettingsData struct {
	Title            string
	Theme            string
	CSRF             string
	FlashType        string
	FlashMsg         string
	Name             string
	Email            string
	AvatarURL        string
	Initials         string
	MaxUploadBytes   int64
	Errors           map[string]string
	Sessions         []domain.Session
	CurrentSessionID string
}
