package handler

import (
	"context"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/sean-ahn/user/backend/model"
	"github.com/sean-ahn/user/backend/server/service"
	userv1 "github.com/sean-ahn/user/proto/gen/go/user/v1"
)

const (
	headerKeyAuthorization = "Authorization"
)

type GetMyPersonalInfoHandlerFunc func(ctx context.Context, req *userv1.GetMyPersonalInfoRequest) (*userv1.GetMyPersonalInfoResponse, error)

func GetMyPersonalInfo(userTokenService service.UserTokenService) GetMyPersonalInfoHandlerFunc {
	return func(ctx context.Context, req *userv1.GetMyPersonalInfoRequest) (*userv1.GetMyPersonalInfoResponse, error) {
		token := extractToken(ctx)
		if token == "" {
			return nil, status.Error(codes.Unauthenticated, "no token")
		}

		user, err := userTokenService.GetUser(ctx, token)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		return &userv1.GetMyPersonalInfoResponse{PersonalInfo: convertToPersonalInfo(user)}, nil
	}
}

func extractToken(ctx context.Context) string {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if authorization := md.Get(headerKeyAuthorization); len(authorization) == 1 {
			split := strings.Split(authorization[0], " ")
			if len(split) == 2 && split[0] == "Bearer" {
				return split[1]
			}
			return ""
		}
	}
	return ""
}

func convertToPersonalInfo(u *model.User) *userv1.PersonalInfo {
	return &userv1.PersonalInfo{
		Name:        u.Name,
		Email:       u.Email,
		PhoneNumber: u.PhoneNumber,
		Nickname:    u.Nickname,
	}
}
