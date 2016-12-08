package util

import (
	"encoding/hex"
	"fmt"
)

func BytesEqual(b1 []byte, b2 []byte) bool {
	if len(b1) != len(b2) {
		return false
	}
	for idx, b := range b1 {
		if b != b2[idx] {
			return false
		}
	}
	return true
}

func BytesToHexstr(bytes []byte) string {
	return fmt.Sprintf("%X", bytes)
}

func HexstrToBytes(hexstr string) []byte {
	bytes, _ := hex.DecodeString(hexstr)
	return bytes
}
