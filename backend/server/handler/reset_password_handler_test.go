package handler

import (
	"context"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/assert"
	"github.com/volatiletech/null/v8"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"

	"github.com/sean-ahn/user/backend/model"
	"github.com/sean-ahn/user/backend/server/service"
	"github.com/sean-ahn/user/backend/test"
	userv1 "github.com/sean-ahn/user/proto/gen/go/user/v1"
)

type userMatcher struct{ id int }

func (m *userMatcher) Matches(x interface{}) bool {
	if u, ok := x.(*model.User); ok {
		return u.UserID == m.id
	}
	return false
}

func (m *userMatcher) String() string {
	return fmt.Sprintf("is user which has id: %d", m.id)
}

var _ gomock.Matcher = (*userMatcher)(nil)

func TestResetPassword(t *testing.T) {
	now := time.Date(2021, 10, 25, 16, 37, 55, 509012743, time.UTC)

	cases := []struct {
		name string
		req  *userv1.ResetPasswordRequest

		dbExpectFunc               func(sqlmock.Sqlmock)
		userTokenServiceExpectFunc func(context.Context) func(*service.MockUserTokenService)

		expectedCode codes.Code
		expectedResp *userv1.ResetPasswordResponse
		expectedErr  string
	}{
		{
			name: "success",
			req: &userv1.ResetPasswordRequest{
				VerificationToken: "verification_token",
				NewPassword:       "P@$$w0rd",
			},
			dbExpectFunc: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(
					"SELECT * FROM `sms_otp_verification` WHERE (`sms_otp_verification`.`verification_token` = ?) LIMIT 1;",
				)).WithArgs(
					"verification_token",
				).WillReturnRows(test.NewSMSOtpVerificationRows([]*model.SMSOtpVerification{
					{PhoneNumber: "+821012345678", VerificationValidUntil: null.TimeFrom(now.Add(defaultSMSOTPValidity))},
				}))

				mock.ExpectQuery(regexp.QuoteMeta(
					"SELECT * FROM `user` WHERE (`user`.`phone_number` = ?) LIMIT 1;",
				)).WithArgs(
					"+821012345678",
				).WillReturnRows(test.NewUserRows([]*model.User{
					{UserID: 1},
				}))

				mock.ExpectBegin()

				mock.ExpectExec(regexp.QuoteMeta(
					"UPDATE `user` SET `password_hash`=? WHERE `user_id`=?",
				)).WithArgs(
					"P@$$w0rd_hash", 1,
				).WillReturnResult(sqlmock.NewResult(0, 1))

				mock.ExpectCommit()
			},
			userTokenServiceExpectFunc: func(ctx context.Context) func(*service.MockUserTokenService) {
				return func(mock *service.MockUserTokenService) {
					mock.EXPECT().RevokeAll(ctx, &userMatcher{id: 1}, gomock.Any()).Return(nil)
				}
			},
			expectedCode: codes.OK,
			expectedResp: &userv1.ResetPasswordResponse{},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			clock := clockwork.NewFakeClockAt(now)

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

			handler := ResetPassword(clock, db, &testHasher{}, mockUserTokenService)

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
