package presence

import (
	"context"
	"testing"
	"time"
)

type fakeStore struct {
	lastSeen map[string]*time.Time
}

func newFakeStore() *fakeStore {
	return &fakeStore{lastSeen: map[string]*time.Time{}}
}

func (f *fakeStore) UpdateLastSeen(ctx context.Context, userID string, at time.Time) error {
	f.lastSeen[userID] = &at
	return nil
}

func (f *fakeStore) GetLastSeen(ctx context.Context, userID string) (*time.Time, error) {
	return f.lastSeen[userID], nil
}

type fakeContacts struct {
	contactsByUser map[string][]string
}

func (f *fakeContacts) ListContactIDs(ctx context.Context, userID string) ([]string, error) {
	return f.contactsByUser[userID], nil
}

type fakeBroadcaster struct {
	sent map[string][][]byte
}

func newFakeBroadcaster() *fakeBroadcaster {
	return &fakeBroadcaster{sent: map[string][][]byte{}}
}

func (f *fakeBroadcaster) SendToUser(userID string, payload []byte) {
	f.sent[userID] = append(f.sent[userID], payload)
}

func TestGetPresence_Online(t *testing.T) {
	store := newFakeStore()
	svc := NewService(store, &fakeContacts{}, newFakeBroadcaster())

	info, err := svc.GetPresence(context.Background(), "alice", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !info.Online || info.LastSeenAt != nil {
		t.Fatalf("expected online with no last_seen_at, got %+v", info)
	}
}

func TestGetPresence_Offline_WithLastSeen(t *testing.T) {
	store := newFakeStore()
	then := time.Now().Add(-time.Hour)
	store.lastSeen["alice"] = &then
	svc := NewService(store, &fakeContacts{}, newFakeBroadcaster())

	info, err := svc.GetPresence(context.Background(), "alice", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Online || info.LastSeenAt == nil {
		t.Fatalf("expected offline with a last_seen_at, got %+v", info)
	}
}

func TestUserWentOffline_PersistsAndNotifiesContacts(t *testing.T) {
	store := newFakeStore()
	contacts := &fakeContacts{contactsByUser: map[string][]string{"alice": {"bob", "carol"}}}
	broadcaster := newFakeBroadcaster()
	svc := NewService(store, contacts, broadcaster)

	svc.UserWentOffline(context.Background(), "alice")

	if store.lastSeen["alice"] == nil {
		t.Fatal("expected last_seen_at to be persisted for alice")
	}
	if len(broadcaster.sent["bob"]) != 1 || len(broadcaster.sent["carol"]) != 1 {
		t.Fatalf("expected both contacts to be notified, got %+v", broadcaster.sent)
	}
}

func TestUserCameOnline_NotifiesContactsOnly_NoPersistence(t *testing.T) {
	store := newFakeStore()
	contacts := &fakeContacts{contactsByUser: map[string][]string{"alice": {"bob"}}}
	broadcaster := newFakeBroadcaster()
	svc := NewService(store, contacts, broadcaster)

	svc.UserWentOnline(context.Background(), "alice")

	if store.lastSeen["alice"] != nil {
		t.Fatal("expected no last_seen_at write when coming online")
	}
	if len(broadcaster.sent["bob"]) != 1 {
		t.Fatalf("expected bob to be notified, got %+v", broadcaster.sent)
	}
}
