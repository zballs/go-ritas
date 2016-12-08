package types

import (
	"bytes"
	"context"
	"encoding/binary"
	// "fmt"
	"github.com/golang/protobuf/proto"
	"github.com/tendermint/go-wire"
	. "github.com/zballs/goRITAS/util"
	"io"
	"log"
	"net"
	"reflect"
)

// Messages

func ToMessageProtocol(args *Args) *Message {
	return &Message{
		&Message_Protocol{
			&MessageProtocol{
				Stage:       args.stage,
				Payload:     args.payload,
				Vector:      args.vector,
				Stages:      args.stages,
				Sender:      args.sender,
				Broadcaster: args.broadcaster,
			},
		},
	}
}

func ToMessageAtomic(args *Args) *Message {
	return &Message{
		&Message_Atomic{
			&MessageAtomic{
				Payload: args.payload,
				Stages:  args.stages,
				Sender:  args.sender,
			},
		},
	}
}

func (m *Message) IsMessageProcol() bool {
	return m.GetProtocol() != nil
}

func (m *Message) IsMessageAtomic() bool {
	return m.GetProtocol() != nil
}

func (m *Message) Digest() []byte {

	args := m.ToArgs()

	data := args.MarshalJSON()

	digest, _ := HashData(data)

	return digest
}

func (m *Message) DigestHexstr() string {
	return BytesToHexstr(m.Digest())
}

func (m *Message) ToArgs() *Args {

	switch m.Value.(type) {

	case *Message_Protocol:

		return &Args{
			stage:       m.GetStage(),
			payload:     m.GetPayload(),
			vector:      m.GetVector(),
			stages:      m.GetStages(),
			sender:      m.GetSender(),
			broadcaster: m.GetBroadcaster(),
		}

	case *Message_Atomic:

		return &Args{
			payload: m.GetPayload(),
			stages:  m.GetStages(),
			sender:  m.GetSender(),
		}

	default:
		return nil
	}
}

func (m *Message) GetStage() *Stage {

	switch m.Value.(type) {

	case *Message_Protocol:
		return m.GetProtocol().GetStage()

	default:
		return nil
	}
}

func (m *Message) GetPayload() *Payload {

	switch m.Value.(type) {

	case *Message_Protocol:
		return m.GetProtocol().GetPayload()

	case *Message_Atomic:
		return m.GetAtomic().GetPayload()

	default:
		return nil
	}
}

func (m *Message) GetStages() *Stages {

	switch m.Value.(type) {

	case *Message_Protocol:
		return m.GetProtocol().GetStages()

	case *Message_Atomic:
		return m.GetAtomic().GetStages()

	default:
		return nil
	}
}

func (m *Message) GetVector() *Vector {

	switch m.Value.(type) {

	case *Message_Protocol:
		return m.GetProtocol().GetVector()

	default:
		return nil
	}
}

func (m *Message) GetSender() *Sender {

	switch m.Value.(type) {

	case *Message_Protocol:
		return m.GetProtocol().GetSender()

	case *Message_Atomic:
		return m.GetAtomic().GetSender()

	default:
		return nil
	}
}

func (m *Message) GetBroadcaster() *Broadcaster {

	switch m.Value.(type) {

	case *Message_Protocol:
		return m.GetProtocol().GetBroadcaster()

	default:
		return nil
	}
}

// Sort messages by sum of payload, sender values
// Deterministic ordering of atomic deliveries

type Messages []*Message

func (slice Messages) Len() int {
	return len(slice)
}

func (slice Messages) Less(i, j int) bool {

	eli, elj := slice[i], slice[j]

	pi, pj := eli.GetPayload(), elj.GetPayload()
	si, sj := eli.GetSender(), elj.GetSender()

	return pi.Value()+si.Value() < pj.Value()+sj.Value()
}

func (slice Messages) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

// Bloom

func (b *Bloom) getIdx(pos uint64) uint64 {
	return uint64(len(b.Data)) - 1 - (pos / 8)
}

func (b *Bloom) length() uint64 {
	return uint64(len(b.Data))
}

func (b *Bloom) size() uint64 {
	return b.length() * 8
}

func (b *Bloom) setBit(pos uint64) {
	idx := b.getIdx(pos)
	bite := b.Data[idx]
	bite |= (1 << (pos % 8))
	b.Data[idx] = bite
}

func (b *Bloom) hasBit(pos uint64) bool {
	idx := b.getIdx(pos)
	bite := b.Data[idx]
	val := bite & (1 << (pos % 8))
	return val > 0
}

func (b *Bloom) HasData(data []byte, n uint64) bool {

	hash := MurmurHash(data)
	size := b.size()
	var i uint64
	for i = 0; i < n; i++ {
		hash += FNVHash(data)
		pos := hash % size
		if !b.hasBit(pos) {
			return false
		}
	}
	return true
}

func (b *Bloom) setBits(data []byte, n uint64) {
	hash := MurmurHash(data)
	size := b.size()
	var i uint64
	for i = 0; i < n; i++ {
		hash += FNVHash(data)
		pos := hash % size
		if !b.hasBit(pos) {
			b.setBit(pos)
		}
	}
}

func (b *Bloom) AddData(data []byte, n uint64) {

	if b.HasData(data, n) {
		return
	}

	b.setBits(data, n)
}

func ToBloom(payloads Payloads) *Bloom {

	data := make([]byte, 6*len(payloads)/5)

	b := &Bloom{data}

	for _, p := range payloads {
		b.AddData(p.Data, 3) //3 iters for now
	}

	return b
}

func (b *Bloom) PayloadWrap() *Payload {
	return ToPayload(b.Data)
}

// Vector

func ToVector(payloads Payloads) *Vector {
	return &Vector{payloads}
}

func VectorFromBytes(bites []byte) *Vector {

	// if err, block is nil

	buf, n, err := new(bytes.Buffer), int(0), error(nil)
	buf.Write(bites)

	datas := wire.ReadByteSlices(buf, 0, &n, &err)

	if err != nil {
		return nil
	}

	payloads := make(Payloads, len(datas))

	for idx, data := range datas {
		payloads[idx] = ToPayload(data)
	}

	return ToVector(payloads)
}

func VectorFromPayload(p *Payload) *Vector {

	// if err, block is nil

	bites := p.Data
	return VectorFromBytes(bites)
}

func (v *Vector) PayloadWrap() *Payload {
	data := v.Bytes()
	return ToPayload(data)
}

func (v *Vector) Bytes() []byte {

	// if err, bytes are nil

	payloads := v.GetPayloads()
	data := make([][]byte, len(payloads))

	for idx, p := range payloads {
		data[idx] = p.Data
	}

	buf, n, err := new(bytes.Buffer), int(0), error(nil)
	wire.WriteByteSlices(data, buf, &n, &err)

	if err != nil {
		return nil
	}

	return buf.Bytes()
}

func (v *Vector) HasPayload(p1 *Payload) bool {

	payloads := v.GetPayloads()

	for _, p2 := range payloads {
		if p1.IsEqual(p2) {
			return true
		}
	}
	return false
}

// HashVector

func ToHashVector(hashloads []*Hashload) *HashVector {
	return &HashVector{hashloads}
}

func HashVectorFromPayloads(payloads Payloads) *HashVector {

	hashloads := make([]*Hashload, len(payloads))

	for idx, p := range payloads {
		hashloads[idx] = HashloadFromPayload(p)
	}

	return ToHashVector(hashloads)
}

func HashVectorFromBytes(bites []byte) *HashVector {

	// if err, block is nil

	buf, n, err := new(bytes.Buffer), int(0), error(nil)
	buf.Write(bites)

	hashes := wire.ReadByteSlices(buf, 0, &n, &err)

	if err != nil {
		return nil
	}

	hashloads := make([]*Hashload, len(hashes))

	for idx, hash := range hashes {
		hashloads[idx] = ToHashload(hash)
	}

	return ToHashVector(hashloads)
}

func HashVectorFromPayload(p *Payload) *HashVector {

	// if err, block is nil
	bites := p.Data
	return HashVectorFromBytes(bites)
}

func (hv *HashVector) PayloadWrap() *Payload {
	data := hv.Bytes()
	return ToPayload(data)
}

func (hv *HashVector) Bytes() []byte {

	// if err, bytes are nil

	hashloads := hv.GetHashloads()
	hashes := make([][]byte, len(hashloads))

	for idx, h := range hashloads {
		hashes[idx] = h.Hash
	}

	buf, n, err := new(bytes.Buffer), int(0), error(nil)
	wire.WriteByteSlices(hashes, buf, &n, &err)

	if err != nil {
		return nil
	}

	return buf.Bytes()
}

func (hv *HashVector) HasHashload(h1 *Hashload) bool {

	hashloads := hv.GetHashloads()

	for _, h2 := range hashloads {

		if h1.IsEqual(h2) {
			return true
		}
	}
	return false
}

func (hv *HashVector) HasPayload(p *Payload) bool {
	h := HashloadFromPayload(p)
	return hv.HasHashload(h)
}

// Hashload

func ToHashload(hash []byte) *Hashload {
	return &Hashload{hash}
}

func HashloadFromBytes(bites []byte) *Hashload {
	hash, _ := HashData(bites)
	return ToHashload(hash)
}

func HashloadFromPayload(p *Payload) *Hashload {
	return HashloadFromBytes(p.Data)
}

func (h *Hashload) ToHexstr() string {
	return BytesToHexstr(h.Hash)
}

func HashloadFromHexstr(hexstr string) *Hashload {
	hash := HexstrToBytes(hexstr)
	return ToHashload(hash)
}

func (h1 *Hashload) IsEqual(h2 *Hashload) bool {
	return BytesEqual(h1.Hash, h2.Hash)
}

func (h *Hashload) Value() uint64 {
	ui64, _ := binary.Uvarint(h.Hash)
	return ui64
}

// Sort Hashloads

type Hashloads []*Hashload

func (slice Hashloads) Len() int {
	return len(slice)
}

func (slice Hashloads) Less(i, j int) bool {
	return slice[i].Value() < slice[j].Value()
}

func (slice Hashloads) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

// Payload

func ToPayload(data []byte) *Payload {
	return &Payload{data}
}

func EmptyPayload() *Payload {
	return &Payload{nil}
}

func (p *Payload) ToUint32() uint32 {
	return binary.BigEndian.Uint32(p.Data)
}

func (p *Payload) ToUint32s() []uint32 {
	var ns []uint32
	var n uint32
	var buf bytes.Buffer
	var err error
	buf.Write(p.Data)
	for {
		err = binary.Read(&buf, binary.BigEndian, &n)
		if err != nil {
			break
		}
		ns = append(ns, n)
	}
	return ns
}

func PayloadFromUint32(n uint32) *Payload {
	data := make([]byte, 4)
	binary.BigEndian.PutUint32(data, n)
	return ToPayload(data)
}

func PayloadFromUint32s(ns ...uint32) *Payload {
	var buf bytes.Buffer
	for _, n := range ns {
		binary.Write(&buf, binary.BigEndian, n)
	}
	return ToPayload(buf.Bytes())
}

func (p *Payload) ToHexstr() string {
	return BytesToHexstr(p.Data)
}

func PayloadFromHexstr(hexstr string) *Payload {
	data := HexstrToBytes(hexstr)
	return ToPayload(data)
}

func (p1 *Payload) IsEqual(p2 *Payload) bool {
	return BytesEqual(p1.Data, p2.Data)
}

func (p *Payload) ToHashload() *Hashload {
	hash, _ := HashData(p.Data)
	return ToHashload(hash)
}

func (p *Payload) Value() uint64 {
	h := p.ToHashload()
	return h.Value()
}

// Sort Payloads

type Payloads []*Payload

func (slice Payloads) Len() int {
	return len(slice)
}

func (slice Payloads) Less(i, j int) bool {
	return slice[i].Value() < slice[j].Value()
}

func (slice Payloads) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

// Stages

func NewStages() *Stages {
	return &Stages{}
}

func (s *Stages) AddStage(stage *Stage) bool {

	switch stage.Value.(type) {

	case *Stage_AB:

		s.AB = stage

	case *Stage_VC:

		s.VC = stage

	case *Stage_MV:

		s.MV = stage

	case *Stage_BC:

		s.BC = stage

	case *Stage_RB:

		if s.EB != nil {
			return false
		}

		s.RB = stage

	case *Stage_EB:

		if s.RB != nil {
			return false
		}

		s.EB = stage

	default:
		return false
	}

	return true
}

func (s1 *Stages) SameStages(s2 *Stages) bool {

	// TODO: add other conditions (if necessary)

	if !s1.AB.SameID(s2.AB) {
		return false
	}

	if !s1.VC.SameID(s2.VC) {
		return false
	}

	if !s1.MV.SameID(s2.MV) {
		return false
	}

	if !s1.BC.SameID(s2.BC) {
		return false
	}

	if !s1.RB.SameID(s2.RB) {
		return false
	}

	if !s1.EB.SameID(s2.EB) {
		return false
	}

	return true
}

// Stage

func ToStageAB(ID, round, step uint32) *Stage {
	return &Stage{
		Value: &Stage_AB{
			&StageAB{ID, round, step},
		},
	}
}

func ToStageVC(ID, round uint32) *Stage {
	return &Stage{
		Value: &Stage_VC{
			&StageVC{ID, round},
		},
	}
}

func ToStageMV(ID, step uint32) *Stage {
	return &Stage{
		Value: &Stage_MV{
			&StageMV{ID, step},
		},
	}
}

func ToStageBC(ID, step uint32) *Stage {
	return &Stage{
		Value: &Stage_BC{
			&StageBC{ID, step},
		},
	}
}

func ToStageRB(ID, step uint32) *Stage {
	return &Stage{
		Value: &Stage_RB{
			&StageRB{ID, step},
		},
	}
}

func ToStageEB(ID, step uint32) *Stage {
	return &Stage{
		Value: &Stage_EB{
			&StageEB{ID, step},
		},
	}
}

func NewStageAB(e *Env) *Stage {
	ID := e.ProtoID()
	return ToStageAB(ID, 0, 1)
}

func NewStageVC(e *Env) *Stage {
	ID := e.ProtoID()
	return ToStageVC(ID, 0)
}

func NewStageMV(e *Env) *Stage {
	ID := e.ProtoID()
	return ToStageMV(ID, 1)
}

func NewStageBC(e *Env) *Stage {
	ID := e.ProtoID()
	return ToStageBC(ID, 1)
}

func NewStageRB(e *Env) *Stage {
	ID := e.ProtoID()
	return ToStageRB(ID, 1)
}

func NewStageEB(e *Env) *Stage {
	ID := e.ProtoID()
	return ToStageEB(ID, 1)
}

func (s *Stage) GetID() uint32 {

	switch s.Value.(type) {

	case *Stage_AB:
		return s.GetAB().ID

	case *Stage_VC:
		return s.GetVC().ID

	case *Stage_MV:
		return s.GetMV().ID

	case *Stage_BC:
		return s.GetBC().ID

	case *Stage_RB:
		return s.GetRB().ID

	case *Stage_EB:
		return s.GetEB().ID

	default:
		return 0
	}
}

func (s *Stage) GetRound() uint32 {

	switch s.Value.(type) {

	case *Stage_AB:
		return s.GetAB().Round

	case *Stage_VC:
		return s.GetVC().Round

	default:
		return 0
	}
}

func (s *Stage) IncrementRound() bool {

	switch s.Value.(type) {

	case *Stage_AB:
		s.GetAB().Round++

	case *Stage_VC:
		s.GetVC().Round++

	default:
		return false
	}

	return true
}

func (s *Stage) GetStep() uint32 {

	switch s.Value.(type) {

	case *Stage_AB:
		return s.GetAB().Step

	case *Stage_MV:
		return s.GetMV().Step

	case *Stage_BC:
		return s.GetBC().Step

	case *Stage_RB:
		return s.GetRB().Step

	case *Stage_EB:
		return s.GetEB().Step

	default:
		return 0
	}
}

func (s *Stage) IncrementStep() bool {

	switch s.Value.(type) {

	case *Stage_AB:
		s.GetAB().Step++

	case *Stage_MV:
		s.GetMV().Step++

	case *Stage_BC:
		s.GetBC().Step++

	case *Stage_RB:
		s.GetRB().Step++

	case *Stage_EB:
		s.GetEB().Step++

	default:
		return false
	}
	return true
}

func (s *Stage) ResetStep() bool {

	switch s.Value.(type) {

	case *Stage_AB:
		s.GetAB().Step = 1

	case *Stage_MV:
		s.GetMV().Step = 1

	case *Stage_BC:
		s.GetBC().Step = 1

	case *Stage_RB:
		s.GetRB().Step = 1

	case *Stage_EB:
		s.GetEB().Step = 1

	default:
		return false
	}
	return true
}

func (s1 *Stage) SameStage(s2 *Stage) bool {

	if s1 == nil && s2 == nil {

		return true

	} else if s1 != nil && s2 != nil {

		v1 := s1.GetValue()
		v2 := s2.GetValue()

		return reflect.TypeOf(v1) == reflect.TypeOf(v2)

	}

	return false
}

func (s1 *Stage) SameID(s2 *Stage) bool {

	if s1 == nil && s2 == nil {

		return true

	} else if s1 != nil && s2 != nil {

		return s1.GetID() == s2.GetID()
	}

	return false
}

func (s1 *Stage) SameStep(s2 *Stage) bool {

	if s1 == nil && s2 == nil {

		return true

	} else if s1 != nil && s2 != nil {

		return s1.GetStep() == s2.GetStep()
	}

	return false
}

func (s1 *Stage) SameRound(s2 *Stage) bool {

	if s1 == nil && s2 == nil {

		return true

	} else if s1 != nil && s2 != nil {

		return s1.GetRound() == s2.GetRound()
	}

	return false
}

// Sender

func ToSender(ida *IDAddr) *Sender {
	return &Sender{ida.ID, ida.Addr}
}

func (s1 *Sender) IsEqual(s2 *Sender) bool {
	return s1.ID == s2.ID && s1.Addr == s2.Addr
}

func (s *Sender) Value() uint64 {
	return uint64(s.ID)
}

// Broadcaster

func ToBroadcaster(ida *IDAddr) *Broadcaster {
	return &Broadcaster{ida.ID, ida.Addr}
}

func (b1 *Broadcaster) IsEqual(b2 *Broadcaster) bool {
	return b1.ID == b2.ID && b1.Addr == b2.Addr
}

// Read, Write Messages

func ReadMessage(r io.Reader) (*Message, error) {
	m := &Message{}
	n, err := int(0), error(nil)
	buf := wire.ReadByteSlice(r, 0, &n, &err)
	if err != nil {
		return nil, err
	}
	err = proto.Unmarshal(buf, m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func WriteMessage(w io.Writer, m *Message) error {
	buf, err := proto.Marshal(m)
	if err != nil {
		return err
	}
	var n int
	wire.WriteByteSlice(buf, w, &n, &err)
	return err
}

// Send Messages

func SendMessage(ctx context.Context, addr string, m *Message) error {

	IfDonePanic(ctx)

	conn, err := net.Dial("tcp", addr)

	if err != nil {
		return err
	}

	err = WriteMessage(conn, m)
	return err
}

func MulticastMessage(ctx context.Context, e *Env, m *Message) {

	for _, addr := range e.Addrs() {

		err := SendMessage(ctx, addr, m)

		if err != nil {

			log.Printf("Couldn't send msg to %s\n", addr)
		}
	}
}
