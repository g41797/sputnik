# Sputnik
Sputnik is tiny golang framework for building of **satellite** or as it's now fashionable to say **side-car** processes.

##  What do satellite processes have in common?
The same sometimes boring flow:
* Initialize process in deterministic order
* Connect to the server process
  * Periodically validate used connection
  * Inform about failed connection and reconnect
* Graceful shutdown
  * Cleanup resources in deterministic order


Usually such processes are used as **adapters**,**bridges** and/or **proxies** for server process - they *translate* foreign protocol to protocol of the server. 
 
Developers also want flexible way to create such processes without changing the code:
  * All adapters in one process
  * Adapter per process
  * Other variants

And it would be nice to write in CV that you developed **modular-monolith**. 


##  Sputnik to the rescue.
Sputnik simplifies creation of satellite/sidecar  processes for servers.

### Modular monolith

Sputnik forces modular-monolith design:
  - process created as set of independent asynchronously running **Blocks**

### Satellites blueprint

Sputnik supports common for all satellite processes functionality::
* Deterministic initialization
* Connect/Reconnect flow
* Server heartbeat
* Convenient negotiation between blocks of the process
* Graceful shutdown

All this with minimal code size - actually the size of README and tests far exceeds size of Sputnik's code.

## Why Sputnik?
* Launched by the Soviet Union on 4 October 1957, **Sputnik** became the first **satellite** in space and changed the world forever.
* Main mindset of Sputnik design was - simplicity and reliability that could be adapted to future projects
* We were both born the same year but I'm a bit older


## Less You Know, the Better You Sleep
Sputnik doesn't use any server information and only assumes that server configuration and connection
are required for functioning. 

### Server Configuration
```go
type ServerConfiguration any
```

In order to get configuration and provide it to the process, Sputnik uses *Configuration Factory*:
```go
type ConfFactory func() ServerConfiguration
```

This function should be supplied by caller of Sputnik during initialization. We will talk about initialization later.

### Server Connection
```go
type ServerConnection any
```

For implementation of connect/disconnect/server health flow, Sputnik uses supplied by caller implementation of following interface:
```go
type ServerConnector interface {
	/*
  Connects to the server and return connection to server
	If connection failed, returns error.
	' Connect' for already connected
	and still not brocken connection should
	return the same value returned in previous
	successful call(s) and nil error
	*/
  Connect(config ServerConfiguration) (conn ServerConnection, err error)

  /*
	Returns false if
	  - was not connected at all
	  - was connected, but connection is brocken
	True returned if
	  - connected and connection is alive
	*/
  IsConnected() bool

	// If connection is alive closes it
	Disconnect()
}
```

### Messages
Sputnik supports asynchronous communication between Blocks of the process.
```go
type Msg map[string]any
```

Possible types of the message:
* command
* query
* event
* update
* .......

Developers of blocks should agree on content of messages.

Sputnik doesn't force specific format of the message.

EXCEPTION: key of the map should not start from "__".

This prefix is used by sputnik for house-keeping values.


## Sputnik's building blocks
Building block of Sputnik called (of-course - do you remember *simplicity*?)   **Block**.

### Block identity
Every Block has descriptor:
```go
type BlockDescriptor struct {
	Name           string
	Responsibility string
}
```

**Name** of the Block should be unique. It is used for creation of the Block.

Good Block names:
* syslogreceiver
* syslogpublisher
* restprocessor

Bad Block names:
* receiver
* processor

Remember - Sputnik based process may support number of protocol adapters. And *receiver* usually is part of everyone.

*Responsibility* of the Block is used for negotiation between blocks. It's possible to create the same block with different responsibilities.


### Block interface

**Block** has set of callbacks/hooks:
* Mandatory:
  * Init
  * Finish
  * Run
* Optional
  * OnServerConnect
  * OnServerDisconnect
  * OnMsg

You can see that these callbacks reflect life cycle/flow of satellite process.  

#### Init

```go
type Init func(conf any) error
```

Init callback is executed by Sputnik once during initialization.
Blocks are initialized in *sequenced order* according to configuration.

Rules of initialization:
 * don't run hard processing within Init
 * don't work with server, wait *OnServerConnect*

If initialization failed (returned error != nil)
 * initialization is terminated
 * already initialized blocks are finished in opposite to init order.


 We will talk later about *conf* (configuration). 

#### Run

```go
type Run func(controller BlockController)
```

Run callback is executed by Sputnik
* after successful initialization of ALL blocks
* on own goroutine

You can consider Run as *main thread* of the block.

Parameter of Run - **BlockController** may be used by block for negotiation with another blocks of the process.

#### Finish

```go
type Finish func(init bool)
```

Finish callback is executed by Sputnik twice:
* during initialization of the process, if Init of another block failed (**init == true**)
  * for this case Finish is called synchronously, on the thread(goroutine) of initialization
* during shutdown of the process 
  * for this case Finish is called asynchronously on own goroutine

For any case, during Finish block
* should clean all resources
* stop all go-routines (don't forget Run's goroutine)

After finish of all blocks Sputnik quits.

#### OnServerConnect

```go
type OnServerConnect func(connection any)
```

*Optional* OnServerConnect callback is executed by Sputnik
* after start of *Run*
* after successful connection to server
* on own goroutine


#### OnServerDisconnect

```go
type OnServerDisconnect func()
```
*Optional* OnServerDisconnect callback is executed by Sputnik
* after start of *Run*
* when previously connected server disconnects
* on own goroutine

#### OnMsg

```go
type OnMsg func(msg Msg)
```
*Optional*  OnMsg callback is executed by Sputnik 
* after start of *Run*
* as result of receiving Msg from another block
* Block also can send message to itself

**UNLIKE OTHER CALLBACKS, OnMsg CALLED SEQUENTIALLY ONE BY ONE FROM THE SAME DEDICATED GOROUTINE**. Frankly speaking - you have the queue of messages.



### Eats own dog food
Sputnik itself consists of 3 **Blocks**:
  * *initiator* - dispatcher of all blocks
  * *finisher*  - listener of external shutdown/exit events
  * *connector* - connects/reconnects with server, provides this
    information to another blocks
