package sputnik

import "fmt"

type activeBlock struct {
	descriptor BlockDescriptor
	block      *Block
	controller *controller
}

func (abl *activeBlock) init(cf ConfFactory) error {
	err := abl.block.init(cf)

	if err != nil {
		err = fmt.Errorf("Init of [%s,%s] failed with error %s", abl.descriptor.Name, abl.descriptor.Responsibility, err.Error())
	}

	return err
}

func (abl *activeBlock) finish() {
	// For interception
	abl.block.finish(true)
}

func newActiveBlock(bd BlockDescriptor, bl *Block) activeBlock {
	return activeBlock{bd, bl, nil}
}

type activeBlocks []*activeBlock

func (abs activeBlocks) getABl(resp string) (*activeBlock, bool) {
	for _, ab := range abs {
		if ab.descriptor.Responsibility == resp {
			return ab, true
		}
	}
	return nil, false
}
