package handler

import (
	"context"
	"database/sql"

	"github.com/jonboulle/clockwork"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/sean-ahn/user/backend/crypto"
	"github.com/sean-ahn/user/backend/model"
	"github.com/sean-ahn/user/backend/persistence/mysql"
	"github.com/sean-ahn/user/backend/server/service"
	userv1 "github.com/sean-ahn/user/proto/gen/go/user/v1"
)

type ResetPasswordHandlerFunc func(ctx context.Context, req *userv1.ResetPasswordRequest) (*userv1.ResetPasswordResponse, error)

func ResetPassword(clock clockwork.Clock, db *sql.DB, hasher crypto.Hasher, userTokenService service.UserTokenService) ResetPasswordHandlerFunc {
	return func(ctx context.Context, req *userv1.ResetPasswordRequest) (*userv1.ResetPasswordResponse, error) {
		now := clock.Now()

		if req.VerificationToken == "" {
			return nil, status.Error(codes.InvalidArgument, "no verification_token")
		}
		if req.NewPassword == "" {
			return nil, status.Error(codes.InvalidArgument, "no new_password")
		}
		if !isValidPassword(req.NewPassword) {
			return nil, status.Error(codes.InvalidArgument, "invalid password")
		}

		newPasswordHash, err := hasher.Hash([]byte(req.NewPassword))
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		verification, err := mysql.FindSMSOTPVerificationByVerificationToken(ctx, db, req.VerificationToken)
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, status.Error(codes.InvalidArgument, "verification not found")
		}
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		phoneNumber, err := normalizePhoneNumber(verification.PhoneNumber)
		if err != nil {
			return nil, status.Error(codes.Internal, "invalid phone_number")
		}

		if !verification.VerificationValidUntil.Valid || verification.VerificationValidUntil.Time.Before(now) {
			return nil, status.Error(codes.InvalidArgument, "invalid verification")
		}

		user, err := mysql.FindUserByPhoneNumber(ctx, db, phoneNumber)
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "no user with given phone number")
		}
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		user.PasswordHash = string(newPasswordHash)

		if err := resetUserPassword(ctx, db, userTokenService, user); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		return &userv1.ResetPasswordResponse{}, nil
	}
}

func resetUserPassword(ctx context.Context, db *sql.DB, userTokenService service.UserTokenService, user *model.User) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return errors.WithStack(err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			logrus.WithError(err).Error()
		}
	}()

	if _, err := user.Update(ctx, tx, boil.Whitelist(model.UserColumns.PasswordHash)); err != nil {
		return errors.WithStack(err)
	}

	if err := userTokenService.RevokeAll(ctx, user, tx); err != nil {
		return errors.WithStack(err)
	}

	if err := tx.Commit(); err != nil {
		return errors.WithStack(err)
	}
	return nil
}
