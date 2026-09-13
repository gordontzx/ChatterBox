package ws

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"ChatterBox/internal/auth"
)

// fakePresence records calls without needing a real presence.Service.
type fakePresence struct {
	onlineCalls  []string
	offlineCalls []string
}

func (f *fakePresence) UserWentOnline(ctx context.Context, userID string) {
	f.onlineCalls = append(f.onlineCalls, userID)
}

func (f *fakePresence) UserWentOffline(ctx context.Context, userID string) {
	f.offlineCalls = append(f.offlineCalls, userID)
}

func TestServeWS_AuthAndDelivery(t *testing.T) {
	secret := []byte("test-secret")
	hub := NewHub()
	handler := NewHandler(hub, secret, nil, nil) // no presence wired for this test

	server := httptest.NewServer(http.HandlerFunc(handler.ServeWS))
	defer server.Close()

	token, err := auth.IssueAccessToken("user-1", secret)
	if err != nil {
		t.Fatalf("failed to issue token: %v", err)
	}

	wsURL := strings.Replace(server.URL, "http://", "ws://", 1) + "/ws?token=" + url.QueryEscape(token)

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer conn.Close()

	time.Sleep(50 * time.Millisecond)

	hub.SendToUser("user-1", []byte("hello over the wire"))

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("expected to receive delivered message, got error: %v", err)
	}
	if string(msg) != "hello over the wire" {
		t.Fatalf("expected 'hello over the wire', got %q", msg)
	}
}

func TestServeWS_FiresPresenceOnConnectAndDisconnect(t *testing.T) {
	secret := []byte("test-secret")
	hub := NewHub()
	presence := &fakePresence{}
	handler := NewHandler(hub, secret, nil, presence)

	server := httptest.NewServer(http.HandlerFunc(handler.ServeWS))
	defer server.Close()

	token, _ := auth.IssueAccessToken("user-1", secret)
	wsURL := strings.Replace(server.URL, "http://", "ws://", 1) + "/ws?token=" + url.QueryEscape(token)

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}

	time.Sleep(50 * time.Millisecond)
	if len(presence.onlineCalls) != 1 || presence.onlineCalls[0] != "user-1" {
		t.Fatalf("expected UserCameOnline called once for user-1, got %v", presence.onlineCalls)
	}

	conn.Close()
	time.Sleep(100 * time.Millisecond) // give ReadPump's cleanup goroutine time to run

	if len(presence.offlineCalls) != 1 || presence.offlineCalls[0] != "user-1" {
		t.Fatalf("expected UserWentOffline called once for user-1, got %v", presence.offlineCalls)
	}
}

func TestServeWS_MissingToken_Rejected(t *testing.T) {
	hub := NewHub()
	handler := NewHandler(hub, []byte("test-secret"), nil, nil)

	server := httptest.NewServer(http.HandlerFunc(handler.ServeWS))
	defer server.Close()

	wsURL := strings.Replace(server.URL, "http://", "ws://", 1) + "/ws"

	_, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err == nil {
		t.Fatal("expected dial to fail without a token")
	}
	if resp == nil || resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %v", resp)
	}
}

func TestServeWS_InvalidToken_Rejected(t *testing.T) {
	hub := NewHub()
	handler := NewHandler(hub, []byte("test-secret"), nil, nil)

	server := httptest.NewServer(http.HandlerFunc(handler.ServeWS))
	defer server.Close()

	wsURL := strings.Replace(server.URL, "http://", "ws://", 1) + "/ws?token=garbage"

	_, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err == nil {
		t.Fatal("expected dial to fail with an invalid token")
	}
	if resp == nil || resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %v", resp)
	}
}
