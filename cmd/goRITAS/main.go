package main

import (
	"context"
	. "github.com/tendermint/go-common"
	. "github.com/zballs/goRITAS/atomic"
	. "github.com/zballs/goRITAS/types"
)

const (
	zero uint32 = 0
	one  uint32 = 1
	two  uint32 = 2
)

var Payload1 = PayloadFromUint32(one)
var Payload2 = PayloadFromUint32(two)

var IDAddrs = []*IDAddr{
	NewIDAddr(uint32(0), "127.0.0.1:10000"),
	NewIDAddr(uint32(1), "127.0.0.1:10001"),
	NewIDAddr(uint32(2), "127.0.0.1:10002"),
	NewIDAddr(uint32(3), "127.0.0.1:10003"),
}

/*
	NewIDAddr(uint32(4), "127.0.0.1:10004"),
	NewIDAddr(uint32(5), "127.0.0.1:10005"),
	NewIDAddr(uint32(6), "127.0.0.1:10006"),
*/

var Payloadz = Payloads{
	PayloadFromUint32(one),
	PayloadFromUint32(one),
	PayloadFromUint32(one),
	PayloadFromUint32(one),
}

func main() {

	for idx, ida := range IDAddrs {

		// Init and run env
		env := InitEnv(ida, IDAddrs)
		env.Run()

		// Cb for atomic broadcast with vector consensus
		// abcb := ABVEC_ControlBlock(env)

		// Cb for atomic broadcast with multivalue consensus
		abcb := ABMV_ControlBlock(env)

		// Context
		ctx := context.Background()

		// Args
		args := NewArgs()
		args.SetArgs(Payloadz[idx])

		// AtomicBroadcastVEC
		go AtomicBroadcastMV(abcb, ctx, env, args)
	}

	// Wait
	TrapSignal(func() {
		//Cleanup
	})
}
