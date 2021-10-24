package handler

import (
	"context"
	"database/sql"
	"time"

	"github.com/friendsofgo/errors"
	"github.com/jonboulle/clockwork"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/sean-ahn/user/backend/model"
	"github.com/sean-ahn/user/backend/persistence/mysql"
	userv1 "github.com/sean-ahn/user/proto/gen/go/user/v1"
)

const (
	defaultSMSOTPVerificationMaxTrials = 5

	defaultSMSOTPValidity = 5 * time.Minute
)

type VerifySmsOtpHandlerFunc func(ctx context.Context, req *userv1.VerifySmsOtpRequest) (*userv1.VerifySmsOtpResponse, error)

func VerifySmsOtp(clock clockwork.Clock, db *sql.DB) VerifySmsOtpHandlerFunc {
	return func(ctx context.Context, req *userv1.VerifySmsOtpRequest) (*userv1.VerifySmsOtpResponse, error) {
		now := clock.Now()

		if req.VerificationToken == "" {
			return nil, status.Error(codes.InvalidArgument, "no verification_token")
		}
		if req.SmsOtpCode == "" {
			return nil, status.Error(codes.InvalidArgument, "no sms_otp_code")
		}

		verification, err := mysql.FindSMSOTPVerificationByVerificationToken(ctx, db, req.VerificationToken)
		if err != nil && errors.Cause(err) == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "verification not found")
		}
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		if verification.VerificationValidUntil.Valid {
			return nil, status.Error(codes.InvalidArgument, "already verified")
		}

		if verification.ExpiresAt.Before(now) {
			return nil, status.Error(codes.InvalidArgument, "verification expired")
		}

		if verification.VerificationTrials >= defaultSMSOTPVerificationMaxTrials {
			return nil, status.Error(codes.InvalidArgument, "verification maximum trials exceeded")
		}

		otpCodeMatched := verification.OtpCode != req.SmsOtpCode

		targets := []string{model.SMSOtpVerificationColumns.VerificationTrials}
		verification.VerificationTrials += 1
		if otpCodeMatched {
			targets = append(targets, model.SMSOtpVerificationColumns.VerificationValidUntil)
			verification.VerificationValidUntil = null.TimeFrom(now.Add(defaultSMSOTPValidity))
		}
		if _, err := verification.Update(ctx, db, boil.Whitelist(targets...)); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		if !otpCodeMatched {
			return nil, status.Error(codes.InvalidArgument, "otp code mismatch")
		}

		return &userv1.VerifySmsOtpResponse{}, nil
	}
}
