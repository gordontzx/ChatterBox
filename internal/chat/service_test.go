package chat

import (
	"context"
	"errors"
	"testing"
)

// fakeStore is an in-memory Store implementation for testing Service
// without touching Postgres.
type fakeStore struct {
	users map[string]bool           // userID -> exists
	chats map[string]*Chat          // chatID -> Chat
}

func newFakeStore() *fakeStore {
	return &fakeStore{
		users: map[string]bool{},
		chats: map[string]*Chat{},
	}
}

func (f *fakeStore) UserExists(ctx context.Context, userID string) (bool, error) {
	return f.users[userID], nil
}

func (f *fakeStore) FindDirectChat(ctx context.Context, userA, userB string) (*Chat, error) {
	for _, c := range f.chats {
		if len(c.MemberIDs) != 2 {
			continue
		}
		hasA := c.MemberIDs[0] == userA || c.MemberIDs[1] == userA
		hasB := c.MemberIDs[0] == userB || c.MemberIDs[1] == userB
		if hasA && hasB {
			return c, nil
		}
	}
	return nil, nil
}

func (f *fakeStore) CreateDirectChat(ctx context.Context, userA, userB string) (*Chat, error) {
	id := "chat-" + userA + "-" + userB
	c := &Chat{ID: id, MemberIDs: []string{userA, userB}}
	f.chats[id] = c
	return c, nil
}

func (f *fakeStore) ListChatsForUser(ctx context.Context, userID string) ([]*Chat, error) {
	var result []*Chat
	for _, c := range f.chats {
		for _, m := range c.MemberIDs {
			if m == userID {
				result = append(result, c)
				break
			}
		}
	}
	return result, nil
}

func TestCreateOrGetDirectChat_CreatesNewChat(t *testing.T) {
	store := newFakeStore()
	store.users["alice"] = true
	store.users["bob"] = true
	svc := NewService(store)

	c, err := svc.CreateOrGetDirectChat(context.Background(), "alice", "bob")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c == nil || len(c.MemberIDs) != 2 {
		t.Fatalf("expected a chat with 2 members, got %+v", c)
	}
}

func TestCreateOrGetDirectChat_ReturnsExistingChat(t *testing.T) {
	store := newFakeStore()
	store.users["alice"] = true
	store.users["bob"] = true
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
	store.users["alice"] = true
	store.users["bob"] = true
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
	store.users["alice"] = true
	svc := NewService(store)

	_, err := svc.CreateOrGetDirectChat(context.Background(), "alice", "alice")
	if !errors.Is(err, ErrCannotChatWithSelf) {
		t.Fatalf("expected ErrCannotChatWithSelf, got %v", err)
	}
}

func TestCreateOrGetDirectChat_OtherUserNotFound(t *testing.T) {
	store := newFakeStore()
	store.users["alice"] = true
	// "bob" deliberately not added to store.users
	svc := NewService(store)

	_, err := svc.CreateOrGetDirectChat(context.Background(), "alice", "bob")
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestListChats_ReturnsOnlyCallersChats(t *testing.T) {
	store := newFakeStore()
	store.users["alice"] = true
	store.users["bob"] = true
	store.users["carol"] = true
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
