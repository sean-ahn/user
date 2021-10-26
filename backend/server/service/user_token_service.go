package service

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/binary"
	"strconv"
	"time"

	mysqldriver "github.com/go-sql-driver/mysql"
	"github.com/golang-jwt/jwt/v4"
	"github.com/jonboulle/clockwork"
	"github.com/pkg/errors"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/sean-ahn/user/backend/model"
	"github.com/sean-ahn/user/backend/persistence/mysql"
	"github.com/sean-ahn/user/backend/server/generator"
)

const (
	issuer = "https://github.com/sean-ahn/user"
)

var (
	ErrTokenRevocationFailed = errors.New("token revocation failed")

	errJWTSecretNotFound   = errors.New("jwt secret not found")
	errInvalidClaimsFormat = errors.New("invalid claims format")
	errRevokedToken        = errors.New("revoked token")
	errExpiredToken        = errors.New("expired token")
)

//go:generate mockgen -package service -destination ./user_token_service_mock.go -mock_names UserTokenService=MockUserTokenService github.com/sean-ahn/user/backend/server/service UserTokenService

type UserTokenService interface {
	Issue(context.Context, *model.User) (string, string, error)
	Refresh(context.Context, string) (string, string, error)
	Revoke(context.Context, string) error
	RevokeAll(context.Context, *model.User, *sql.Tx) error
	GetUser(context.Context, string) (*model.User, error)
}

type UserJWTTokenService struct {
	clock       clockwork.Clock
	db          *sql.DB
	idGenerator generator.Generator

	accessTokenExpiresIn  time.Duration
	refreshTokenExpiresIn time.Duration
}

var _ UserTokenService = (*UserJWTTokenService)(nil)

type JWTClaims struct {
	jwt.RegisteredClaims

	UserID string `json:"user_id"`
}

func NewJWTTokenService(clock clockwork.Clock, db *sql.DB, accessTokenExpiresIn, refreshTokenExpiresIn time.Duration) *UserJWTTokenService {
	return &UserJWTTokenService{
		clock:                 clock,
		db:                    db,
		idGenerator:           &generator.UUIDGenerator{},
		accessTokenExpiresIn:  accessTokenExpiresIn,
		refreshTokenExpiresIn: refreshTokenExpiresIn,
	}
}

func (s *UserJWTTokenService) Issue(ctx context.Context, user *model.User) (string, string, error) {
	secret, err := s.getSecret(ctx, user)
	if errors.Cause(err) == errJWTSecretNotFound {
		secret, err = s.createSecret(ctx, user)
	}
	if err != nil {
		return "", "", err
	}

	return s.generateTokens(user, secret)
}

func (s *UserJWTTokenService) Refresh(ctx context.Context, refreshToken string) (string, string, error) {
	token, err := s.parseToken(ctx, refreshToken)
	if err != nil {
		return "", "", err
	}

	claims := token.Claims.(*JWTClaims)
	if claims.ID == "" {
		return "", "", errors.WithStack(errInvalidClaimsFormat)
	}

	userID64, err := strconv.ParseInt(claims.UserID, 10, 32)
	if err != nil {
		return "", "", errors.WithStack(errInvalidClaimsFormat)
	}

	jd, err := mysql.FindJWTDenylistByJTI(ctx, s.db, claims.ID)
	if err != nil && errors.Cause(err) != sql.ErrNoRows {
		return "", "", err
	}
	if jd != nil {
		return "", "", errors.WithStack(errRevokedToken)
	}

	user, err := mysql.GetUser(ctx, s.db, int(userID64))
	if err != nil {
		return "", "", err
	}

	secret, err := s.getSecret(ctx, user)
	if err != nil {
		return "", "", err
	}

	if err := s.Revoke(ctx, refreshToken); err != nil {
		return "", "", err
	}

	return s.generateTokens(user, secret)
}

func (s *UserJWTTokenService) Revoke(ctx context.Context, refreshToken string) error {
	token, err := s.parseToken(ctx, refreshToken)
	if err != nil {
		return errors.Wrap(ErrTokenRevocationFailed, err.Error())
	}

	claims := token.Claims.(*JWTClaims)
	if claims.ID == "" {
		return errors.WithStack(errInvalidClaimsFormat)
	}

	userID64, err := strconv.ParseInt(claims.UserID, 10, 32)
	if err != nil {
		return errors.WithStack(errInvalidClaimsFormat)
	}

	jd := &model.JWTDenylist{
		UserID: int(userID64),
		Jti:    claims.ID,
	}
	if err := jd.Insert(ctx, s.db, boil.Infer()); err != nil {
		if merr, ok := errors.Cause(err).(*mysqldriver.MySQLError); !ok || merr.Number != mysql.ErrorCodeDuplicateEntry {
			return errors.Wrap(ErrTokenRevocationFailed, err.Error())
		}
	}

	return nil
}

func (s *UserJWTTokenService) RevokeAll(ctx context.Context, user *model.User, tx *sql.Tx) error {
	var exec boil.ContextExecutor = s.db
	if tx != nil {
		exec = tx
	}

	if _, err := model.JWTAudienceSecrets(
		model.JWTAudienceSecretWhere.Audience.EQ(s.getAudience(user)),
	).DeleteAll(ctx, exec); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (s *UserJWTTokenService) GetUser(ctx context.Context, accessToken string) (*model.User, error) {
	token, err := s.parseToken(ctx, accessToken)
	if err != nil {
		return nil, err
	}

	claims := token.Claims.(*JWTClaims)
	userID64, err := strconv.ParseInt(claims.UserID, 10, 32)
	if err != nil {
		return nil, errors.WithStack(errInvalidClaimsFormat)
	}

	user, err := mysql.GetUser(ctx, s.db, int(userID64))
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserJWTTokenService) getAudience(user *model.User) string {
	return "user:" + strconv.FormatInt(int64(user.UserID), 10)
}

func (s *UserJWTTokenService) parseToken(ctx context.Context, token string) (*jwt.Token, error) {
	tk, err := jwt.ParseWithClaims(token, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		claims, err := s.parseClaims(token.Claims)
		if err != nil {
			return nil, errors.WithStack(errInvalidClaimsFormat)
		}

		sec, err := s.getSecretByAudience(ctx, claims.Audience[0])
		if err != nil {
			return nil, err
		}
		return sec, nil
	})
	if err != nil {
		switch x := errors.Cause(err).(type) {
		case nil:
		case *jwt.ValidationError:
			if x.Errors == jwt.ValidationErrorExpired {
				return nil, errors.WithStack(errExpiredToken)
			}
			if x.Errors == jwt.ValidationErrorUnverifiable {
				return nil, errors.WithStack(x.Inner)
			}
			return nil, x
		default:
			return nil, err
		}
	}

	return tk, nil
}

func (s *UserJWTTokenService) parseClaims(c jwt.Claims) (*JWTClaims, error) {
	claims, ok := c.(*JWTClaims)
	if !ok {
		return nil, errors.WithStack(errInvalidClaimsFormat)
	}
	if claims.Issuer != issuer {
		return nil, errors.WithStack(errInvalidClaimsFormat)
	}
	if len(claims.Audience) != 1 {
		return nil, errors.WithStack(errInvalidClaimsFormat)
	}
	if _, err := strconv.ParseInt(claims.UserID, 10, 32); err != nil {
		return nil, errors.WithStack(errInvalidClaimsFormat)
	}

	return claims, nil
}

func (s *UserJWTTokenService) newClaimsPair(user *model.User) (JWTClaims, JWTClaims) {
	now := s.clock.Now()
	return JWTClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				Audience:  []string{s.getAudience(user)},
				ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTokenExpiresIn)),
				IssuedAt:  jwt.NewNumericDate(now),
				Issuer:    issuer,
			},
			UserID: strconv.FormatInt(int64(user.UserID), 10),
		}, JWTClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				ID:        s.idGenerator.Generate(),
				Audience:  []string{s.getAudience(user)},
				ExpiresAt: jwt.NewNumericDate(now.Add(s.refreshTokenExpiresIn)),
				IssuedAt:  jwt.NewNumericDate(now),
				Issuer:    issuer,
			},
			UserID: strconv.FormatInt(int64(user.UserID), 10),
		}
}

func (s *UserJWTTokenService) generateTokens(user *model.User, secret []byte) (string, string, error) {
	accessTokenClaims, refreshTokenClaims := s.newClaimsPair(user)

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims).SignedString(secret)
	if err != nil {
		return "", "", errors.WithStack(err)
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaims).SignedString(secret)
	if err != nil {
		return "", "", errors.WithStack(err)
	}

	return accessToken, refreshToken, nil
}

func (s *UserJWTTokenService) getSecret(ctx context.Context, user *model.User) ([]byte, error) {
	return s.getSecretByAudience(ctx, s.getAudience(user))
}

func (s *UserJWTTokenService) getSecretByAudience(ctx context.Context, aud string) ([]byte, error) {
	jas, err := model.JWTAudienceSecrets(model.JWTAudienceSecretWhere.Audience.EQ(aud)).One(ctx, s.db)
	if errors.Cause(err) == sql.ErrNoRows {
		return nil, errors.WithStack(errJWTSecretNotFound)
	}
	if err != nil {
		return nil, errors.WithStack(err)
	}

	secret, err := base64.StdEncoding.DecodeString(jas.Secret)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return secret, nil
}

func (s *UserJWTTokenService) createSecret(ctx context.Context, user *model.User) ([]byte, error) {
	secret, err := s.newSecret(user)
	if err != nil {
		return nil, err
	}

	jas := &model.JWTAudienceSecret{
		Audience: s.getAudience(user),
		Secret:   base64.StdEncoding.EncodeToString(secret),
	}
	if err := jas.Insert(ctx, s.db, boil.Infer()); err != nil {
		return nil, errors.WithStack(err)
	}
	return secret, nil
}

func (s *UserJWTTokenService) newSecret(user *model.User) ([]byte, error) {
	h := sha256.New()

	h.Write([]byte(user.PasswordHash))

	ts := make([]byte, 8)
	binary.LittleEndian.PutUint64(ts, uint64(s.clock.Now().UnixNano()))
	h.Write(ts)

	return h.Sum(nil), nil
}
