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
Everything you read above is not fantasy.
Sputnik supports:
* Flexibly creation of satellite processes
* Modular architecture
* Convenient negotiation between parts of the process
* Deterministic initialization
* Connect/Reconnect flow
* Graceful shutdown

All this with minimal code size - actually the size of README, tests and comments far exceeds size of Sputnik's code.

## Why Sputnik?
* Launched by the Soviet Union on 4 October 1957, **Sputnik** became the first **satellite** in space and changed the world forever.
* Main mindset of Sputnik design was - simplicity and reliability that could be adapted to future projects
* We were both born the same year but I'm a bit older


## Sputnik's building blocks
Building block of Sputnik called (of-course - do you remember *simplicity*?)   **Block**.

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

### Init

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

### Run

```go
type Run func(controller BlockController)
```

Run callback is executed by Sputnik
* after successful initialization of ALL blocks
* on own goroutine

You can consider Run as *main thread* of the block.

Parameter of Run - **BlockController** may be used by block for negotiation with another blocks of the process.

### Finish

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

### OnServerConnect

```go
type OnServerConnect func(connection any)
```

*Optional* OnServerConnect callback is executed by Sputnik
* after start of *Run*
* after successful connection to server
* on own goroutine


### OnServerDisconnect

```go
type OnServerDisconnect func()
```
*Optional* OnServerDisconnect callback is executed by Sputnik
* after start of *Run*
* when previously connected server disconnects
* on own goroutine

### OnMsg

```go
type OnMsg func(msg Msg)
```
*Optional*  OnMsg callback is executed by Sputnik 
* after start of *Run*
* as result of receiving Msg from another block
* Block also can send message to itself

**UNLIKE OTHER CALLBACKS, OnMsg CALLED SEQUENTIALLY ONE BY ONE FROM THE SAME DEDICATED GOROUTINE**. Frankly speaking - you have the queue of messages.