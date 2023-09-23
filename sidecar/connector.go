package sidecar

import "github.com/g41797/sputnik"

// Connector provides possibility for negotiation between sputnik based
// software and external broker process
type Connector interface {
	// Connect to the broker or attach to existing shared
	Connect(cf sputnik.ConfFactory, shared sputnik.ServerConnection) error

	// For shared connection - detach, for own - close
	Disconnect()
}

type MessageProducer interface {
	Connector

	// Translate message to format of the broker and send it
	Produce(msg sputnik.Msg) error
}

type MessageConsumer interface {
	Connector

	// Receive event from broker, convert to sputnik.Msg
	// and send to another block
	Consume(sender sputnik.BlockCommunicator) error
}
