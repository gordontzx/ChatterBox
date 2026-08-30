package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

const refreshTokenBytes = 32
const minPasswordLength = 8

type Service struct {
	store  Store
	secret []byte
}

func NewService(store Store, secret []byte) *Service {
	return &Service{store: store, secret: secret}
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) error {
	if err := validatePasswordStrength(req.Password); err != nil {
		return err
	}

	hash, err := HashPassword(req.Password)
	if err != nil {
		return err
	}

	_, err = s.store.CreateUser(ctx, req.Email, hash, req.Name)
	return err
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (*TokenPair, error) {
	user, err := s.store.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}

	// User does not exist or wrong password
	if user == nil || !PasswordMatches(user.PasswordHash, req.Password) {
		return nil, ErrInvalidCredentials
	}

	refreshToken := generateRefreshToken(refreshTokenBytes)
	refreshHash := hashToken(refreshToken)

	if err := s.store.CreateSession(ctx, user.ID, req.DeviceName, refreshHash); err != nil {
		return nil, err
	}

	accessToken, err := IssueAccessToken(user.ID, s.secret)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func validatePasswordStrength(password string) error {
	if len(password) < minPasswordLength {
		return ErrPasswordTooShort
	}
	return nil
}

func generateRefreshToken(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		panic("crypto/rand failed: " + err.Error())
	}
	return hex.EncodeToString(b)
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
