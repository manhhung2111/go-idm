package logic

import (
	"context"
	"errors"

	"github.com/manhhung2111/go-idm/internal/config"
	"golang.org/x/crypto/bcrypt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Hash interface {
	Hash(ctx context.Context, data string) (string, error)
	IsHashEqual(ctx context.Context, data string, hashedData string) (bool, error)
}

type hash struct {
	authConfig config.Auth
}

func NewHash(authConfig config.Auth) Hash {
	return &hash{
		authConfig: authConfig,
	}
}


// Hash implements Hash.
func (h *hash) Hash(_ context.Context, data string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(data), h.authConfig.Hash.Cost)
	if err != nil {
		return "", status.Errorf(codes.Internal, "failed to hash data: %+v", err)
	}

	return string(hashed), nil
}

// IsHashEqual implements Hash.
func (h *hash) IsHashEqual(_ context.Context, data string, hashedData string) (bool, error) {
	if err := bcrypt.CompareHashAndPassword([]byte(hashedData), []byte(data)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, nil
		}

		return false, status.Errorf(codes.Internal, "failed to check if data equal hash: %+v", err)
	}

	return true, nil 
}

