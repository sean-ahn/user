package handler

import (
	"context"
	"database/sql"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/sean-ahn/user/backend/server/service"
	userv1 "github.com/sean-ahn/user/proto/gen/go/user/v1"
)

type VerifyEmailHandlerFunc func(ctx context.Context, req *userv1.VerifyEmailRequest) (*userv1.VerifyEmailResponse, error)

func VerifyEmail(db *sql.DB) VerifyEmailHandlerFunc {
	return func(ctx context.Context, req *userv1.VerifyEmailRequest) (*userv1.VerifyEmailResponse, error) {
		if req.VerificationToken == "" {
			return nil, status.Error(codes.InvalidArgument, "no verification_token")
		}

		return &userv1.VerifyEmailResponse{AccessToken: accessToken, RefreshToken: refreshToken}, nil
	}
}
