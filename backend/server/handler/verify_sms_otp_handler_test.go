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
	"google.golang.org/protobuf/testing/protocmp"

	"github.com/sean-ahn/user/backend/model"
	"github.com/sean-ahn/user/backend/test"
	userv1 "github.com/sean-ahn/user/proto/gen/go/user/v1"
)

func TestVerifySmsOtp(t *testing.T) {
	now := time.Date(2021, 10, 24, 17, 18, 16, 304850171, time.UTC)

	cases := []struct {
		name string
		req  *userv1.VerifySmsOtpRequest

		dbExpectFunc func(sqlmock.Sqlmock)

		expectedCode codes.Code
		expectedResp *userv1.VerifySmsOtpResponse
		expectedErr  string
	}{
		{
			name: "success",
			req: &userv1.VerifySmsOtpRequest{
				VerificationToken: "verification_token",
				SmsOtpCode:        "123456",
			},
			dbExpectFunc: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(
					"SELECT * FROM `sms_otp_verification` WHERE (`sms_otp_verification`.`verification_token` = ?) LIMIT 1;",
				)).WithArgs(
					"verification_token",
				).WillReturnRows(test.NewSMSOtpVerificationRows([]*model.SMSOtpVerification{
					{SMSOtpVerificationID: 2, VerificationToken: "verification_token", ExpiresAt: now.Add(defaultSMSOTPExpiration), OtpCode: "123456"},
				}))

				mock.ExpectExec(regexp.QuoteMeta(
					"UPDATE `sms_otp_verification` SET `verification_trials`=?,`verification_valid_until`=? WHERE `sms_otp_verification_id`=?",
				)).WithArgs(
					1, now.Add(defaultSMSOTPValidity), 2,
				).WillReturnResult(
					sqlmock.NewResult(0, 1),
				)
			},
			expectedCode: codes.OK,
			expectedResp: &userv1.VerifySmsOtpResponse{},
		},
		{
			name: "otp code mismatch",
			req: &userv1.VerifySmsOtpRequest{
				VerificationToken: "verification_token",
				SmsOtpCode:        "000000",
			},
			dbExpectFunc: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(
					"SELECT * FROM `sms_otp_verification` WHERE (`sms_otp_verification`.`verification_token` = ?) LIMIT 1;",
				)).WithArgs(
					"verification_token",
				).WillReturnRows(test.NewSMSOtpVerificationRows([]*model.SMSOtpVerification{
					{SMSOtpVerificationID: 2, VerificationToken: "verification_token", ExpiresAt: now.Add(defaultSMSOTPExpiration), OtpCode: "123456", VerificationTrials: 2},
				}))

				mock.ExpectExec(regexp.QuoteMeta(
					"UPDATE `sms_otp_verification` SET `verification_trials`=? WHERE `sms_otp_verification_id`=?",
				)).WithArgs(
					3, 2,
				).WillReturnResult(
					sqlmock.NewResult(0, 1),
				)
			},
			expectedCode: codes.InvalidArgument,
			expectedErr:  "rpc error: code = InvalidArgument desc = otp code mismatch",
		},
		{
			name: "already verified",
			req: &userv1.VerifySmsOtpRequest{
				VerificationToken: "verification_token",
				SmsOtpCode:        "000000",
			},
			dbExpectFunc: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(
					"SELECT * FROM `sms_otp_verification` WHERE (`sms_otp_verification`.`verification_token` = ?) LIMIT 1;",
				)).WithArgs(
					"verification_token",
				).WillReturnRows(test.NewSMSOtpVerificationRows([]*model.SMSOtpVerification{
					{VerificationToken: "verification_token", ExpiresAt: now.Add(defaultSMSOTPExpiration), VerificationValidUntil: null.TimeFrom(now.Add(defaultSMSOTPValidity + 1*time.Minute))},
				}))
			},
			expectedCode: codes.InvalidArgument,
			expectedErr:  "rpc error: code = InvalidArgument desc = already verified",
		},
		{
			name: "verification expired",
			req: &userv1.VerifySmsOtpRequest{
				VerificationToken: "verification_token",
				SmsOtpCode:        "000000",
			},
			dbExpectFunc: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(
					"SELECT * FROM `sms_otp_verification` WHERE (`sms_otp_verification`.`verification_token` = ?) LIMIT 1;",
				)).WithArgs(
					"verification_token",
				).WillReturnRows(test.NewSMSOtpVerificationRows([]*model.SMSOtpVerification{
					{VerificationToken: "verification_token", ExpiresAt: now.Add(-1 * time.Minute)},
				}))
			},
			expectedCode: codes.InvalidArgument,
			expectedErr:  "rpc error: code = InvalidArgument desc = verification expired",
		},
		{
			name: "verification max trials exceeded",
			req: &userv1.VerifySmsOtpRequest{
				VerificationToken: "verification_token",
				SmsOtpCode:        "123456",
			},
			dbExpectFunc: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(
					"SELECT * FROM `sms_otp_verification` WHERE (`sms_otp_verification`.`verification_token` = ?) LIMIT 1;",
				)).WithArgs(
					"verification_token",
				).WillReturnRows(test.NewSMSOtpVerificationRows([]*model.SMSOtpVerification{
					{VerificationToken: "verification_token", ExpiresAt: now.Add(defaultSMSOTPExpiration), VerificationTrials: defaultSMSOTPVerificationMaxTrials, OtpCode: "123456"},
				}))
			},
			expectedCode: codes.InvalidArgument,
			expectedErr:  "rpc error: code = InvalidArgument desc = verification maximum trials exceeded",
		},
		{
			name: "verification not found",
			req: &userv1.VerifySmsOtpRequest{
				VerificationToken: "verification_token",
				SmsOtpCode:        "123456",
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
			expectedCode: codes.NotFound,
			expectedErr:  "rpc error: code = NotFound desc = verification not found",
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

			handler := VerifySmsOtp(clock, db)

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
