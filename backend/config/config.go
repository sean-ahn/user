package config

import (
	"database/sql"

	"github.com/jonboulle/clockwork"

	"github.com/sean-ahn/user/backend/crypto"
	"github.com/sean-ahn/user/backend/server/service"
	smsv1 "github.com/sean-ahn/user/proto/gen/go/sms/v1"
)

type Config interface {
	Setting() Setting

	Clock() clockwork.Clock
	DB() *sql.DB
	PasswordHasher() crypto.Hasher
	SmsV1Client() smsv1.SmsServiceClient
	UserTokenService() service.UserTokenService
}

type DefaultConfig struct {
	setting          Setting
	clock            clockwork.Clock
	db               *sql.DB
	passwordHasher   crypto.Hasher
	smsv1Cli         smsv1.SmsServiceClient
	userTokenService service.UserTokenService
}

var _ Config = (*DefaultConfig)(nil)

func (c *DefaultConfig) Setting() Setting {
	return c.setting
}

func (c *DefaultConfig) Clock() clockwork.Clock {
	return c.clock
}

func (c *DefaultConfig) DB() *sql.DB {
	return c.db
}

func (c *DefaultConfig) PasswordHasher() crypto.Hasher {
	return c.passwordHasher
}

func (c *DefaultConfig) SmsV1Client() smsv1.SmsServiceClient {
	return c.smsv1Cli
}

func (c *DefaultConfig) UserTokenService() service.UserTokenService {
	return c.userTokenService
}

func New(
	setting Setting,
	clock clockwork.Clock,
	db *sql.DB,
	passwordHasher crypto.Hasher,
	smsv1Cli smsv1.SmsServiceClient,
	userTokenService service.UserTokenService,
) *DefaultConfig {
	return &DefaultConfig{
		setting:          setting,
		clock:            clock,
		db:               db,
		passwordHasher:   passwordHasher,
		smsv1Cli:         smsv1Cli,
		userTokenService: userTokenService,
	}
}
