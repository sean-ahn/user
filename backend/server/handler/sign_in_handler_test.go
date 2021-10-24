package handler

import (
	"context"
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/sean-ahn/user/backend/model"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/testing/protocmp"

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

		dbExpectFunc func(sqlmock.Sqlmock)

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
				).WillReturnRows(newUserRows([]*model.User{
					{UserID: 1, Email: "john.doe@naver.com", IsEmailVerified: true, PhoneNumber: "+821012345678", PasswordHash: "P@ssw0rd_hash"},
				}))
			},
			expectedCode: codes.OK,
			expectedResp: &userv1.SignInResponse{},
		},
		{
			name: "sign in with verified email",
			req:  &userv1.SignInRequest{Id: "john.doe@naver.com", Password: "P@ssw0rd"},
			dbExpectFunc: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(regexp.QuoteMeta(
					"SELECT * FROM `user` WHERE (`user`.`email` = ?) LIMIT 1;",
				)).WithArgs(
					"john.doe@naver.com",
				).WillReturnRows(newUserRows([]*model.User{
					{UserID: 1, Email: "john.doe@naver.com", IsEmailVerified: true, PhoneNumber: "+821012345678", PasswordHash: "P@ssw0rd_hash"},
				}))
			},
			expectedCode: codes.OK,
			expectedResp: &userv1.SignInResponse{},
		},
		{
			name: "sign in with unverified email",
			req:  &userv1.SignInRequest{Id: "john.doe@naver.com", Password: "P@ssw0rd"},
			dbExpectFunc: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(regexp.QuoteMeta(
					"SELECT * FROM `user` WHERE (`user`.`email` = ?) LIMIT 1;",
				)).WithArgs(
					"john.doe@naver.com",
				).WillReturnRows(newUserRows([]*model.User{
					{UserID: 1, Email: "john.doe@naver.com", IsEmailVerified: false, PhoneNumber: "+821012345678", PasswordHash: "P@ssw0rd_hash"},
				}))
			},
			expectedCode: codes.Unauthenticated,
			expectedErr:  "rpc error: code = Unauthenticated desc = email not verified yet",
		},
		{
			name: "sign in with not existing email",
			req:  &userv1.SignInRequest{Id: "john.doe@naver.com", Password: "P@ssw0rd"},
			dbExpectFunc: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(regexp.QuoteMeta(
					"SELECT * FROM `user` WHERE (`user`.`email` = ?) LIMIT 1;",
				)).WithArgs(
					"john.doe@naver.com",
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
				).WillReturnRows(newUserRows([]*model.User{
					{UserID: 1, Email: "john.doe@naver.com", IsEmailVerified: true, PhoneNumber: "+821012345678", PasswordHash: "P@ssw0rd_hash"},
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

			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fail()
			}
			if tc.dbExpectFunc != nil {
				tc.dbExpectFunc(mock)
			}
			defer func() {
				mock.ExpectClose()
				if err := db.Close(); err != nil {
					t.Error(err)
				}
			}()

			handler := SignIn(&testHasher{}, db)

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

func newUserRows(users []*model.User) *sqlmock.Rows {
	rows := sqlmock.NewRows([]string{
		model.UserColumns.UserID,
		model.UserColumns.Name,
		model.UserColumns.Email,
		model.UserColumns.IsEmailVerified,
		model.UserColumns.PhoneNumber,
		model.UserColumns.Nickname,
		model.UserColumns.PasswordHash,
		model.UserColumns.CreatedAt,
		model.UserColumns.UpdatedAt,
	})
	for _, u := range users {
		rows.AddRow(
			u.UserID,
			u.Name,
			u.Email,
			u.IsEmailVerified,
			u.PhoneNumber,
			u.Nickname,
			u.PasswordHash,
			u.CreatedAt,
			u.UpdatedAt,
		)
	}
	return rows
}
