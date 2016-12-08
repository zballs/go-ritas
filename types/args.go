package types

import (
	"encoding/json"
	"log"
)

type Args struct {
	stage       *Stage
	payload     *Payload
	vector      *Vector
	stages      *Stages
	sender      *Sender
	broadcaster *Broadcaster
}

func NewArgs() *Args {

	return &Args{
		stages: NewStages(),
	}
}

// Get

func (args *Args) GetStage() *Stage {

	return args.stage
}

func (args *Args) GetPayload() *Payload {

	return args.payload
}

func (args *Args) GetStages() *Stages {

	return args.stages
}

func (args *Args) GetSender() *Sender {

	return args.sender
}

func (args *Args) GetBroadcaster() *Broadcaster {

	return args.broadcaster
}

// Set

func (args *Args) SetArg(arg interface{}) bool {

	switch arg.(type) {

	case *Stage:
		args.stage = arg.(*Stage)
		return args.stages.AddStage(args.stage)

	case *Stages:
		args.stages = arg.(*Stages)

	case *Payload:
		args.payload = arg.(*Payload)

	case *Vector:
		args.vector = arg.(*Vector)

	case *Sender:
		args.sender = arg.(*Sender)

	case *Broadcaster:
		args.broadcaster = arg.(*Broadcaster)

	default:
		log.Println("Could not set arg!")
		return false
	}

	return true
}

func (args *Args) SetArgs(argz ...interface{}) {

	for _, arg := range argz {
		args.SetArg(arg)
	}
}

func (args *Args) MarshalJSON() []byte {

	stage := *args.stage
	payload := *args.payload
	sender := *args.sender

	v := struct {
		Stage   `json:"stage"`
		Payload `json:"payload"`
		Sender  `json:"sender"`
	}{stage, payload, sender}

	data, _ := json.Marshal(v)

	return data
}
