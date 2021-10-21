package config

import (
	"os"
	"strconv"

	"github.com/sirupsen/logrus"
)

type Setting struct {
	GRPCServerPort            int
	HTTPServerPort            int
	GracefulShutdownTimeoutMs int
}

func NewSetting() Setting {
	return Setting{
		GRPCServerPort:            mustAtoi(getEnv("GRPC_SERVER_PORT", "8080")),
		HTTPServerPort:            mustAtoi(getEnv("HTTP_SERVER_PORT", "8081")),
		GracefulShutdownTimeoutMs: mustAtoi(getEnv("GRACEFUL_SHUTDOWN_TIMEOUT_MS", "3000")),
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