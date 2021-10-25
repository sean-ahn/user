package mysql

import (
	"context"

	"github.com/friendsofgo/errors"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/sean-ahn/user/backend/model"
)

func GetUser(ctx context.Context, exec boil.ContextExecutor, id int) (*model.User, error) {
	u, err := model.Users(model.UserWhere.UserID.EQ(id)).One(ctx, exec)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return u, nil
}

func FindUserByEmail(ctx context.Context, exec boil.ContextExecutor, email string) (*model.User, error) {
	u, err := model.Users(model.UserWhere.Email.EQ(email)).One(ctx, exec)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return u, nil
}

func FindUserByPhoneNumber(ctx context.Context, exec boil.ContextExecutor, phoneNumber string) (*model.User, error) {
	u, err := model.Users(model.UserWhere.PhoneNumber.EQ(phoneNumber)).One(ctx, exec)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return u, nil
}

func FindJWTDenylistByJTI(ctx context.Context, exec boil.ContextExecutor, jti string) (*model.JWTDenylist, error) {
	d, err := model.JWTDenylists(model.JWTDenylistWhere.Jti.EQ(jti)).One(ctx, exec)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return d, nil
}

func FindSMSOTPVerificationByVerificationToken(ctx context.Context, exec boil.ContextExecutor, token string) (*model.SMSOtpVerification, error) {
	v, err := model.SMSOtpVerifications(model.SMSOtpVerificationWhere.VerificationToken.EQ(token)).One(ctx, exec)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return v, nil
}
