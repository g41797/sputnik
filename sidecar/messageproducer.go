package sidecar

import "github.com/g41797/sputnik"

// MessageProducer provides possibility for negotiation between sputnik based
// software and external broker process
type MessageProducer interface {
	// Connect to the broker
	Connect(cf sputnik.ConfFactory, shared sputnik.ServerConnection) error

	// Translate message to format of the broker and send it
	Produce(msg sputnik.Msg) error

	// If connection is alive closes it
	Disconnect()
}
