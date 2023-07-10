package sputnik

type Configurator interface {
	// Returns configuration of running process
	// If implementation supports multithreading,
	// every call may return the same "object"
	// Otherwise copy/clone of configuration should be returned
	Configuration() (conf any, err error)
}

type Server interface {
	Configurator

	// Connects to the server and return connection to server
	// If connection failed, returns error.
	// ' Connect' for already connected
	// and still not brocken connection should
	// return the same value returned in previous
	// successful call(s) and nil error
	//
	Connect() (conn, err error)

	// Returns false if
	//  - was not connected at all
	//  - was connected, but connection is brocken
	// True returned if
	//  - connected and connection is alive
	IsConnected() bool
}
