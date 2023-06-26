package sputnik

//
// Block has Name (analog of golang type) and Responsibility (instance of specific block)
// This separation allows to run simultaneously blocks with the same Name.
// Block has set of the callbacks:
//   - mandatory:	Init|Run|Finish
//   - optional:	OnServerConnect|OnServerDisconnect|OnUpdate
//
//
// Simplified Block life cycle:
//   - Init
//   - Run
//   - OnServerConnect
//   - [*]OnUpdate
//   - OnServerDisconnect
//   - Finish
//
// Please pay attention, that after Run order of callbacks will be unpredictable.
//

type Block struct {
	Name           string
	Responsibility string

	Init         Init
	Run          Run
	Finish       Finish
	OnConnect    OnServerConnect
	OnDisconnect OnServerDisconnect
	OnUpdate     OnUpdate
}

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

// Finish callback is executed by sputnik once during shutdown of the process.
// Blocks are finished in reverse order.
//
// For tests sputnik supports possibility to run Init/Finish blocks per test.
type Finish func() error

// Optional OnServerConnect callback is executed by sputnik after successful
// connection to server.
type OnServerConnect func(sc any, lp any)

// Optional OnServerDisconnect callback is executed by sputnik when previously
// connected server disconnects.
type OnServerDisconnect func(sc any)

// Because asynchronous nature of blocks, negotiation between blocks done using 'bags'
// Developers of blocks should agree on content of bags.
// sputnik doesn't force specific format of the bag
// with one exception: key of the map should not start from "__".
// This prefix is used by sputnik for house-keeping values.
// You can call Update with nil bug, but on the side of the receiver it may be not nil,
// because data added by sputnik
type Bag map[string]any

// Optional OnUpdate callback is executed by sputnik as result of receiving Bag.
// Block can send bag to itself.
type OnUpdate func(bb Bag)

// BlockController provides possibility for negotiation between blocks
// Block gets own controller as parameter of Run
type BlockController interface {
	Name() string
	Responsibility() string

	// Asynchronously send bag to controlled block
	// true is returned if controlled block has OnUpdate callback
	Update(bb Bag) bool

	// Asynchronously notify controlled block about server status
	// true is returned if controlled block has OnServerConnect callback
	ServerConnected(sc any, lp any) bool

	// Asynchronously notify controlled block about server status
	// true is returned if controlled block has OnServerDisconnect callback
	ServerDisconnected(sc any) bool

	// Asynchronously call Finish callback of controlled block
	//
	Finish()

	// Get controller of block
	//
	Controller(resp string) (bc BlockController, exists bool)

	// For using in another goroutine
	//
	Clone() (BlockController, error)
}
