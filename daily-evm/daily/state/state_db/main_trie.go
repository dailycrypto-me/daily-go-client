package state_db

import (
	"math/big"
	"sync"

	"github.com/dailycrypto-me/daily-evm/daily/util/bigutil"
	"github.com/dailycrypto-me/daily-evm/daily/util/bin"

	"github.com/dailycrypto-me/daily-evm/daily/state/state_common"

	"github.com/dailycrypto-me/daily-evm/daily/util/keccak256"
	"github.com/dailycrypto-me/daily-evm/daily/util/util_rlp"

	"github.com/dailycrypto-me/daily-evm/daily/util"

	"github.com/dailycrypto-me/daily-evm/common"
	"github.com/dailycrypto-me/daily-evm/crypto"
	"github.com/dailycrypto-me/daily-evm/rlp"
)

type MainTrieSchema struct{}

func (MainTrieSchema) ValueStorageToHashEncoding(enc_storage []byte) (enc_hash []byte) {
	encoder := take_acc_encoder()
	rlp_list := encoder.ListStart()
	next, curr, err := rlp.SplitList(enc_storage)
	util.PanicIfNotNil(err)
	curr, next, err = rlp.SplitString(next)
	util.PanicIfNotNil(err)
	encoder.AppendString(curr)
	curr, next, err = rlp.SplitString(next)
	util.PanicIfNotNil(err)
	encoder.AppendString(curr)
	curr, next, err = rlp.SplitString(next)
	util.PanicIfNotNil(err)
	if len(curr) != 0 {
		encoder.AppendString(curr)
	} else {
		encoder.AppendString(state_common.EmptyRLPListHash[:])
	}
	curr, next, err = rlp.SplitString(next)
	util.PanicIfNotNil(err)
	if len(curr) != 0 {
		encoder.AppendString(curr)
	} else {
		encoder.AppendString(crypto.EmptyBytesKeccak256[:])
	}
	encoder.ListEnd(rlp_list)
	enc_hash = encoder.ToBytes(-1)
	return_acc_encoder(encoder)
	return
}

func (MainTrieSchema) MaxValueSizeToStoreInTrie() int { return 8 }

type Account struct {
	Nonce           *big.Int
	Balance         *big.Int
	StorageRootHash *common.Hash
	CodeHash        *common.Hash
	CodeSize        uint64
}

func DecodeAccountFromTrie(enc_storage []byte) (ret Account) {
	rest, tmp := rlp.MustSplitList(enc_storage)
	tmp, rest = rlp.MustSplitString(rest)
	ret.Nonce = bigutil.FromBytes(tmp)
	tmp, rest = rlp.MustSplitString(rest)
	ret.Balance = bigutil.FromBytes(tmp)
	if tmp, rest = rlp.MustSplitString(rest); len(tmp) != 0 {
		ret.StorageRootHash = new(common.Hash).SetBytes(tmp)
	}
	if tmp, rest = rlp.MustSplitString(rest); len(tmp) != 0 {
		ret.CodeHash = new(common.Hash).SetBytes(tmp)
	}
	tmp, rest = rlp.MustSplitString(rest)
	ret.CodeSize = bin.DEC_b_endian_compact_64(tmp)
	return
}

func (self *Account) EncodeForTrie() (enc_storage, enc_hash []byte) {
	encoder := take_acc_encoder()
	storage_rlp_list := encoder.ListStart()
	encoder.AppendBigInt(self.Nonce)
	encoder.AppendBigInt(self.Balance)
	if self.StorageRootHash != nil {
		encoder.AppendString(self.StorageRootHash[:])
	} else {
		encoder.AppendEmptyString()
	}
	if self.CodeHash != nil {
		encoder.AppendString(self.CodeHash[:])
	} else {
		encoder.AppendEmptyString()
	}
	encoder.AppendUint(self.CodeSize)
	encoder.ListEnd(storage_rlp_list)
	enc_storage = encoder.ToBytes(-1)
	encoder.Reset()
	hash_rlp_list := encoder.ListStart()
	encoder.AppendBigInt(self.Nonce)
	encoder.AppendBigInt(self.Balance)
	if self.StorageRootHash != nil {
		encoder.AppendString(self.StorageRootHash[:])
	} else {
		encoder.AppendString(state_common.EmptyRLPListHash[:])
	}
	if self.CodeHash != nil {
		encoder.AppendString(self.CodeHash[:])
	} else {
		encoder.AppendString(crypto.EmptyBytesKeccak256[:])
	}
	encoder.ListEnd(hash_rlp_list)
	enc_hash = encoder.ToBytes(-1)
	return_acc_encoder(encoder)
	return
}

func StorageRoot(acc_enc_storage []byte) *common.Hash {
	return keccak256.HashView(util_rlp.RLPListAt(acc_enc_storage, 2))
}

func CodeHash(acc_enc_storage []byte) *common.Hash {
	return keccak256.HashView(util_rlp.RLPListAt(acc_enc_storage, 3))
}

func take_acc_encoder() (ret *rlp.Encoder) {
	ret = acc_encoder_pool.Get().(*rlp.Encoder)
	ret.Reset()
	return ret
}

func return_acc_encoder(encoder *rlp.Encoder) {
	acc_encoder_pool.Put(encoder)
}

var acc_encoder_pool = sync.Pool{New: func() interface{} {
	ret := new(rlp.Encoder)
	ret.ResizeReset(1<<8, 1)
	return ret
}}

type MainTrieInputAdapter struct{ Reader }

func (self MainTrieInputAdapter) GetValue(key *common.Hash, cb func(v []byte)) {
	self.Get(COL_main_trie_value, key, cb)
}

func (self MainTrieInputAdapter) GetNode(node_hash *common.Hash, cb func([]byte)) {
	self.Get(COL_main_trie_node, node_hash, cb)
}

type MainTrieIOAdapter struct{ ReadWriter }

func (self MainTrieIOAdapter) GetValue(key *common.Hash, cb func(v []byte)) {
	self.Get(COL_main_trie_value, key, cb)
}

func (self MainTrieIOAdapter) GetNode(node_hash *common.Hash, cb func([]byte)) {
	self.Get(COL_main_trie_node, node_hash, cb)
}

func (self MainTrieIOAdapter) PutValue(key *common.Hash, v []byte) {
	self.Put(COL_main_trie_value, key, v)
}

func (self MainTrieIOAdapter) PutNode(node_hash *common.Hash, node []byte) {
	self.Put(COL_main_trie_node, node_hash, node)
}
