package store

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	wsTicketLength   = 32
	wsTicketLifetime = 30 * time.Second // Short-lived for security
)

// WSTicket represents a one-time WebSocket authentication ticket
type WSTicket struct {
	Ticket    string
	UserID    uuid.UUID
	BaseID    uuid.UUID
	ExpiresAt time.Time
	Used      bool
}

// WSTicketStore manages WebSocket authentication tickets
type WSTicketStore struct {
	mu      sync.Mutex
	tickets map[string]*WSTicket
}

// NewWSTicketStore creates a new WebSocket ticket store
func NewWSTicketStore() *WSTicketStore {
	store := &WSTicketStore{
		tickets: make(map[string]*WSTicket),
	}
	// Start cleanup goroutine
	go store.cleanup()
	return store
}

// cleanup removes expired tickets periodically
func (s *WSTicketStore) cleanup() {
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for ticket, t := range s.tickets {
			if t.ExpiresAt.Before(now) || t.Used {
				delete(s.tickets, ticket)
			}
		}
		s.mu.Unlock()
	}
}

// GenerateTicket creates a new one-time WebSocket ticket for a user
func (s *WSTicketStore) GenerateTicket(ctx context.Context, userID, baseID uuid.UUID) (string, error) {
	// Generate random ticket
	ticketBytes := make([]byte, wsTicketLength)
	if _, err := rand.Read(ticketBytes); err != nil {
		return "", err
	}
	ticket := base64.URLEncoding.EncodeToString(ticketBytes)

	s.mu.Lock()
	defer s.mu.Unlock()

	s.tickets[ticket] = &WSTicket{
		Ticket:    ticket,
		UserID:    userID,
		BaseID:    baseID,
		ExpiresAt: time.Now().Add(wsTicketLifetime),
		Used:      false,
	}

	return ticket, nil
}

// ValidateTicket validates and consumes a WebSocket ticket (one-time use)
func (s *WSTicketStore) ValidateTicket(ctx context.Context, ticket string, baseID uuid.UUID) (uuid.UUID, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	t, exists := s.tickets[ticket]
	if !exists {
		return uuid.Nil, errors.New("ticket not found")
	}

	if t.Used {
		return uuid.Nil, errors.New("ticket already used")
	}

	if time.Now().After(t.ExpiresAt) {
		delete(s.tickets, ticket)
		return uuid.Nil, errors.New("ticket expired")
	}

	if t.BaseID != baseID {
		return uuid.Nil, errors.New("ticket not valid for this base")
	}

	// Mark as used and delete
	t.Used = true
	delete(s.tickets, ticket)

	return t.UserID, nil
}
