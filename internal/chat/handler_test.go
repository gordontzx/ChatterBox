package chat

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ChatterBox/internal/middleware"
)

// withAuthenticatedUser mimics what middleware.RequireAuth would inject,
// so handler tests can run without the full middleware in the chain.
func withAuthenticatedUser(req *http.Request, userID string) *http.Request {
	ctx := context.WithValue(req.Context(), middleware.UserIDContextKeyForTests, userID)
	return req.WithContext(ctx)
}

func TestCreateDirect_Success(t *testing.T) {
	store := newFakeStore()
	store.users["alice"] = true
	store.users["bob"] = true
	svc := NewService(store)
	h := NewHandler(svc)

	body, _ := json.Marshal(CreateDirectChatRequest{OtherUserID: "bob"})
	req := httptest.NewRequest("POST", "/chats/direct", bytes.NewReader(body))
	req = withAuthenticatedUser(req, "alice")
	rec := httptest.NewRecorder()

	h.CreateDirect(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var got Chat
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if got.ID == "" {
		t.Fatal("expected a non-empty chat ID")
	}
}

func TestCreateDirect_UnknownUser_Returns404(t *testing.T) {
	store := newFakeStore()
	store.users["alice"] = true
	svc := NewService(store)
	h := NewHandler(svc)

	body, _ := json.Marshal(CreateDirectChatRequest{OtherUserID: "ghost"})
	req := httptest.NewRequest("POST", "/chats/direct", bytes.NewReader(body))
	req = withAuthenticatedUser(req, "alice")
	rec := httptest.NewRecorder()

	h.CreateDirect(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestCreateDirect_Self_Returns400(t *testing.T) {
	store := newFakeStore()
	store.users["alice"] = true
	svc := NewService(store)
	h := NewHandler(svc)

	body, _ := json.Marshal(CreateDirectChatRequest{OtherUserID: "alice"})
	req := httptest.NewRequest("POST", "/chats/direct", bytes.NewReader(body))
	req = withAuthenticatedUser(req, "alice")
	rec := httptest.NewRecorder()

	h.CreateDirect(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestCreateDirect_MalformedBody_Returns400(t *testing.T) {
	store := newFakeStore()
	svc := NewService(store)
	h := NewHandler(svc)

	req := httptest.NewRequest("POST", "/chats/direct", bytes.NewReader([]byte("not json")))
	req = withAuthenticatedUser(req, "alice")
	rec := httptest.NewRecorder()

	h.CreateDirect(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestList_ReturnsCallersChats(t *testing.T) {
	store := newFakeStore()
	store.users["alice"] = true
	store.users["bob"] = true
	svc := NewService(store)
	svc.CreateOrGetDirectChat(context.Background(), "alice", "bob")

	h := NewHandler(svc)

	req := httptest.NewRequest("GET", "/chats", nil)
	req = withAuthenticatedUser(req, "alice")
	rec := httptest.NewRecorder()

	h.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var chats []*Chat
	if err := json.NewDecoder(rec.Body).Decode(&chats); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(chats) != 1 {
		t.Fatalf("expected 1 chat, got %d", len(chats))
	}
}
