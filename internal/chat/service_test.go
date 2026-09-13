package chat

import (
	"context"
	"errors"
	"testing"
)

// fakeStore is an in-memory Store implementation for testing Service
// without touching Postgres.
type fakeStore struct {
	users map[string]bool   // userID -> exists
	names map[string]string // userID -> display name
	chats map[string]*Chat  // chatID -> Chat
	nextID int
}

func newFakeStore() *fakeStore {
	return &fakeStore{
		users: map[string]bool{},
		names: map[string]string{},
		chats: map[string]*Chat{},
	}
}

func (f *fakeStore) seedUser(id, name string) {
	f.users[id] = true
	f.names[id] = name
}

func (f *fakeStore) newChatID() string {
	f.nextID++
	return "chat-" + string(rune('a'+f.nextID))
}

func (f *fakeStore) buildMembers(ids []string) []Member {
	members := make([]Member, 0, len(ids))
	for _, id := range ids {
		members = append(members, Member{ID: id, Name: f.names[id]})
	}
	return members
}

func (f *fakeStore) UserExists(ctx context.Context, userID string) (bool, error) {
	return f.users[userID], nil
}

func (f *fakeStore) IsChatMember(ctx context.Context, chatID, userID string) (bool, error) {
	c, ok := f.chats[chatID]
	if !ok {
		return false, nil
	}
	for _, m := range c.Members {
		if m.ID == userID {
			return true, nil
		}
	}
	return false, nil
}

func (f *fakeStore) GetChat(ctx context.Context, chatID string) (*Chat, error) {
	return f.chats[chatID], nil // nil, nil if not found — matches real store's contract
}

func (f *fakeStore) FindDirectChat(ctx context.Context, userA, userB string) (*Chat, error) {
	for _, c := range f.chats {
		if c.Type != ChatTypeDirect || len(c.Members) != 2 {
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
	id := f.newChatID()
	c := &Chat{
		ID:      id,
		Type:    ChatTypeDirect,
		Members: f.buildMembers([]string{userA, userB}),
	}
	f.chats[id] = c
	return c, nil
}

func (f *fakeStore) CreateGroupChat(ctx context.Context, name string, memberIDs []string) (*Chat, error) {
	id := f.newChatID()
	nameCopy := name
	c := &Chat{
		ID:      id,
		Type:    ChatTypeGroup,
		Name:    &nameCopy,
		Members: f.buildMembers(memberIDs),
	}
	f.chats[id] = c
	return c, nil
}

func (f *fakeStore) AddMember(ctx context.Context, chatID, userID string) error {
	c, ok := f.chats[chatID]
	if !ok {
		return nil // matches real store's ON CONFLICT DO NOTHING no-op behavior on missing chat
	}
	for _, m := range c.Members {
		if m.ID == userID {
			return nil // already a member — idempotent no-op
		}
	}
	c.Members = append(c.Members, Member{ID: userID, Name: f.names[userID]})
	return nil
}

func (f *fakeStore) RemoveMember(ctx context.Context, chatID, userID string) error {
	c, ok := f.chats[chatID]
	if !ok {
		return nil
	}
	var filtered []Member
	for _, m := range c.Members {
		if m.ID != userID {
			filtered = append(filtered, m)
		}
	}
	c.Members = filtered
	return nil
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

func (f *fakeStore) ShareAnyChat(ctx context.Context, user1, user2 string) (bool, error) {
	for _, c := range f.chats {
		var contains1, contains2 bool
		for _, m := range c.Members {
			if m.ID == user1 {
				contains1 = true
			}
			if m.ID == user2 {
				contains2 = true
			}
		}
		if contains1 && contains2 {
			return true, nil
		}
	}
	return false, nil
}

// ---------- Direct chat tests ----------

func TestCreateOrGetDirectChat_CreatesNewChat(t *testing.T) {
	store := newFakeStore()
	store.seedUser("alice", "Alice")
	store.seedUser("bob", "Bob")
	svc := NewService(store)

	c, wasCreated, err := svc.CreateOrGetDirectChat(context.Background(), "alice", "bob")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !wasCreated {
		t.Fatal("expected wasCreated = true for a brand new chat")
	}
	if c.Type != ChatTypeDirect {
		t.Fatalf("expected direct chat type, got %s", c.Type)
	}
	if len(c.Members) != 2 {
		t.Fatalf("expected a chat with 2 members, got %+v", c)
	}
}

func TestCreateOrGetDirectChat_IncludesMemberNames(t *testing.T) {
	store := newFakeStore()
	store.seedUser("alice", "Alice")
	store.seedUser("bob", "Bob")
	svc := NewService(store)

	c, _, err := svc.CreateOrGetDirectChat(context.Background(), "alice", "bob")
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

	first, wasCreatedFirst, err := svc.CreateOrGetDirectChat(ctx, "alice", "bob")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !wasCreatedFirst {
		t.Fatal("expected first call to report wasCreated = true")
	}

	second, wasCreatedSecond, err := svc.CreateOrGetDirectChat(ctx, "alice", "bob")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if wasCreatedSecond {
		t.Fatal("expected second call to report wasCreated = false")
	}
	if first.ID != second.ID {
		t.Fatalf("expected the same chat to be returned, got %s and %s", first.ID, second.ID)
	}
}

func TestCreateOrGetDirectChat_ReturnsExistingChat_ReverseOrder(t *testing.T) {
	store := newFakeStore()
	store.seedUser("alice", "Alice")
	store.seedUser("bob", "Bob")
	svc := NewService(store)
	ctx := context.Background()

	first, _, _ := svc.CreateOrGetDirectChat(ctx, "alice", "bob")
	second, _, err := svc.CreateOrGetDirectChat(ctx, "bob", "alice")
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

	_, _, err := svc.CreateOrGetDirectChat(context.Background(), "alice", "alice")
	if !errors.Is(err, ErrCannotChatWithSelf) {
		t.Fatalf("expected ErrCannotChatWithSelf, got %v", err)
	}
}

func TestCreateOrGetDirectChat_OtherUserNotFound(t *testing.T) {
	store := newFakeStore()
	store.seedUser("alice", "Alice")
	svc := NewService(store)

	_, _, err := svc.CreateOrGetDirectChat(context.Background(), "alice", "bob")
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestCreateOrGetDirectChat_DoesNotCollideWithGroupChatOfSameTwoUsers(t *testing.T) {
	// A 2-person group chat between alice and bob must NOT be returned
	// when alice asks for her *direct* chat with bob.
	store := newFakeStore()
	store.seedUser("alice", "Alice")
	store.seedUser("bob", "Bob")
	svc := NewService(store)
	ctx := context.Background()

	group, err := svc.CreateGroupChat(ctx, "alice", CreateGroupChatRequest{
		Name:      "Just us two, but a group",
		MemberIDs: []string{"bob"},
	})
	if err != nil {
		t.Fatalf("unexpected error creating group: %v", err)
	}

	direct, _, err := svc.CreateOrGetDirectChat(ctx, "alice", "bob")
	if err != nil {
		t.Fatalf("unexpected error creating direct: %v", err)
	}

	if direct.ID == group.ID {
		t.Fatal("expected direct chat and group chat to be distinct, got the same chat")
	}
	if direct.Type != ChatTypeDirect {
		t.Fatalf("expected direct chat type, got %s", direct.Type)
	}
}

// ---------- Group chat tests ----------

func TestCreateGroupChat_Success(t *testing.T) {
	store := newFakeStore()
	store.seedUser("alice", "Alice")
	store.seedUser("bob", "Bob")
	store.seedUser("carol", "Carol")
	svc := NewService(store)

	c, err := svc.CreateGroupChat(context.Background(), "alice", CreateGroupChatRequest{
		Name:      "Weekend Trip",
		MemberIDs: []string{"bob", "carol"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Type != ChatTypeGroup {
		t.Fatalf("expected group chat type, got %s", c.Type)
	}
	if c.Name == nil || *c.Name != "Weekend Trip" {
		t.Fatalf("expected name 'Weekend Trip', got %+v", c.Name)
	}
	if len(c.Members) != 3 {
		t.Fatalf("expected 3 members (alice, bob, carol), got %d", len(c.Members))
	}
}

func TestCreateGroupChat_AllowsTwoTotalMembers(t *testing.T) {
	store := newFakeStore()
	store.seedUser("alice", "Alice")
	store.seedUser("bob", "Bob")
	svc := NewService(store)

	c, err := svc.CreateGroupChat(context.Background(), "alice", CreateGroupChatRequest{
		Name:      "Just us two",
		MemberIDs: []string{"bob"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(c.Members) != 2 {
		t.Fatalf("expected 2 total members, got %d", len(c.Members))
	}
}

func TestCreateGroupChat_RequiresAtLeastOneOtherMember(t *testing.T) {
	store := newFakeStore()
	store.seedUser("alice", "Alice")
	svc := NewService(store)

	_, err := svc.CreateGroupChat(context.Background(), "alice", CreateGroupChatRequest{
		Name:      "Solo group?",
		MemberIDs: []string{},
	})
	if !errors.Is(err, ErrGroupTooSmall) {
		t.Fatalf("expected ErrGroupTooSmall, got %v", err)
	}
}

func TestCreateGroupChat_RequiresName(t *testing.T) {
	store := newFakeStore()
	store.seedUser("alice", "Alice")
	store.seedUser("bob", "Bob")
	svc := NewService(store)

	_, err := svc.CreateGroupChat(context.Background(), "alice", CreateGroupChatRequest{
		Name:      "   ", // whitespace only
		MemberIDs: []string{"bob"},
	})
	if !errors.Is(err, ErrGroupNameRequired) {
		t.Fatalf("expected ErrGroupNameRequired, got %v", err)
	}
}

func TestCreateGroupChat_UnknownMember(t *testing.T) {
	store := newFakeStore()
	store.seedUser("alice", "Alice")
	svc := NewService(store)

	_, err := svc.CreateGroupChat(context.Background(), "alice", CreateGroupChatRequest{
		Name:      "Trip",
		MemberIDs: []string{"ghost"},
	})
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestCreateGroupChat_DeduplicatesAndExcludesCallerFromInput(t *testing.T) {
	store := newFakeStore()
	store.seedUser("alice", "Alice")
	store.seedUser("bob", "Bob")
	svc := NewService(store)

	// alice includes herself and duplicates bob in the request — should
	// still just end up with [alice, bob], not error or double-add bob.
	c, err := svc.CreateGroupChat(context.Background(), "alice", CreateGroupChatRequest{
		Name:      "Trip",
		MemberIDs: []string{"alice", "bob", "bob"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(c.Members) != 2 {
		t.Fatalf("expected exactly 2 deduplicated members, got %d: %+v", len(c.Members), c.Members)
	}
}

// ---------- AddMember tests ----------

func TestAddMember_Success(t *testing.T) {
	store := newFakeStore()
	store.seedUser("alice", "Alice")
	store.seedUser("bob", "Bob")
	store.seedUser("carol", "Carol")
	svc := NewService(store)
	ctx := context.Background()

	group, _ := svc.CreateGroupChat(ctx, "alice", CreateGroupChatRequest{
		Name:      "Trip",
		MemberIDs: []string{"bob"},
	})

	err := svc.AddMember(ctx, "alice", group.ID, "carol")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	updated, _ := store.GetChat(ctx, group.ID)
	if len(updated.Members) != 3 {
		t.Fatalf("expected 3 members after add, got %d", len(updated.Members))
	}
}

func TestAddMember_NonMemberCannotAdd(t *testing.T) {
	store := newFakeStore()
	store.seedUser("alice", "Alice")
	store.seedUser("bob", "Bob")
	store.seedUser("mallory", "Mallory")
	svc := NewService(store)
	ctx := context.Background()

	group, _ := svc.CreateGroupChat(ctx, "alice", CreateGroupChatRequest{
		Name:      "Trip",
		MemberIDs: []string{"bob"},
	})

	err := svc.AddMember(ctx, "mallory", group.ID, "mallory")
	if !errors.Is(err, ErrNotChatMember) {
		t.Fatalf("expected ErrNotChatMember, got %v", err)
	}
}

func TestAddMember_CannotAddToDirectChat(t *testing.T) {
	store := newFakeStore()
	store.seedUser("alice", "Alice")
	store.seedUser("bob", "Bob")
	store.seedUser("carol", "Carol")
	svc := NewService(store)
	ctx := context.Background()

	direct, _, _ := svc.CreateOrGetDirectChat(ctx, "alice", "bob")

	err := svc.AddMember(ctx, "alice", direct.ID, "carol")
	if !errors.Is(err, ErrNotGroupChat) {
		t.Fatalf("expected ErrNotGroupChat, got %v", err)
	}
}

func TestAddMember_AlreadyMember(t *testing.T) {
	store := newFakeStore()
	store.seedUser("alice", "Alice")
	store.seedUser("bob", "Bob")
	svc := NewService(store)
	ctx := context.Background()

	group, _ := svc.CreateGroupChat(ctx, "alice", CreateGroupChatRequest{
		Name:      "Trip",
		MemberIDs: []string{"bob"},
	})

	err := svc.AddMember(ctx, "alice", group.ID, "bob")
	if !errors.Is(err, ErrAlreadyMember) {
		t.Fatalf("expected ErrAlreadyMember, got %v", err)
	}
}

func TestAddMember_UnknownUser(t *testing.T) {
	store := newFakeStore()
	store.seedUser("alice", "Alice")
	store.seedUser("bob", "Bob")
	svc := NewService(store)
	ctx := context.Background()

	group, _ := svc.CreateGroupChat(ctx, "alice", CreateGroupChatRequest{
		Name:      "Trip",
		MemberIDs: []string{"bob"},
	})

	err := svc.AddMember(ctx, "alice", group.ID, "ghost")
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

// ---------- ListChats tests ----------

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

func TestListChats_IncludesBothDirectAndGroupChats(t *testing.T) {
	store := newFakeStore()
	store.seedUser("alice", "Alice")
	store.seedUser("bob", "Bob")
	store.seedUser("carol", "Carol")
	svc := NewService(store)
	ctx := context.Background()

	svc.CreateOrGetDirectChat(ctx, "alice", "bob")
	svc.CreateGroupChat(ctx, "alice", CreateGroupChatRequest{Name: "Trip", MemberIDs: []string{"bob", "carol"}})

	chats, err := svc.ListChats(ctx, "alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chats) != 2 {
		t.Fatalf("expected 2 chats for alice, got %d", len(chats))
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

func TestLeaveChat_CannotLeaveDirectChat(t *testing.T) {
	store := newFakeStore()
	store.seedUser("alice", "Alice")
	store.seedUser("bob", "Bob")
	svc := NewService(store)
	ctx := context.Background()

	direct, _, _ := svc.CreateOrGetDirectChat(ctx, "alice", "bob")

	err := svc.LeaveChat(ctx, "alice", direct.ID)
	if !errors.Is(err, ErrNotGroupChat) {
		t.Fatalf("expected ErrNotGroupChat, got %v", err)
	}
}

func TestLeaveChat_NonMemberCannotLeave(t *testing.T) {
	store := newFakeStore()
	store.seedUser("alice", "Alice")
	store.seedUser("bob", "Bob")
	store.seedUser("mallory", "Mallory")
	svc := NewService(store)
	ctx := context.Background()

	group, _ := svc.CreateGroupChat(ctx, "alice", CreateGroupChatRequest{
		Name:      "Trip",
		MemberIDs: []string{"bob"},
	})

	err := svc.LeaveChat(ctx, "mallory", group.ID)
	if !errors.Is(err, ErrNotChatMember) {
		t.Fatalf("expected ErrNotChatMember, got %v", err)
	}
}
