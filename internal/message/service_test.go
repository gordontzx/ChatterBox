package message

import (
	"context"
	"errors"
	"slices"
	"testing"
)

type fakeStore struct {
	members  map[string][]string // chatID -> member userIDs
	messages []Message
	receipts map[string]map[string]int64
}

func newFakeStore() *fakeStore {
	return &fakeStore{members: map[string][]string{}}
}

func (f *fakeStore) IsChatMember(ctx context.Context, chatID, userID string) (bool, error) {
	if slices.Contains(f.members[chatID], userID) {
		return true, nil
	}
	return false, nil
}

func (f *fakeStore) OtherChatMemberIDs(ctx context.Context, chatID, excludeUserID string) ([]string, error) {
	var others []string
	for _, id := range f.members[chatID] {
		if id != excludeUserID {
			others = append(others, id)
		}
	}
	return others, nil
}

func (f *fakeStore) CreateMessage(ctx context.Context, chatID, senderID, content string) (*Message, error) {
	m := Message{
		ID:       "msg-1",
		ChatID:   chatID,
		SenderID: senderID,
		Seq:      int64(len(f.messages) + 1),
		Content:  content,
	}
	f.messages = append(f.messages, m)
	return &m, nil
}

func (f *fakeStore) GetMessagesSince(ctx context.Context, chatID string, sinceSeq int64, limit int) ([]Message, error) {
	var result []Message
	for _, m := range f.messages {
		if m.ChatID == chatID && m.Seq > sinceSeq {
			if len(result) >= limit {
				break
			}
			result = append(result, m)
		}
	}
	return result, nil
}

func (f *fakeStore) UpsertReadReceipt(ctx context.Context, chatID, userID string, upToSeq int64) error {
	if f.receipts == nil {
		f.receipts = map[string]map[string]int64{}
	}
	if f.receipts[chatID] == nil {
		f.receipts[chatID] = map[string]int64{}
	}
	if upToSeq > f.receipts[chatID][userID] {
		f.receipts[chatID][userID] = upToSeq
	}
	return nil
}

func (f *fakeStore) GetReadReceiptsForChat(ctx context.Context, chatID string) ([]ReadReceipt, error) {
	var result []ReadReceipt
	for userID, seq := range f.receipts[chatID] {
		result = append(result, ReadReceipt{UserID: userID, LastReadSeq: seq})
	}
	return result, nil
}

func TestSendMessage_Success(t *testing.T) {
	store := newFakeStore()
	store.members["chat-1"] = []string{"alice", "bob"}
	svc := NewService(store)

	result, err := svc.SendMessage(context.Background(), "alice", "chat-1", "hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Message.Content != "hello" {
		t.Fatalf("expected content 'hello', got %q", result.Message.Content)
	}
	if len(result.OtherMemberIDs) != 1 || result.OtherMemberIDs[0] != "bob" {
		t.Fatalf("expected [bob], got %v", result.OtherMemberIDs)
	}
}

func TestSendMessage_NotAMember(t *testing.T) {
	store := newFakeStore()
	store.members["chat-1"] = []string{"alice", "bob"}
	svc := NewService(store)

	_, err := svc.SendMessage(context.Background(), "carol", "chat-1", "hello")
	if !errors.Is(err, ErrNotChatMember) {
		t.Fatalf("expected ErrNotChatMember, got %v", err)
	}
}

func TestSendMessage_EmptyContent(t *testing.T) {
	store := newFakeStore()
	store.members["chat-1"] = []string{"alice", "bob"}
	svc := NewService(store)

	_, err := svc.SendMessage(context.Background(), "alice", "chat-1", "   ")
	if !errors.Is(err, ErrEmptyContent) {
		t.Fatalf("expected ErrEmptyContent, got %v", err)
	}
}

func TestSendMessage_ContentTooLong(t *testing.T) {
	store := newFakeStore()
	store.members["chat-1"] = []string{"alice", "bob"}
	svc := NewService(store)

	longContent := make([]byte, maxContentLength+1)
	for i := range longContent {
		longContent[i] = 'a'
	}

	_, err := svc.SendMessage(context.Background(), "alice", "chat-1", string(longContent))
	if !errors.Is(err, ErrContentTooLong) {
		t.Fatalf("expected ErrContentTooLong, got %v", err)
	}
}

func TestGetMessagesSince_ReturnsOnlyNewer(t *testing.T) {
	store := newFakeStore()
	store.members["chat-1"] = []string{"alice", "bob"}
	svc := NewService(store)
	ctx := context.Background()

	svc.SendMessage(ctx, "alice", "chat-1", "first")
	svc.SendMessage(ctx, "bob", "chat-1", "second")
	svc.SendMessage(ctx, "alice", "chat-1", "third")

	messages, err := svc.GetMessagesSince(ctx, "alice", "chat-1", 1, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(messages) != 2 {
		t.Fatalf("expected 2 messages (seq 2 and 3), got %d", len(messages))
	}
	if messages[0].Content != "second" || messages[1].Content != "third" {
		t.Fatalf("expected messages in order, got %+v", messages)
	}
}

func TestGetMessagesSince_NonMember_Forbidden(t *testing.T) {
	store := newFakeStore()
	store.members["chat-1"] = []string{"alice", "bob"}
	svc := NewService(store)

	_, err := svc.GetMessagesSince(context.Background(), "carol", "chat-1", 0, 0)
	if !errors.Is(err, ErrNotChatMember) {
		t.Fatalf("expected ErrNotChatMember, got %v", err)
	}
}

func TestGetMessagesSince_SinceZero_ReturnsAll(t *testing.T) {
	store := newFakeStore()
	store.members["chat-1"] = []string{"alice", "bob"}
	svc := NewService(store)
	ctx := context.Background()

	svc.SendMessage(ctx, "alice", "chat-1", "one")
	svc.SendMessage(ctx, "bob", "chat-1", "two")

	messages, err := svc.GetMessagesSince(ctx, "alice", "chat-1", 0, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(messages) != 2 {
		t.Fatalf("expected all 2 messages, got %d", len(messages))
	}
}

func TestGetMessagesSince_DefaultLimit_WhenZero(t *testing.T) {
	store := newFakeStore()
	store.members["chat-1"] = []string{"alice", "bob"}
	svc := NewService(store)
	ctx := context.Background()

	for range 200 {
		svc.SendMessage(ctx, "alice", "chat-1", "msg")
	}

	messages, err := svc.GetMessagesSince(ctx, "alice", "chat-1", 0, 0) // limit=0 → default
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(messages) != defaultMessageLimit {
		t.Fatalf("expected %d messages (default limit), got %d", defaultMessageLimit, len(messages))
	}
}

func TestGetMessagesSince_RespectsExplicitLimit(t *testing.T) {
	store := newFakeStore()
	store.members["chat-1"] = []string{"alice", "bob"}
	svc := NewService(store)
	ctx := context.Background()

	for range 10 {
		svc.SendMessage(ctx, "alice", "chat-1", "msg")
	}

	messages, err := svc.GetMessagesSince(ctx, "alice", "chat-1", 0, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(messages) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(messages))
	}
}

func TestGetMessagesSince_ClampsExcessiveLimit(t *testing.T) {
	store := newFakeStore()
	store.members["chat-1"] = []string{"alice", "bob"}
	svc := NewService(store)
	ctx := context.Background()

	for range 250 {
		svc.SendMessage(ctx, "alice", "chat-1", "msg")
	}

	messages, err := svc.GetMessagesSince(ctx, "alice", "chat-1", 0, 999999)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(messages) != maxMessageLimit {
		t.Fatalf("expected clamped to %d messages, got %d", maxMessageLimit, len(messages))
	}
}

func TestSendTypingIndicator_Success(t *testing.T) {
	store := newFakeStore()
	store.members["chat-1"] = []string{"alice", "bob"}
	svc := NewService(store)

	others, err := svc.SendTypingIndicator(context.Background(), "alice", "chat-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(others) != 1 || others[0] != "bob" {
		t.Fatalf("expected [bob], got %v", others)
	}
}

func TestSendTypingIndicator_NotAMember(t *testing.T) {
	store := newFakeStore()
	store.members["chat-1"] = []string{"alice", "bob"}
	svc := NewService(store)

	_, err := svc.SendTypingIndicator(context.Background(), "carol", "chat-1")
	if !errors.Is(err, ErrNotChatMember) {
		t.Fatalf("expected ErrNotChatMember, got %v", err)
	}
}
