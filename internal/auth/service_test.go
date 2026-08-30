package auth

import (
	"context"
	"errors"
	"testing"
)

type fakeStore struct {
	usersByEmail map[string]*User
}

func newFakeStore() *fakeStore {
	return &fakeStore{usersByEmail: map[string]*User{}}
}

func (f *fakeStore) CreateUser(ctx context.Context, email, passwordHash, name string) (*User, error) {
	if _, exists := f.usersByEmail[email]; exists {
		return nil, ErrEmailAlreadyRegistered
	}
	u := &User{ID: "user-1", Email: email, PasswordHash: passwordHash, Name: name}
	f.usersByEmail[email] = u
	return u, nil
}

func (f *fakeStore) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	return f.usersByEmail[email], nil // nil, nil if not found — matches real store's contract
}

func (f *fakeStore) CreateSession(ctx context.Context, userID, deviceName, refreshTokenHash string) error {
	return nil
}

func TestRegisterAndLogin(t *testing.T) {
	svc := NewService(newFakeStore(), []byte("test-secret"))
	ctx := context.Background()

	err := svc.Register(ctx, RegisterRequest{Email: "a@b.com", Password: "password123", Name: "Alice"})
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	tokens, err := svc.Login(ctx, LoginRequest{Email: "a@b.com", Password: "password123", DeviceName: "test-device"})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		t.Fatal("expected non-empty tokens")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	svc := NewService(newFakeStore(), []byte("test-secret"))
	ctx := context.Background()

	svc.Register(ctx, RegisterRequest{Email: "a@b.com", Password: "password123", Name: "Alice"})

	_, err := svc.Login(ctx, LoginRequest{Email: "a@b.com", Password: "wrong-password"})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLogin_UnknownEmail(t *testing.T) {
	svc := NewService(newFakeStore(), []byte("test-secret"))

	_, err := svc.Login(context.Background(), LoginRequest{Email: "nobody@b.com", Password: "whatever1"})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}
