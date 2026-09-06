package ws

import (
	"testing"
	"time"
)

func TestHub_RegisterAndSendToUser(t *testing.T) {
	hub := NewHub()
	client := NewClient("user-1", nil, hub, nil)

	hub.Register(client)

	hub.SendToUser("user-1", []byte("hello"))

	select {
	case msg := <-client.send:
		if string(msg) != "hello" {
			t.Fatalf("expected 'hello', got %q", msg)
		}
	case <-time.After(time.Second):
		t.Fatal("expected message to be delivered to client.send, timed out")
	}
}

func TestHub_SendToUser_MultiDevice(t *testing.T) {
	hub := NewHub()
	phone := NewClient("user-1", nil, hub, nil)
	laptop := NewClient("user-1", nil, hub, nil)

	hub.Register(phone)
	hub.Register(laptop)

	hub.SendToUser("user-1", []byte("hi both devices"))

	for _, c := range []*Client{phone, laptop} {
		select {
		case msg := <-c.send:
			if string(msg) != "hi both devices" {
				t.Fatalf("expected message on every device, got %q", msg)
			}
		case <-time.After(time.Second):
			t.Fatal("expected message delivered to every registered device")
		}
	}
}

func TestHub_SendToUser_UnknownUser_NoOp(t *testing.T) {
	hub := NewHub()
	// no clients registered at all

	// should not panic or block, even though "ghost" has no connections
	hub.SendToUser("ghost", []byte("nobody home"))
}

func TestHub_SendToUser_OnlyTargetUser(t *testing.T) {
	hub := NewHub()
	alice := NewClient("alice", nil, hub, nil)
	bob := NewClient("bob", nil, hub, nil)

	hub.Register(alice)
	hub.Register(bob)

	hub.SendToUser("alice", []byte("for alice only"))

	select {
	case <-alice.send:
		// expected
	case <-time.After(time.Second):
		t.Fatal("expected alice to receive her message")
	}

	select {
	case msg := <-bob.send:
		t.Fatalf("bob should not have received anything, got %q", msg)
	case <-time.After(100 * time.Millisecond):
		// expected — bob's channel stays empty
	}
}

func TestHub_Unregister_RemovesClientAndClosesChannel(t *testing.T) {
	hub := NewHub()
	client := NewClient("user-1", nil, hub, nil)

	hub.Register(client)
	hub.Unregister(client)

	// after unregistering, sending should be a no-op (no receivers, no panic)
	hub.SendToUser("user-1", []byte("should go nowhere"))

	// the client's send channel should now be closed
	_, ok := <-client.send
	if ok {
		t.Fatal("expected client.send to be closed after Unregister")
	}
}

func TestHub_Unregister_LastDeviceRemovesUserEntirely(t *testing.T) {
	hub := NewHub()
	client := NewClient("user-1", nil, hub, nil)
	hub.Register(client)

	hub.mut.RLock()
	_, exists := hub.clients["user-1"]
	hub.mut.RUnlock()
	if !exists {
		t.Fatal("expected user-1 to be present after Register")
	}

	hub.Unregister(client)

	hub.mut.RLock()
	_, exists = hub.clients["user-1"]
	hub.mut.RUnlock()
	if exists {
		t.Fatal("expected user-1 entry to be removed once their last device disconnects")
	}
}

func TestHub_SendToUser_FullBuffer_DoesNotBlock(t *testing.T) {
	hub := NewHub()
	client := NewClient("user-1", nil, hub, nil)
	hub.Register(client)

	// fill the buffer completely
	for i := 0; i < sendBufferSize; i++ {
		hub.SendToUser("user-1", []byte("filler"))
	}

	// one more send should be dropped, not block — this call must return promptly
	done := make(chan struct{})
	go func() {
		hub.SendToUser("user-1", []byte("overflow"))
		close(done)
	}()

	select {
	case <-done:
		// expected — non-blocking drop worked
	case <-time.After(time.Second):
		t.Fatal("SendToUser blocked on a full buffer instead of dropping the message")
	}
}
