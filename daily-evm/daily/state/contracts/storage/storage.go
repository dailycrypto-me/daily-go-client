package contract_storage

import (
	"fmt"
	"math/big"

	"github.com/dailycrypto-me/daily-evm/common"
	"github.com/dailycrypto-me/daily-evm/daily/util/asserts"
	"github.com/dailycrypto-me/daily-evm/daily/util/bin"
	"github.com/dailycrypto-me/daily-evm/daily/util/keccak256"
)

type StorageReader interface {
	GetAccountStorage(addr *common.Address, key *common.Hash, cb func([]byte))
}
type StorageWriter interface {
	SubBalance(*common.Address, *big.Int) bool
	AddBalance(*common.Address, *big.Int)
	Put(*common.Address, *common.Hash, []byte)
	GetNonce(address *common.Address) *big.Int
	IncrementNonce(address *common.Address)
}
type Storage interface {
	StorageReader
	StorageWriter
}

type StorageReaderWrapper struct {
	StorageReader
	cache            map[common.Hash][]byte
	contract_address *common.Address
}

func (self *StorageReaderWrapper) Init(contract_address *common.Address, backend StorageReader) *StorageReaderWrapper {
	self.StorageReader = backend
	self.contract_address = contract_address
	return self
}

func (self *StorageReaderWrapper) ClearCache() {
	self.cache = nil
}

func (self *StorageReaderWrapper) Get(k *common.Hash, cb func([]byte)) {
	if val, present := self.cache[*k]; present {
		if len(val) != 0 {
			cb(val)
		}
		return
	}
	self.StorageReader.GetAccountStorage(self.contract_address, k, func(bytes []byte) {
		if self.cache == nil {
			self.cache = make(map[common.Hash][]byte)
		}
		bytes = common.CopyBytes(bytes)
		self.cache[*k] = bytes
		cb(bytes)
	})
}

func (self *StorageReaderWrapper) ListForEach(prefix []byte, cb func([]byte)) {
	len_k := Stor_k_2(prefix, bin.BytesView("length"))
	var length uint64
	self.Get(&len_k, func(bytes []byte) {
		length = bin.DEC_b_endian_compact_64(bytes)
	})
	for i := uint64(0); i < length; i++ {
		self.Get(Stor_k_1(prefix, bin.ENC_b_endian_64(i)), cb)
	}
}

// debug-only
func (self *StorageReaderWrapper) ListPrint(prefix []byte) {
	fmt.Println("[")
	self.ListForEach(prefix, func(bytes []byte) {
		fmt.Println("   ", bytes)
	})
	fmt.Println("]")
}

type StorageWrapper struct {
	StorageReaderWrapper
	StorageWriter
	contract_address *common.Address
}

func (self *StorageWrapper) Init(contract_address *common.Address, storage Storage) *StorageWrapper {
	self.StorageReaderWrapper.Init(contract_address, storage)
	self.StorageWriter = storage
	self.contract_address = contract_address
	return self
}

func (self *StorageWrapper) Put(k *common.Hash, v []byte) {
	if self.cache == nil {
		self.cache = make(map[common.Hash][]byte)
	}
	self.cache[*k] = v
	self.StorageWriter.Put(self.contract_address, k, v)
}

func (self *StorageWrapper) ListAppend(prefix []byte, val []byte) (pos uint64) {
	len_k := Stor_k_2(prefix, bin.BytesView("length"))
	self.Get(&len_k, func(bytes []byte) {
		pos = bin.DEC_b_endian_compact_64(bytes)
	})
	self.Put(Stor_k_1(prefix, bin.ENC_b_endian_64(pos)), val)
	self.Put(&len_k, bin.ENC_b_endian_compact_64_1(pos+1))
	return
}

func (self *StorageWrapper) ListRemove(prefix []byte, pos uint64) (last_element []byte) {
	len_k := Stor_k_2(prefix, bin.BytesView("length"))
	var length uint64
	self.Get(&len_k, func(bytes []byte) {
		length = bin.DEC_b_endian_compact_64(bytes)
	})
	asserts.Holds(0 < length)
	if length != 1 {
		lastpos_k := Stor_k_2(prefix, bin.ENC_b_endian_64(length-1))
		self.Get(&lastpos_k, func(bytes []byte) {
			last_element = bytes
		})
		self.Put(&lastpos_k, nil)
	}
	self.Put(Stor_k_1(prefix, bin.ENC_b_endian_64(pos)), last_element)
	self.Put(&len_k, bin.ENC_b_endian_compact_64_1(length-1))
	return
}

func Stor_k_1(parts ...[]byte) *common.Hash {
	return keccak256.Hash(parts...)
}

func Stor_k_2(parts ...[]byte) common.Hash {
	return keccak256.HashAndReturnByValue(parts...)
}
