package handler

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jonboulle/clockwork"
	"github.com/nyaruka/phonenumbers"
	"github.com/pkg/errors"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/sean-ahn/user/backend/model"
	"github.com/sean-ahn/user/backend/server/generator"
	smsv1 "github.com/sean-ahn/user/proto/gen/go/sms/v1"
	userv1 "github.com/sean-ahn/user/proto/gen/go/user/v1"
)

const (
	defaultSMSOTPExpiration = 3 * time.Minute
)

type RequestSmsOtpHandlerFunc func(ctx context.Context, req *userv1.RequestSmsOtpRequest) (*userv1.RequestSmsOtpResponse, error)

func RequestSmsOtp(clock clockwork.Clock, db *sql.DB, idGenerator, otpGenerator generator.Generator, smsv1ServiceCli smsv1.SmsServiceClient) RequestSmsOtpHandlerFunc {
	return func(ctx context.Context, req *userv1.RequestSmsOtpRequest) (*userv1.RequestSmsOtpResponse, error) {
		if req.PhoneNumber == "" {
			return nil, status.Error(codes.InvalidArgument, "no phone_number")
		}

		normalizedPhoneNumber, err := normalizePhoneNumber(req.PhoneNumber)
		if err != nil || normalizedPhoneNumber != req.PhoneNumber {
			return nil, status.Error(codes.InvalidArgument, "invalid phone_number")
		}

		verification := model.SMSOtpVerification{
			VerificationToken: idGenerator.Generate(),
			PhoneNumber:       normalizedPhoneNumber,
			OtpCode:           otpGenerator.Generate(),
			ExpiresAt:         clock.Now().Add(defaultSMSOTPExpiration),
		}
		if err := verification.Insert(ctx, db, boil.Infer()); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		if err := sendOTPVerificationSMS(ctx, smsv1ServiceCli, normalizedPhoneNumber, verification.OtpCode); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		return &userv1.RequestSmsOtpResponse{
			VerificationToken: verification.VerificationToken,
			ExpiresInMs:       int32(defaultSMSOTPExpiration.Milliseconds()), // safe
		}, nil
	}
}

func normalizePhoneNumber(s string) (string, error) {
	p, err := phonenumbers.Parse(s, "KR")
	if err != nil {
		return "", errors.WithStack(err)
	}
	return phonenumbers.Format(p, phonenumbers.E164), nil
}

func sendOTPVerificationSMS(ctx context.Context, cli smsv1.SmsServiceClient, phoneNumber, code string) error {
	_, err := cli.Send(ctx, &smsv1.SendRequest{
		To:      phoneNumber,
		Message: fmt.Sprintf("%s is your authentication code.", code),
	})
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}
