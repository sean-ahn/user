package handler

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
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
			name:         "no refresh_token",
			req:          &userv1.RefreshTokenRequest{},
			expectedCode: codes.InvalidArgument,
			expectedErr:  "rpc error: code = InvalidArgument desc = no refresh_token",
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			ctrl := gomock.NewController(t)

			mockUserTokenService := service.NewMockUserTokenService(ctrl)

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
