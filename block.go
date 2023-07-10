package sputnik

// Block has Name (analog of golang type) and Responsibility (instance of specific block)
// This separation allows to run simultaneously blocks with the same Name.
// Other possibility - blocks with different name but with the same responsibility,
// e.g. different implementation of "finisher" depends on environment.
//
// initiator block has predetermined Responsibility - "initiator"
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
type OnServerConnect func(connection any)

// Optional OnServerDisconnect callback is executed by sputnik when previously
// connected server disconnects.
type OnServerDisconnect func()

// Because asynchronous nature of blocks, negotiation between blocks done using 'messages'
// Message may be command|query|event|update|...
// Developers of blocks should agree on content of messages.
// sputnik doesn't force specific format of the message
// with one exception: key of the map should not start from "__".
// This prefix is used by sputnik for house-keeping values.
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
// After Run order of callbacks will be unpredictable.
type Block struct {
	init         Init
	run          Run
	finish       Finish
	onConnect    OnServerConnect
	onDisconnect OnServerDisconnect
	onMsg        OnMsg
}

type BlockOption = func(b *Block)

func NewBlock(opts ...BlockOption) *Block {
	blk := new(Block)
	for _, opt := range opts {
		opt(blk)
	}
	return blk
}

func WithInit(f Init) BlockOption {
	return func(b *Block) {
		b.init = f
	}
}

func WithRun(f Run) BlockOption {
	return func(b *Block) {
		b.run = f
	}
}

func WithFinish(f Finish) BlockOption {
	return func(b *Block) {
		b.finish = f
	}
}

func WithOnConnect(f OnServerConnect) BlockOption {
	return func(b *Block) {
		b.onConnect = f
	}
}

func WithOnDisConnect(f OnServerDisconnect) BlockOption {
	return func(b *Block) {
		b.onDisconnect = f
	}
}

func WithOnMsg(f OnMsg) BlockOption {
	return func(b *Block) {
		b.onMsg = f
	}
}

// Check presence of mandatory callbacks
func (bl *Block) isValid() bool {
	return bl.init != nil && bl.run != nil && bl.finish != nil
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
	//  - msg != nil
	Send(msg Msg) bool

	// Asynchronously notify controlled block about server status
	// true is returned if if controlled block has OnServerConnect callback
	ServerConnected(sc any) bool

	// Asynchronously notify controlled block about server status
	// true is returned if controlled block has OnServerDisconnect callback
	ServerDisconnected() bool

	// Asynchronously call Finish callback of controlled block
	//
	Finish()
}
