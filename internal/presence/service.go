package presence

import (
	"context"
	"encoding/json"
	"log"
	"time"
)

// chat.service
type ContactsLookup interface {
	ListContactIDs(ctx context.Context, userID string) ([]string, error)
}

// ws.hub
type Broadcaster interface {
	SendToUser(userID string, payload []byte)
}

type Service struct {
	store	 Store
	contacts ContactsLookup
	hub 	 Broadcaster
}

func NewService(store Store, contacts ContactsLookup, hub Broadcaster) *Service {
	return &Service{store: store, contacts: contacts, hub: hub}
}

func (s *Service) GetPresence(ctx context.Context, userID string, isOnline bool) (*Info, error) {
	if isOnline {
		return &Info{UserID: userID, Online: true}, nil
	}

	lastSeen, err := s.store.GetLastSeen(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &Info{UserID: userID, Online: false, LastSeenAt: lastSeen}, nil
}

func (s *Service) UserWentOffline(ctx context.Context, userID string) {
	now := time.Now()
	if err := s.store.UpdateLastSeen(ctx, userID, now); err != nil {
		log.Printf("failed to update last_seen for user %s: %v", userID, err)
	}
	s.broadcast(ctx, userID, false, &now)
}

func (s *Service) UserWentOnline(ctx context.Context, userID string) {
	s.broadcast(ctx, userID, true, nil)
}

func (s *Service) broadcast(ctx context.Context, userID string, online bool, lastSeen *time.Time) {
	contactIDs, err := s.contacts.ListContactIDs(ctx, userID)
	if err != nil {
		log.Printf("failed to list contacts for presence broadcast, user %s: %v", userID, err)
		return
	}

	event := Event{Type: "presence", UserID: userID, Online: online, LastSeenAt: lastSeen}
	payload, err := json.Marshal(event)
	if err != nil {
		log.Printf("failed to marshal presence event: %v", err)
		return
	}

	for _, contactID := range contactIDs {
		s.hub.SendToUser(contactID, payload)
	}
}
