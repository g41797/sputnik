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
 
Developers should be able to flexibly create such processes without changing the code:
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

### Finish

```go
type Init func(conf any) error
```
### Run

```go
type Init func(conf any) error
```
### OnServerConnect

```go
type Init func(conf any) error
```
### OnServerDisconnect

```go
type Init func(conf any) error
```
### OnMsg

```go
type Init func(conf any) error
```
