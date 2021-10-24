package handler

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/sean-ahn/user/backend/server/service"
	userv1 "github.com/sean-ahn/user/proto/gen/go/user/v1"
)

type RefreshTokenHandlerFunc func(ctx context.Context, req *userv1.RefreshTokenRequest) (*userv1.RefreshTokenResponse, error)

func RefreshToken(userTokenService service.UserTokenService) RefreshTokenHandlerFunc {
	return func(ctx context.Context, req *userv1.RefreshTokenRequest) (*userv1.RefreshTokenResponse, error) {
		if req.RefreshToken == "" {
			return nil, status.Error(codes.InvalidArgument, "no refresh_token")
		}

		accessToken, refreshToken, err := userTokenService.Refresh(ctx, req.RefreshToken)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid refresh token")
		}

		return &userv1.RefreshTokenResponse{AccessToken: accessToken, RefreshToken: refreshToken}, nil
	}
}
