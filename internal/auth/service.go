package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

const (
	refreshTokenBytes = 32
	minPasswordLength = 8
)

type Service struct {
	store  Store
	secret []byte
}

func NewService(store Store, secret []byte) *Service {
	return &Service{store: store, secret: secret}
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) error {
	email, err := normalizeEmail(req.Email)
	if err != nil {
		return err
	}

	if err := validatePasswordStrength(req.Password); err != nil {
		return err
	}

	hash, err := HashPassword(req.Password)
	if err != nil {
		return err
	}

	_, err = s.store.CreateUser(ctx, email, hash, req.Name)
	return err
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (*TokenPair, error) {
	email, err := normalizeEmail(req.Email)
	if err != nil {
		return nil, ErrInvalidEmail
	}

	user, err := s.store.GetUserByEmail(ctx, email)
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

func normalizeEmail(email string) (string, error) {
	trimmed := strings.ToLower(strings.TrimSpace(email))
	if trimmed == "" || !strings.Contains(trimmed, "@") || strings.ContainsAny(trimmed, " \t\n") {
		return "", ErrInvalidEmail
	}
	return trimmed, nil
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
