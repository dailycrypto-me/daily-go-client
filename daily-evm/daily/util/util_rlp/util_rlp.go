package util_rlp

import (
	"github.com/dailycrypto-me/daily-evm/rlp"
)

func RLPListAt(list []byte, pos uint) (ret []byte) {
	_, rest, _ := rlp.MustSplit(list)
	for i := uint(0); i <= pos; i++ {
		_, ret, rest = rlp.MustSplit(rest)
	}
	return
}
