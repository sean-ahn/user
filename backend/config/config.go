package config

import (
	"database/sql"

	"github.com/sean-ahn/user/backend/server/service"

	"github.com/sean-ahn/user/backend/crypto"
)

type Config interface {
	Setting() Setting

	DB() *sql.DB
	PasswordHasher() crypto.Hasher
	UserTokenService() service.UserTokenService
}

type DefaultConfig struct {
	setting          Setting
	db               *sql.DB
	passwordHasher   crypto.Hasher
	userTokenService service.UserTokenService
}

var _ Config = (*DefaultConfig)(nil)

func (c *DefaultConfig) Setting() Setting {
	return c.setting
}

func (c *DefaultConfig) DB() *sql.DB {
	return c.db
}

func (c *DefaultConfig) PasswordHasher() crypto.Hasher {
	return c.passwordHasher
}

func (c *DefaultConfig) UserTokenService() service.UserTokenService {
	return c.userTokenService
}

func New(setting Setting, db *sql.DB, passwordHasher crypto.Hasher, userTokenService service.UserTokenService) *DefaultConfig {
	return &DefaultConfig{
		setting:          setting,
		db:               db,
		passwordHasher:   passwordHasher,
		userTokenService: userTokenService,
	}
}
