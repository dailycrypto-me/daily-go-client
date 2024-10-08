package state_db

import (
	"github.com/dailycrypto-me/daily-evm/common"
	"github.com/dailycrypto-me/daily-evm/daily/trie"
	"github.com/dailycrypto-me/daily-evm/daily/util/keccak256"
)

type ExtendedReader struct{ Reader }

func (self ExtendedReader) GetCode(hash *common.Hash) (ret []byte) {
	self.Get(COL_code, hash, func(bytes []byte) {
		ret = common.CopyBytes(bytes)
	})
	return
}

func (self ExtendedReader) GetAccountStorage(addr *common.Address, key *common.Hash, cb func([]byte)) {
	AccountTrieInputAdapter{addr, self}.GetValue(keccak256.Hash(key[:]), cb)
}

func (self ExtendedReader) GetAccount(addr *common.Address, cb func(Account)) {
	self.GetRawAccount(addr, func(v []byte) {
		cb(DecodeAccountFromTrie(v))
	})
}

func (self ExtendedReader) GetRawAccount(addr *common.Address, cb func([]byte)) {
	MainTrieInputAdapter{self}.GetValue(keccak256.Hash(addr[:]), cb)
}

func (self ExtendedReader) GetCodeByAddress(addr *common.Address) (ret []byte) {
	self.GetRawAccount(addr, func(acc []byte) {
		if code_hash := CodeHash(acc); code_hash != nil {
			ret = self.GetCode(code_hash)
		}
	})
	return
}

func (self ExtendedReader) ForEachStorage(addr *common.Address, f func(*common.Hash, []byte)) {
	self.GetRawAccount(addr, func(acc []byte) {
		storage_root := StorageRoot(acc)
		if storage_root == nil {
			return
		}
		trie.Reader{AccountTrieSchema{}}.ForEach(
			AccountTrieInputAdapter{addr, self},
			storage_root,
			true,
			func(hash *common.Hash, val trie.Value) {
				enc_storage, _ := val.EncodeForTrie()
				f(hash, enc_storage)
			})
	})
}

func (self ExtendedReader) ForEachAccountNodeHashByRoot(storage_root *common.Hash, f func(*common.Hash, []byte)) {
	if storage_root == nil {
		return
	}
	no_addr := common.Address{}
	trie.Reader{AccountTrieSchema{}}.ForEachNodeHash(
		AccountTrieInputAdapter{&no_addr, self},
		storage_root,
		f)
}

func (self ExtendedReader) ForEachMainNodeHashByRoot(storage_root *common.Hash, f func(*common.Hash, []byte)) {
	if storage_root == nil {
		return
	}
	trie.Reader{MainTrieSchema{}}.ForEachNodeHash(
		MainTrieInputAdapter{self},
		storage_root,
		f)
}
