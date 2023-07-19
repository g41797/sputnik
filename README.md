![](_logo/logo.png)

# sputnik [![GoDev](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white)](https://pkg.go.dev/github.com/g41797/sputnik) [![Wiki](https://img.shields.io/badge/Wikipedia-%23000000.svg?style=for-the-badge&logo=wikipedia&logoColor=white)](https://en.wikipedia.org/wiki/Sputnik_1)
**sputnik** is tiny golang framework for building of **satellite** or as it's now fashionable to say **side-car** processes.

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


##  sputnik to the rescue.
sputnik simplifies creation of satellite/sidecar  processes for servers.

### Modular monolith

sputnik forces modular-monolith design:
  - process created as set of independent asynchronously running **Blocks**

### Satellites blueprint

sputnik supports common for all satellite processes functionality::
* Deterministic initialization
* Connect/Reconnect flow
* Server heartbeat
* Convenient negotiation between blocks of the process
* Graceful shutdown

All this with minimal code size - actually the size of README and tests far exceeds size of sputnik's code.

## Why Sputnik?
* Launched by the Soviet Union on 4 October 1957, **Sputnik** became the first **satellite** in space and changed the world forever.
* Main mindset of Sputnik design was - simplicity and reliability that could be adapted to future projects
* We were both born the same year but I'm a bit older


## Less You Know, the Better You Sleep
sputnik doesn't use any server information and only assumes that *server configuration* and *server connection*
are required for functioning. 

### Server Configuration
```go
type ServerConfiguration any
```

In order to get configuration and provide it to the process, sputnik uses *Configuration Factory*:
```go
type ConfFactory func() ServerConfiguration
```

This function should be supplied by caller of sputnik during initialization. We will talk about initialization later.

### Server Connection
```go
type ServerConnection any
```

For implementation of connect/disconnect/server health flow, sputnik uses supplied by caller implementation of following interface:
```go
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
```

### Messages
sputnik supports asynchronous communication between Blocks of the process.
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

sputnik doesn't force specific format of the message.

EXCEPTION: key of the map should not start from "__".

This prefix is used by sputnik for house-keeping values.


## sputnik's building blocks
sputnik based process consists of *infrastructure* and *application* **Blocks**


### Eats own dog food
Infrastructure **Blocks**:
  * *initiator* - dispatcher of all blocks
  * *finisher*  - listener of external shutdown/exit events
  * *connector* - connects/reconnects with server, provides this
    information to another blocks


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

Remember - sputnik based process may support number of protocol adapters. And *receiver* usually is part of everyone.

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

Init callback is executed by sputnik once during initialization.
Blocks are initialized in *sequenced order* according to configuration.

Rules of initialization:
 * don't run hard processing within Init
 * don't work with server, wait *OnServerConnect*

If initialization failed (returned error != nil)
 * initialization is terminated
 * already initialized blocks are finished in opposite to init order.

#### Run

```go
type Run func(controller BlockController)
```

Run callback is executed by sputnik
* after successful initialization of ALL blocks
* on own goroutine

You can consider Run as *main thread* of the block.

Parameter of Run - **BlockController** may be used by block for negotiation with another blocks of the process.

#### Finish

```go
type Finish func(init bool)
```

Finish callback is executed by sputnik twice:
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

*Optional* OnServerConnect callback is executed by sputnik
* after start of *Run*
* after successful connection to server
* on own goroutine


#### OnServerDisconnect

```go
type OnServerDisconnect func()
```
*Optional* OnServerDisconnect callback is executed by sputnik
* after start of *Run*
* when previously connected server disconnects
* on own goroutine

#### OnMsg

```go
type OnMsg func(msg Msg)
```
*Optional*  OnMsg callback is executed by sputnik 
* after start of *Run*
* as result of receiving Msg from another block
* Block also can send message to itself

**UNLIKE OTHER CALLBACKS, OnMsg CALLED SEQUENTIALLY ONE BY ONE FROM THE SAME DEDICATED GOROUTINE**. Frankly speaking - you have the queue of messages.

### Block creation

Developer supplies *BlockFactory* function:
```go
type BlockFactory func() *Block
```
*BlockFactory* registered in the process via RegisterBlockFactory:
```go
func RegisterBlockFactory(blockName string, blockFactory BlockFactory)
```
Use *init()* for this registration:
```go
func init() { // Registration of finisher block factory
	RegisterBlockFactory(DefaultFinisherName, finisherBlockFactory)
}
```

Where finisherBlockFactory is :
```go
func finisherBlockFactory() *Block {
	finisher := new(finisher)
	block := NewBlock(
		WithInit(finisher.init),
		WithRun(finisher.run),
		WithFinish(finisher.finish),
		WithOnMsg(finisher.debug))
	return block
}
```
You can see that factory called *NewBlock function* using [functional options pattern](https://golang.cafe/blog/golang-functional-options-pattern.html):

List of options:
```go
WithInit(f Init)
WithFinish(f Finish)
WithRun(f Run)
WithOnConnect(f OnServerConnect)
WithOnDisConnect(f OnServerDisconnect)
WithOnMsg(f OnMsg)
```
where *f* is related callback/hook

### Block control
Block control is provided via interface *BlockController*. Block gets own controller as parameter of **Run**.
```go
type BlockController interface {
	//
	// Get controller of block by block's responsibility
	// Example - get BlockController of initiator:
	// initbl, ok := bc.Controller(sputnik.InitiatorResponsibility)
	//
	Controller(resp string) (bc BlockController, exists bool)

	// Identification of controlled block
	Descriptor() BlockDescriptor

	// Asynchronously send message to controlled block
	// true is returned if
	//  - controlled block has OnMsg callback
	//  - recipient of messages was not cancelled
	//  - msg != nil
	Send(msg Msg) bool

	// Asynchronously notify controlled block about server status
	// true is returned if if controlled block has OnServerConnect callback
	ServerConnected(sc ServerConnection) bool

	// Asynchronously notify controlled block about server status
	// true is returned if controlled block has OnServerDisconnect callback
	ServerDisconnected() bool

	// Asynchronously call Finish callback of controlled block
	//
	Finish()
}
```

Main usage of own BlockController:
* get BlockController of another block
  * send message to this block of
  * call callback on this block

Example 1 - how 'finisher' informs 'initiator' about process termination:
```go
	ibc, _ := fns.owncontroller.Controller("initiator")
	ibc.Finish()
```
Example 2 - how 'initiator' sends setup settings to 'connector':
```go
	setupMsg := make(Msg)
	setupMsg["__connector"] = connectorPlugin
	setupMsg["__timeout"] = 10000

	connController.Send(setupMsg)
```

## sputnik flight

### Create sputnik

Use *NewSputnik function* for creation of sputnik.
It supports following options:
```go
WithConfFactory(cf ConfFactory)                      // Set Configuration Factory. Mandatory
WithAppBlocks(appBlocks []BlockDescriptor)           // List of descriptors for application blocks
WithBlockFactories(blkFacts BlockFactories)          // List of block factories. Optional. If was not set, used list of factories registrated during init()
WithFinisher(fbd BlockDescriptor)                    // Descriptor of finisher. Optional. If was not set, default supplied finished will be used.
WithConnector(cnt ServerConnector, to time.Duration) // Server Connector plug-in and timeout for connect/reconnect. Optional
```

Example: creation of sputnik for tests:
```go
	testSputnik, _ := sputnik.NewSputnik(
		sputnik.WithConfFactory(dumbConf),
		sputnik.WithAppBlocks(blkList),
		sputnik.WithBlockFactories(tb.factories()),
		sputnik.WithConnector(&tb.conntr, tb.to),
	)
```

### Preparing for flight

After creation of sputnik, call *Prepare*:
```go
// Creates and initializes all blocks.
//
// If creation and initialization of any block failed:
//
//   - Finish is called on all already initialized blocks
//
//   - Order of finish - reversal of initialization
//
//     = Returned error describes reason of the failure
//
// Otherwise returned 2 functions for sputnik management:
//
//   - lfn - Launch of the sputnik , exit from this function will be
//     after signal for shutdown of the process  or after call of
//     second returned function (see below)
//
//   - st - ShootDown of sputnik - abort flight
func (sputnik Sputnik) Prepare() (lfn Launch, st ShootDown, err error) 
```

Example :
```go
	launch, kill, err := testSputnik.Prepare()
```

### sputnik launch

Very simple - just call returned *launch* function.

This call is synchronous. sputnik continues to run on current goroutine till
* process termination (it means that launch may be last line in main)
* call of second returned function

In order to use kill(ShootDown of sputnik) function, launch and kill should run
on different go-routines.


## Never Asked Questions

- [I'd like to create stand alone process without any server. Is it possible?](#id-like-to-create-stand-alone-process-without-any-server-is-it-possible)
- [I'd like to embed sputnik to my process? Is it possible?](#id-like-to-embed-sputnik-to-my-process-is-it-possible)
- [You wrote that *finisher* can be replaced. For what?](#you-wrote-that-finisher-can-be-replaced-for-what)

### I'd like to create stand alone process without any server? Is it possible?

Don't use *WithConnector* option and sputnik will not run any server connector.
But for this case your blocks should not have  *OnServerConnect* and *OnServerDisconnect* callbacks.

### I'd like to embed sputnik to my process? Is it possible?

Of course, supply *ServerConnector* for in-proc communication.

### You wrote that *finisher* can be replaced. For what?

For example in case above, you will need to coordinate exit with code of the host.

## Contributing

Feel free to report bugs and suggest improvements.

## License

[MIT](LICENSE)


