package sputnik

import (
	"errors"
)

var _ ServerConnector = &DummyConnector{}

// Connector's plugin used for debugging/testing
type DummyConnector struct {
	connected bool
}

func (c *DummyConnector) Connect(config ServerConfiguration) (conn ServerConnection, err error) {
	if !c.connected {
		return nil, errors.New("connection failed")
	}
	return "connected", nil
}

func (c *DummyConnector) IsConnected() bool {
	return c.connected
}

func (c *DummyConnector) Disconnect() {
	c.connected = false
}

// Allows to simulate state of the connection
func (c *DummyConnector) SetState(connected bool) {
	c.connected = connected
}
