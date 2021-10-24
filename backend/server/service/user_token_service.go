package service

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/binary"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/jonboulle/clockwork"
	"github.com/pkg/errors"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/sean-ahn/user/backend/model"
)

const (
	issuer = "https://github.com/sean-ahn/user"
)

var (
	errJWTSecretNotFound = errors.New("jwt secret not found")
)

//go:generate mockgen -package service -destination ./user_token_service_mock.go -mock_names UserTokenService=MockUserTokenService github.com/sean-ahn/user/backend/server/service UserTokenService

type UserTokenService interface {
	Issue(context.Context, *model.User) (string, string, error)
	Refresh(context.Context, *model.User, string) (string, string, error)
	Revoke(context.Context, *model.User, string) error
}

type UserJWTTokenService struct {
	clock clockwork.Clock
	db    *sql.DB

	accessTokenExpiresIn  time.Duration
	refreshTokenExpiresIn time.Duration
}

var _ UserTokenService = (*UserJWTTokenService)(nil)

type JWTClaims struct {
	jwt.StandardClaims

	UserID string `json:"user_id"`
}

func NewJWTTokenService(clock clockwork.Clock, db *sql.DB, accessTokenExpiresIn, refreshTokenExpiresIn time.Duration) *UserJWTTokenService {
	return &UserJWTTokenService{clock: clock, db: db, accessTokenExpiresIn: accessTokenExpiresIn, refreshTokenExpiresIn: refreshTokenExpiresIn}
}

func (s *UserJWTTokenService) Issue(ctx context.Context, user *model.User) (string, string, error) {
	secret, err := s.getSecret(ctx, user)
	if errors.Cause(err) == errJWTSecretNotFound {
		secret, err = s.createSecret(ctx, user)
	}
	if err != nil {
		return "", "", err
	}

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

func (s *UserJWTTokenService) Refresh(ctx context.Context, user *model.User, refreshToken string) (string, string, error) {
	panic("implement me")
}

func (s *UserJWTTokenService) Revoke(ctx context.Context, user *model.User, refreshToken string) error {
	panic("implement me")
}

func (s *UserJWTTokenService) newClaimsPair(user *model.User) (JWTClaims, JWTClaims) {
	now := s.clock.Now()
	return JWTClaims{
			StandardClaims: jwt.StandardClaims{
				Audience:  s.getAudience(user),
				ExpiresAt: now.Add(s.accessTokenExpiresIn).Unix(),
				IssuedAt:  now.Unix(),
				Issuer:    issuer,
			},
			UserID: strconv.FormatInt(int64(user.UserID), 10),
		}, JWTClaims{
			StandardClaims: jwt.StandardClaims{
				Id:        uuid.New().String(),
				Audience:  s.getAudience(user),
				ExpiresAt: now.Add(s.refreshTokenExpiresIn).Unix(),
				IssuedAt:  now.Unix(),
				Issuer:    issuer,
			},
			UserID: strconv.FormatInt(int64(user.UserID), 10),
		}
}

func (s *UserJWTTokenService) getSecret(ctx context.Context, user *model.User) ([]byte, error) {
	jas, err := model.JWTAudienceSecrets(model.JWTAudienceSecretWhere.Audience.EQ(s.getAudience(user))).One(ctx, s.db)
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

func (s *UserJWTTokenService) getAudience(user *model.User) string {
	return "user:" + strconv.FormatInt(int64(user.UserID), 10)
}
