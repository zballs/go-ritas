package consensus

import (
	"context"
	// "fmt"
	. "github.com/zballs/goRITAS/broadcast"
	. "github.com/zballs/goRITAS/types"
)

func VC_ControlBlock(parent *ControlBlock, env *Env) *ControlBlock {

	stage := NewStageVC(env)

	vccb := NewControlBlock(stage, parent, "vector")

	go VC_AddChildren(vccb, env)

	return vccb
}

func VC_AddChildren(vccb *ControlBlock, env *Env) {

	// Add RB, MV control blocks in order

	rbcb := RB_ControlBlock(vccb, env)
	vccb.Children <- rbcb

	mvcb := MV_ControlBlock(vccb, env)
	vccb.Children <- mvcb
}

func VC_AddChildMV(vccb *ControlBlock, env *Env) {

	// Add MV control block

	mvcb := MV_ControlBlock(vccb, env)

	go func() {
		vccb.Children <- mvcb
	}()

}

func VectorConsensus(vccb *ControlBlock, ctx context.Context, env *Env, args *Args) *Args {

	// Set stage
	args.Set(vccb.Stage)

	// Create child context
	childCtx, cancel := context.WithCancel(ctx)

	// Reliable broadcast
	vccb.Info("VC -> RB", "vccb", vccb.GetID())

	rbcb := vccb.GetChild()
	go ReliableBroadcast(rbcb, childCtx, env, args)

	var idx int
	var count uint32
	var vector *Vector
	var msg *Message
	var msgs Messages
	var payloads Payloads

	for {

		count = zero

		for {

			msgs = env.ReplicaDelivered(ctx, vccb)

			count = uint32(len(msgs))

			if count == env.N()-env.F()+vccb.GetRound() {
				break
			}

			// Wait for update before looping
			env.ReplicaWait(ctx, vccb)
		}

		// Cancel child context
		cancel()

		idx = 0

		payloads = make(Payloads, len(msgs))

		for idx, msg = range msgs {
			payloads[idx] = msg.GetPayload()
		}

		// Make vector of payloads
		vector = ToVector(payloads)

		// Set payload wrap of vector
		args.Set(vector.PayloadWrap())

		// Create new child context
		childCtx, cancel = context.WithCancel(ctx)

		// Multivalue consensus
		mvcb := vccb.GetChild()

		vccb.Info("VC -> MV", "vccb", vccb.GetID())
		args = MultivalueConsensus(mvcb, childCtx, env, args)

		if args.GetPayload() != EmptyPayload() {
			break
		}

		vccb.Info("Next round", "vccb", vccb.GetID())

		// Increment cb.Stage round
		vccb.IncrementRound()

		// Add MV child only
		VC_AddChildMV(vccb, env)

		// Set stage
		args.Set(vccb.Stage)
	}

	// Cancel child context
	cancel()

	vccb.Info("Finished", "vccb", vccb.GetID())

	return args
}
