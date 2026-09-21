package database

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"boilerplate/internal/config"
	"boilerplate/internal/domain"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Seed inserts demo accounts (and a few sign-in records for the demo user) the
// first time the app boots against an empty database. Passwords below are for
// local development only; override the admin account via BARA_SEED_ADMIN_*.
func Seed(db *gorm.DB, cfg config.Config, log *slog.Logger) error {
	var count int64
	if err := db.Model(&domain.User{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	log.Info("empty database, seeding demo data")

	adminHash, err := bcrypt.GenerateFromPassword([]byte(cfg.SeedAdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	demoHash, err := bcrypt.GenerateFromPassword([]byte("User1234"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	now := time.Now()
	users := []domain.User{
		{Name: "Raka Ardiansyah", Email: strings.ToLower(cfg.SeedAdminEmail), Password: string(adminHash), Role: domain.RoleAdmin, IsActive: true, CreatedAt: now.Add(-120 * 24 * time.Hour), UpdatedAt: now},
		{Name: "Aulia Rahmani", Email: "aulia@bara.dev", Password: string(demoHash), Role: domain.RoleUser, IsActive: true, CreatedAt: now.Add(-84 * 24 * time.Hour), UpdatedAt: now},
		{Name: "Bima Nugraha", Email: "bima@bara.dev", Password: string(demoHash), Role: domain.RoleUser, IsActive: true, CreatedAt: now.Add(-41 * 24 * time.Hour), UpdatedAt: now},
		{Name: "Clara Wibisono", Email: "clara@bara.dev", Password: string(demoHash), Role: domain.RoleUser, IsActive: true, CreatedAt: now.Add(-12 * 24 * time.Hour), UpdatedAt: now},
		{Name: "Dimas Prasetya", Email: "dimas@bara.dev", Password: string(demoHash), Role: domain.RoleUser, IsActive: false, CreatedAt: now.Add(-5 * 24 * time.Hour), UpdatedAt: now},
	}
	if err := db.Create(&users).Error; err != nil {
		return err
	}

	var aulia domain.User
	if err := db.Where("email = ?", "aulia@bara.dev").First(&aulia).Error; err != nil {
		return fmt.Errorf("seed: find demo user: %w", err)
	}

	type seedSession struct {
		userAgent string
		ip        string
		age       time.Duration
	}
	seeds := []seedSession{
		{"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36", "114.10.44.210", 2 * time.Hour},
		{"Mozilla/5.0 (iPhone; CPU iPhone OS 17_5 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.5 Mobile/15E148 Safari/604.1", "118.136.90.7", 3 * 24 * time.Hour},
		{"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:127.0) Gecko/20100101 Firefox/127.0", "103.11.20.88", 9 * 24 * time.Hour},
	}
	for _, s := range seeds {
		id, err := randomID()
		if err != nil {
			return err
		}
		created := now.Add(-s.age)
		err = db.Create(&domain.Session{
			ID:         id,
			UserID:     aulia.ID,
			UserAgent:  s.userAgent,
			IP:         s.ip,
			CreatedAt:  created,
			LastSeenAt: created.Add(20 * time.Minute),
			ExpiresAt:  now.Add(7 * 24 * time.Hour),
		}).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func randomID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
