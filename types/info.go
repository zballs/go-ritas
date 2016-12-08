package types

/*
import (
	"crypto"
	"crypto/ecdsa"
	. "github.com/zballs/goRITAS/util"
)

type Info struct {
	Addr string
	ID   uint32

	PubKey  crypto.PublicKey
	privKey *ecdsa.PrivateKey
}

func NewInfo(addr string, ID uint32) *Info {
	return &Info{
		Addr:    addr,
		ID:      ID,
		privKey: GeneratePrivKey(),
	}
}

func (info *Info) ShareInfo() *Info {
	pubKey := info.privKey.Public()
	return &Info{
		Addr:   info.Addr,
		ID:     info.ID,
		PubKey: pubKey,
	}
}

func (info *Info) SignMessage(msg *Message) []byte {

	if info.privKey == nil {
		return nil
	}

	digest := msg.Digest()
	if digest == nil {
		return nil
	}

	signature, err := SignHash(digest, info.privKey)
	if err != nil {
		return nil
	}

	return signature
}

func (info *Info) VerifySignature(signature []byte, msg *Message) bool {

	if info.PubKey == nil {
		return false
	}

	digest := msg.Digest()
	if digest == nil {
		return false
	}

	return VerifySignature(signature, digest, info.PubKey)
}
*/
