package test

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/sean-ahn/user/backend/model"
)

func CloseSqlmock(t *testing.T, db *sql.DB, mock sqlmock.Sqlmock) {
	mock.ExpectClose()
	if err := db.Close(); err != nil {
		t.Error(err)
	}
}

func NewUserRows(users []*model.User) *sqlmock.Rows {
	rows := sqlmock.NewRows([]string{
		model.UserColumns.UserID,
		model.UserColumns.Name,
		model.UserColumns.Email,
		model.UserColumns.IsEmailVerified,
		model.UserColumns.PhoneNumber,
		model.UserColumns.Nickname,
		model.UserColumns.PasswordHash,
		model.UserColumns.CreatedAt,
		model.UserColumns.UpdatedAt,
	})
	for _, u := range users {
		rows.AddRow(
			u.UserID,
			u.Name,
			u.Email,
			u.IsEmailVerified,
			u.PhoneNumber,
			u.Nickname,
			u.PasswordHash,
			u.CreatedAt,
			u.UpdatedAt,
		)
	}
	return rows
}

func NewJWTAudienceSecretRows(secrets []*model.JWTAudienceSecret) *sqlmock.Rows {
	rows := sqlmock.NewRows([]string{
		model.JWTAudienceSecretColumns.JWTAudienceSecretID,
		model.JWTAudienceSecretColumns.Audience,
		model.JWTAudienceSecretColumns.Secret,
		model.JWTAudienceSecretColumns.CreatedAt,
		model.JWTAudienceSecretColumns.UpdatedAt,
	})
	for _, s := range secrets {
		rows.AddRow(
			s.JWTAudienceSecretID,
			s.Audience,
			s.Secret,
			s.CreatedAt,
			s.UpdatedAt,
		)
	}
	return rows
}

func NewJWTDenylistRows(denylists []*model.JWTDenylist) *sqlmock.Rows {
	rows := sqlmock.NewRows([]string{
		model.JWTDenylistColumns.JWTDenylistID,
		model.JWTDenylistColumns.UserID,
		model.JWTDenylistColumns.Jti,
		model.JWTDenylistColumns.CreatedAt,
		model.JWTDenylistColumns.UpdatedAt,
	})
	for _, d := range denylists {
		rows.AddRow(
			d.JWTDenylistID,
			d.UserID,
			d.Jti,
			d.CreatedAt,
			d.UpdatedAt,
		)
	}
	return rows
}
