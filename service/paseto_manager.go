package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/o1egl/paseto"
)

const (
	authorizationHeader = "authorization"
	authorizationBearer = "bearer"
	TokenKey            = "cd56e76e8bf6a1c32eb26966c864e983"
	TokenDuration       = 15 * time.Minute
)

type PasetoManager struct {
	paseto        *paseto.V2
	symmetricKey  []byte
	tokenDuration time.Duration
}

type UserPayload struct {
	paseto.JSONToken
	Username string `json:"username"`
	Role     string `json:"role"`
}

func NewPasetoManager(secretKey string, tokenDuration time.Duration) *PasetoManager {
	return &PasetoManager{
		paseto:        paseto.NewV2(),
		symmetricKey:  []byte(secretKey),
		tokenDuration: tokenDuration,
	}
}

func (manager *PasetoManager) CreateToken(user *User) (string, error) {
	payload := &UserPayload{
		JSONToken: paseto.JSONToken{
			Expiration: time.Now().Add(manager.tokenDuration),
		},
	}
	payload.Set("username", user.Username)
	payload.Set("role", user.Role)

	token, err := manager.paseto.Encrypt(manager.symmetricKey, payload, nil)
	if err != nil {
		return "", fmt.Errorf("token encryption failed: %w", err)
	}

	return token, nil
}

func (manager *PasetoManager) VerifyToken(token string) (*UserPayload, error) {
	var newPayload paseto.JSONToken
	err := manager.paseto.Decrypt(token, manager.symmetricKey, &newPayload, nil)
	if err != nil {
		if strings.Contains(err.Error(), "expired") {
			return nil, fmt.Errorf("token expired: %w", err)
		}

		return nil, fmt.Errorf("token decryption failed: %w", err)
	}

	username := newPayload.Get("username")
	role := newPayload.Get("role")

	payload := &UserPayload{
		Username:  username,
		Role:      role,
		JSONToken: newPayload,
	}

	return payload, nil
}
