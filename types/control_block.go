package types

import (
	log "github.com/inconshreveable/log15"
)

type ControlBlock struct {
	*Stage

	Parent   *ControlBlock
	Children chan *ControlBlock

	log.Logger
}

func NewControlBlock(stage *Stage, parent *ControlBlock, module string) *ControlBlock {

	return &ControlBlock{
		Stage:    stage,
		Parent:   parent,
		Children: make(chan *ControlBlock),
		Logger:   log.New("module", module),
	}
}

func (cb *ControlBlock) GetControlBlock() *ControlBlock {
	return cb
}

func (cb *ControlBlock) GetChild() *ControlBlock {

	child, ok := <-cb.Children

	if !ok {
		return nil
	}

	return child
}

func (cb *ControlBlock) OutOfContext(msg *Message, stages *Stages) bool {

	if !stages.SameStages(msg.GetStages()) {
		return true
	}

	if !cb.SameStage(msg.GetStage()) {
		return true
	}

	return false
}
