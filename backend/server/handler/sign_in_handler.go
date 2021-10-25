package handler

import (
	"context"
	"database/sql"
	"strings"

	"github.com/nyaruka/phonenumbers"
	"github.com/pkg/errors"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/sean-ahn/user/backend/crypto"
	"github.com/sean-ahn/user/backend/model"
	"github.com/sean-ahn/user/backend/server/service"
	userv1 "github.com/sean-ahn/user/proto/gen/go/user/v1"
)

type IDType int

const (
	IDTypeEmail IDType = iota + 1
	IDTypePhoneNumber
)

const (
	signInFailureMessage = "id or password incorrect"
)

type SignInHandlerFunc func(ctx context.Context, req *userv1.SignInRequest) (*userv1.SignInResponse, error)

func SignIn(hasher crypto.Hasher, db *sql.DB, userTokenService service.UserTokenService) SignInHandlerFunc {
	return func(ctx context.Context, req *userv1.SignInRequest) (*userv1.SignInResponse, error) {
		if req.Id == "" {
			return nil, status.Error(codes.InvalidArgument, "no id")
		}
		if req.Password == "" {
			return nil, status.Error(codes.InvalidArgument, "no password")
		}

		var (
			id       = req.Id
			findByID func(context.Context, boil.ContextExecutor, string) (*model.User, error)
		)
		switch detectIDType(req.Id) {
		case IDTypeEmail:
			findByID = findUserByEmail
			if !isValidEmail(req.Id) {
				return nil, status.Error(codes.InvalidArgument, "invalid id format")
			}
		case IDTypePhoneNumber:
			findByID = findUserByPhoneNumber
			norm, err := normalizePhoneNumber(id)
			if err != nil {
				return nil, status.Error(codes.InvalidArgument, "invalid id format")
			}
			id = norm
		default:
			return nil, status.Error(codes.InvalidArgument, "unknown id format")
		}
		if findByID == nil {
			return nil, status.Error(codes.Unknown, "should not be here")
		}

		user, err := findByID(ctx, db, id)
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, status.Error(codes.Unauthenticated, signInFailureMessage) // for security reason
		}
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		passwordHash, err := hasher.Hash([]byte(req.Password))
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		if user.PasswordHash != string(passwordHash) {
			return nil, status.Error(codes.Unauthenticated, signInFailureMessage)
		}

		if !user.IsEmailVerified {
			return nil, status.Error(codes.Unauthenticated, "email not verified yet")
		}

		accessToken, refreshToken, err := userTokenService.Issue(ctx, user)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		return &userv1.SignInResponse{AccessToken: accessToken, RefreshToken: refreshToken}, nil // TODO: Add token
	}
}

func detectIDType(id string) IDType {
	if p, err := phonenumbers.Parse(id, "KR"); err == nil && phonenumbers.IsValidNumber(p) {
		return IDTypePhoneNumber
	}
	if strings.Contains(id, "@") { // TODO: Use regexp
		return IDTypeEmail
	}
	return 0
}

func findUserByEmail(ctx context.Context, exec boil.ContextExecutor, email string) (*model.User, error) {
	u, err := model.Users(model.UserWhere.Email.EQ(email)).One(ctx, exec)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return u, nil
}

func findUserByPhoneNumber(ctx context.Context, exec boil.ContextExecutor, phoneNumber string) (*model.User, error) {
	u, err := model.Users(model.UserWhere.PhoneNumber.EQ(phoneNumber)).One(ctx, exec)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return u, nil
}
