package logic

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/manhhung2111/go-idm/internal/config"
	"github.com/manhhung2111/go-idm/internal/dataaccess/database"
	"github.com/manhhung2111/go-idm/internal/utils"
	"go.uber.org/zap"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const privateKey = "manhhung2111-secret"

var (
	errUnexpectedSigningMethod = status.Error(codes.Unauthenticated, "unexpected signing method")
	errCannotGetTokensClaims   = status.Error(codes.Unauthenticated, "cannot get token's claims")
	errCannotGetTokensSubClaim = status.Error(codes.Unauthenticated, "cannot get token's sub claim")
	errCannotGetTokensExpClaim = status.Error(codes.Unauthenticated, "cannot get token's exp claim")
	errInvalidToken            = status.Error(codes.Unauthenticated, "invalid token")
	errFailedToSignToken       = status.Error(codes.Internal, "failed to sign token")
)

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

	tokenString, err := token.SignedString([]byte(privateKey))
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to sign token")
		return "", time.Time{}, errFailedToSignToken
	}

	return tokenString, expireTime, nil
}

func (t *token) GetAccountIDAndExpireTime(ctx context.Context, token string) (uint64, time.Time, error) {
	logger := utils.LoggerWithContext(ctx, t.logger)

	parsedToken, err := jwt.Parse(token, func(parsedToken *jwt.Token) (interface{}, error) {
		// Ensure it's signed with HMAC
		if _, ok := parsedToken.Method.(*jwt.SigningMethodHMAC); !ok {
			logger.Error("unexpected signing method")
			return nil, errUnexpectedSigningMethod
		}

		return []byte(privateKey), nil
	})
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to parse token")
		return 0, time.Time{}, err
	}

	if !parsedToken.Valid {
		logger.Error("invalid token")
		return 0, time.Time{}, errInvalidToken
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		logger.Error("cannot get token's claims")
		return 0, time.Time{}, errCannotGetTokensClaims
	}

	accountId, ok := claims["sub"].(float64)
	if !ok {
		logger.Error("cannot get token's sub claim")
		return 0, time.Time{}, errCannotGetTokensSubClaim
	}

	expireTimeUnix, ok := claims["exp"].(float64)
	if !ok {
		logger.Error("cannot get token's exp claim")
		return 0, time.Time{}, errCannotGetTokensExpClaim
	}

	return uint64(accountId), time.Unix(int64(expireTimeUnix), 0), nil
}


func (t *token) WithDatabase(database database.IDatabase) Token {
	t.accountDataAccessor = t.accountDataAccessor.WithDatabase(database, t.logger)
	return t
}
