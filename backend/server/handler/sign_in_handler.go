package handler

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	userv1 "github.com/sean-ahn/user/proto/gen/go/user/v1"
)

type SignInHandlerFunc func(ctx context.Context, req *userv1.SignInRequest) (*userv1.SignInResponse, error)

func SignIn() SignInHandlerFunc {
	return func(ctx context.Context, req *userv1.SignInRequest) (*userv1.SignInResponse, error) {
		if req.Id == "" {
			return nil, status.Error(codes.InvalidArgument, "no id")
		}
		if req.Password == "" {
			return nil, status.Error(codes.InvalidArgument, "no password")
		}

		return &userv1.SignInResponse{}, nil
	}
}
