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
	Errors           map[string]string
	Sessions         []domain.Session
	CurrentSessionID string
}
