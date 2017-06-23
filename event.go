package events

import "errors"

var (
	// ErrorCNF Could not get address from peer id
	ErrorCNF = errors.New("Could not resolve peer ID")
)

const (
	// EventTypePDU is an event that should be persisted
	EventTypePDU = "PDU"
	// EventTypeEDU is an ephemeral event that can be handled and discarded
	EventTypeEDU = "EDU"
)

// Event is the basic wrapper for everything that is being sent
// around in the network
type Event struct {
	// Type can be either ACK or PDU atm
	Type string `json:"type"`
	// ID of the event
	ID string `json:"id"`
	// OwnerID is the ID of the peer that created the event
	OwnerID string `json:"owner_id"`
	// Payload can be anything
	Payload interface{} `json:"payload"`
	// Recipients of the event, as of now does not include the sender
	Recipients []string `json:"recipients"` // array of peer IDs
	// SenderID is the ID of the peer that created the event
	SenderID string `json:"sender_id"`
}
