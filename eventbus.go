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

type EventBus struct {
	protocolID string
	peer       ps.Peer
	network    net.Network
	handlers   []func(ev *Event) error
	events     map[string]*Event
	streams    map[string]*mux.Stream
}

func New(protocolID string, network net.Network, peer ps.Peer) (*EventBus, error) {
	eb := &EventBus{
		protocolID: protocolID,
		network:    network,
		peer:       peer,
		handlers:   []func(ev *Event) error{},
		events:     map[string]*Event{},
		streams:    map[string]*mux.Stream{},
	}
	if err := network.HandleStream(protocolID, eb.streamHander); err != nil {
		return nil, err
	}
	return eb, nil
}

func (eb *EventBus) HandleEvent(handler func(ev *Event) error) error {
	eb.handlers = append(eb.handlers, handler)
	return nil
}

func (eb *EventBus) streamHander(protocolID string, stream io.ReadWriteCloser) error {
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

func (eb *EventBus) getStream(peerID string) (*mux.Stream, error) {
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
func (eb *EventBus) send(peerID string, ev *Event) error {
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
func (p *EventBus) Send(oev *Event) error {
	switch oev.Type {
	case EventTypePDU:
		// persist the PDU events we send
		if _, ok := p.events[oev.ID]; !ok {
			p.events[oev.ID] = oev
		}
	}

	if re, err := oev.GetRecipient(string(p.peer.GetID())); err == nil {
		re.Ack = true
	}

	// go through all recipiends
	for _, re := range oev.Recipients {
		// don't send events to ourselves
		if re.PeerID == string(p.peer.GetID()) {
			continue
		}

		// don't send events to peple who have already ACKed
		if re.Ack == true {
			continue
		}

		// since right now we do everything in memory, copy the
		// event so we don't share it between the various peers
		ev := &Event{}
		jsonCopy(oev, ev)

		// attach sender id
		ev.SenderID = string(p.peer.GetID())

		// send the event to the peer
		if err := p.send(re.PeerID, ev); err != nil {
			fmt.Printf("! Could not send to peer: %s\n", err.Error())
		}
	}
	return nil
}

// Handle an incoming event
func (p *EventBus) handle(ev *Event) error {
	// p.Lock()
	// defer p.Unlock()

	switch ev.Type {
	// if the event is a PDU
	case EventTypePDU:
		if eev, ok := p.events[ev.ID]; ok {
			if re, err := eev.GetRecipient(ev.SenderID); err == nil {
				re.Ack = true
			}
			break
		}
		// we assume the owner already knows about this
		if re, err := ev.GetRecipient(ev.OwnerID); err == nil {
			re.Ack = true
		}
		// we assume that the sender already knows about this
		if re, err := ev.GetRecipient(ev.SenderID); err == nil {
			re.Ack = true
		}
		// we assume that we know about this
		// TODO(geoah) Ack for our own stuff should probably happen when sending as well
		if re, err := ev.GetRecipient(string(p.peer.GetID())); err == nil {
			// fmt.Printf("** Peer=%s: Setting peer=%s ack to true\n", p.ID, p.ID)
			re.Ack = true
		}

		// persist the event
		p.events[ev.ID] = ev

		for _, h := range p.handlers {
			h(ev)
		}

		// and finally send it
		go p.Send(ev)
	}
	return nil
}
