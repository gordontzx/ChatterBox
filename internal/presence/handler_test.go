package presence

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"ChatterBox/internal/middleware"
)

type fakeOnlineChecker struct {
	online map[string]bool
}

func (f *fakeOnlineChecker) IsOnline(userID string) bool {
	return f.online[userID]
}

type fakeContactChecker struct {
	shared map[string]bool // key: "userA|userB", either order
}

func (f *fakeContactChecker) ShareAnyChat(ctx context.Context, userA, userB string) (bool, error) {
	return f.shared[userA+"|"+userB] || f.shared[userB+"|"+userA], nil
}

func withAuthenticatedUser(req *http.Request, userID string) *http.Request {
	ctx := context.WithValue(req.Context(), middleware.UserIDContextKeyForTests, userID)
	return req.WithContext(ctx)
}

func TestGetUserPresence_OwnPresence_AlwaysAllowed(t *testing.T) {
	store := newFakeStore()
	svc := NewService(store, &fakeContacts{}, newFakeBroadcaster())
	h := NewHandler(svc, &fakeOnlineChecker{online: map[string]bool{"alice": true}}, &fakeContactChecker{})

	req := httptest.NewRequest("GET", "/users/alice/presence", nil)
	req.SetPathValue("userID", "alice")
	req = withAuthenticatedUser(req, "alice") // checking own presence

	rec := httptest.NewRecorder()
	h.GetUserPresence(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for own presence, got %d", rec.Code)
	}
}

func TestGetUserPresence_Contact_Allowed(t *testing.T) {
	store := newFakeStore()
	svc := NewService(store, &fakeContacts{}, newFakeBroadcaster())
	h := NewHandler(
		svc,
		&fakeOnlineChecker{online: map[string]bool{"bob": true}},
		&fakeContactChecker{shared: map[string]bool{"alice|bob": true}},
	)

	req := httptest.NewRequest("GET", "/users/bob/presence", nil)
	req.SetPathValue("userID", "bob")
	req = withAuthenticatedUser(req, "alice")

	rec := httptest.NewRecorder()
	h.GetUserPresence(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for a contact's presence, got %d", rec.Code)
	}
}

func TestGetUserPresence_NonContact_Forbidden(t *testing.T) {
	store := newFakeStore()
	svc := NewService(store, &fakeContacts{}, newFakeBroadcaster())
	h := NewHandler(
		svc,
		&fakeOnlineChecker{online: map[string]bool{"stranger": true}},
		&fakeContactChecker{}, // no shared chats recorded — always false
	)

	req := httptest.NewRequest("GET", "/users/stranger/presence", nil)
	req.SetPathValue("userID", "stranger")
	req = withAuthenticatedUser(req, "alice")

	rec := httptest.NewRecorder()
	h.GetUserPresence(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for a non-contact, got %d", rec.Code)
	}
}
