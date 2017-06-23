package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	uuid "github.com/google/uuid"

	events "github.com/nimona/go-nimona-events"
	nimnet "github.com/nimona/go-nimona-net"
	peerstore "github.com/nimona/go-nimona-peerstore"
)

const (
	protocolID = "/echo/v0"
	eventType  = "echo-message"
)

var addr = flag.String("addr", ":2180", "echo service address")

func main() {
	n1Port := 2160
	n1PeerID := "n1"

	n2Port := 2161
	n2PeerID := "n2"

	// create networks
	// newNode will return a peer, a peerstore, a network, and an event bus
	p1, n1ps, _, n1eb, err := newNode(n1Port, n1PeerID)
	if err != nil {
		log.Fatal("Could not create n1", err)
	}
	p2, n2ps, _, n2eb, err := newNode(n2Port, n2PeerID)
	if err != nil {
		log.Fatal("Could not create n2", err)
	}

	// we now need to let each network now of the other peer
	// this is not required if we attach a dht or other discovery method

	// add p2 to n1
	if err := n1ps.Put(p2); err != nil {
		log.Fatal("Could not add p2 to n1ps")
	}

	// add p1 to n2
	if err := n2ps.Put(p1); err != nil {
		log.Fatal("Could not add p1 to n2ps")
	}

	// send a message to p2 from p1
	if err := n1eb.Send(&events.Event{
		Type:       eventType,
		ID:         uuid.New().String(),
		OwnerID:    string(p1.GetID()),
		SenderID:   string(p1.GetID()),
		Payload:    "Hello",
		Recipients: []string{string(p2.GetID())},
	}); err != nil {
		fmt.Println("Could not send event", err)
	}

	// send a message to p1 from p2
	if err := n2eb.Send(&events.Event{
		Type:       eventType,
		ID:         uuid.New().String(),
		OwnerID:    string(p2.GetID()),
		SenderID:   string(p2.GetID()),
		Payload:    "Hi there",
		Recipients: []string{string(p1.GetID())},
	}); err != nil {
		fmt.Println("Could not send event", err)
	}

	// wait a bit to receive both messages
	time.Sleep(500 * time.Millisecond)
}

func eventHandler(ev *events.Event) error {
	fmt.Println("Received event:", ev)
	return nil
}

func newNode(port int, peerID string) (peerstore.Peer, peerstore.Peerstore, nimnet.Network, *events.EventBus, error) {
	// create new peerstore
	ps := peerstore.New()

	// create local peer
	host := fmt.Sprintf("0.0.0.0:%d", port)
	pr := &peerstore.BasicPeer{
		ID:        peerstore.ID(peerID),
		Addresses: []string{host},
	}

	// initialize network
	mn, err := nimnet.NewTCPNetwork(pr, ps)
	if err != nil {
		fmt.Println("Could not initialize network", err)
	}

	hn := func(event *events.Event) error {
		fmt.Printf("Peer %s received event: %+v\n", peerID, event)
		return nil
	}

	// initialize event bus
	eb, err := events.New(protocolID, mn, pr)
	if err != nil {
		fmt.Println("Could not initialize event bus", err)
	}

	// handle incomming events
	if err := eb.HandleEvent(hn); err != nil {
		fmt.Println("Could not attache event handler", err)
		return nil, nil, nil, nil, err
	}

	// print some info
	ip := "127.0.0.1"
	fmt.Printf("* New node: host=%s:%d id=%s\n", ip, port, peerID)

	return pr, ps, mn, eb, nil
}
