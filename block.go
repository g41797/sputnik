package sputnik

// Block has Name (analog of golang type) and Responsibility (instance of specific block)
// This separation allows to run simultaneously blocks with the same Name.
// Other possibility - blocks with different name but with the same responsibility,
// e.g. different implementation of "finisher" depends on environment.
//
// initiator block has predetermined Responsibility
const InitiatorResponsibility = "initiator"

type BlockDescriptor struct {
	Name           string
	Responsibility string
}

// Block has set of the callbacks:
//   - mandatory:	Init|Run|Finish
//   - optional:	OnServerConnect|OnServerDisconnect|OnMsg
//
// Init callback is executed by sputnik once during initialization.
// Blocks are initialized in sequenced order according to configuration.
// Some rules :
//   - don't run hard processing within Init
//   - don't work with server till call of OnServerConnect
type Init func(conf any) error

// After successful initialization of ALL blocks, sputnik creates goroutine and calls Run
// Other callbacks will be executed on another goroutines
// After Run block is allowed to negotiate with another blocks of the process
// via BlockController
type Run func(self BlockController)

// Finish callback is executed by sputnik:
//   - during initialization  of the process if init of another block failed (init == true)
//   - during shutdown of the process (init == false)
//
// Blocks are finished in reverse of initialization order.
type Finish func(init bool)

// Optional OnServerConnect callback is executed by sputnik after successful
// connection to server.
type OnServerConnect func(connection any, logger any)

// Optional OnServerDisconnect callback is executed by sputnik when previously
// connected server disconnects.
type OnServerDisconnect func(connection any)

// Because asynchronous nature of blocks, negotiation between blocks done using 'messages'
// Message may be command|query|event|update|...
// Developers of blocks should agree on content of messages.
// sputnik doesn't force specific format of the message
// with one exception: key of the map should not start from "__".
// This prefix is used by sputnik for house-keeping values.
// You can call Send with nil message, but on the side of the receiver it may be not nil,
// because data added by sputnik
type Msg map[string]any

// Optional OnMsg callback is executed by sputnik as result of receiving Msg.
// Block can send event to itself.
// Unlike other callbacks, OnMsg called sequentially one by one from the same goroutine.
type OnMsg func(msg Msg)

// Simplified Block life cycle:
//   - Init
//   - Run
//   - OnServerConnect
//   - [*]OnMsg
//   - OnServerDisconnect
//   - Finish
//
// Please pay attention, that after Run order of callbacks will be unpredictable.
type Block struct {
	Init         Init
	Run          Run
	Finish       Finish
	OnConnect    OnServerConnect
	OnDisconnect OnServerDisconnect
	OnMsg        OnMsg
}

// Check presence of mandatory callbacks
func (bl *Block) isValid() bool {
	return bl.Init != nil && bl.Run != nil && bl.Finish != nil
}

// BlockController provides possibility for negotiation between blocks
// Block gets own controller as parameter of Run
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
	Send(msg Msg) bool

	// Asynchronously notify controlled block about server status
	// true is returned if if controlled block has OnServerConnect callback
	ServerConnected(sc any, lp any) bool

	// Asynchronously notify controlled block about server status
	// true is returned if controlled block has OnServerDisconnect callback
	ServerDisconnected(sc any) bool

	// Asynchronously call Finish callback of controlled block
	//
	Finish()
}
