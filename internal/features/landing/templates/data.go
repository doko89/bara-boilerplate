package templates

import "boilerplate/internal/domain"

// LandingData carries everything the landing page needs.
type LandingData struct {
	Title string
	Theme string // "light", "dark" or "" (follow system)
	User  *domain.User
}
