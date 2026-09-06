package message

import (
	"context"
	"errors"
	"testing"
)

type fakeStore struct {
	members  map[string][]string // chatID -> member userIDs
	messages []Message
}

func newFakeStore() *fakeStore {
	return &fakeStore{members: map[string][]string{}}
}

func (f *fakeStore) IsChatMember(ctx context.Context, chatID, userID string) (bool, error) {
	for _, id := range f.members[chatID] {
		if id == userID {
			return true, nil
		}
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
