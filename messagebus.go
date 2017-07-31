package messagebus

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"

	sha3 "golang.org/x/crypto/sha3"

	mux "github.com/nimona/go-nimona-mux"
	net "github.com/nimona/go-nimona-net"
)

type MessageBus interface {
	HandleMessage(handler func(hash []byte, msg Message) error) error
	Send(payload *Payload, peerIDs []string) error
}

type messageBus struct {
	protocolID      string
	peer            net.Peer
	network         net.Network
	handlers        []func(hash []byte, msg Message) error
	handledMessages []string
	streams         map[string]*mux.Stream
}

func New(protocolID string, network net.Network, peer net.Peer) (MessageBus, error) {
	eb := &messageBus{
		protocolID:      protocolID,
		network:         network,
		peer:            peer,
		handlers:        []func(hash []byte, msg Message) error{},
		handledMessages: []string{},
		streams:         map[string]*mux.Stream{},
	}
	if err := network.RegisterStreamHandler(protocolID, eb.streamHander); err != nil {
		return nil, err
	}
	return eb, nil
}

func (eb *messageBus) HandleMessage(handler func(hash []byte, msg Message) error) error {
	eb.handlers = append(eb.handlers, handler)
	return nil
}

func (eb *messageBus) streamHander(protocolID string, stream io.ReadWriteCloser) error {
	sr := bufio.NewReader(stream)
	for {
		// read line
		// TODO replaces with proper stream decoder
		line, err := sr.ReadString('\n')
		if err != nil {
			fmt.Println("Could not read") // TODO Fix logging
			return err                    // TODO(geoah) Return?
		}

		// decode envelope
		ev := &Envelope{}
		if err := json.Unmarshal([]byte(line), &ev); err != nil {
			fmt.Println("Could not decode envelope") // TODO Fix logging
			return err
		}

		// verify hash
		// TODO verify hash

		// decode message
		if err := json.Unmarshal(ev.MessageRaw, &ev.Message); err != nil {
			fmt.Println("Could not decode message") // TODO Fix logging
			return err
		}

		// decode payload
		if err := json.Unmarshal([]byte(ev.Message.PayloadRaw), &ev.Message.Payload); err != nil {
			fmt.Println("Could not decode payload") // TODO Fix logging
			return nil
		}

		// get creator
		// TODO Make verification optional ena re-enable
		// creator, err := eb.network.GetPeer(ev.Message.Payload.Creator)
		// if err != nil {
		// 	// TODO attempt to retrieve creator and retry?
		// 	fmt.Println("Unknown creator")
		// 	return err
		// }

		// verify signature
		// valid, err := creator.Verify(ev.Message.PayloadRaw, ev.Message.Signature)
		// if err != nil {
		// 	return err
		// }
		// if valid == false {
		// 	return errors.New("Invalid signature") // TODO Better error
		// }

		// sent message to handlers
		eb.handle(ev.Hash, ev.Message)
	}
}

func (eb *messageBus) getStream(peerID string) (*mux.Stream, error) {
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
func (eb *messageBus) send(ev *Envelope, peerID string) error {
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
func (eb *messageBus) Send(payload *Payload, peerIDs []string) error {
	// encode payload
	bpay, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// sign message
	spay, err := eb.peer.Sign(bpay)
	if err != nil {
		return err
	}

	// create signed message
	msg := &Message{
		PayloadRaw: bpay,
		Signature:  spay,
	}

	// encode message
	bmsg, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	// hash bmsg
	shash := sha3.Sum256(bmsg)

	// create envelope
	ev := &Envelope{
		MessageRaw: bmsg,
		Hash:       shash[:],
	}

	// go through all recipiends
	for _, pid := range peerIDs {
		// don't send messagebus to ourselves
		if pid == string(eb.peer.ID) {
			continue
		}

		// send the event to the peer
		if err := eb.send(ev, pid); err != nil {
			// TODO Handle error
			fmt.Printf("Could not send to peer: %s\n", err.Error())
		}
	}

	return nil
}

// Handle an incoming event
func (eb *messageBus) handle(hash []byte, msg Message) error {
	// TODO event checking and adding are not thread safe

	// check if we have already handled this event
	shash := fmt.Sprintf("%x", hash)
	for _, msgHash := range eb.handledMessages {
		if msgHash == shash {
			return nil
		}
	}

	// else add to handled messagebus
	eb.handledMessages = append(eb.handledMessages, shash)

	// and trigger handlers
	for _, h := range eb.handlers {
		h(hash, msg)
	}

	return nil
}
