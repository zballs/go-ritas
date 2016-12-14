package consensus

import (
	"context"
	. "github.com/zballs/goRITAS/broadcast"
	. "github.com/zballs/goRITAS/types"
	"log"
)

func MV_ControlBlock(parent *ControlBlock, env *Env) *ControlBlock {

	stage := NewStageMV(env)

	mvcb := NewControlBlock(stage, parent, "multivalue")

	go MV_AddChildren(mvcb, env)

	return mvcb
}

func MV_AddChildren(mvcb *ControlBlock, env *Env) {

	// Add RB, EB, BC control blocks in order
	rbcb := RB_ControlBlock(mvcb, env)
	mvcb.Children <- rbcb

	ebcb := EB_ControlBlock(mvcb, env)
	mvcb.Children <- ebcb

	bccb := BC_ControlBlock(mvcb, env)
	mvcb.Children <- bccb
}

func ValidPayloads(env *Env, payload *Payload, vector *Vector) bool {

	if payload == nil {
		return true
	}

	var count uint32

	for _, p := range vector.GetPayloads() {

		if !payload.IsEqual(p) {
			log.Printf("%v != %v", payload.String(), p.String())
			continue
		}

		count++

		if count == env.N()-2*env.F() {
			return true
		}
	}

	return false
}

func MultivalueConsensus(mvcb *ControlBlock, ctx context.Context, env *Env, args *Args) *Args {

	// Set stage
	args.Set(mvcb.Stage)

	// Create child context
	childCtx, cancel := context.WithCancel(ctx)

	// Reliable broadcast
	rbcb := mvcb.GetChild()

	mvcb.Info("MV -> RB", "mvcb", mvcb.GetID())
	go ReliableBroadcast(rbcb, childCtx, env, args)

	var count, value uint32
	var counts map[string]uint32

	var payload *Payload
	var payloads Payloads

	var msg *Message
	var msgs Messages

	for {

		msgs = env.ReplicaDelivered(ctx, mvcb)

		count = uint32(len(msgs))

		if count == env.N()-env.F() {
			break
		}

		// Wait for update before looping
		env.ReplicaWait(ctx, mvcb)
	}

	counts = make(map[string]uint32)

	for _, msg = range msgs {

		p := msg.GetPayload()

		hexstr := p.ToHexstr()

		counts[hexstr]++

		if counts[hexstr] == env.N()-2*env.F() {
			payload = p
			break
		}
	}

	// Cancel child context
	cancel()

	// Increment cb.Stage step
	mvcb.IncrementStep()

	// Set stage, payload, vector
	args.SetMultiple(mvcb.Stage, payload, ToVector(payloads))

	// Create new child context
	childCtx, cancel = context.WithCancel(ctx)

	// Echo broadcast
	ebcb := mvcb.GetChild()

	mvcb.Info("MV -> EB", "mvcb", mvcb.GetID())
	go EchoBroadcast(ebcb, childCtx, env, args)

FOR_LOOP_1:
	for {

		msgs = env.ReplicaDelivered(ctx, mvcb)

		payloads = nil

		// TODO: make not redundant

		for _, msg = range msgs {

			p := msg.GetPayload()
			v := msg.GetVector()

			if ValidPayloads(env, p, v) {

				payloads = append(payloads, p)

				count = uint32(len(payloads))

				if count == env.N()-env.F() {
					break FOR_LOOP_1
				}
			} else {
				mvcb.Info("Invalid Payloads")
			}
		}

		env.ReplicaWait(ctx, mvcb)
	}

	// Cancel child context
	cancel()

	value = one
	counts = make(map[string]uint32)

	mvcb.Info("Payloads", "length", len(payloads))

FOR_LOOP_2:
	for _, p := range payloads {

		hexstr := p.ToHexstr()

		counts[hexstr]++

		if counts[hexstr] == env.N()-2*env.F() {

			payload = p

			for _, p := range payloads {

				if p != nil && p != payload {
					mvcb.Info("Broken")
					break FOR_LOOP_2

				}
			}

			value = two
		}
	}

	// Set payload
	args.Set(PayloadFromUint32(value))

	// Create new child context
	childCtx, cancel = context.WithCancel(ctx)
	defer cancel()

	// Binary consensus
	bccb := mvcb.GetChild()

	mvcb.Info("MV -> BC", "mvcb", mvcb.GetID())
	args = BinaryConsensus(bccb, childCtx, env, args)

	if args.GetPayload().ToUint32() == one {

		// set empty payload
		args.Set(EmptyPayload())

		return args
	}

	if value == two {

		// set payload
		args.Set(payload)

		return args
	}

	counts = make(map[string]uint32)

FOR_LOOP_3:
	for {

		msgs = env.ReplicaDelivered(ctx, mvcb)

		payloads = nil

		for _, msg = range msgs {

			p := msg.GetPayload()
			v := msg.GetVector()

			if ValidPayloads(env, p, v) {

				hexstr := p.ToHexstr()
				counts[hexstr]++

				if counts[hexstr] == env.N()-2*env.F() {
					payload = p
					break FOR_LOOP_3
				}
			}
		}

		env.ReplicaWait(ctx, mvcb)
	}

	// set payload
	args.Set(payload)

	return args
}
