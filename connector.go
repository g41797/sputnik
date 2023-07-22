package sputnik

import "time"

type ServerConnection any

type ServerConnector interface {
	// Connects to the server and return connection to server
	// If connection failed, returns error.
	// ' Connect' for already connected
	// and still not brocken connection should
	// return the same value returned in previous
	// successful call(s) and nil error
	Connect(config ServerConfiguration) (conn ServerConnection, err error)

	// Returns false if
	//  - was not connected at all
	//  - was connected, but connection is brocken
	// True returned if
	//  - connected and connection is alive
	IsConnected() bool

	// If connection is alive closes it
	Disconnect()
}

func ConnectorDescriptor() BlockDescriptor {
	return BlockDescriptor{DefaultConnectorName, DefaultConnectorResponsibility}
}

func init() {
	RegisterBlockFactory(DefaultConnectorName, connectorBlockFactory)
}

const DefaultConnectorTimeout = time.Second * 5

func connectorBlockFactory() *Block {
	connector := new(connector)
	block := NewBlock(
		WithInit(connector.init),
		WithRun(connector.run),
		WithFinish(connector.finish),
		WithOnMsg(connector.onSetup))
	return block
}

type doIt func()

type connector struct {
	conf ServerConfiguration

	mc  chan Msg
	cnr ServerConnector
	to  time.Duration

	cbc BlockCommunicator
	ibc BlockCommunicator

	bgfin  chan struct{}
	endfin chan struct{}

	next      doIt
	connected bool
}

func (c *connector) init(conf ServerConfiguration) error {

	c.conf = conf
	c.next = c.connect
	c.mc = make(chan Msg, 1)
	c.bgfin = make(chan struct{}, 1)

	return nil
}

func (c *connector) run(self BlockCommunicator) {
	defer close(c.mc)

	c.cbc = self
	c.ibc, _ = self.Communicator(InitiatorResponsibility)

	c.endfin = make(chan struct{}, 1)
	defer close(c.endfin)

	enableloop := false

	select {
	case <-c.bgfin:
		break

	case msg := <-c.mc:
		c.setup(msg)
		enableloop = true
	}

	if enableloop {

		ticker := time.NewTicker(c.to)

	runloop:
		for {
			select {
			case <-c.bgfin:
				break runloop

			case <-ticker.C:
				c.next()
			}
		}

		c.close()

		ticker.Stop()
	}

	return
}

func (c *connector) finish(init bool) {
	if init {
		return
	}

	close(c.bgfin)
	<-c.endfin

	return
}

func (c *connector) onSetup(msg Msg) {
	c.mc <- msg
	return
}

func (c *connector) setup(msg Msg) {

	cntr, exists := msg["__connector"]

	if !exists {
		return
	}

	c.cnr = cntr.(ServerConnector)

	to, _ := msg["__timeout"]

	timeout, _ := to.(time.Duration)
	c.to = timeout

	return
}

func (c *connector) connect() {
	if c.cnr == nil {
		return
	}

	if !c.connected {
		conn, err := c.cnr.Connect(c.conf)

		if err != nil {
			return
		}
		c.notifyConnected(conn)
		return
	}
}

func (c *connector) checkConnection() {
	if c.cnr == nil {
		return
	}

	if c.cnr.IsConnected() {
		return
	}

	c.notifyDisonnected()

	return
}

func (c *connector) close() {
	if c.cnr == nil {
		return
	}

	c.cnr.Disconnect()
	c.cnr = nil
	c.next = c.nop
	c.connected = false
}

func (c *connector) nop() {
	return
}

func (c *connector) notifyConnected(conn any) {
	c.ibc.Send(serverconnectedmsg((conn)))
	c.connected = true
	c.next = c.checkConnection
	return
}

func (c *connector) notifyDisonnected() {
	c.ibc.Send(serverdisconnectedmsg())
	c.connected = false
	c.next = c.connect
	return
}
