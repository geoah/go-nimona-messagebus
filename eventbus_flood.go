package events

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"

	mux "github.com/nimona/go-nimona-mux"
	net "github.com/nimona/go-nimona-net"
	ps "github.com/nimona/go-nimona-peerstore"
)

type FloodEventBus struct {
	protocolID    string
	peer          ps.Peer
	network       net.Network
	handlers      []func(ev *Event) error
	handledEvents []string
	streams       map[string]*mux.Stream
}

func New(protocolID string, network net.Network, peer ps.Peer) (*FloodEventBus, error) {
	eb := &FloodEventBus{
		protocolID:    protocolID,
		network:       network,
		peer:          peer,
		handlers:      []func(ev *Event) error{},
		handledEvents: []string{},
		streams:       map[string]*mux.Stream{},
	}
	if err := network.HandleStream(protocolID, eb.streamHander); err != nil {
		return nil, err
	}
	return eb, nil
}

func (eb *FloodEventBus) HandleEvent(handler func(ev *Event) error) error {
	eb.handlers = append(eb.handlers, handler)
	return nil
}

func (eb *FloodEventBus) streamHander(protocolID string, stream io.ReadWriteCloser) error {
	sr := bufio.NewReader(stream)
	for {
		line, err := sr.ReadString('\n')
		if err != nil {
			fmt.Println("Could not read")
			return nil // TODO(geoah) Return?
		}
		ev := &Event{}
		if err := json.Unmarshal([]byte(line), &ev); err != nil {
			fmt.Println("Could not decode JSON")
			return nil
		}
		eb.handle(ev)
	}
}

func (eb *FloodEventBus) getStream(peerID string) (*mux.Stream, error) {
	if stream, ok := eb.streams[peerID]; ok {
		// TODO Check if stream is still ok
		return stream, nil
	}
	stream, err := eb.network.NewStream(eb.protocolID, peerID)
	if err != nil {
		return nil, err
	}
	eb.streams[peerID] = stream
	return stream, nil
}

// send an event to a peer
func (eb *FloodEventBus) send(peerID string, ev *Event) error {
	stream, err := eb.getStream(peerID)
	if err != nil {
		return err
	}
	evbs, _ := json.Marshal(ev)
	evbs = append(evbs, '\n')
	if _, err := stream.Write(evbs); err != nil {
		return err
	}
	return nil
}

// Send takes an Event and will send it to its intended recipients
func (eb *FloodEventBus) Send(ev *Event) error {
	// go through all recipiends
	for _, pid := range ev.Recipients {
		// don't send events to ourselves
		if pid == string(eb.peer.GetID()) {
			continue
		}

		// attach sender id
		ev.SenderID = string(eb.peer.GetID())

		// send the event to the peer
		if err := eb.send(pid, ev); err != nil {
			fmt.Printf("! Could not send to peer: %s\n", err.Error())
		}
	}

	return nil
}

// Handle an incoming event
func (eb *FloodEventBus) handle(ev *Event) error {
	// TODO event checking and adding are not thread safe

	// check if we have already handled this event
	for _, evID := range eb.handledEvents {
		if evID == ev.ID {
			return nil
		}
	}

	// else add to handled events
	eb.handledEvents = append(eb.handledEvents, ev.ID)

	// and trigger handlers
	for _, h := range eb.handlers {
		h(ev)
	}

	return nil
}
