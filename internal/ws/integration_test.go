package ws

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"ChatterBox/internal/auth"
)

func TestServeWS_AuthAndDelivery(t *testing.T) {
	secret := []byte("test-secret")
	hub := NewHub()
	handler := NewHandler(hub, secret)

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

	// give the server a brief moment to finish registering the client in the Hub
	// before we try to deliver to it (Register happens synchronously in ServeWS
	// before the goroutines start, so this is mostly a safety margin).
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

func TestServeWS_MissingToken_Rejected(t *testing.T) {
	hub := NewHub()
	handler := NewHandler(hub, []byte("test-secret"))

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
	handler := NewHandler(hub, []byte("test-secret"))

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
