package main

import (
	"context"
	"encoding/base64"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/jonboulle/clockwork"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/sean-ahn/user/backend/client"
	"github.com/sean-ahn/user/backend/config"
	"github.com/sean-ahn/user/backend/crypto"
	"github.com/sean-ahn/user/backend/persistence/mysql"
	"github.com/sean-ahn/user/backend/server"
	"github.com/sean-ahn/user/backend/server/service"
)

const (
	passwordSalt = "DdUvj62VZaFEJkHOQkxT1A=="
)

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})

	if err := run(); err != nil {
		logrus.Panic(err)
	}
}

func run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clock := clockwork.NewRealClock()
	rand.Seed(clock.Now().UnixNano())

	setting := config.NewSetting()

	db := mysql.MustGetDB(setting.DB)

	salt, err := base64.StdEncoding.DecodeString(passwordSalt)
	if err != nil {
		logrus.Panic(err)
	}

	userTokenService := service.NewJWTTokenService(
		clock,
		db,
		time.Duration(setting.AccessTokenExpiresInMs)*time.Millisecond,
		time.Duration(setting.RefreshTokenExpiresInMs)*time.Millisecond,
	)

	cfg := config.New(
		setting,
		clock,
		db,
		crypto.NewScryptHasher(salt),
		client.GetMockSmsV1Service(setting.SMSV1ServiceEndpoint),
		userTokenService,
	)

	grpcServer, err := server.NewGRPCServer(cfg)
	if err != nil {
		logrus.Panic(err)
	}

	httpServer, err := server.NewHTTPServer(ctx, cfg)
	if err != nil {
		logrus.Panic(err)
	}

	go func() {
		lis, err := net.Listen("tcp", ":"+strconv.Itoa(setting.GRPCServerPort))
		if err != nil {
			logrus.Panic(err)
		}

		logrus.WithField("port", setting.GRPCServerPort).Info("starting gRPC server")
		if err := grpcServer.Serve(lis); err != nil && err != grpc.ErrServerStopped {
			logrus.Panic(err)
		}
	}()

	go func() {
		logrus.WithField("port", setting.HTTPServerPort).Info("starting HTTP server")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Panic(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)

	<-quit

	time.Sleep(time.Duration(setting.GracefulShutdownTimeoutMs) * time.Millisecond)

	logrus.Info("stopping HTTP server")
	if err := httpServer.Shutdown(ctx); err != nil {
		logrus.Errorf("shutdown http server: %s", err)
	}

	logrus.Info("stopping gRPC server")
	grpcServer.GracefulStop()

	return nil
}
