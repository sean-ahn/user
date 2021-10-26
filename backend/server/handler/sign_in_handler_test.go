package handler

import (
	"context"
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"

	"github.com/sean-ahn/user/backend/model"
	"github.com/sean-ahn/user/backend/server/service"
	"github.com/sean-ahn/user/backend/test"
	userv1 "github.com/sean-ahn/user/proto/gen/go/user/v1"
)

type testHasher struct{}

func (h *testHasher) Hash(s []byte) ([]byte, error) {
	return []byte(string(s) + "_hash"), nil
}

func TestSignIn(t *testing.T) {
	cases := []struct {
		name string
		req  *userv1.SignInRequest

		dbExpectFunc               func(sqlmock.Sqlmock)
		userTokenServiceExpectFunc func(context.Context) func(*service.MockUserTokenService)

		expectedCode codes.Code
		expectedResp *userv1.SignInResponse
		expectedErr  string
	}{
		{
			name: "sign in with phone number",
			req:  &userv1.SignInRequest{Id: "821012345678", Password: "P@ssw0rd"},
			dbExpectFunc: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(regexp.QuoteMeta(
					"SELECT * FROM `user` WHERE (`user`.`phone_number` = ?) LIMIT 1;",
				)).WithArgs(
					"+821012345678",
				).WillReturnRows(test.NewUserRows([]*model.User{
					{UserID: 1, Email: "john.doe@example.com", IsEmailConfirmed: true, PhoneNumber: "+821012345678", PasswordHash: "P@ssw0rd_hash"},
				}))
			},
			userTokenServiceExpectFunc: func(ctx context.Context) func(*service.MockUserTokenService) {
				return func(mock *service.MockUserTokenService) {
					mock.EXPECT().
						Issue(ctx, &model.User{UserID: 1, Email: "john.doe@example.com", IsEmailConfirmed: true, PhoneNumber: "+821012345678", PasswordHash: "P@ssw0rd_hash"}).
						Return("access_token", "refresh_token", nil)
				}
			},
			expectedCode: codes.OK,
			expectedResp: &userv1.SignInResponse{AccessToken: "access_token", RefreshToken: "refresh_token"},
		},
		{
			name: "sign in with verified email",
			req:  &userv1.SignInRequest{Id: "john.doe@example.com", Password: "P@ssw0rd"},
			dbExpectFunc: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(regexp.QuoteMeta(
					"SELECT * FROM `user` WHERE (`user`.`email` = ?) LIMIT 1;",
				)).WithArgs(
					"john.doe@example.com",
				).WillReturnRows(test.NewUserRows([]*model.User{
					{UserID: 1, Email: "john.doe@example.com", IsEmailConfirmed: true, PhoneNumber: "+821012345678", PasswordHash: "P@ssw0rd_hash"},
				}))
			},
			userTokenServiceExpectFunc: func(ctx context.Context) func(*service.MockUserTokenService) {
				return func(mock *service.MockUserTokenService) {
					mock.EXPECT().
						Issue(ctx, &model.User{UserID: 1, Email: "john.doe@example.com", IsEmailConfirmed: true, PhoneNumber: "+821012345678", PasswordHash: "P@ssw0rd_hash"}).
						Return("access_token", "refresh_token", nil)
				}
			},
			expectedCode: codes.OK,
			expectedResp: &userv1.SignInResponse{AccessToken: "access_token", RefreshToken: "refresh_token"},
		},
		{
			name: "sign in with phone number when email has unverified",
			req:  &userv1.SignInRequest{Id: "+821012345678", Password: "P@ssw0rd"},
			dbExpectFunc: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(regexp.QuoteMeta(
					"SELECT * FROM `user` WHERE (`user`.`phone_number` = ?) LIMIT 1;",
				)).WithArgs(
					"+821012345678",
				).WillReturnRows(test.NewUserRows([]*model.User{
					{UserID: 1, Email: "john.doe@example.com", IsEmailConfirmed: false, PhoneNumber: "+821012345678", PasswordHash: "P@ssw0rd_hash"},
				}))
			},
			userTokenServiceExpectFunc: func(ctx context.Context) func(*service.MockUserTokenService) {
				return func(mock *service.MockUserTokenService) {
					mock.EXPECT().
						Issue(ctx, &model.User{UserID: 1, Email: "john.doe@example.com", IsEmailConfirmed: false, PhoneNumber: "+821012345678", PasswordHash: "P@ssw0rd_hash"}).
						Return("access_token", "refresh_token", nil)
				}
			},
			expectedCode: codes.OK,
			expectedResp: &userv1.SignInResponse{AccessToken: "access_token", RefreshToken: "refresh_token"},
		},
		{
			name: "sign in with unverified email",
			req:  &userv1.SignInRequest{Id: "john.doe@example.com", Password: "P@ssw0rd"},
			dbExpectFunc: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(regexp.QuoteMeta(
					"SELECT * FROM `user` WHERE (`user`.`email` = ?) LIMIT 1;",
				)).WithArgs(
					"john.doe@example.com",
				).WillReturnRows(test.NewUserRows([]*model.User{
					{UserID: 1, Email: "john.doe@example.com", IsEmailConfirmed: false, PhoneNumber: "+821012345678", PasswordHash: "P@ssw0rd_hash"},
				}))
			},
			expectedCode: codes.Unauthenticated,
			expectedErr:  "rpc error: code = Unauthenticated desc = email not verified yet",
		},
		{
			name: "sign in with not existing email",
			req:  &userv1.SignInRequest{Id: "john.doe@example.com", Password: "P@ssw0rd"},
			dbExpectFunc: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(regexp.QuoteMeta(
					"SELECT * FROM `user` WHERE (`user`.`email` = ?) LIMIT 1;",
				)).WithArgs(
					"john.doe@example.com",
				).WillReturnError(
					sql.ErrNoRows,
				)
			},
			expectedCode: codes.Unauthenticated,
			expectedErr:  "rpc error: code = Unauthenticated desc = id or password incorrect",
		},
		{
			name: "sign in with incorrect password",
			req:  &userv1.SignInRequest{Id: "+821012345678", Password: "qwerty123"},
			dbExpectFunc: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(regexp.QuoteMeta(
					"SELECT * FROM `user` WHERE (`user`.`phone_number` = ?) LIMIT 1;",
				)).WithArgs(
					"+821012345678",
				).WillReturnRows(test.NewUserRows([]*model.User{
					{UserID: 1, Email: "john.doe@example.com", IsEmailConfirmed: true, PhoneNumber: "+821012345678", PasswordHash: "P@ssw0rd_hash"},
				}))
			},
			expectedCode: codes.Unauthenticated,
			expectedErr:  "rpc error: code = Unauthenticated desc = id or password incorrect",
		},
		{
			name:         "no id",
			req:          &userv1.SignInRequest{Password: "P@ssw0rd"},
			expectedCode: codes.InvalidArgument,
			expectedErr:  "rpc error: code = InvalidArgument desc = no id",
		},
		{
			name:         "no password",
			req:          &userv1.SignInRequest{Id: "821012345678"},
			expectedCode: codes.InvalidArgument,
			expectedErr:  "rpc error: code = InvalidArgument desc = no password",
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			ctrl := gomock.NewController(t)

			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fail()
			}
			if tc.dbExpectFunc != nil {
				tc.dbExpectFunc(mock)
			}
			defer test.CloseSqlmock(t, db, mock)

			mockUserTokenService := service.NewMockUserTokenService(ctrl)
			if tc.userTokenServiceExpectFunc != nil {
				tc.userTokenServiceExpectFunc(ctx)(mockUserTokenService)
			}

			handler := SignIn(&testHasher{}, db, mockUserTokenService)

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

func Test_detectIDType(t *testing.T) {
	cases := []struct {
		given    string
		expected IDType
	}{
		{given: "@", expected: IDTypeEmail},
		{given: "821012345678", expected: IDTypePhoneNumber},
		{given: "1012345678", expected: IDTypePhoneNumber},
		{given: "abc123", expected: 0},
		{given: "1234", expected: 0},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.given, func(t *testing.T) {
			got := detectIDType(tc.given)

			assert.Equal(t, tc.expected, got)
		})
	}
}

func Test_normalizePhoneNumber(t *testing.T) {
	cases := []struct {
		given, expected string
	}{
		{given: "1012345678", expected: "+821012345678"},
		{given: "821012345678", expected: "+821012345678"},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.given, func(t *testing.T) {
			normalized, err := normalizePhoneNumber(tc.given)

			assert.NoError(t, err)
			assert.Equal(t, tc.expected, normalized)
		})
	}
}
