package auth

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

type fakeStore struct {
	usersByEmail   map[string]*User
	sessionsByHash map[string]*Session
}

func newFakeStore() *fakeStore {
	return &fakeStore{
		usersByEmail:   map[string]*User{},
		sessionsByHash: map[string]*Session{},
	}
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
	f.sessionsByHash[refreshTokenHash] = &Session{ID: "session-1", UserID: userID}
	return nil
}

func (f *fakeStore) GetSessionByRefreshHash(ctx context.Context, hash string) (*Session, error) {
	return f.sessionsByHash[hash], nil
}

func (f *fakeStore) RevokeSession(ctx context.Context, hash string) error {
	if sess, ok := f.sessionsByHash[hash]; ok {
		now := time.Now()
		sess.RevokedAt = &now
	}
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

func TestRegister_NormalizesEmail(t *testing.T) {
	store := newFakeStore()
	svc := NewService(store, []byte("test-secret"))
	ctx := context.Background()

	svc.Register(ctx, RegisterRequest{Email: "  Alice@Example.com  ", Password: "password123", Name: "Alice"})
	err := svc.Register(ctx, RegisterRequest{Email: "alice@example.com", Password: "password123", Name: "Alice"})

	if !errors.Is(err, ErrEmailAlreadyRegistered) {
		t.Fatalf("expected duplicate detection across case/whitespace variants, got %v", err)
	}
}

func TestRegister_PasswordTooLong(t *testing.T) {
	svc := NewService(newFakeStore(), []byte("test-secret"))
	longPassword := strings.Repeat("a", 73)

	err := svc.Register(context.Background(), RegisterRequest{Email: "a@b.com", Password: longPassword, Name: "A"})
	if !errors.Is(err, ErrPasswordTooLong) {
		t.Fatalf("expected ErrPasswordTooLong, got %v", err)
	}
}

func TestRegister_InvalidEmail(t *testing.T) {
	svc := NewService(newFakeStore(), []byte("test-secret"))

	err := svc.Register(context.Background(), RegisterRequest{Email: "not-an-email", Password: "password123", Name: "A"})
	if !errors.Is(err, ErrInvalidEmail) {
		t.Fatalf("expected ErrInvalidEmail, got %v", err)
	}
}

func TestRefreshAndLogout(t *testing.T) {
	svc := NewService(newFakeStore(), []byte("test-secret"))
	ctx := context.Background()

	svc.Register(ctx, RegisterRequest{Email: "a@b.com", Password: "password123", Name: "A"})
	tokens, _ := svc.Login(ctx, LoginRequest{Email: "a@b.com", Password: "password123", DeviceName: "test"})

	// refresh should succeed
	resp, err := svc.RefreshAccessToken(ctx, RefreshRequest{RefreshToken: tokens.RefreshToken})
	if err != nil || resp.AccessToken == "" {
		t.Fatalf("expected valid refresh, got err=%v resp=%v", err, resp)
	}

	// logout
	if err := svc.Logout(ctx, LogoutRequest{RefreshToken: tokens.RefreshToken}); err != nil {
		t.Fatalf("logout failed: %v", err)
	}

	// refresh after logout should fail
	_, err = svc.RefreshAccessToken(ctx, RefreshRequest{RefreshToken: tokens.RefreshToken})
	if !errors.Is(err, ErrInvalidRefreshToken) {
		t.Fatalf("expected ErrInvalidRefreshToken after logout, got %v", err)
	}
}
