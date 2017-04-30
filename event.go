package events

import "errors"

var (
	// ErrorCNF Could not get address from peer id
	ErrorCNF = errors.New("Could not resolve peer ID")
)

const (
	// EventTypePDU is an event that should be persisted
	EventTypePDU = "PDU"
)

// Event is the basic wrapper for everything that is being sent
// around in the network
type Event struct {
	// Type can be either ACK or PDU atm
	Type string `json:"type"`
	// ID of the event
	ID string `json:"id"`
	// ParentID of an existing Event, optional for PDUs, required for ACKs
	ParentID string `json:"parent_id"`
	// OwnerID is the ID of the peer that created the event
	OwnerID string `json:"owner_id"`
	// Payload is anything
	Payload interface{} `json:"payload"`
	// Recipients of the event, as of now does not include the sender
	Recipients []*Recipient `json:"recipients"` // map[<peer-id>]Recipient
	// SenderID // TODO This should be moved away from the event
	SenderID string
}

// GetRecipient from a peer id
func (ev *Event) GetRecipient(peerID string) (*Recipient, error) {
	for _, re := range ev.Recipients {
		if re.PeerID == peerID {
			return re, nil
		}
	}
	return nil, ErrorCNF
}

// Recipient information for events
type Recipient struct {
	PeerID string `json:"peer_id"`
	Ack    bool   `json:"-"`
}
