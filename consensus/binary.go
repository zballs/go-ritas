package consensus

import (
	"context"
	"crypto/rand"
	. "github.com/zballs/goRITAS/broadcast"
	. "github.com/zballs/goRITAS/types"
	"math/big"
)

const (
	zero uint32 = 0
	one  uint32 = 1
	two  uint32 = 2
)

func BC_ControlBlock(parent *ControlBlock, env *Env) *ControlBlock {

	stage := NewStageBC(env)

	bccb := NewControlBlock(stage, parent, "binary")

	go BC_AddChildren(bccb, env)

	return bccb
}

func BC_AddChildren(bccb *ControlBlock, env *Env) {

	rbcb1 := RB_ControlBlock(bccb, env)
	bccb.Children <- rbcb1

	rbcb2 := RB_ControlBlock(bccb, env)
	bccb.Children <- rbcb2

	rbcb3 := RB_ControlBlock(bccb, env)
	bccb.Children <- rbcb3
}

func BinaryConsensus(bccb *ControlBlock, ctx context.Context, env *Env, args *Args) *Args {

	var msg *Message
	var msgs Messages

	var payload *Payload
	var done bool

	var childCtx context.Context
	var cancel context.CancelFunc

	var count, result, ones, twos uint32

	for {

		// Set stage
		args.Set(bccb.Stage)

		// Create child context
		childCtx, cancel = context.WithCancel(ctx)

		// Reliable broadcast
		rbcb := bccb.GetChild()

		bccb.Info("BC -> RB1", "bccb", bccb.GetID())
		go ReliableBroadcast(rbcb, childCtx, env, args)

		for {

			// Get msgs with cb.Stage(ID, step)
			msgs = env.ReplicaDelivered(ctx, bccb)

			count = uint32(len(msgs))

			if count == env.N()-env.F() {
				break
			}

			// Wait for update before looping
			env.ReplicaWait(ctx, bccb)
		}

		// Cancel child context
		cancel()

		result = one //set to one
		ones = zero
		twos = zero

		for _, msg = range msgs {

			payload = msg.GetPayload()

			if payload == nil {
				continue
			}

			if payload.ToUint32() == one {
				ones++
				if ones == (env.N()-env.F())/2 {
					break
				}
			}

			if payload.ToUint32() == two {
				twos++
				if twos == (env.N()-env.F())/2 {
					result = two
					break
				}
			}
		}

		// Increment cb.Stage step
		bccb.IncrementStep()

		// Set stage and payload
		args.SetMultiple(bccb.Stage, PayloadFromUint32(result))

		// Create new child context
		childCtx, cancel = context.WithCancel(ctx)

		// Reliable broadcast
		rbcb = bccb.GetChild()

		bccb.Info("BC -> RB2", "bccb", bccb.GetID())
		go ReliableBroadcast(rbcb, childCtx, env, args)

		for {

			msgs = env.ReplicaDelivered(ctx, bccb)

			count = uint32(len(msgs))

			if count == env.N()-env.F() {
				break
			}

			env.ReplicaWait(ctx, bccb)
		}

		// Cancel child context
		cancel()

		result = zero //set to zero
		ones = zero
		twos = zero

		for _, msg = range msgs {

			payload = msg.GetPayload()

			if payload == nil {
				continue
			}

			if payload.ToUint32() == one {
				ones++
				if ones == env.N()/2 {
					result = one
					break
				}
			}

			if payload.ToUint32() == two {
				twos++
				if twos == env.N()/2 {
					result = two
					break
				}
			}
		}

		// Increment cb.Stage step
		bccb.IncrementStep()

		// Set stage and payload
		args.SetMultiple(bccb.Stage, PayloadFromUint32(result))

		// Create new child context
		childCtx, cancel = context.WithCancel(ctx)

		// Output to reliable broadcast
		rbcb = bccb.GetChild()

		bccb.Info("BC -> RB3", "bccb", bccb.GetID())
		go ReliableBroadcast(rbcb, childCtx, env, args)

		// randomly set result to one or two
		randnum, _ := rand.Int(rand.Reader, big.NewInt(2))
		result = uint32(randnum.Int64() + 1)

		// BC3
		for {

			msgs = env.ReplicaDelivered(ctx, bccb)

			count = uint32(len(msgs))

			if count == env.N()-env.F() {
				break
			}

			env.ReplicaWait(ctx, bccb)
		}

		ones = zero
		twos = zero

		for _, msg = range msgs {

			payload = msg.GetPayload()

			if payload == nil {
				continue
			}

			if payload.ToUint32() == one {
				ones++
				if ones == 2*env.F()+1 {
					result = one
					done = true
					break
				}
			}

			if payload.ToUint32() == two {
				twos++
				if twos == 2*env.F()+1 {
					result = two
					done = true
					break
				}
			}
		}

		if done {

			// Set payload with result value
			args.Set(PayloadFromUint32(result))

			return args
		}

		// Reset step
		bccb.ResetStep()

		// Add children again
		go BC_AddChildren(bccb, env)
	}
}
