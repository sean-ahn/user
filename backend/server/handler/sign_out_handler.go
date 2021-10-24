package handler

import (
	"context"

	"github.com/friendsofgo/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/sean-ahn/user/backend/server/service"
	userv1 "github.com/sean-ahn/user/proto/gen/go/user/v1"
)

type SignOutHandlerFunc func(ctx context.Context, req *userv1.SignOutRequest) (*userv1.SignOutResponse, error)

func SignOut(userTokenService service.UserTokenService) SignOutHandlerFunc {
	return func(ctx context.Context, req *userv1.SignOutRequest) (*userv1.SignOutResponse, error) {
		if req.RefreshToken == "" {
			return nil, status.Error(codes.InvalidArgument, "no refresh_token")
		}

		if err := userTokenService.Revoke(ctx, req.RefreshToken); err != nil {
			switch errors.Cause(err) {
			case service.ErrTokenRevocationFailed:
				return nil, status.Error(codes.Internal, err.Error())
			default:
				return nil, status.Error(codes.InvalidArgument, err.Error())
			}
		}

		return &userv1.SignOutResponse{}, nil
	}
}
