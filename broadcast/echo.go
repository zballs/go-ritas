package broadcast

import (
	"context"
	. "github.com/zballs/goRITAS/types"
)

func EB_ControlBlock(parent *ControlBlock, env *Env) *ControlBlock {

	stage := NewStageEB(env)

	ebcb := NewControlBlock(stage, parent, "echo")
	//no children

	return ebcb
}

// Echo broadcast

func EchoBroadcast(ebcb *ControlBlock, ctx context.Context, env *Env, args *Args) {

	// Panic recovery
	defer func() {
		if r := recover(); r != nil {
			ebcb.Info(r.(string), "ebcb", ebcb.GetID())
		}
	}()

	// Get stage
	stage := &*(args.GetStage())

	// Set stage, sender, broadcaster
	args.SetMultiple(ebcb.Stage, env.Sender(), env.Broadcaster())

	// Get stages
	stages := &*(args.GetStages())

	// Create initial message
	msg := ToMessageProtocol(args)

	// Multicast message
	MulticastMessage(ctx, env, msg)

	echos := make(map[string][]uint32)

	for {

		// Keep reading/delivering until ctx is done..

	FOR_LOOP_1:
		for {

			msg = env.ServerGetMessage(ctx)

			if ebcb.OutOfContext(msg, stages) {
				// handle out of context message
				// ebcb.Warn("Out of context", "msg", msg.String())
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

					break FOR_LOOP_1
				}

			default:
				ebcb.Error("Invalid step", "stage", msg.GetStage().String())
			}
		}

		// Increment cb.Stage step
		ebcb.IncrementStep()

		// Set stage, broadcaster
		args = msg.ToArgs()
		args.SetMultiple(ebcb.Stage, env.Broadcaster())

		// Multicast message
		msg = ToMessageProtocol(args)

		MulticastMessage(ctx, env, msg)

		ebcb.Info("Step 2")

	FOR_LOOP_2:
		for {

			msg = env.ServerGetMessage(ctx)

			if ebcb.OutOfContext(msg, stages) {
				// handle out of context message
				// ebcb.Warn("Out of context", "msg", msg.String())
				env.ServerPutMessage(ctx, msg) //for now
				continue
			}

			step := msg.GetStage().GetStep()

			switch step {

			case 1:

				// Set stage, broadcaster
				args = msg.ToArgs()
				args.SetMultiple(ebcb.Stage, env.Broadcaster())

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

				if count == 2*env.F()+1 {

					// Deliver message to replica
					env.ReplicaDeliver(ctx, msg, stage)

					// ebcb.Info("Delivery", "stage", stage.GetValue(), "ebcb", ebcb.GetID())

					break FOR_LOOP_2
				}

			default:
				ebcb.Error("Invalid step", "stage", msg.GetStage().String())
			}
		}

		// Reset cb.Stage step
		ebcb.ResetStep()

		ebcb.Info("Looping")
	}
}
