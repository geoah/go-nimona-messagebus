package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	logrus "github.com/sirupsen/logrus"

	mb "github.com/nimona/go-nimona-messagebus"
	net "github.com/nimona/go-nimona-net"
)

const (
	protocolID = "echo"
)

var addr = flag.String("addr", ":2180", "echo service address")

func main() {
	logrus.SetLevel(logrus.DebugLevel)

	n1Port := 2160
	n1PeerID := "n1"

	n2Port := 2161
	n2PeerID := "n2"

	// create networks
	// newNode will return a peer, a peerstore, a network, and an event bus
	p1, p1net, n1eb, err := newNode(n1Port, n1PeerID, "160A2921.asc")
	if err != nil {
		log.Fatal("Could not create n1", err)
	}
	p2, p2net, n2eb, err := newNode(n2Port, n2PeerID, "63A0147B.asc")
	if err != nil {
		log.Fatal("Could not create n2", err)
	}

	// we now need to let each network now of the other peer
	// this is not required if we attach a dht or other discovery method

	// add p2 to n1
	if err := p1net.PutPeer(*p2); err != nil {
		log.Fatal("Could not add p2 to n1ps")
	}

	// add p1 to n2
	if err := p2net.PutPeer(*p1); err != nil {
		log.Fatal("Could not add p1 to n2ps")
	}

	// send a message to p2 from p1
	payload := &mb.Payload{
		Creator: p1.ID,
		Type:    protocolID + "/message",
		Data:    []byte("Hello!"),
	}
	if err := n1eb.Send(payload, []string{p2.ID}); err != nil {
		fmt.Println("Could not send event", err)
	}

	// send a message to p1 from p2
	payload = &mb.Payload{
		Creator: p2.ID,
		Type:    protocolID + "echo",
		Data:    []byte("Hi there!"),
	}
	if err := n2eb.Send(payload, []string{p1.ID}); err != nil {
		fmt.Println("Could not send event", err)
	}

	// wait a bit to receive both messages
	time.Sleep(500 * time.Millisecond)
}

func newNode(port int, peerID string, keyPath string) (*net.Peer, net.Network, mb.MessageBus, error) {
	// create local peer
	// host := fmt.Sprintf("0.0.0.0:%d", port)
	pr, err := net.NewPeerFromArmorFile(keyPath)
	if err != nil {
		return nil, nil, nil, err
	}

	// pr.Addresses = append(pr.Addresses, host)

	// initialize network
	mn, err := net.NewNetwork(pr, port)
	if err != nil {
		fmt.Println("Could not initialize network", err)
		return nil, nil, nil, err
	}

	hn := func(hash []byte, msg mb.Message) error {
		fmt.Printf("Peer %s received event hash=%x, signature=%x, payload=%s\n", peerID, hash, msg.Signature, string(msg.PayloadRaw))
		return err
	}

	// initialize event bus
	eb, err := mb.New(protocolID, mn)
	if err != nil {
		fmt.Println("Could not initialize event bus", err)
	}

	// handle incomming messagebus
	if err := eb.HandleMessage(hn); err != nil {
		fmt.Println("Could not attache event handler", err)
		return nil, nil, nil, err
	}

	// print some info
	ip := "127.0.0.1"
	fmt.Printf("* New node: host=%s:%d id=%s\n", ip, port, peerID)

	return pr, mn, eb, nil
}
