package handler

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"

	"github.com/sean-ahn/user/backend/server/service"
	userv1 "github.com/sean-ahn/user/proto/gen/go/user/v1"
)

func TestRefreshToken(t *testing.T) {
	cases := []struct {
		name string
		req  *userv1.RefreshTokenRequest

		userTokenServiceExpectFunc func(context.Context) func(*service.MockUserTokenService)

		expectedCode codes.Code
		expectedResp *userv1.RefreshTokenResponse
		expectedErr  string
	}{
		{
			name: "success",
			req:  &userv1.RefreshTokenRequest{RefreshToken: "old_refresh_token"},
			userTokenServiceExpectFunc: func(ctx context.Context) func(*service.MockUserTokenService) {
				return func(mock *service.MockUserTokenService) {
					mock.EXPECT().
						Refresh(ctx, "old_refresh_token").
						Return("new_access_token", "new_refresh_token", nil)
				}
			},
			expectedCode: codes.OK,
			expectedResp: &userv1.RefreshTokenResponse{
				AccessToken:  "new_access_token",
				RefreshToken: "new_refresh_token",
			},
		},
		{
			name:         "no refresh_token",
			req:          &userv1.RefreshTokenRequest{},
			expectedCode: codes.InvalidArgument,
			expectedErr:  "rpc error: code = InvalidArgument desc = no refresh_token",
		},
		{
			name: "refresh token expired",
			req:  &userv1.RefreshTokenRequest{RefreshToken: "expired_refresh_token"},
			userTokenServiceExpectFunc: func(ctx context.Context) func(*service.MockUserTokenService) {
				return func(mock *service.MockUserTokenService) {
					mock.EXPECT().
						Refresh(ctx, "expired_refresh_token").
						Return("", "", errors.New("token expired"))
				}
			},
			expectedCode: codes.Unauthenticated,
			expectedErr:  "rpc error: code = Unauthenticated desc = token expired",
		},
		{
			name: "refresh token validation is delegated to service",
			req:  &userv1.RefreshTokenRequest{RefreshToken: "1234"},
			userTokenServiceExpectFunc: func(ctx context.Context) func(*service.MockUserTokenService) {
				return func(mock *service.MockUserTokenService) {
					mock.EXPECT().
						Refresh(ctx, "1234").
						Return("", "", errors.New("invalid token"))
				}
			},
			expectedCode: codes.Unauthenticated,
			expectedErr:  "rpc error: code = Unauthenticated desc = invalid token",
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			ctrl := gomock.NewController(t)

			mockUserTokenService := service.NewMockUserTokenService(ctrl)
			if tc.userTokenServiceExpectFunc != nil {
				tc.userTokenServiceExpectFunc(ctx)(mockUserTokenService)
			}

			handler := RefreshToken(mockUserTokenService)

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
