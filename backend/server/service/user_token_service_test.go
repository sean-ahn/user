package service

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	mysqldriver "github.com/go-sql-driver/mysql"
	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/assert"

	"github.com/sean-ahn/user/backend/model"
	"github.com/sean-ahn/user/backend/persistence/mysql"
	"github.com/sean-ahn/user/backend/server/generator"
	"github.com/sean-ahn/user/backend/test"
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
				).WillReturnRows(test.NewJWTAudienceSecretRows([]*model.JWTAudienceSecret{
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
			defer test.CloseSqlmock(t, db, mock)

			if tc.dbExpectFunc != nil {
				tc.dbExpectFunc(mock)
			}

			svc := UserJWTTokenService{
				clock:                 clockwork.NewFakeClockAt(now),
				db:                    db,
				idGenerator:           &generator.UUIDGenerator{},
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

func TestUserJWTTokenService_Refresh(t *testing.T) {
	now := time.Date(2021, 10, 24, 7, 39, 46, 127956672, time.UTC)

	cases := []struct {
		name string

		token string

		dbExpectFunc func(sqlmock.Sqlmock)

		expectedErr string
	}{
		{
			name:  "refresh even if secret has changed",
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJodHRwczovL2dpdGh1Yi5jb20vc2Vhbi1haG4vdXNlciIsImF1ZCI6WyJ1c2VyOjEiXSwiZXhwIjoxNjM2Mjk2ODYwLCJpYXQiOjE2MzUwODcyNjAsImp0aSI6ImQzOTE0MTZjLWMyZDItNDRkNS1iM2VjLTE0N2E3NzEzNjA2ZCIsInVzZXJfaWQiOiIxIn0.4KVLf9emjWK2GB9vXj1A-7KXHGq8HLnExUv54XoiNnU",
			dbExpectFunc: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(
					"SELECT * FROM `jwt_audience_secret` WHERE (`jwt_audience_secret`.`audience` = ?) LIMIT 1;",
				)).WithArgs(
					"user:1",
				).WillReturnRows(test.NewJWTAudienceSecretRows([]*model.JWTAudienceSecret{
					{Audience: "user:1", Secret: "DBetxLyZOcHw3gQ+ozOyg+c6N1j2xG2yPTSVRrnXsaE="},
				}))

				mock.ExpectQuery(regexp.QuoteMeta(
					"SELECT * FROM `jwt_denylist` WHERE (`jwt_denylist`.`jti` = ?) LIMIT 1;",
				)).WithArgs(
					"d391416c-c2d2-44d5-b3ec-147a7713606d",
				).WillReturnError(
					sql.ErrNoRows,
				)

				mock.ExpectQuery(regexp.QuoteMeta(
					"SELECT * FROM `user` WHERE (`user`.`user_id` = ?) LIMIT 1;",
				)).WithArgs(
					1,
				).WillReturnRows(test.NewUserRows([]*model.User{
					{UserID: 1, PasswordHash: "password_hash"},
				}))

				mock.ExpectQuery(regexp.QuoteMeta(
					"SELECT * FROM `jwt_audience_secret` WHERE (`jwt_audience_secret`.`audience` = ?) LIMIT 1;",
				)).WithArgs(
					"user:1",
				).WillReturnRows(test.NewJWTAudienceSecretRows([]*model.JWTAudienceSecret{
					{Audience: "user:1", Secret: "DBetxLyZOcHw3gQ+ozOyg+c6N1j2xG2yPTSVRrnXsaE="},
				}))

				mock.ExpectQuery(regexp.QuoteMeta(
					"SELECT * FROM `jwt_audience_secret` WHERE (`jwt_audience_secret`.`audience` = ?) LIMIT 1;",
				)).WithArgs(
					"user:1",
				).WillReturnRows(test.NewJWTAudienceSecretRows([]*model.JWTAudienceSecret{
					{Audience: "user:1", Secret: "DBetxLyZOcHw3gQ+ozOyg+c6N1j2xG2yPTSVRrnXsaE="},
				}))

				mock.ExpectExec(regexp.QuoteMeta(
					"INSERT INTO `jwt_denylist` (`user_id`,`jti`) VALUES (?,?)",
				)).WithArgs(
					1, "d391416c-c2d2-44d5-b3ec-147a7713606d",
				).WillReturnResult(
					sqlmock.NewResult(3, 1),
				)

				mock.ExpectQuery(regexp.QuoteMeta(
					"SELECT `jwt_denylist_id`,`created_at`,`updated_at` FROM `jwt_denylist` WHERE `jwt_denylist_id`=?",
				)).WithArgs(
					3,
				).WillReturnRows(sqlmock.NewRows([]string{"jwt_denylist_id", "created_at", "updated_at"}).
					AddRow(3, now, now),
				)
			},
		},
		{
			name:  "revoked token",
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJodHRwczovL2dpdGh1Yi5jb20vc2Vhbi1haG4vdXNlciIsImF1ZCI6WyJ1c2VyOjEiXSwiZXhwIjoxNjM2Mjk2ODYwLCJpYXQiOjE2MzUwODcyNjAsImp0aSI6ImQzOTE0MTZjLWMyZDItNDRkNS1iM2VjLTE0N2E3NzEzNjA2ZCIsInVzZXJfaWQiOiIxIn0.4KVLf9emjWK2GB9vXj1A-7KXHGq8HLnExUv54XoiNnU",
			dbExpectFunc: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(
					"SELECT * FROM `jwt_audience_secret` WHERE (`jwt_audience_secret`.`audience` = ?) LIMIT 1;",
				)).WithArgs(
					"user:1",
				).WillReturnRows(test.NewJWTAudienceSecretRows([]*model.JWTAudienceSecret{
					{Audience: "user:1", Secret: "DBetxLyZOcHw3gQ+ozOyg+c6N1j2xG2yPTSVRrnXsaE="},
				}))

				mock.ExpectQuery(regexp.QuoteMeta(
					"SELECT * FROM `jwt_denylist` WHERE (`jwt_denylist`.`jti` = ?) LIMIT 1;",
				)).WithArgs(
					"d391416c-c2d2-44d5-b3ec-147a7713606d",
				).WillReturnRows(test.NewJWTDenylistRows([]*model.JWTDenylist{
					{UserID: 1, Jti: "d391416c-c2d2-44d5-b3ec-147a7713606d"},
				}))
			},
			expectedErr: "revoked token",
		},
		{
			name:  "expired",
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJodHRwczovL2dpdGh1Yi5jb20vc2Vhbi1haG4vdXNlciIsImF1ZCI6WyJ1c2VyOjEiXSwiZXhwIjoxNjMzMjk2ODYwLCJpYXQiOjE2MzIwODcyNjAsImp0aSI6ImQzOTE0MTZjLWMyZDItNDRkNS1iM2VjLTE0N2E3NzEzNjA2ZCIsInVzZXJfaWQiOiIxIn0.gQNGyErIA4iS-WfYWHa7JhcCuJEymqQu-kVsD-w86Jo",
			dbExpectFunc: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(
					"SELECT * FROM `jwt_audience_secret` WHERE (`jwt_audience_secret`.`audience` = ?) LIMIT 1;",
				)).WithArgs(
					"user:1",
				).WillReturnRows(test.NewJWTAudienceSecretRows([]*model.JWTAudienceSecret{
					{Audience: "user:1", Secret: "DBetxLyZOcHw3gQ+ozOyg+c6N1j2xG2yPTSVRrnXsaE="},
				}))
			},
			expectedErr: "expired token",
		},
		{
			name:        "invalid format",
			token:       "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJodHRwczovL2dpdGh1Yi5jb20vc2Vhbi1haG4vdXNlciIsImF1ZCI6WyJ1c2VyOjEiLCJkZXZpY2U6eCJdLCJleHAiOjE2MzMyOTY4NjAsImlhdCI6MTYzMjA4NzI2MCwianRpIjoiZDM5MTQxNmMtYzJkMi00NGQ1LWIzZWMtMTQ3YTc3MTM2MDZkIiwidXNlcl9pZCI6IjEifQ.v0kmjF3XIRTcKvNzCWtl6VJ7em1lLZ19OLCe3q2Ofm0",
			expectedErr: "invalid claims format",
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
			defer test.CloseSqlmock(t, db, mock)

			if tc.dbExpectFunc != nil {
				tc.dbExpectFunc(mock)
			}

			svc := UserJWTTokenService{
				clock:                 clockwork.NewFakeClockAt(now),
				db:                    db,
				idGenerator:           &generator.UUIDGenerator{},
				accessTokenExpiresIn:  10 * time.Second,
				refreshTokenExpiresIn: 14 * 24 * time.Hour,
			}

			accessToken, refreshToken, err := svc.Refresh(ctx, tc.token)

			if tc.expectedErr != "" {
				assert.EqualError(t, err, tc.expectedErr)
				assert.Empty(t, accessToken)
				assert.Empty(t, refreshToken)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, accessToken)
				assert.NotEmpty(t, refreshToken)
			}
		})
	}

}

func TestUserJWTTokenService_Revoke(t *testing.T) {
	now := time.Date(2021, 10, 24, 7, 39, 46, 127956672, time.UTC)

	cases := []struct {
		name string

		token string

		dbExpectFunc func(sqlmock.Sqlmock)

		expectedErr string
	}{
		{
			name:  "revoke",
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJodHRwczovL2dpdGh1Yi5jb20vc2Vhbi1haG4vdXNlciIsImF1ZCI6WyJ1c2VyOjEiXSwiZXhwIjoxNjM2Mjk2ODYwLCJpYXQiOjE2MzUwODcyNjAsImp0aSI6ImQzOTE0MTZjLWMyZDItNDRkNS1iM2VjLTE0N2E3NzEzNjA2ZCIsInVzZXJfaWQiOiIxIn0.4KVLf9emjWK2GB9vXj1A-7KXHGq8HLnExUv54XoiNnU",
			dbExpectFunc: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(
					"SELECT * FROM `jwt_audience_secret` WHERE (`jwt_audience_secret`.`audience` = ?) LIMIT 1;",
				)).WithArgs(
					"user:1",
				).WillReturnRows(test.NewJWTAudienceSecretRows([]*model.JWTAudienceSecret{
					{Audience: "user:1", Secret: "DBetxLyZOcHw3gQ+ozOyg+c6N1j2xG2yPTSVRrnXsaE="},
				}))

				mock.ExpectExec(regexp.QuoteMeta(
					"INSERT INTO `jwt_denylist` (`user_id`,`jti`) VALUES (?,?)",
				)).WithArgs(
					1, "d391416c-c2d2-44d5-b3ec-147a7713606d",
				).WillReturnResult(
					sqlmock.NewResult(3, 1),
				)

				mock.ExpectQuery(regexp.QuoteMeta(
					"SELECT `jwt_denylist_id`,`created_at`,`updated_at` FROM `jwt_denylist` WHERE `jwt_denylist_id`=?",
				)).WithArgs(
					3,
				).WillReturnRows(sqlmock.NewRows([]string{"jwt_denylist_id", "created_at", "updated_at"}).
					AddRow(3, now, now),
				)
			},
		},
		{
			name:  "revoke without jti",
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJodHRwczovL2dpdGh1Yi5jb20vc2Vhbi1haG4vdXNlciIsImF1ZCI6WyJ1c2VyOjEiXSwiZXhwIjoxNjM2Mjk2ODYwLCJpYXQiOjE2MzUwODcyNjAsInVzZXJfaWQiOiIxIn0.zFRHLPXD7dmU7M_IEPiNz1ytiWtBoZFoHeli2cswz9g",
			dbExpectFunc: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(
					"SELECT * FROM `jwt_audience_secret` WHERE (`jwt_audience_secret`.`audience` = ?) LIMIT 1;",
				)).WithArgs(
					"user:1",
				).WillReturnRows(test.NewJWTAudienceSecretRows([]*model.JWTAudienceSecret{
					{Audience: "user:1", Secret: "DBetxLyZOcHw3gQ+ozOyg+c6N1j2xG2yPTSVRrnXsaE="},
				}))
			},
			expectedErr: "invalid claims format",
		},
		{
			name:  "already revoked token",
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJodHRwczovL2dpdGh1Yi5jb20vc2Vhbi1haG4vdXNlciIsImF1ZCI6WyJ1c2VyOjEiXSwiZXhwIjoxNjM2Mjk2ODYwLCJpYXQiOjE2MzUwODcyNjAsImp0aSI6ImQzOTE0MTZjLWMyZDItNDRkNS1iM2VjLTE0N2E3NzEzNjA2ZCIsInVzZXJfaWQiOiIxIn0.4KVLf9emjWK2GB9vXj1A-7KXHGq8HLnExUv54XoiNnU",
			dbExpectFunc: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(
					"SELECT * FROM `jwt_audience_secret` WHERE (`jwt_audience_secret`.`audience` = ?) LIMIT 1;",
				)).WithArgs(
					"user:1",
				).WillReturnRows(test.NewJWTAudienceSecretRows([]*model.JWTAudienceSecret{
					{Audience: "user:1", Secret: "DBetxLyZOcHw3gQ+ozOyg+c6N1j2xG2yPTSVRrnXsaE="},
				}))

				mock.ExpectExec(regexp.QuoteMeta(
					"INSERT INTO `jwt_denylist` (`user_id`,`jti`) VALUES (?,?)",
				)).WithArgs(
					1, "d391416c-c2d2-44d5-b3ec-147a7713606d",
				).WillReturnError(
					&mysqldriver.MySQLError{Number: mysql.ErrorCodeDuplicateEntry},
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
			defer test.CloseSqlmock(t, db, mock)

			if tc.dbExpectFunc != nil {
				tc.dbExpectFunc(mock)
			}

			svc := UserJWTTokenService{
				clock:                 clockwork.NewFakeClockAt(now),
				db:                    db,
				idGenerator:           &generator.UUIDGenerator{},
				accessTokenExpiresIn:  10 * time.Second,
				refreshTokenExpiresIn: 14 * 24 * time.Hour,
			}

			err = svc.Revoke(ctx, tc.token)

			if tc.expectedErr != "" {
				assert.EqualError(t, err, tc.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
