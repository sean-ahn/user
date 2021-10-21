package server

import (
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"github.com/sean-ahn/user/backend/config"
	userv1 "github.com/sean-ahn/user/proto/gen/go/user/v1"
)

type UserServer struct {
	userv1.UnimplementedUserServiceServer

	cfg config.Config
}

func NewUserServer(cfg config.Config) (*UserServer, error) {
	return &UserServer{
		cfg: cfg,
	}, nil
}

func (s *UserServer) Config() config.Config {
	return s.cfg
}

func NewGRPCServer(cfg config.Config) (*grpc.Server, error) {
	logrus.ErrorKey = "grpc.error"
	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})

	srv := grpc.NewServer()

	userServer, err := NewUserServer(cfg)
	if err != nil {
		return nil, err
	}

	userv1.RegisterUserServiceServer(srv, userServer)

	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(srv, healthServer)

	reflection.Register(srv)

	return srv, nil
}
