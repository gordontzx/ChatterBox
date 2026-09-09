package chat

import (
	"context"
	"errors"
	"testing"
)

// fakeStore is an in-memory Store implementation for testing Service
// without touching Postgres.
type fakeStore struct {
	users map[string]bool          // userID -> exists
	names map[string]string        // userID -> display name
	chats map[string]*Chat         // chatID -> Chat
}

func newFakeStore() *fakeStore {
	return &fakeStore{
		users: map[string]bool{},
		names: map[string]string{},
		chats: map[string]*Chat{},
	}
}

// seedUser registers a user as existing, with a display name, for use
// in FindDirectChat/CreateDirectChat's member construction.
func (f *fakeStore) seedUser(id, name string) {
	f.users[id] = true
	f.names[id] = name
}

func (f *fakeStore) UserExists(ctx context.Context, userID string) (bool, error) {
	return f.users[userID], nil
}

func (f *fakeStore) FindDirectChat(ctx context.Context, userA, userB string) (*Chat, error) {
	for _, c := range f.chats {
		if len(c.Members) != 2 {
			continue
		}
		hasA := c.Members[0].ID == userA || c.Members[1].ID == userA
		hasB := c.Members[0].ID == userB || c.Members[1].ID == userB
		if hasA && hasB {
			return c, nil
		}
	}
	return nil, nil
}

func (f *fakeStore) CreateDirectChat(ctx context.Context, userA, userB string) (*Chat, error) {
	id := "chat-" + userA + "-" + userB
	c := &Chat{
		ID: id,
		Members: []Member{
			{ID: userA, Name: f.names[userA]},
			{ID: userB, Name: f.names[userB]},
		},
	}
	f.chats[id] = c
	return c, nil
}

func (f *fakeStore) ListChatsForUser(ctx context.Context, userID string) ([]*Chat, error) {
	var result []*Chat
	for _, c := range f.chats {
		for _, m := range c.Members {
			if m.ID == userID {
				result = append(result, c)
				break
			}
		}
	}
	return result, nil
}

func TestCreateOrGetDirectChat_CreatesNewChat(t *testing.T) {
	store := newFakeStore()
	store.seedUser("alice", "Alice")
	store.seedUser("bob", "Bob")
	svc := NewService(store)

	c, err := svc.CreateOrGetDirectChat(context.Background(), "alice", "bob")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c == nil || len(c.Members) != 2 {
		t.Fatalf("expected a chat with 2 members, got %+v", c)
	}
}

func TestCreateOrGetDirectChat_IncludesMemberNames(t *testing.T) {
	store := newFakeStore()
	store.seedUser("alice", "Alice")
	store.seedUser("bob", "Bob")
	svc := NewService(store)

	c, err := svc.CreateOrGetDirectChat(context.Background(), "alice", "bob")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var gotBobName string
	for _, m := range c.Members {
		if m.ID == "bob" {
			gotBobName = m.Name
		}
	}
	if gotBobName != "Bob" {
		t.Fatalf("expected member name 'Bob', got %q", gotBobName)
	}
}

func TestCreateOrGetDirectChat_ReturnsExistingChat(t *testing.T) {
	store := newFakeStore()
	store.seedUser("alice", "Alice")
	store.seedUser("bob", "Bob")
	svc := NewService(store)
	ctx := context.Background()

	first, err := svc.CreateOrGetDirectChat(ctx, "alice", "bob")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	second, err := svc.CreateOrGetDirectChat(ctx, "alice", "bob")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if first.ID != second.ID {
		t.Fatalf("expected the same chat to be returned, got %s and %s", first.ID, second.ID)
	}
}

func TestCreateOrGetDirectChat_ReturnsExistingChat_ReverseOrder(t *testing.T) {
	// Calling with (bob, alice) should still find the chat created as (alice, bob).
	store := newFakeStore()
	store.seedUser("alice", "Alice")
	store.seedUser("bob", "Bob")
	svc := NewService(store)
	ctx := context.Background()

	first, _ := svc.CreateOrGetDirectChat(ctx, "alice", "bob")
	second, err := svc.CreateOrGetDirectChat(ctx, "bob", "alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if first.ID != second.ID {
		t.Fatalf("expected same chat regardless of argument order, got %s and %s", first.ID, second.ID)
	}
}

func TestCreateOrGetDirectChat_CannotChatWithSelf(t *testing.T) {
	store := newFakeStore()
	store.seedUser("alice", "Alice")
	svc := NewService(store)

	_, err := svc.CreateOrGetDirectChat(context.Background(), "alice", "alice")
	if !errors.Is(err, ErrCannotChatWithSelf) {
		t.Fatalf("expected ErrCannotChatWithSelf, got %v", err)
	}
}

func TestCreateOrGetDirectChat_OtherUserNotFound(t *testing.T) {
	store := newFakeStore()
	store.seedUser("alice", "Alice")
	// "bob" deliberately not seeded
	svc := NewService(store)

	_, err := svc.CreateOrGetDirectChat(context.Background(), "alice", "bob")
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestListChats_ReturnsOnlyCallersChats(t *testing.T) {
	store := newFakeStore()
	store.seedUser("alice", "Alice")
	store.seedUser("bob", "Bob")
	store.seedUser("carol", "Carol")
	svc := NewService(store)
	ctx := context.Background()

	svc.CreateOrGetDirectChat(ctx, "alice", "bob")
	svc.CreateOrGetDirectChat(ctx, "bob", "carol") // alice is not part of this one

	chats, err := svc.ListChats(ctx, "alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chats) != 1 {
		t.Fatalf("expected exactly 1 chat for alice, got %d", len(chats))
	}
}

func TestListChats_IncludesMembersWithNames(t *testing.T) {
	store := newFakeStore()
	store.seedUser("alice", "Alice")
	store.seedUser("bob", "Bob")
	svc := NewService(store)
	ctx := context.Background()

	svc.CreateOrGetDirectChat(ctx, "alice", "bob")

	chats, err := svc.ListChats(ctx, "alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chats) != 1 || len(chats[0].Members) != 2 {
		t.Fatalf("expected 1 chat with 2 members, got %+v", chats)
	}

	names := map[string]bool{}
	for _, m := range chats[0].Members {
		names[m.Name] = true
	}
	if !names["Alice"] || !names["Bob"] {
		t.Fatalf("expected both Alice and Bob in member names, got %+v", chats[0].Members)
	}
}

func TestListChats_NoChats_ReturnsEmpty(t *testing.T) {
	store := newFakeStore()
	svc := NewService(store)

	chats, err := svc.ListChats(context.Background(), "nobody")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chats) != 0 {
		t.Fatalf("expected no chats, got %d", len(chats))
	}
}
