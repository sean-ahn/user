package server

import (
	"context"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"github.com/sean-ahn/user/backend/config"
	"github.com/sean-ahn/user/backend/server/generator"
	"github.com/sean-ahn/user/backend/server/handler"
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

func (s *UserServer) RequestSmsOtp(ctx context.Context, req *userv1.RequestSmsOtpRequest) (*userv1.RequestSmsOtpResponse, error) {
	return handler.RequestSmsOtp(s.cfg.Clock(), s.cfg.DB(), &generator.UUIDGenerator{}, &generator.OTPGenerator{Len: s.cfg.Setting().SMSOTPCodeLength}, s.cfg.SmsV1Client())(ctx, req)
}

func (s *UserServer) VerifySmsOtp(ctx context.Context, req *userv1.VerifySmsOtpRequest) (*userv1.VerifySmsOtpResponse, error) {
	return handler.VerifySmsOtp(s.cfg.Clock(), s.cfg.DB())(ctx, req)
}

func (s *UserServer) ConfirmEmail(ctx context.Context, req *userv1.ConfirmEmailRequest) (*userv1.ConfirmEmailResponse, error) {
	return handler.ConfirmEmail(s.cfg.DB())(ctx, req)
}

func (s *UserServer) Register(ctx context.Context, req *userv1.RegisterRequest) (*userv1.RegisterResponse, error) {
	return handler.Register(s.cfg.Clock(), s.cfg.DB(), s.cfg.PasswordHasher())(ctx, req)
}

func (s *UserServer) SignIn(ctx context.Context, req *userv1.SignInRequest) (*userv1.SignInResponse, error) {
	return handler.SignIn(s.cfg.PasswordHasher(), s.cfg.DB(), s.cfg.UserTokenService())(ctx, req)
}

func (s *UserServer) SignOut(ctx context.Context, req *userv1.SignOutRequest) (*userv1.SignOutResponse, error) {
	return handler.SignOut(s.cfg.UserTokenService())(ctx, req)
}

func (s *UserServer) RefreshToken(ctx context.Context, req *userv1.RefreshTokenRequest) (*userv1.RefreshTokenResponse, error) {
	return handler.RefreshToken(s.cfg.UserTokenService())(ctx, req)
}

func (s *UserServer) ResetPassword(ctx context.Context, req *userv1.ResetPasswordRequest) (*userv1.ResetPasswordResponse, error) {
	return handler.ResetPassword(s.cfg.Clock(), s.cfg.DB(), s.cfg.PasswordHasher(), s.cfg.UserTokenService())(ctx, req)
}

func (s *UserServer) GetMyPersonalInfo(ctx context.Context, req *userv1.GetMyPersonalInfoRequest) (*userv1.GetMyPersonalInfoResponse, error) {
	return handler.GetMyPersonalInfo(s.cfg.UserTokenService())(ctx, req)
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
