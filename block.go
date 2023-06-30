package sputnik

import "fmt"

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
//   - optional:	OnServerConnect|OnServerDisconnect|OnEvent
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

// Finish callback is executed by sputnik once during shutdown of the process.
// Blocks are finished in reverse order.
//
// For tests sputnik supports possibility to run Init/Finish blocks per test.
type Finish func()

// Optional OnServerConnect callback is executed by sputnik after successful
// connection to server.
type OnServerConnect func(connection any, logger any)

// Optional OnServerDisconnect callback is executed by sputnik when previously
// connected server disconnects.
type OnServerDisconnect func(connection any)

// Because asynchronous nature of blocks, negotiation between blocks done using 'events'
// Developers of blocks should agree on content of events.
// sputnik doesn't force specific format of the event
// with one exception: key of the map should not start from "__".
// This prefix is used by sputnik for house-keeping values.
// You can call Notify with nil event, but on the side of the receiver it may be not nil,
// because data added by sputnik
type Event map[string]any

// Optional OnEvent callback is executed by sputnik as result of receiving Event.
// Block can send event to itself.
// Unlike other callbacks, OnEvent called sequentially one by one from the same goroutine.
type OnEvent func(ev Event)

// Simplified Block life cycle:
//   - Init
//   - Run
//   - OnServerConnect
//   - [*]OnEvent
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
	OnEvent      OnEvent
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

	// Asynchronously send event to controlled block
	// true is returned if controlled block has OnEvent callback
	Notify(ev Event) bool

	// Asynchronously notify controlled block about server status
	// true is returned if controlled block has OnServerConnect callback
	ServerConnected(sc any, lp any) bool

	// Asynchronously notify controlled block about server status
	// true is returned if controlled block has OnServerDisconnect callback
	ServerDisconnected(sc any) bool

	// Asynchronously call Finish callback of controlled block
	//
	Finish()
}

type cloneableBlockController interface {
	BlockController
	Clone() cloneableBlockController
}

type activeBlock struct {
	bd  BlockDescriptor
	bl  Block
	cbc cloneableBlockController
}

func (abl *activeBlock) init(cnf any) error {
	err := abl.bl.Init(cnf)

	if err != nil {
		err = fmt.Errorf("Init of [%s,%s] failed with error %s", abl.bd.Name, abl.bd.Responsibility, err.Error())
	}

	return err
}

func (abl *activeBlock) finish() {
	// For interception
	abl.bl.Finish()
}

func newActiveBlock(bd BlockDescriptor, bl Block) activeBlock {
	return activeBlock{bd, bl, nil}
}

type activeBlocks []activeBlock

func (abs activeBlocks) getABl(resp string) *activeBlock {
	var res *activeBlock
	for _, ab := range abs {
		if ab.bd.Responsibility == resp {
			return &ab
		}
	}
	return res
}
