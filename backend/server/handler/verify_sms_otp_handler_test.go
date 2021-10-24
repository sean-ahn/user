package handler

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/go-cmp/cmp"
	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"

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
		},
		{
			name: "otp code mismatch",
		},
		{
			name: "already verified",
		},
		{
			name: "verification expired",
		},
		{
			name: "verification max trials exceeded",
		},
		{
			name: "verification not found",
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
