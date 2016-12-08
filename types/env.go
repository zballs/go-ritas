package types

import (
	"context"
	"sync"
)

// IDAddr

type IDAddr struct {
	ID   uint32
	Addr string
}

func NewIDAddr(ID uint32, addr string) *IDAddr {
	return &IDAddr{ID, addr}
}

func (ida1 *IDAddr) Different(ida2 *IDAddr) bool {

	if ida1.ID == ida2.ID || ida1.Addr == ida2.Addr {
		return false
	}

	return true
}

// Env

type Env struct {
	ida *IDAddr

	replica *Replica
	server  *Server

	n uint32
	f uint32

	protoID uint32
	mtx     sync.Mutex

	group []*IDAddr
}

func InitEnv(ida *IDAddr, idas []*IDAddr) *Env {

	env := &Env{
		ida:     ida,
		replica: NewReplica(ida.ID),
		server:  NewServer(ida.Addr),
	}

	for _, ida = range idas {

		if !env.AddToGroup(ida) {
			//Log
		}
	}

	env.n = uint32(len(env.group))
	env.f = (env.n - 1) / 3

	return env
}

func (env *Env) Run() error {

	// Start server
	err := env.server.Start()

	if err != nil {
		return err
	}

	// Accept conns, read messages
	go env.server.ReadRoutine()

	return nil
}

func (env *Env) AddToGroup(ida1 *IDAddr) bool {

	if len(env.group) > 0 {

		for _, ida2 := range env.group {

			if !ida1.Different(ida2) {
				return false
			}
		}
	}

	env.group = append(env.group, ida1)
	return true
}

func (env *Env) N() uint32 {
	return env.n
}

func (env *Env) F() uint32 {
	return env.f
}

func (env *Env) ProtoID() uint32 {

	env.mtx.Lock()
	defer env.mtx.Unlock()

	protoID := env.protoID
	env.protoID++

	return protoID
}

func (env *Env) Sender() *Sender {
	return ToSender(env.ida)
}

func (env *Env) Broadcaster() *Broadcaster {
	return ToBroadcaster(env.ida)
}

func (env *Env) Addrs() []string {

	addrs := make([]string, env.n)

	for idx, ida := range env.group {

		addrs[idx] = ida.Addr
	}

	return addrs
}

// Panic

func IfDonePanic(ctx context.Context) {

	select {

	case <-ctx.Done():

		panic("Context canceled")

	default:
		//Do nothing
	}
}

// Server ops

func (env *Env) ServerGetMessage(ctx context.Context) *Message {
	IfDonePanic(ctx)
	return <-env.server.Msgs
}

func (env *Env) ServerPutMessage(ctx context.Context, msg *Message) {
	IfDonePanic(ctx)
	go func() { env.server.Msgs <- msg }()
}

// Replica ops

func (env *Env) ReplicaDeliver(ctx context.Context, msg *Message, stage *Stage) {
	IfDonePanic(ctx)
	env.replica.Deliver(msg, stage)
}

func (env *Env) ReplicaDelivered(ctx context.Context, cb *ControlBlock) Messages {
	IfDonePanic(ctx)
	return env.replica.Delivered(cb)
}

func (env *Env) ReplicaAtomic(ctx context.Context) Messages {
	IfDonePanic(ctx)
	return env.replica.DeliveredAtomic()
}

func (env *Env) ReplicaWait(ctx context.Context, cb *ControlBlock) {
	IfDonePanic(ctx)
	env.replica.Wait(cb)
}
