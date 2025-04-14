package Utils

import (
	"bytes"

	"github.com/ethereum/go-ethereum/common/bitutil"
)

func XOR(m []byte, key []byte) (c []byte) {
	c = make([]byte, len(m))
	if counts := (len(m)-1)/len(key) + 1; counts > 1 {
		key = bytes.Repeat(key, counts)
	}
	bitutil.XORBytes(c, m, key)
	return
}
