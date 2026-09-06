package user

import (
	"context"
	"errors"
	"testing"
)

type fakeStore struct {
	byEmail map[string]*PublicUser
	byID    map[string]*PublicUser
}

func newFakeStore() *fakeStore {
	return &fakeStore{
		byEmail: map[string]*PublicUser{},
		byID:    map[string]*PublicUser{},
	}
}

func (f *fakeStore) FindByEmail(ctx context.Context, email string) (*PublicUser, error) {
	return f.byEmail[email], nil
}

func (f *fakeStore) FindByID(ctx context.Context, userID string) (*PublicUser, error) {
	return f.byID[userID], nil
}

func seedUser(store *fakeStore, id, name, email string) {
	u := &PublicUser{ID: id, Name: name, Email: email}
	store.byID[id] = u
	store.byEmail[email] = u
}

func TestFindByEmail_Found(t *testing.T) {
	store := newFakeStore()
	seedUser(store, "bob-id", "Bob", "bob@example.com")
	svc := NewService(store)

	u, err := svc.FindByEmail(context.Background(), "Bob@Example.com") // mismatched case
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.ID != "bob-id" {
		t.Fatalf("expected bob-id, got %s", u.ID)
	}
}

func TestFindByEmail_NotFound(t *testing.T) {
	store := newFakeStore()
	svc := NewService(store)

	_, err := svc.FindByEmail(context.Background(), "nobody@example.com")
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestFindByEmail_InvalidFormat(t *testing.T) {
	store := newFakeStore()
	svc := NewService(store)

	_, err := svc.FindByEmail(context.Background(), "not-an-email")
	if !errors.Is(err, ErrInvalidEmail) {
		t.Fatalf("expected ErrInvalidEmail, got %v", err)
	}
}

func TestGetByID_Found(t *testing.T) {
	store := newFakeStore()
	seedUser(store, "alice-id", "Alice", "alice@example.com")
	svc := NewService(store)

	u, err := svc.FindByID(context.Background(), "alice-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.Name != "Alice" {
		t.Fatalf("expected Alice, got %s", u.Name)
	}
}

func TestGetByID_NotFound(t *testing.T) {
	store := newFakeStore()
	svc := NewService(store)

	_, err := svc.FindByID(context.Background(), "ghost-id")
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}
