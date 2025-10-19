package usecase

import (
	"encoding/json"
	"sync"
)

// EventType represents different types of events
type EventType string

const (
	EventSessionUpdate    EventType = "session_update"
	EventSessionCreated   EventType = "session_created"
	EventPaymentConfirmed EventType = "payment_confirmed"
	EventStatsUpdate      EventType = "stats_update"
)

// Event represents a server-sent event
type Event struct {
	Type EventType   `json:"type"`
	Data interface{} `json:"data"`
}

// SessionUpdateEvent represents a session status change
type SessionUpdateEvent struct {
	SessionID    uint    `json:"session_id"`
	PlatNomor    string  `json:"plat_nomor,omitempty"`
	VehicleType  string  `json:"vehicle_type"`
	OldStatus    string  `json:"old_status,omitempty"`
	NewStatus    string  `json:"new_status"`
	TotalCost    float64 `json:"total_cost,omitempty"`
	CheckoutTime string  `json:"checkout_time,omitempty"`
	CheckinTime  string  `json:"checkin_time,omitempty"`
}

// PaymentConfirmedEvent represents a payment confirmation
type PaymentConfirmedEvent struct {
	SessionID     uint    `json:"session_id"`
	PaymentID     uint    `json:"payment_id"`
	PaymentMethod string  `json:"payment_method"`
	Amount        float64 `json:"amount"`
	ConfirmedBy   string  `json:"confirmed_by"`
	ConfirmedAt   string  `json:"confirmed_at"`
}

// StatsUpdateEvent represents dashboard statistics update
type StatsUpdateEvent struct {
	ActiveSessions    int     `json:"active_sessions"`
	PendingPayments   int     `json:"pending_payments"`
	TodayRevenue      float64 `json:"today_revenue"`
	TodayTransactions int     `json:"today_transactions"`
}

// JukirClient represents a connected jukir client
type JukirClient struct {
	JukirID uint
	Channel chan string
}

// EventManager manages SSE connections and broadcasts events
type EventManager struct {
	clients map[uint]*JukirClient // map[jukirID]*JukirClient
	mu      sync.RWMutex
}

// NewEventManager creates a new event manager
func NewEventManager() *EventManager {
	return &EventManager{
		clients: make(map[uint]*JukirClient),
	}
}

// RegisterJukir registers a jukir for SSE updates
func (em *EventManager) RegisterJukir(jukirID uint) chan string {
	em.mu.Lock()
	defer em.mu.Unlock()

	// Create channel for this jukir
	ch := make(chan string, 10) // Buffer of 10 events

	client := &JukirClient{
		JukirID: jukirID,
		Channel: ch,
	}

	em.clients[jukirID] = client
	return ch
}

// UnregisterJukir removes a jukir from SSE updates
func (em *EventManager) UnregisterJukir(jukirID uint) {
	em.mu.Lock()
	defer em.mu.Unlock()

	if client, ok := em.clients[jukirID]; ok {
		close(client.Channel)
		delete(em.clients, jukirID)
	}
}

// NotifyJukir sends an event to a specific jukir
func (em *EventManager) NotifyJukir(jukirID uint, eventType EventType, data interface{}) {
	em.mu.RLock()
	defer em.mu.RUnlock()

	if client, ok := em.clients[jukirID]; ok {
		event := Event{
			Type: eventType,
			Data: data,
		}

		eventJSON, err := json.Marshal(event)
		if err != nil {
			return
		}

		// Non-blocking send
		select {
		case client.Channel <- string(eventJSON):
			// Event sent successfully
		default:
			// Channel full, skip this event
		}
	}
}

// BroadcastToArea sends an event to all jukirs in a specific area
func (em *EventManager) BroadcastToArea(areaID uint, jukirIDs []uint, eventType EventType, data interface{}) {
	event := Event{
		Type: eventType,
		Data: data,
	}

	eventJSON, err := json.Marshal(event)
	if err != nil {
		return
	}

	em.mu.RLock()
	defer em.mu.RUnlock()

	for _, jukirID := range jukirIDs {
		if client, ok := em.clients[jukirID]; ok {
			select {
			case client.Channel <- string(eventJSON):
				// Event sent successfully
			default:
				// Channel full, skip
			}
		}
	}
}

// GetConnectedJukirs returns the number of connected jukirs
func (em *EventManager) GetConnectedJukirs() int {
	em.mu.RLock()
	defer em.mu.RUnlock()
	return len(em.clients)
}

// IsJukirConnected checks if a jukir is connected
func (em *EventManager) IsJukirConnected(jukirID uint) bool {
	em.mu.RLock()
	defer em.mu.RUnlock()
	_, ok := em.clients[jukirID]
	return ok
}
