package handler

import (
	"context"
	"database/sql"

	"github.com/volatiletech/sqlboiler/v4/boil"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/sean-ahn/user/backend/model"
	"github.com/sean-ahn/user/backend/persistence/mysql"
	userv1 "github.com/sean-ahn/user/proto/gen/go/user/v1"
)

type ConfirmEmailHandlerFunc func(ctx context.Context, req *userv1.ConfirmEmailRequest) (*userv1.ConfirmEmailResponse, error)

func ConfirmEmail(db *sql.DB) ConfirmEmailHandlerFunc {
	return func(ctx context.Context, req *userv1.ConfirmEmailRequest) (*userv1.ConfirmEmailResponse, error) {
		if req.ConfirmationToken == "" {
			return nil, status.Error(codes.InvalidArgument, "no confirmation_code")
		}

		user, err := mysql.FindUserByEmail(ctx, db, req.ConfirmationToken)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid confirmation code")
		}

		if user.IsEmailVerified {
			return nil, status.Error(codes.InvalidArgument, "already confirmed email")
		}

		user.IsEmailVerified = true
		if _, err := user.Update(ctx, db, boil.Whitelist(model.UserColumns.IsEmailVerified)); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		return &userv1.ConfirmEmailResponse{}, nil
	}
}
