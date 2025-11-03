package logic

import (
	"context"
	"errors"

	"github.com/manhhung2111/go-idm/internal/config"
	"golang.org/x/crypto/bcrypt"
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
func (h *hash) Hash(ctx context.Context, data string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(data), h.authConfig.Hash.Cost)
	if err != nil {
		return "", err
	}

	return string(hashed), nil
}

// IsHashEqual implements Hash.
func (h *hash) IsHashEqual(ctx context.Context, data string, hashedData string) (bool, error) {
	if err := bcrypt.CompareHashAndPassword([]byte(hashedData), []byte(data)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, nil
		}

		return false, err
	}

	return true, nil 
}

