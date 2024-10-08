package slashing

import (
	"github.com/dailycrypto-me/daily-evm/common"
	"github.com/dailycrypto-me/daily-evm/core/types"
	"github.com/dailycrypto-me/daily-evm/rlp"
	"github.com/dailycrypto-me/daily-evm/daily/state/chain_config"
	contract_storage "github.com/dailycrypto-me/daily-evm/daily/state/contracts/storage"
)

type Reader struct {
	cfg     *chain_config.ChainConfig
	storage *contract_storage.StorageReaderWrapper
}

func (self *Reader) Init(cfg *chain_config.ChainConfig, blk_n types.BlockNum, storage_factory func(types.BlockNum) contract_storage.StorageReader) *Reader {
	self.cfg = cfg

	blk_n_actual := uint64(0)
	if uint64(self.cfg.DPOS.DelegationDelay) < blk_n {
		blk_n_actual = blk_n - uint64(self.cfg.DPOS.DelegationDelay)
	}

	self.storage = new(contract_storage.StorageReaderWrapper).Init(slashing_contract_address, storage_factory(blk_n_actual))
	return self
}

func (self *Reader) getJailBlock(addr *common.Address) (jailed bool, block types.BlockNum) {
	block = 0
	jailed = false

	db_key := contract_storage.Stor_k_1(field_validators_jail_block, addr.Bytes())
	self.storage.Get(db_key, func(bytes []byte) {
		rlp.MustDecodeBytes(bytes, &block)
		jailed = true
	})

	return
}

func (self Reader) IsJailed(block types.BlockNum, addr *common.Address) bool {
	if !self.cfg.Hardforks.IsMagnoliaHardfork(block) {
		return false
	}

	jailed, jail_block := self.getJailBlock(addr)
	if !jailed {
		return false
	}

	if jail_block < block {
		return false
	}

	return true
}
