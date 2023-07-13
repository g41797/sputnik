package sputnik_test

import (
	"errors"

	"github.com/g41797/sputnik"
)

var _ sputnik.Connector = &dummyConnector{}

type dummyConnector struct {
	connected bool
}

func (c *dummyConnector) Connect(config any) (conn any, err error) {
	if !c.connected {
		return nil, errors.New("connection failed")
	}
	return "connected", nil
}

func (c *dummyConnector) IsConnected() bool {
	return c.connected
}

func (c *dummyConnector) Disconnect() {
	c.connected = false
}

func (c *dummyConnector) setState(connected bool) {
	c.connected = connected
}
