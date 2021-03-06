package config

import (
	"os"
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/sean-ahn/user/backend/persistence/mysql"
)

type Setting struct {
	GRPCServerPort            int
	HTTPServerPort            int
	GracefulShutdownTimeoutMs int

	DB mysql.Setting

	AccessTokenExpiresInMs  int
	RefreshTokenExpiresInMs int

	SMSOTPCodeLength int

	SMSV1ServiceEndpoint string
}

func NewSetting() Setting {
	return Setting{
		GRPCServerPort:            mustAtoi(getEnv("GRPC_SERVER_PORT", "8080")),
		HTTPServerPort:            mustAtoi(getEnv("HTTP_SERVER_PORT", "8081")),
		GracefulShutdownTimeoutMs: mustAtoi(getEnv("GRACEFUL_SHUTDOWN_TIMEOUT_MS", "3000")),

		DB: mysql.Setting{
			Host:              getEnv("DB_HOST", "localhost"),
			Port:              mustAtoi(getEnv("DB_PORT", "3306")),
			Name:              getEnv("DB_NAME", "user"),
			User:              getEnv("DB_USER", ""),
			Password:          getEnv("DB_PASSWORD", ""),
			MaxIdleConns:      mustAtoi(getEnv("DB_MAX_IDLE_CONNS", "2")),
			MaxOpenConns:      mustAtoi(getEnv("DB_MAX_OPEN_CONNS", "5")),
			ConnMaxLifetimeMs: mustAtoi(getEnv("DB_CONN_MAX_LIFETIME_MS", "14400000")),
		},

		AccessTokenExpiresInMs:  mustAtoi(getEnv("ACCESS_TOKEN_EXPIRES_IN_MS", "600000")),      // 10 min
		RefreshTokenExpiresInMs: mustAtoi(getEnv("REFRESH_TOKEN_EXPIRES_IN_MS", "1209600000")), // 14 days

		SMSOTPCodeLength:     mustAtoi(getEnv("SMS_OTP_CODE_LENGTH", "6")),
		SMSV1ServiceEndpoint: getEnv("SMS_V1_SERVICE_ENDPOINT", ""),
	}
}

func getEnv(key string, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	if defaultValue == "" {
		logrus.Panicf("no environment variable: %s", key)
	}
	return defaultValue
}

func mustAtoi(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		logrus.Panic(err)
	}
	return i
}
