package handler

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"

	"github.com/sean-ahn/user/backend/model"
	"github.com/sean-ahn/user/backend/server/service"
	userv1 "github.com/sean-ahn/user/proto/gen/go/user/v1"
)

func TestSignOut(t *testing.T) {
	cases := []struct {
		name string
		req  *userv1.SignOutRequest

		userTokenServiceExpectFunc func(context.Context) func(*service.MockUserTokenService)

		expectedCode codes.Code
		expectedResp *userv1.SignOutResponse
		expectedErr  string
	}{
		{
			name: "success",
			req:  &userv1.SignOutRequest{RefreshToken: "refresh_token"},
			userTokenServiceExpectFunc: func(ctx context.Context) func(*service.MockUserTokenService) {
				return func(mock *service.MockUserTokenService) {
					mock.EXPECT().
						GetUser(ctx, "access_token").
						Return(&model.User{UserID: 1}, nil)

					mock.EXPECT().
						Revoke(ctx, "refresh_token").
						Return(nil)
				}
			},
			expectedCode: codes.OK,
			expectedResp: &userv1.SignOutResponse{},
		},
		{
			name: "unauthorized",
			req:  &userv1.SignOutRequest{RefreshToken: "refresh_token"},
			userTokenServiceExpectFunc: func(ctx context.Context) func(*service.MockUserTokenService) {
				return func(mock *service.MockUserTokenService) {
					mock.EXPECT().
						GetUser(ctx, "access_token").
						Return(nil, errors.New("invalid token"))
				}
			},
			expectedCode: codes.Unauthenticated,
			expectedErr:  "rpc error: code = Unauthenticated desc = invalid token",
		},
		{
			name: "no refresh_token",
			req:  &userv1.SignOutRequest{},
			userTokenServiceExpectFunc: func(ctx context.Context) func(*service.MockUserTokenService) {
				return func(mock *service.MockUserTokenService) {
					mock.EXPECT().
						GetUser(ctx, "access_token").
						Return(&model.User{UserID: 1}, nil)
				}
			},
			expectedCode: codes.InvalidArgument,
			expectedErr:  "rpc error: code = InvalidArgument desc = no refresh_token",
		},
		{
			name: "refresh token revocation failed",
			req:  &userv1.SignOutRequest{RefreshToken: "refresh_token"},
			userTokenServiceExpectFunc: func(ctx context.Context) func(*service.MockUserTokenService) {
				return func(mock *service.MockUserTokenService) {
					mock.EXPECT().
						GetUser(ctx, "access_token").
						Return(&model.User{UserID: 1}, nil)

					mock.EXPECT().
						Revoke(ctx, "refresh_token").
						Return(service.ErrTokenRevocationFailed)
				}
			},
			expectedCode: codes.Internal,
			expectedErr:  "rpc error: code = Internal desc = token revocation failed",
		},
		{
			name: "unexpected error",
			req:  &userv1.SignOutRequest{RefreshToken: "refresh_token"},
			userTokenServiceExpectFunc: func(ctx context.Context) func(*service.MockUserTokenService) {
				return func(mock *service.MockUserTokenService) {
					mock.EXPECT().
						GetUser(ctx, "access_token").
						Return(&model.User{UserID: 1}, nil)

					mock.EXPECT().
						Revoke(ctx, "refresh_token").
						Return(errors.New("unexpected error"))
				}
			},
			expectedCode: codes.InvalidArgument,
			expectedErr:  "rpc error: code = InvalidArgument desc = unexpected error",
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			ctx := metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				headerKeyAuthorization: "Bearer access_token",
			}))

			ctrl := gomock.NewController(t)

			mockUserTokenService := service.NewMockUserTokenService(ctrl)
			if tc.userTokenServiceExpectFunc != nil {
				tc.userTokenServiceExpectFunc(ctx)(mockUserTokenService)
			}

			handler := SignOut(mockUserTokenService)

			resp, err := handler(ctx, tc.req)

			if tc.expectedErr != "" {
				assert.EqualError(t, err, tc.expectedErr)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.expectedCode, status.Code(err))
			assert.Empty(t, cmp.Diff(resp, tc.expectedResp, protocmp.Transform()))
		})
	}

}
