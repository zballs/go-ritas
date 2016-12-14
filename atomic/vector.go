package atomic

import (
	// "fmt"
	"context"
	. "github.com/zballs/goRITAS/broadcast"
	. "github.com/zballs/goRITAS/consensus"
	. "github.com/zballs/goRITAS/types"
	"sort"
)

const (
	zero uint32 = 0
	one  uint32 = 1
	two  uint32 = 2
)

func ABVEC_ControlBlock(env *Env) *ControlBlock {

	stage := NewStageAB(env)

	abcb := NewControlBlock(stage, nil, "atomic-vec")

	go ABVEC_AddChildren(abcb, env)

	return abcb
}

func ABVEC_AddChildren(abcb *ControlBlock, env *Env) {

	// Add RB, VC control blocks in order
	rbcb := RB_ControlBlock(abcb, env)
	abcb.Children <- rbcb

	vccb := VC_ControlBlock(abcb, env)
	abcb.Children <- vccb
}

func ABVEC_AddChildVC(abcb *ControlBlock, env *Env) {

	// Add VC control block
	vccb := VC_ControlBlock(abcb, env)
	abcb.Children <- vccb
}

// Atomic Broadcast with vector consensus

func AtomicBroadcastVEC(abcb *ControlBlock, ctx context.Context, env *Env, args *Args) {

	var childCtx context.Context
	var cancel context.CancelFunc

	var msgs Messages
	var payloads Payloads

	var vector *Vector
	var hashvector *HashVector

	var idx, lastLength int
	var counts map[string]uint32
	var added map[string]struct{}

	for {

		// Set stage
		args.Set(abcb.Stage)

		// Create child context
		childCtx, cancel = context.WithCancel(ctx)

		// Reliable broadcast
		abcb.Info("AB -> RB", "abcb", abcb.GetID())

		rbcb := abcb.GetChild()
		go ReliableBroadcast(rbcb, childCtx, env, args)

		idx, lastLength = 0, 0

	FOR_LOOP:
		for {

			for {

				msgs = env.ReplicaDelivered(ctx, abcb)

				if len(msgs) > lastLength {
					lastLength = len(msgs)
					break
				}

				// wait on update before looping
				env.ReplicaWait(ctx, abcb)
			}

			// Cancel child context
			cancel()

			count := uint32(len(msgs))

			payloads = make(Payloads, count)

			for idx, m := range msgs {
				payloads[idx] = m.GetPayload()
			}

			// Create hashvector from payloads
			hashvector = HashVectorFromPayloads(payloads)

			// Increment cb.Stage step
			abcb.IncrementStep()

			// Set stage and payload
			args.SetMultiple(abcb.Stage, hashvector.PayloadWrap())

			// Create new child context
			childCtx, cancel = context.WithCancel(ctx)

			// Vector Consensus
			abcb.Info("AB -> VC", "abcb", abcb.GetID())

			vccb := abcb.GetChild()
			args = VectorConsensus(vccb, childCtx, env, args)

			abcb.Info("VC finished", "abcb", abcb.GetID())

			vector = VectorFromPayload(args.GetPayload())

			var hashloads Hashloads
			counts = make(map[string]uint32)

			for _, p := range vector.GetPayloads() {

				// each payload is a hashvector
				hashvector = HashVectorFromPayload(p)
				added = make(map[string]struct{})

				for _, h := range hashvector.GetHashloads() {

					hexstr := h.ToHexstr()

					if _, exists := added[hexstr]; exists {
						continue
					}

					added[hexstr] = struct{}{}
					counts[hexstr]++

					if counts[hexstr] == env.F()+1 {
						hashloads = append(hashloads, h)
					}
				}
			}

			if len(hashloads) == 0 {

				continue

			}

			// Deterministic ordering of hashloads
			sort.Sort(hashloads)

			deliverAtomic := make(Messages, len(hashloads))

			for _, h1 := range hashloads {

				for idx, _ = range msgs {

					p := msgs[idx].GetPayload()

					h2 := p.ToHashload()

					if h1.IsEqual(h2) {

						args.SetMultiple(
							p, // payload
							msgs[idx].GetStages(),
							msgs[idx].GetSender())

						msg := ToMessageAtomic(args)

						deliverAtomic = append(deliverAtomic, msg)

						break
					}
				}

				if idx < len(msgs) {
					msgs = append(msgs[:idx], msgs[idx+1:]...)

				} else {

					// Add only the VC child
					ABVEC_AddChildVC(abcb, env)

					// Reset cb.Stage step
					abcb.ResetStep()

					continue FOR_LOOP
				}
			}

			for _, msg := range deliverAtomic {

				// Atomically deliver message
				env.ReplicaDeliver(ctx, msg, nil)
			}

			abcb.Info("Atomically delivered messages woohoo!", "abcb", abcb.GetID())

			// Add children again
			go ABVEC_AddChildren(abcb, env)

			// Increment cb.Stage round
			abcb.IncrementRound()

			// Reset cb.Stage step
			abcb.ResetStep()

			// Clear deliveries before advancing to next round??

			break
		}
	}
}
