package types

import (
	// "fmt"
	. "github.com/zballs/goRITAS/util"
	"sync"
)

type Replica struct {
	ID uint32

	AB1       Messages
	AB2       Messages
	AB_Update chan struct{}
	AB_Mtx    sync.Mutex

	VC        Messages
	VC_Update chan struct{}
	VC_Mtx    sync.Mutex

	MV1       Messages
	MV2       Messages
	MV_Update chan struct{}
	MV_Mtx    sync.Mutex

	BC1       Messages
	BC2       Messages
	BC3       Messages
	BC_Update chan struct{}
	BC_Mtx    sync.Mutex

	Atomic     Messages
	Atomic_Mtx sync.Mutex
}

func NewReplica(ID uint32) *Replica {

	return &Replica{
		ID:        ID,
		AB_Update: make(chan struct{}),
		VC_Update: make(chan struct{}),
		MV_Update: make(chan struct{}),
		BC_Update: make(chan struct{}),
	}
}

// DELIVER

func (r *Replica) DeliverBC1(msg *Message) {

	for _, m := range r.BC1 {
		if BytesEqual(msg.Digest(), m.Digest()) {
			return
		}
	}

	r.BC_Mtx.Lock()
	r.BC1 = append(r.BC1, msg)
	r.BC_Mtx.Unlock()

	r.BC_Update <- struct{}{}
}

func (r *Replica) DeliverBC2(msg *Message) {

	for _, m := range r.BC2 {
		if BytesEqual(msg.Digest(), m.Digest()) {
			return
		}
	}

	r.BC_Mtx.Lock()
	r.BC2 = append(r.BC2, msg)
	r.BC_Mtx.Unlock()

	r.BC_Update <- struct{}{}
}

func (r *Replica) DeliverBC3(msg *Message) {

	for _, m := range r.BC3 {
		if BytesEqual(msg.Digest(), m.Digest()) {
			return
		}
	}

	r.BC_Mtx.Lock()
	r.BC3 = append(r.BC3, msg)
	r.BC_Mtx.Unlock()

	r.BC_Update <- struct{}{}
}

func (r *Replica) DeliverMV1(msg *Message) {

	for _, m := range r.MV1 {
		if BytesEqual(msg.Digest(), m.Digest()) {
			return
		}
	}

	r.MV_Mtx.Lock()
	r.MV1 = append(r.MV1, msg)
	r.MV_Mtx.Unlock()

	r.MV_Update <- struct{}{}
}

func (r *Replica) DeliverMV2(msg *Message) {

	for _, m := range r.MV2 {
		if BytesEqual(msg.Digest(), m.Digest()) {
			return
		}
	}

	r.MV_Mtx.Lock()
	r.MV2 = append(r.MV2, msg)
	r.MV_Mtx.Unlock()

	r.MV_Update <- struct{}{}
}

func (r *Replica) DeliverVC(msg *Message) {

	for _, m := range r.VC {
		if BytesEqual(msg.Digest(), m.Digest()) {
			return
		}
	}

	r.VC_Mtx.Lock()
	r.VC = append(r.VC, msg)
	r.VC_Mtx.Unlock()

	r.VC_Update <- struct{}{}
}

func (r *Replica) DeliverAB1(msg *Message) {

	for _, m := range r.AB1 {
		if BytesEqual(msg.Digest(), m.Digest()) {
			return
		}
	}

	r.AB_Mtx.Lock()
	r.AB1 = append(r.AB1, msg)
	r.AB_Mtx.Unlock()

	r.AB_Update <- struct{}{}
}

func (r *Replica) DeliverAB2(msg *Message) {

	for _, m := range r.AB2 {
		if BytesEqual(msg.Digest(), m.Digest()) {
			return
		}
	}

	r.AB_Mtx.Lock()
	r.AB2 = append(r.AB2, msg)
	r.AB_Mtx.Unlock()

	r.AB_Update <- struct{}{}
}

func (r *Replica) DeliverAtomic(msg *Message) {

	for _, m := range r.Atomic {
		if BytesEqual(msg.Digest(), m.Digest()) {
			return
		}
	}

	r.Atomic_Mtx.Lock()
	defer r.Atomic_Mtx.Unlock()
	r.Atomic = append(r.Atomic, msg)
}

func (r *Replica) Deliver(msg *Message, stage *Stage) {

	switch msg.Value.(type) {

	case *Message_Protocol:

		switch stage.Value.(type) {

		case *Stage_AB:

			switch stage.GetStep() {

			case 1:
				r.DeliverAB1(msg)

			case 2:
				r.DeliverAB2(msg)

			default:
				//Log
			}

		case *Stage_VC:

			r.DeliverVC(msg)

		case *Stage_MV:

			switch stage.GetStep() {

			case 1:
				r.DeliverMV1(msg)

			case 2:
				r.DeliverMV2(msg)

			default:
				//Log
			}

		case *Stage_BC:

			switch stage.GetStep() {

			case 1:
				r.DeliverBC1(msg)

			case 2:
				r.DeliverBC2(msg)

			case 3:
				r.DeliverBC3(msg)

			default:
				//Log
			}

		default:
			//Log
		}

	case *Message_Atomic:
		r.DeliverAtomic(msg)

	default:
		//Log
	}
}

// REMOVE

/*
func (r *Replica) RemoveAB(args *Args) {

	r.AB_Mtx.Lock()
	defer r.AB_Mtx.Unlock()

	for idx, m := range r.AB {

		i := m.GetInstance()
		s := m.GetSender()

		if args.Stages.IsEqual(i) && args.Sender.IsEqual(s) {
			r.AB = append(r.AB[:idx], r.AB[idx+1:]...)
			return
		}
	}
}

func (r *Replica) Remove(args *Args) {

	switch args.Stage {

	case *Stage_AB:

		r.RemoveAB(args)

	default:
		//Log
	}
}
*/

// DELIVERED

func (r *Replica) DeliveredBC1() Messages {
	r.BC_Mtx.Lock()
	defer r.BC_Mtx.Unlock()
	return r.BC1
}

func (r *Replica) DeliveredBC2() Messages {

	r.BC_Mtx.Lock()
	defer r.BC_Mtx.Unlock()
	return r.BC2
}

func (r *Replica) DeliveredBC3() Messages {
	r.BC_Mtx.Lock()
	defer r.BC_Mtx.Unlock()
	return r.BC3
}

func (r *Replica) DeliveredMV1() Messages {
	r.MV_Mtx.Lock()
	defer r.MV_Mtx.Unlock()
	return r.MV1
}

func (r *Replica) DeliveredMV2() Messages {
	r.MV_Mtx.Lock()
	defer r.MV_Mtx.Unlock()
	return r.MV2
}

func (r *Replica) DeliveredVC() Messages {
	r.VC_Mtx.Lock()
	defer r.VC_Mtx.Unlock()
	return r.VC
}

func (r *Replica) DeliveredAB1() Messages {
	r.AB_Mtx.Lock()
	defer r.AB_Mtx.Unlock()
	return r.AB1
}

func (r *Replica) DeliveredAB2() Messages {
	r.AB_Mtx.Lock()
	defer r.AB_Mtx.Unlock()
	return r.AB2
}

func (r *Replica) Delivered(cb *ControlBlock) Messages {

	switch cb.Value.(type) {

	case *Stage_AB:

		switch cb.GetStep() {

		case 1:
			return r.DeliveredAB1()

		case 2:
			return r.DeliveredAB2()

		default:
		}

	case *Stage_VC:

		return r.DeliveredVC()

	case *Stage_MV:

		switch cb.GetStep() {

		case 1:
			return r.DeliveredMV1()

		case 2:
			return r.DeliveredMV2()

		default:
		}

	case *Stage_BC:

		switch cb.GetStep() {

		case 1:
			return r.DeliveredBC1()

		case 2:
			return r.DeliveredBC2()

		case 3:
			return r.DeliveredBC3()

		default:
		}

	default:
	}

	return nil
}

func (r *Replica) DeliveredAtomic() Messages {
	r.Atomic_Mtx.Lock()
	defer r.Atomic_Mtx.Unlock()
	return r.Atomic
}

// UPDATES

func (r *Replica) Wait(cb *ControlBlock) {

	switch cb.Value.(type) {

	case *Stage_AB:

		r.WaitForUpdateAB()

	case *Stage_VC:

		r.WaitForUpdateVC()

	case *Stage_MV:

		r.WaitForUpdateMV()

	case *Stage_BC:

		r.WaitForUpdateBC()

	default:
		//
	}

}

func (r *Replica) WaitForUpdateAB() {
	<-r.AB_Update
}

func (r *Replica) WaitForUpdateVC() {
	<-r.VC_Update
}

func (r *Replica) WaitForUpdateMV() {
	<-r.MV_Update
}

func (r *Replica) WaitForUpdateBC() {
	<-r.BC_Update
}
