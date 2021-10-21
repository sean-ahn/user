package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/sean-ahn/user/backend/config"
	"github.com/sean-ahn/user/backend/server"
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

	setting := config.NewSetting()

	cfg := config.New(setting)

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
