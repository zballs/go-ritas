package util

import (
	"crypto/sha256"
)

// cryptographic

func HashData(data []byte) ([]byte, error) {

	h := sha256.New()

	_, err := h.Write(data)
	if err != nil {
		return nil, err
	}

	hash := h.Sum(nil)
	return hash, nil
}

// non-cryptographic // for bloom filter

const (

	// Murmur
	c1     uint64 = 0xcc9e2d51
	c2     uint64 = 0x1b873593
	n      uint64 = 0xe6546b64
	round4 uint64 = 0xfffffffc
	seed   uint64 = 0x0 //

	// FNV
	FNV_offset_basis uint64 = 0xcbf29ce484222325
	FNV_prime        uint64 = 0x100000001b3
)

func MurmurHash(bytes []byte) uint64 {
	h := seed
	length := uint64(len(bytes))
	roundedEnd := length & round4
	var i uint64
	var k uint64
	for i = 0; i < roundedEnd; i += 4 {
		b0, b1, b2, b3 := bytes[i], bytes[i+1], bytes[i+2], bytes[i+3]
		k := uint64(b0 | (b1 << 8) | (b2 << 16) | (b3 << 24))
		k *= c1
		k = (k << 15) | (k >> 17)
		k *= c2
		h ^= k
		h = (h << 13) | (h >> 19)
		h = h*5 + n
	}
	k = 0
	val := length & 0x03
	if val == 3 {
		k = uint64(bytes[roundedEnd+2] << 16)
	}
	if val >= 2 {
		k |= uint64(bytes[roundedEnd+1] << 8)
	}
	if val >= 1 {
		k |= uint64(bytes[roundedEnd])
		k *= c1
		k = (k << 15) | (k >> 17)
		k *= c2
		h ^= k
	}
	h ^= length
	h ^= h >> 16
	h *= 0x85ebca6b
	h ^= h >> 13
	h *= 0xc2b2ae35
	h ^= h >> 16
	return h
}

func FNVHash(bytes []byte) uint64 {
	hash := FNV_offset_basis
	for _, b := range bytes {
		hash *= FNV_prime
		hash |= uint64(b)
	}
	return hash
}
