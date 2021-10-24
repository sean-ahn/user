package config

import (
	"database/sql"

	"github.com/sean-ahn/user/backend/crypto"
)

type Config interface {
	Setting() Setting

	DB() *sql.DB
	PasswordHasher() crypto.Hasher
}

type DefaultConfig struct {
	setting        Setting
	db             *sql.DB
	passwordHasher crypto.Hasher
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

func New(setting Setting, db *sql.DB, passwordHasher crypto.Hasher) *DefaultConfig {
	return &DefaultConfig{
		setting:        setting,
		db:             db,
		passwordHasher: passwordHasher,
	}
}
