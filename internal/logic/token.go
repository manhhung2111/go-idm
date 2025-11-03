package logic

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/manhhung2111/go-idm/internal/config"
	"github.com/manhhung2111/go-idm/internal/dataaccess/database"
	"github.com/manhhung2111/go-idm/internal/utils"
	"go.uber.org/zap"
)

const privateKey = "manhhung2111-secret"

type Token interface {
	GetToken(ctx context.Context, accountId uint64) (string, time.Time, error)
	GetAccountIDAndExpireTime(ctx context.Context, token string) (uint64, time.Time, error)
	WithDatabase(database database.IDatabase) Token
}

type token struct {
	accountDataAccessor database.AccountDataAccessor
	expiresIn           time.Duration
	authConfig          config.Auth
	logger              *zap.Logger
}

func NewToken(
	accountDataAccessor database.AccountDataAccessor,
	authConfig config.Auth,
	logger *zap.Logger,
) (Token, error) {
	expiresIn, err := authConfig.Token.GetExpiresInDuration()
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to parse expires_in")
		return nil, err
	}

	return &token{
		accountDataAccessor: accountDataAccessor,
		expiresIn:           expiresIn,
		authConfig:          authConfig,
		logger:              logger,
	}, nil
}

func (t *token) GetToken(ctx context.Context, accountId uint64) (string, time.Time, error) {
	logger := utils.LoggerWithContext(ctx, t.logger)

	expireTime := time.Now().Add(t.expiresIn)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": accountId,
		"exp": expireTime.Unix(),
	})

	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to sign token")
		return "", time.Time{}, err
	}

	return tokenString, expireTime, nil
}

func (t *token) GetAccountIDAndExpireTime(ctx context.Context, token string) (uint64, time.Time, error) {
	logger := utils.LoggerWithContext(ctx, t.logger)

	parsedToken, err := jwt.Parse(token, func(parsedToken *jwt.Token) (interface{}, error) {
		if _, ok := parsedToken.Method.(*jwt.SigningMethodHMAC); !ok {
			logger.Error("unexpected signing method")
			return nil, errors.New("unexpected signing method")
		}

		claims, ok := parsedToken.Claims.(jwt.MapClaims)
		if !ok {
			logger.Error("cannot get token's claims")
			return nil, errors.New("cannot get token's claims")
		}

		return claims, nil
	})

	if err != nil {
		logger.With(zap.Error(err)).Error("failed to parse token")
		return 0, time.Time{}, err
	}

	if !parsedToken.Valid {
		logger.Error("invalid token")
		return 0, time.Time{}, errors.New("invalid token")
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		logger.Error("cannot get token's claims")
		return 0, time.Time{}, errors.New("cannot get token's claims")
	}

	accountId, ok := claims["sub"].(float64)
	if !ok {
		logger.Error("cannot get token's sub claim")
		return 0, time.Time{}, errors.New("cannot get token's sub claim")
	}

	expireTimeUnix, ok := claims["exp"].(float64)
	if !ok {
		logger.Error("cannot get token's exp claim")
		return 0, time.Time{}, errors.New("cannot get token's exp claim")
	}

	return uint64(accountId), time.Unix(int64(expireTimeUnix), 0), nil
}

func (t *token) WithDatabase(database database.IDatabase) Token {
	t.accountDataAccessor = t.accountDataAccessor.WithDatabase(database)
	return t
}
