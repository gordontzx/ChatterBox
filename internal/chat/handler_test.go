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

type fakeBroadcaster struct {
	sent map[string][][]byte
}

func newFakeBroadcaster() *fakeBroadcaster {
	return &fakeBroadcaster{sent: map[string][][]byte{}}
}

func (f *fakeBroadcaster) SendToUser(userID string, payload []byte) {
	f.sent[userID] = append(f.sent[userID], payload)
}

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
	broadcaster := newFakeBroadcaster()
	h := NewHandler(svc, broadcaster)

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
	broadcaster := newFakeBroadcaster()
	h := NewHandler(svc, broadcaster)

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
	broadcaster := newFakeBroadcaster()
	h := NewHandler(svc, broadcaster)

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
	broadcaster := newFakeBroadcaster()
	h := NewHandler(svc, broadcaster)

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

	broadcaster := newFakeBroadcaster()
	h := NewHandler(svc, broadcaster)

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

func TestCreateDirect_BroadcastsToOtherUser(t *testing.T) {
	store := newFakeStore()
	store.seedUser("alice", "Alice")
	store.seedUser("bob", "Bob")
	svc := NewService(store)
	broadcaster := newFakeBroadcaster()
	h := NewHandler(svc, broadcaster)

	body, _ := json.Marshal(CreateDirectChatRequest{OtherUserID: "bob"})
	req := httptest.NewRequest("POST", "/chats/direct", bytes.NewReader(body))
	req = withAuthenticatedUser(req, "alice")
	rec := httptest.NewRecorder()

	h.CreateDirect(rec, req)

	if len(broadcaster.sent["bob"]) != 1 {
		t.Fatalf("expected bob to receive 1 chat_created event, got %d", len(broadcaster.sent["bob"]))
	}
	if len(broadcaster.sent["alice"]) != 0 {
		t.Fatal("expected alice (the caller) NOT to receive her own broadcast")
	}
}

func TestAddMember_BroadcastsToExistingAndNewMember(t *testing.T) {
	store := newFakeStore()
	store.seedUser("alice", "Alice")
	store.seedUser("bob", "Bob")
	store.seedUser("carol", "Carol")
	svc := NewService(store)
	ctx := context.Background()
	group, _ := svc.CreateGroupChat(ctx, "alice", CreateGroupChatRequest{Name: "Trip", MemberIDs: []string{"bob"}})

	broadcaster := newFakeBroadcaster()
	h := NewHandler(svc, broadcaster)

	body, _ := json.Marshal(AddMemberRequest{UserID: "carol"})
	req := httptest.NewRequest("POST", "/chats/"+group.ID+"/members", bytes.NewReader(body))
	req.SetPathValue("chatID", group.ID)
	req = withAuthenticatedUser(req, "alice")
	rec := httptest.NewRecorder()

	h.AddMember(rec, req)

	if len(broadcaster.sent["bob"]) != 1 {
		t.Fatalf("expected bob (existing member) to be notified, got %d", len(broadcaster.sent["bob"]))
	}
	if len(broadcaster.sent["carol"]) != 1 {
		t.Fatalf("expected carol (newly added) to be notified, got %d", len(broadcaster.sent["carol"]))
	}
}

func TestLeave_BroadcastsToRemainingMembers(t *testing.T) {
	store := newFakeStore()
	store.seedUser("alice", "Alice")
	store.seedUser("bob", "Bob")
	svc := NewService(store)
	ctx := context.Background()
	group, _ := svc.CreateGroupChat(ctx, "alice", CreateGroupChatRequest{Name: "Trip", MemberIDs: []string{"bob"}})

	broadcaster := newFakeBroadcaster()
	h := NewHandler(svc, broadcaster)

	req := httptest.NewRequest("POST", "/chats/"+group.ID+"/leave", nil)
	req.SetPathValue("chatID", group.ID)
	req = withAuthenticatedUser(req, "bob") // bob leaves
	rec := httptest.NewRecorder()

	h.Leave(rec, req)

	if len(broadcaster.sent["alice"]) != 1 {
		t.Fatalf("expected alice (remaining member) to be notified, got %d", len(broadcaster.sent["alice"]))
	}
	if len(broadcaster.sent["bob"]) != 0 {
		t.Fatal("expected bob (who just left) NOT to be notified of his own departure")
	}
}
