package handler

import (
	"context"
	"database/sql"
	"net/mail"
	"regexp"
	"unicode"

	"github.com/jonboulle/clockwork"
	"github.com/pkg/errors"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/sean-ahn/user/backend/crypto"
	"github.com/sean-ahn/user/backend/model"
	"github.com/sean-ahn/user/backend/persistence/mysql"
	userv1 "github.com/sean-ahn/user/proto/gen/go/user/v1"
)

var (
	regexpEmail = regexp.MustCompile(`^[a-z0-9._%+\-$]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
)

type RegisterHandlerFunc func(ctx context.Context, req *userv1.RegisterRequest) (*userv1.RegisterResponse, error)

func Register(clock clockwork.Clock, db *sql.DB, hasher crypto.Hasher) RegisterHandlerFunc {
	return func(ctx context.Context, req *userv1.RegisterRequest) (*userv1.RegisterResponse, error) {
		now := clock.Now()

		if err := validateRegisterRequest(req); err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
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

		if _, err := mysql.FindUserByPhoneNumber(ctx, db, phoneNumber); errors.Cause(err) != sql.ErrNoRows {
			if err == nil {
				return nil, status.Error(codes.InvalidArgument, "already used phone number")
			}
			return nil, status.Error(codes.Internal, err.Error())
		}

		if _, err := mysql.FindUserByEmail(ctx, db, req.Email); errors.Cause(err) != sql.ErrNoRows {
			if err == nil {
				return nil, status.Error(codes.InvalidArgument, "already used email")
			}
			return nil, status.Error(codes.Internal, err.Error())
		}

		nickname := req.Name
		if req.Nickname != nil {
			nickname = *req.Nickname
		}

		passwordHash, err := hasher.Hash([]byte(req.Password))
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		user := &model.User{
			Name:             req.Name,
			Email:            req.Email,
			IsEmailConfirmed: false,
			PhoneNumber:      phoneNumber,
			Nickname:         nickname,
			PasswordHash:     string(passwordHash),
		}
		if err := user.Insert(ctx, db, boil.Infer()); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		return &userv1.RegisterResponse{}, nil
	}
}

func validateRegisterRequest(req *userv1.RegisterRequest) error {
	if req.VerificationToken == "" {
		return errors.New("no verification_token")
	}
	if req.Name == "" {
		return errors.New("no name")
	}
	if req.Email == "" {
		return errors.New("no email")
	}
	if !isValidEmail(req.Email) {
		return errors.New("invalid email")
	}
	if req.Password == "" {
		return errors.New("no password")
	}
	if !isValidPassword(req.Password) {
		return errors.New("invalid password")
	}
	if req.Nickname != nil && *req.Nickname == "" {
		return errors.New("empty nickname")
	}
	return nil
}

func isValidEmail(email string) bool {
	if _, err := mail.ParseAddress(email); err != nil {
		return false
	}
	if !regexpEmail.MatchString(email) {
		return false
	}
	return true
}

func isValidPassword(pw string) bool {
	const (
		minLen, maxLen = 8, 20
	)

	rr := []rune(pw)
	if len(rr) < minLen || len(rr) > maxLen {
		return false
	}

	var hasAlpha, hasDigit, hasAllowedSpecial bool
	for _, r := range rr {
		switch {
		case isAlpha(r):
			hasAlpha = true
		case isAllowedSpecial(r):
			hasAllowedSpecial = true
		case unicode.IsDigit(r):
			hasDigit = true
		default:
			return false
		}
	}

	return hasAlpha && hasDigit && hasAllowedSpecial
}

func isAlpha(c rune) bool {
	return ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z')
}

// isAllowedSpecial returns whether c is valid special character for password.
// See https://en.wikipedia.org/wiki/List_of_Special_Characters_for_Passwords
func isAllowedSpecial(c rune) bool {
	return unicode.IsSymbol(c) || unicode.IsPunct(c)
}
