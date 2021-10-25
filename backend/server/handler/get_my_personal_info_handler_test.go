package handler

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"

	"github.com/sean-ahn/user/backend/model"
	"github.com/sean-ahn/user/backend/server/service"
	userv1 "github.com/sean-ahn/user/proto/gen/go/user/v1"
)

func TestGetMyPersonalInfo(t *testing.T) {
	cases := []struct {
		name string
		req  *userv1.GetMyPersonalInfoRequest

		userTokenServiceExpectFunc func(context.Context) func(*service.MockUserTokenService)

		expectedCode codes.Code
		expectedResp *userv1.GetMyPersonalInfoResponse
		expectedErr  string
	}{
		{
			name: "success",
			req:  &userv1.GetMyPersonalInfoRequest{},
			userTokenServiceExpectFunc: func(ctx context.Context) func(*service.MockUserTokenService) {
				return func(mock *service.MockUserTokenService) {
					mock.EXPECT().
						GetUser(ctx, "access_token").
						Return(&model.User{
							Name:        "name",
							Email:       "john.doe@example.com",
							PhoneNumber: "+821012345678",
							Nickname:    "nickname",
						}, nil)
				}
			},
			expectedCode: codes.OK,
			expectedResp: &userv1.GetMyPersonalInfoResponse{
				PersonalInfo: &userv1.PersonalInfo{
					Name:        "name",
					Email:       "john.doe@example.com",
					PhoneNumber: "+821012345678",
					Nickname:    "nickname",
				},
			},
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			ctx := metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{headerKeyAuthorization: "access_token"}))

			ctrl := gomock.NewController(t)

			mockUserTokenService := service.NewMockUserTokenService(ctrl)
			if tc.userTokenServiceExpectFunc != nil {
				tc.userTokenServiceExpectFunc(ctx)(mockUserTokenService)
			}

			handler := GetMyPersonalInfo(mockUserTokenService)

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
