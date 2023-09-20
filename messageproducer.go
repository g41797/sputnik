package sputnik

// MessageProducer provides possibility for negotiation between sputnik based
// software and external broker process
type MessageProducer interface {
	// Connect to the broker
	Connect(cf ConfFactory) error

	// Translate message to format of the broker and send it
	Produce(msg Msg) error

	// If connection is alive closes it
	Disconnect()
}
