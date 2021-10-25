package handler

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/jonboulle/clockwork"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"

	"github.com/sean-ahn/user/backend/client"
	"github.com/sean-ahn/user/backend/test"
	smsv1 "github.com/sean-ahn/user/proto/gen/go/sms/v1"
	userv1 "github.com/sean-ahn/user/proto/gen/go/user/v1"
)

type testGenerator struct {
	mock string
}

func (g *testGenerator) Generate() string {
	return g.mock
}

func TestRequestSmsOtp(t *testing.T) {
	now := time.Date(2021, 10, 24, 17, 18, 16, 304850171, time.UTC)

	cases := []struct {
		name string
		req  *userv1.RequestSmsOtpRequest

		dbExpectFunc                   func(sqlmock.Sqlmock)
		mockSmsServiceClientExpectFunc func(context.Context) func(*client.MockSmsServiceClient)

		expectedCode codes.Code
		expectedResp *userv1.RequestSmsOtpResponse
		expectedErr  string
	}{
		{
			name: "success",
			req:  &userv1.RequestSmsOtpRequest{PhoneNumber: "+821012345678"},
			dbExpectFunc: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(regexp.QuoteMeta(
					"INSERT INTO `sms_otp_verification` (`verification_token`,`phone_number`,`otp_code`,`expires_at`,`verification_trials`,`verification_valid_until`) VALUES (?,?,?,?,?,?)",
				)).WithArgs(
					"verification_token", "+821012345678", "123456", now.Add(defaultSMSOTPExpiration), 0, nil,
				).WillReturnResult(
					sqlmock.NewResult(1, 1),
				)

				mock.ExpectQuery(regexp.QuoteMeta(
					"SELECT `sms_otp_verification_id`,`created_at`,`updated_at` FROM `sms_otp_verification` WHERE `sms_otp_verification_id`=?",
				)).WithArgs(
					1,
				).WillReturnRows(sqlmock.NewRows([]string{"sms_otp_verification_id", "created_at", "updated_at"}).AddRow(
					1, now, now,
				))
			},
			mockSmsServiceClientExpectFunc: func(ctx context.Context) func(*client.MockSmsServiceClient) {
				return func(mock *client.MockSmsServiceClient) {
					mock.EXPECT().
						Send(ctx, &smsv1.SendRequest{
							To:      "+821012345678",
							Message: "123456 is your authentication code.",
						}).
						Return(
							&smsv1.SendResponse{}, nil,
						)
				}
			},
			expectedCode: codes.OK,
			expectedResp: &userv1.RequestSmsOtpResponse{
				VerificationToken: "verification_token",
				ExpiresInMs:       180000,
			},
		},
		{
			name: "fail to sms send",
			req:  &userv1.RequestSmsOtpRequest{PhoneNumber: "+821012345678"},
			dbExpectFunc: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(regexp.QuoteMeta(
					"INSERT INTO `sms_otp_verification` (`verification_token`,`phone_number`,`otp_code`,`expires_at`,`verification_trials`,`verification_valid_until`) VALUES (?,?,?,?,?,?)",
				)).WithArgs(
					"verification_token", "+821012345678", "123456", now.Add(defaultSMSOTPExpiration), 0, nil,
				).WillReturnResult(
					sqlmock.NewResult(1, 1),
				)

				mock.ExpectQuery(regexp.QuoteMeta(
					"SELECT `sms_otp_verification_id`,`created_at`,`updated_at` FROM `sms_otp_verification` WHERE `sms_otp_verification_id`=?",
				)).WithArgs(
					1,
				).WillReturnRows(sqlmock.NewRows([]string{"sms_otp_verification_id", "created_at", "updated_at"}).AddRow(
					1, now, now,
				))
			},
			mockSmsServiceClientExpectFunc: func(ctx context.Context) func(*client.MockSmsServiceClient) {
				return func(mock *client.MockSmsServiceClient) {
					mock.EXPECT().
						Send(ctx, &smsv1.SendRequest{
							To:      "+821012345678",
							Message: "123456 is your authentication code.",
						}).
						Return(
							nil, errors.New("unexpected error"),
						)
				}
			},
			expectedCode: codes.Internal,
			expectedErr:  "rpc error: code = Internal desc = unexpected error",
		},
		{
			name:         "invalid E.167 phone number",
			req:          &userv1.RequestSmsOtpRequest{PhoneNumber: "821012345678"},
			expectedCode: codes.InvalidArgument,
			expectedErr:  "rpc error: code = InvalidArgument desc = invalid phone_number",
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			ctrl := gomock.NewController(t)

			clock := clockwork.NewFakeClockAt(now)

			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fail()
			}
			if tc.dbExpectFunc != nil {
				tc.dbExpectFunc(mock)
			}
			defer test.CloseSqlmock(t, db, mock)

			mockSmsServiceCli := client.NewMockSmsServiceClient(ctrl)
			if tc.mockSmsServiceClientExpectFunc != nil {
				tc.mockSmsServiceClientExpectFunc(ctx)(mockSmsServiceCli)
			}

			handler := RequestSmsOtp(clock, db, &testGenerator{mock: "verification_token"}, &testGenerator{mock: "123456"}, mockSmsServiceCli)

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
