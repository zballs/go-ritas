package broadcast

import (
	"context"
	// "fmt"
	. "github.com/zballs/goRITAS/types"
)

func RB_ControlBlock(parent *ControlBlock, env *Env) *ControlBlock {

	stage := NewStageRB(env)

	rbcb := NewControlBlock(stage, parent, "reliable")
	//no children

	return rbcb
}

// Reliable Broadcast

func ReliableBroadcast(rbcb *ControlBlock, ctx context.Context, env *Env, args *Args) {

	// Panic recovery
	defer func() {
		if r := recover(); r != nil {
			rbcb.Info(r.(string), "rbcb", rbcb.GetID())
		}
	}()

	// Get stage
	stage := &*(args.GetStage())

	// Set stage, sender, broadcaster
	args.SetMultiple(rbcb.Stage, env.Sender(), env.Broadcaster())

	// Get stages
	stages := &*(args.GetStages())

	// Create initial message
	msg := ToMessageProtocol(args)

	// Multicast
	MulticastMessage(ctx, env, msg)

	echos := make(map[string][]uint32)
	readys := make(map[string][]uint32)

	var ready bool

	for {

		// Keep reading/delivering until ctx is done..

		ready = false

	FOR_LOOP_1:
		for {

			msg = env.ServerGetMessage(ctx)

			if rbcb.OutOfContext(msg, stages) {
				// handle out of context message
				env.ServerPutMessage(ctx, msg) //for now
				continue
			}

			step := msg.GetStage().GetStep()

			switch step {

			case 1:

				break FOR_LOOP_1

			case 2:

				hexstr := msg.DigestHexstr()
				broadcaster := msg.GetBroadcaster()

				if len(echos[hexstr]) > 0 {
					for _, broadcasterID := range echos[hexstr] {
						if broadcaster.ID == broadcasterID {
							continue FOR_LOOP_1
						}
					}
				}

				echos[hexstr] = append(echos[hexstr], broadcaster.ID)

				count := uint32(len(echos[hexstr]))

				if count == (env.N()+env.F())/2 {

					ready = true

					break FOR_LOOP_1
				}

			case 3:

				hexstr := msg.DigestHexstr()
				broadcaster := msg.GetBroadcaster()

				if len(readys[hexstr]) > 0 {
					for _, broadcasterID := range readys[hexstr] {
						if broadcaster.ID == broadcasterID {
							continue FOR_LOOP_1
						}
					}
				}

				readys[hexstr] = append(readys[hexstr], broadcaster.ID)

				count := uint32(len(readys[hexstr]))

				if count == env.F()+1 {

					ready = true

					break FOR_LOOP_1
				}

			default:
				rbcb.Error("Invalid step", "stage", msg.GetStage().String())
			}
		}

		// Increment cb.Stage step
		rbcb.IncrementStep()

		// Set stage, broadcaster
		args = msg.ToArgs()
		args.SetMultiple(rbcb.Stage, env.Broadcaster())

		// Multicast message
		msg = ToMessageProtocol(args)

		MulticastMessage(ctx, env, msg)

		if ready {

			// Increment cb.Stage step
			rbcb.IncrementStep()

			// Set stage
			args.Set(rbcb.Stage)

			// Multicast message
			msg = ToMessageProtocol(args)

			MulticastMessage(ctx, env, msg)

		} else {

		FOR_LOOP_2:
			for {

				msg = env.ServerGetMessage(ctx)

				if rbcb.OutOfContext(msg, stages) {
					// handle out of context message
					env.ServerPutMessage(ctx, msg) //for now
					continue
				}

				step := msg.GetStage().GetStep()

				switch step {

				case 1:

					// Set stage, broadcaster
					args = msg.ToArgs()
					args.SetMultiple(rbcb.Stage, env.Broadcaster())

					// Multicast message
					msg = ToMessageProtocol(args)

					MulticastMessage(ctx, env, msg)

				case 2:

					hexstr := msg.DigestHexstr()
					broadcaster := msg.GetBroadcaster()

					if len(echos[hexstr]) > 0 {
						for _, broadcasterID := range echos[hexstr] {
							if broadcaster.ID == broadcasterID {
								continue FOR_LOOP_2
							}
						}
					}

					echos[hexstr] = append(echos[hexstr], broadcaster.ID)

					count := uint32(len(echos[hexstr]))

					if count == (env.N()+env.F())/2 {
						break FOR_LOOP_2
					}

				case 3:

					hexstr := msg.DigestHexstr()
					broadcaster := msg.GetBroadcaster()

					if len(readys[hexstr]) > 0 {
						for _, broadcasterID := range readys[hexstr] {
							if broadcaster.ID == broadcasterID {
								continue FOR_LOOP_2
							}
						}
					}

					readys[hexstr] = append(readys[hexstr], broadcaster.ID)

					count := uint32(len(readys[hexstr]))

					if count == env.F()+1 {

						break FOR_LOOP_2
					}

				default:
					rbcb.Error("Invalid step", "stage", msg.GetStage().String())
				}
			}

			// Increment cb.Stage step
			rbcb.IncrementStep()

			// Set stage, broadcaster
			args = msg.ToArgs()
			args.SetMultiple(rbcb.Stage, env.Broadcaster())

			// Multicast message
			msg = ToMessageProtocol(args)

			MulticastMessage(ctx, env, msg)
		}

	FOR_LOOP_3:
		for {

			msg = env.ServerGetMessage(ctx)

			if rbcb.OutOfContext(msg, stages) {
				// handle out of context message
				// rbcb.Warn("Out of context", "msg", msg.String())
				env.ServerPutMessage(ctx, msg) //for now
				continue
			}

			step := msg.GetStage().GetStep()

			switch step {

			case 1:

				env.ServerPutMessage(ctx, msg)

			case 2:

				hexstr := msg.DigestHexstr()
				broadcaster := msg.GetBroadcaster()

				if len(echos[hexstr]) > 0 {
					for _, broadcasterID := range echos[hexstr] {
						if broadcaster.ID == broadcasterID {
							continue FOR_LOOP_3
						}
					}
				}

				echos[hexstr] = append(echos[hexstr], broadcaster.ID)

				count := uint32(len(echos[hexstr]))

				if count == (env.N()+env.F())/2 {

					// Set stage, broadcaster
					args = msg.ToArgs()
					args.SetMultiple(rbcb.Stage, env.Broadcaster())

					// Multicast message
					msg = ToMessageProtocol(args)

					MulticastMessage(ctx, env, msg)
				}

			case 3:

				hexstr := msg.DigestHexstr()
				broadcaster := msg.GetBroadcaster()

				if len(readys[hexstr]) > 0 {
					for _, broadcasterID := range readys[hexstr] {
						if broadcaster.ID == broadcasterID {
							continue FOR_LOOP_3
						}
					}
				}

				readys[hexstr] = append(readys[hexstr], broadcaster.ID)

				count := uint32(len(readys[hexstr]))

				if count == 2*env.F()+1 {

					// Deliver message to replica
					env.ReplicaDeliver(ctx, msg, stage)

					// rbcb.Info("Delivery", "stage", stage.GetValue(), "rbcb", rbcb.GetID())

					break FOR_LOOP_3
				}

			default:
				rbcb.Error("Invalid step", "stage", msg.GetStage().String())
			}
		}

		// Reset cb.Stage step
		rbcb.ResetStep()
	}
}
