package handler

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/go-cmp/cmp"
	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/assert"
	"github.com/volatiletech/null/v8"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"

	"github.com/sean-ahn/user/backend/model"
	"github.com/sean-ahn/user/backend/test"
	userv1 "github.com/sean-ahn/user/proto/gen/go/user/v1"
)

func TestRegister(t *testing.T) {
	now := time.Date(2021, 10, 25, 9, 40, 47, 42395845, time.UTC)

	cases := []struct {
		name string
		req  *userv1.RegisterRequest

		dbExpectFunc func(sqlmock.Sqlmock)

		expectedCode codes.Code
		expectedResp *userv1.RegisterResponse
		expectedErr  string
	}{
		{
			name: "success",
			req: &userv1.RegisterRequest{
				VerificationToken: "verification_token",
				Name:              "name",
				Email:             "john.doe@example.com",
				Password:          "P@ssw0rd",
				Nickname:          proto.String("nickname"),
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
				).WillReturnError(
					sql.ErrNoRows,
				)

				mock.ExpectQuery(regexp.QuoteMeta(
					"SELECT * FROM `user` WHERE (`user`.`email` = ?) LIMIT 1;",
				)).WithArgs(
					"john.doe@example.com",
				).WillReturnError(
					sql.ErrNoRows,
				)

				mock.ExpectExec(regexp.QuoteMeta(
					"INSERT INTO `user` (`name`,`email`,`is_email_confirmed`,`phone_number`,`nickname`,`password_hash`) VALUES (?,?,?,?,?,?)",
				)).WithArgs(
					"name", "john.doe@example.com", false, "+821012345678", "nickname", "P@ssw0rd_hash",
				).WillReturnResult(
					sqlmock.NewResult(2, 1),
				)

				mock.ExpectQuery(regexp.QuoteMeta(
					"SELECT `user_id`,`created_at`,`updated_at` FROM `user` WHERE `user_id`=?",
				)).WithArgs(
					2,
				).WillReturnRows(sqlmock.NewRows([]string{"user_id", "created_at", "updated_at"}).AddRow(
					2, now, now,
				))
			},
			expectedCode: codes.OK,
			expectedResp: &userv1.RegisterResponse{},
		},
		{
			name: "fail even if email not confirmed",
			req: &userv1.RegisterRequest{
				VerificationToken: "verification_token",
				Name:              "name",
				Email:             "john.doe@example.com",
				Password:          "P@ssw0rd",
				Nickname:          proto.String("nickname"),
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
				).WillReturnError(
					sql.ErrNoRows,
				)

				mock.ExpectQuery(regexp.QuoteMeta(
					"SELECT * FROM `user` WHERE (`user`.`email` = ?) LIMIT 1;",
				)).WithArgs(
					"john.doe@example.com",
				).WillReturnRows(test.NewUserRows([]*model.User{
					{UserID: 2, Email: "john.doe@example.com", IsEmailConfirmed: false},
				}))
			},
			expectedCode: codes.InvalidArgument,
			expectedErr:  "rpc error: code = InvalidArgument desc = already used email",
		},
		{
			name: "verification not found",
			req: &userv1.RegisterRequest{
				VerificationToken: "verification_token",
				Name:              "name",
				Email:             "john.doe@example.com",
				Password:          "P@ssw0rd",
				Nickname:          proto.String("nickname"),
			},
			dbExpectFunc: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(
					"SELECT * FROM `sms_otp_verification` WHERE (`sms_otp_verification`.`verification_token` = ?) LIMIT 1;",
				)).WithArgs(
					"verification_token",
				).WillReturnError(
					sql.ErrNoRows,
				)
			},
			expectedCode: codes.InvalidArgument,
			expectedErr:  "rpc error: code = InvalidArgument desc = verification not found",
		},
		{
			name: "invalid verification",
			req: &userv1.RegisterRequest{
				VerificationToken: "verification_token",
				Name:              "name",
				Email:             "john.doe@example.com",
				Password:          "P@ssw0rd",
				Nickname:          proto.String("nickname"),
			},
			dbExpectFunc: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(
					"SELECT * FROM `sms_otp_verification` WHERE (`sms_otp_verification`.`verification_token` = ?) LIMIT 1;",
				)).WithArgs(
					"verification_token",
				).WillReturnRows(test.NewSMSOtpVerificationRows([]*model.SMSOtpVerification{
					{PhoneNumber: "+821012345678", VerificationValidUntil: null.TimeFrom(now.Add(-1 * time.Minute))},
				}))
			},
			expectedCode: codes.InvalidArgument,
			expectedErr:  "rpc error: code = InvalidArgument desc = invalid verification",
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			clock := clockwork.NewFakeClockAt(now)

			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fail()
			}
			if tc.dbExpectFunc != nil {
				tc.dbExpectFunc(mock)
			}
			defer test.CloseSqlmock(t, db, mock)

			handler := Register(clock, db, &testHasher{})

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

func Test_isValidPassword(t *testing.T) {
	cases := []struct {
		given    string
		expected bool
	}{
		{given: "p@assw0rd", expected: true},
		{given: "P@assw0rd", expected: true},
		{given: "p@assw rd", expected: false},
		{given: "qwerty123", expected: false},
		{given: "비밀번호486!", expected: false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.given, func(t *testing.T) {
			got := isValidPassword(tc.given)

			assert.Equal(t, tc.expected, got)
		})
	}
}
