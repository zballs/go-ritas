package atomic

import (
	"context"
	. "github.com/zballs/goRITAS/broadcast"
	. "github.com/zballs/goRITAS/consensus"
	. "github.com/zballs/goRITAS/types"
	"sort"
)

func ABMV_ControlBlock(env *Env) *ControlBlock {

	stage := NewStageAB(env)

	abcb := NewControlBlock(stage, nil, "atomic-mv")

	go ABMV_AddChildren(abcb, env)

	return abcb
}

func ABMV_AddChildren(abcb *ControlBlock, env *Env) {

	rbcb1 := RB_ControlBlock(abcb, env)
	abcb.Children <- rbcb1

	rbcb2 := RB_ControlBlock(abcb, env)
	abcb.Children <- rbcb2

	mvcb := MV_ControlBlock(abcb, env)
	abcb.Children <- mvcb
}

// Atomic Broadcast with multivalue consensus

func AtomicBroadcastMV(abcb *ControlBlock, ctx context.Context, env *Env, args *Args) {

	var msg *Message
	var msgs Messages

	var payload *Payload
	var payloads Payloads

	var vector *Vector

	var idx, lastLength int
	var count uint32
	var counts map[string]uint32
	var added map[string]struct{}

	var childCtx context.Context
	var cancel context.CancelFunc

	windows := make([]uint32, env.N())

FOR_LOOP_1:
	for {

		// Set stage
		args.SetArg(abcb.Stage)

		// Create child context
		childCtx, cancel = context.WithCancel(ctx)

		// Reliable broadcast
		rbcb := abcb.GetChild()

		abcb.Info("AB -> RB1", "abcb", abcb.GetID())
		go ReliableBroadcast(rbcb, childCtx, env, args)

		lastLength = 0

		for {

			for {

				msgs = env.ReplicaDelivered(ctx, abcb)

				if len(msgs) > lastLength {
					lastLength = len(msgs)
					break
				}

				env.ReplicaWait(ctx, abcb)
			}

			// Cancel child context
			cancel()

			payloads = nil

			for _, msg = range msgs {

				round := msg.GetStage().GetRound()
				sender := msg.GetSender()

				if windows[sender.ID] > round {
					continue
				}

				payloads = append(payloads, PayloadFromUint32s(round, sender.ID))
			}

			// If payloads is nil??

			vector = ToVector(payloads)

			// Increment step
			abcb.IncrementStep()

			// Create new child context
			childCtx, cancel = context.WithCancel(ctx)

			// Set stage and payload
			args.SetArgs(abcb.Stage, vector.PayloadWrap())

			// Reliable broadcast
			rbcb = abcb.GetChild()

			abcb.Info("AB -> RB2", "abcb", abcb.GetID())
			go ReliableBroadcast(rbcb, childCtx, env, args)

			for {

				msgs = env.ReplicaDelivered(ctx, abcb)

				count = uint32(len(msgs))

				if count == env.N()-env.F() {
					break
				}

				env.ReplicaWait(ctx, abcb)
			}

			// Cancel child context
			cancel()

			payloads = nil

			counts = make(map[string]uint32)

			for _, msg := range msgs {

				payload = msg.GetPayload()
				vector = VectorFromPayload(payload)

				added = make(map[string]struct{})

				for _, p := range vector.GetPayloads() {

					hexstr := p.ToHexstr()

					if _, exists := added[hexstr]; exists {
						// if vector has payload = (senderID, round)
						// only count it once..
						continue
					}

					added[hexstr] = struct{}{}
					counts[hexstr]++

					if counts[hexstr] == env.F()+1 {

						uints := p.ToUint32s()
						round, senderID := uints[0], uints[1]

						if windows[senderID] <= round { // && round < L
							payloads = append(payloads, p)
						}
					}
				}
			}

			vector = ToVector(payloads)

			// Increment step
			abcb.IncrementStep()

			// Create new child context
			childCtx, cancel = context.WithCancel(ctx)

			// Set stage and payload
			args.SetArgs(abcb.Stage, vector.PayloadWrap())

			// Output to multivalue consensus
			mvcb := abcb.GetChild()

			abcb.Info("AB -> MV", "abcb", abcb.GetID())
			args = MultivalueConsensus(mvcb, childCtx, env, args)

			payload = args.GetPayload()
			vector = VectorFromPayload(payload)

			var deliverAtomic Messages

		FOR_LOOP_2:
			for {

				msgs = env.ReplicaDelivered(ctx, abcb)

				for _, p := range vector.GetPayloads() {

					uints := p.ToUint32s()

					round, senderID := uints[0], uints[1]

					for idx, _ = range msgs {

						r := msgs[idx].GetStage().GetRound()
						sender := msgs[idx].GetSender()

						if r == round && sender.ID == senderID {

							payload = msgs[idx].GetPayload()
							stages := msgs[idx].GetStages()

							// Set args
							args.SetArgs(payload, stages, sender)

							msg = ToMessageAtomic(args)

							deliverAtomic = append(deliverAtomic, msg)

							break
						}
					}

					if idx < len(msgs) {

						msgs = append(msgs[:idx], msgs[idx+1:]...)

					} else {

						env.ReplicaWait(ctx, abcb)

						continue FOR_LOOP_2
					}
				}

				break
			}

			if len(deliverAtomic) == 0 {
				continue
			}

			// Deterministic ordering
			sort.Sort(deliverAtomic)

			for _, msg = range deliverAtomic {

				// Atomically deliver msg
				env.ReplicaDeliver(ctx, msg, nil) //no stage
			}

			abcb.Info("Atomically delivered messages woohoo!", "abcb", abcb.GetID())

			// Increment windows

			msgs = env.ReplicaAtomic(ctx)

			for _, msg = range msgs {

				sender := msg.GetSender()
				windows[sender.ID]++
			}

			// Add children again
			go ABMV_AddChildren(abcb, env)

			// Increment cb.Stage round
			abcb.IncrementRound()

			// Reset cb.Stage step
			abcb.ResetStep()

			// Clear deliveries before advancing to next round??

			continue FOR_LOOP_1
		}
	}
}
