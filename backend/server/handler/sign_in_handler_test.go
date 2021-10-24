package handler

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/testing/protocmp"

	userv1 "github.com/sean-ahn/user/proto/gen/go/user/v1"
)

func TestSignIn(t *testing.T) {
	cases := []struct {
		name string
		req  *userv1.SignInRequest

		expectedCode codes.Code
		expectedResp *userv1.SignInResponse
		expectedErr  string
	}{
		{
			name: "success",
			req:  &userv1.SignInRequest{Id: "id", Password: "p@ssw0rd"},
			expectedCode: codes.OK,
			expectedResp: &userv1.SignInResponse{},
		},
		{
			name: "no id",
			req:  &userv1.SignInRequest{Password: "p@ssw0rd"},
			expectedCode: codes.InvalidArgument,
			expectedErr: "rpc error: code = InvalidArgument desc = no id",
		},
		{
			name: "no password",
			req:  &userv1.SignInRequest{Id: "id"},
			expectedCode: codes.InvalidArgument,
			expectedErr: "rpc error: code = InvalidArgument desc = no password",
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			handler := SignIn()

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
