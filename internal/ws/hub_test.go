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
		t.Fatal("expected message to be delivered, timed out")
	}
}

func TestHub_Register_ReportsFirstConnection(t *testing.T) {
	hub := NewHub()
	phone := NewClient("user-1", nil, hub, nil)
	laptop := NewClient("user-1", nil, hub, nil)

	isFirst := hub.Register(phone)
	if !isFirst {
		t.Fatal("expected first registration to report isFirstConnection = true")
	}

	isFirstAgain := hub.Register(laptop)
	if isFirstAgain {
		t.Fatal("expected second device's registration to report isFirstConnection = false")
	}
}

func TestHub_Unregister_ReportsFullyOffline_OnlyOnLastDevice(t *testing.T) {
	hub := NewHub()
	phone := NewClient("user-1", nil, hub, nil)
	laptop := NewClient("user-1", nil, hub, nil)

	hub.Register(phone)
	hub.Register(laptop)

	fullyOffline := hub.Unregister(phone)
	if fullyOffline {
		t.Fatal("expected NOT fully offline while laptop is still connected")
	}

	fullyOfflineNow := hub.Unregister(laptop)
	if !fullyOfflineNow {
		t.Fatal("expected fully offline after last device disconnects")
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
	case <-time.After(time.Second):
		t.Fatal("expected alice to receive her message")
	}

	select {
	case msg := <-bob.send:
		t.Fatalf("bob should not have received anything, got %q", msg)
	case <-time.After(100 * time.Millisecond):
	}
}

func TestHub_Unregister_ClosesChannel(t *testing.T) {
	hub := NewHub()
	client := NewClient("user-1", nil, hub, nil)

	hub.Register(client)
	hub.Unregister(client)

	hub.SendToUser("user-1", []byte("should go nowhere"))

	_, ok := <-client.send
	if ok {
		t.Fatal("expected client.send to be closed after Unregister")
	}
}

func TestHub_IsOnline(t *testing.T) {
	hub := NewHub()
	client := NewClient("user-1", nil, hub, nil)

	if hub.IsOnline("user-1") {
		t.Fatal("expected user-1 to be offline before registering")
	}

	hub.Register(client)
	if !hub.IsOnline("user-1") {
		t.Fatal("expected user-1 to be online after registering")
	}

	hub.Unregister(client)
	if hub.IsOnline("user-1") {
		t.Fatal("expected user-1 to be offline after unregistering")
	}
}

func TestHub_SendToUser_FullBuffer_DoesNotBlock(t *testing.T) {
	hub := NewHub()
	client := NewClient("user-1", nil, hub, nil)
	hub.Register(client)

	for i := 0; i < sendBufferSize; i++ {
		hub.SendToUser("user-1", []byte("filler"))
	}

	done := make(chan struct{})
	go func() {
		hub.SendToUser("user-1", []byte("overflow"))
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("SendToUser blocked on a full buffer instead of dropping the message")
	}
}
