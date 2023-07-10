package sputnik

import "fmt"

type activeBlock struct {
	bd BlockDescriptor
	bl *Block
	bc BlockController
}

func (abl *activeBlock) init(cnf any) error {
	err := abl.bl.init(cnf)

	if err != nil {
		err = fmt.Errorf("Init of [%s,%s] failed with error %s", abl.bd.Name, abl.bd.Responsibility, err.Error())
	}

	return err
}

func (abl *activeBlock) finish() {
	// For interception
	abl.bl.finish(true)
}

func newActiveBlock(bd BlockDescriptor, bl *Block) activeBlock {
	return activeBlock{bd, bl, nil}
}

type activeBlocks []*activeBlock

func (abs activeBlocks) getABl(resp string) (*activeBlock, bool) {
	for _, ab := range abs {
		if ab.bd.Responsibility == resp {
			return ab, true
		}
	}
	return nil, false
}
