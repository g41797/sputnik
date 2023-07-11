package sputnik

type Connector interface {
	// Connects to the server and return connection to server
	// If connection failed, returns error.
	// ' Connect' for already connected
	// and still not brocken connection should
	// return the same value returned in previous
	// successful call(s) and nil error
	Connect(config any) (conn any, err error)

	// Returns false if
	//  - was not connected at all
	//  - was connected, but connection is brocken
	// True returned if
	//  - connected and connection is alive
	IsConnected() bool
}
