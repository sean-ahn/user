package service

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/assert"

	"github.com/sean-ahn/user/backend/model"
)

func TestUserJWTTokenService_Issue(t *testing.T) {
	now := time.Date(2021, 10, 24, 7, 39, 46, 127956672, time.UTC)

	user := &model.User{UserID: 1, PasswordHash: "password_hash"}

	mockSecret := "Zwl61lLdI5fAWlSD9AK1wwjb44W6PjVZUFgPf++pvmo="

	cases := []struct {
		name string

		dbExpectFunc func(sqlmock.Sqlmock)
	}{
		{
			name: "secret exists",

			dbExpectFunc: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(
					"SELECT * FROM `jwt_audience_secret` WHERE (`jwt_audience_secret`.`audience` = ?) LIMIT 1;",
				)).WithArgs(
					"user:1",
				).WillReturnRows(newJWTAudienceSecretRows([]*model.JWTAudienceSecret{
					{Audience: "user:1", Secret: mockSecret},
				}))
			},
		},
		{
			name: "secret not exists",

			dbExpectFunc: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(
					"SELECT * FROM `jwt_audience_secret` WHERE (`jwt_audience_secret`.`audience` = ?) LIMIT 1;",
				)).WithArgs(
					"user:1",
				).WillReturnError(
					sql.ErrNoRows,
				)

				mock.ExpectExec(regexp.QuoteMeta(
					"INSERT INTO `jwt_audience_secret` (`audience`,`secret`) VALUES (?,?)",
				)).WithArgs(
					"user:1", mockSecret,
				).WillReturnResult(
					sqlmock.NewResult(2, 1),
				)

				mock.ExpectQuery(regexp.QuoteMeta(
					"SELECT `jwt_audience_secret_id`,`created_at`,`updated_at` FROM `jwt_audience_secret` WHERE `jwt_audience_secret_id`=?",
				)).WithArgs(
					2,
				).WillReturnRows(sqlmock.NewRows([]string{"jwt_audience_secret_id", "created_at", "updated_at"}).
					AddRow(2, now, now),
				)
			},
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fail()
			}
			if tc.dbExpectFunc != nil {
				tc.dbExpectFunc(mock)
			}
			defer func() {
				mock.ExpectClose()
				if err := db.Close(); err != nil {
					t.Error(err)
				}
			}()

			svc := UserJWTTokenService{
				clock:                 clockwork.NewFakeClockAt(now),
				db:                    db,
				accessTokenExpiresIn:  10 * time.Second,
				refreshTokenExpiresIn: 14 * 24 * time.Hour,
			}

			accessToken, refreshToken, err := svc.Issue(ctx, user)

			assert.NoError(t, err)
			assert.NotEmpty(t, accessToken)
			assert.NotEmpty(t, refreshToken)
		})
	}
}

func newJWTAudienceSecretRows(secrets []*model.JWTAudienceSecret) *sqlmock.Rows {
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
