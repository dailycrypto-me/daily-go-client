package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/big"
	"os"
	"path"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
	"strconv"
	"time"
	"unsafe"

	"github.com/dailycrypto-me/daily-evm/daily/util/files"

	"github.com/dailycrypto-me/daily-evm/daily/state"

	"github.com/dailycrypto-me/daily-evm/daily/state/chain_config"
	"github.com/dailycrypto-me/daily-evm/daily/state/state_db_rocksdb"

	"github.com/dailycrypto-me/daily-evm/common"
	"github.com/dailycrypto-me/daily-evm/common/hexutil"
	"github.com/dailycrypto-me/daily-evm/core"
	"github.com/dailycrypto-me/daily-evm/core/types"
	"github.com/dailycrypto-me/daily-evm/core/vm"
	"github.com/dailycrypto-me/daily-evm/params"
	"github.com/dailycrypto-me/daily-evm/daily/util"
	"github.com/dailycrypto-me/daily-evm/daily/util/asserts"
	"github.com/dailycrypto-me/daily-evm/daily/util/bin"
	"github.com/linxGnu/grocksdb"
)

func main() {
	const profiling = false
	const desired_num_trx_per_block = 20000
	//const desired_num_trx_per_block = 0

	usr_dir, e1 := os.UserHomeDir()
	util.PanicIfNotNil(e1)
	dest_data_dir := files.CreateDirectories(usr_dir, "daily_evm_test")

	profile_basedir := files.CreateDirectories(dest_data_dir, "profiles")
	new_prof_file := func(time time.Time, kind string) *os.File {
		ret, err := os.Create(path.Join(profile_basedir, strconv.FormatInt(time.Unix(), 10)+"_"+kind+".prof"))
		util.PanicIfNotNil(err)
		return ret
	}
	last_profile_snapshot_time := time.Now()
	if profiling {
		debug.SetGCPercent(-1)
		pprof.StartCPUProfile(new_prof_file(last_profile_snapshot_time, "cpu"))
	}

	blk_db_opts := grocksdb.NewDefaultOptions()
	blk_db_opts.SetErrorIfExists(false)
	blk_db_opts.SetCreateIfMissing(true)
	blk_db_opts.SetCreateIfMissingColumnFamilies(true)
	blk_db_opts.IncreaseParallelism(runtime.NumCPU())
	blk_db_opts.SetMaxFileOpeningThreads(runtime.NumCPU())
	blk_db_opts.OptimizeForPointLookup(256)
	blk_db_opts.SetMaxOpenFiles(32)
	blk_db, e0 := grocksdb.OpenDbForReadOnly(blk_db_opts, "/home/oleg/win10/ubuntu/blockchain", false)
	util.PanicIfNotNil(e0)
	defer blk_db.Close()

	type Transaction struct {
		From     common.Address  `json:"from" gencodec:"required"`
		GasPrice *hexutil.Big    `json:"gasPrice" gencodec:"required"`
		To       *common.Address `json:"to,omitempty"`
		Nonce    *hexutil.Big    `json:"nonce" gencodec:"required"`
		Value    *hexutil.Big    `json:"value" gencodec:"required"`
		Gas      hexutil.Uint64  `json:"gas" gencodec:"required"`
		Input    hexutil.Bytes   `json:"input" gencodec:"required"`
	}
	type Log struct {
		Address common.Address `json:"address"  gencodec:"required"`
		Topics  []common.Hash  `json:"topics"  gencodec:"required"`
		Data    hexutil.Bytes  `json:"data"  gencodec:"required"`
	}
	type Receipt struct {
		ContractAddress *common.Address `json:"contractAddress"`
		GasUsed         hexutil.Uint64  `json:"gasUsed"  gencodec:"required"`
		Logs            []Log           `json:"logs"  gencodec:"required"`
	}
	type TransactionAndReceipt struct {
		Transaction
		Receipt Receipt `json:"receipt"  gencodec:"required"`
	}
	type UncleBlock struct {
		Number hexutil.Uint64 `json:"number"  gencodec:"required"`
		Miner  common.Address `json:"miner"  gencodec:"required"`
	}
	type VmBlock struct {
		Miner      common.Address `json:"miner" gencodec:"required"`
		GasLimit   hexutil.Uint64 `json:"gasLimit"  gencodec:"required"`
		Time       hexutil.Uint64 `json:"timestamp"  gencodec:"required"`
		Difficulty *hexutil.Big   `json:"difficulty"  gencodec:"required"`
	}
	type BlockInfo struct {
		VmBlock
		UncleBlocks  []UncleBlock            `json:"uncleBlocks"  gencodec:"required"`
		Transactions []TransactionAndReceipt `json:"transactions"  gencodec:"required"`
		Hash         common.Hash             `json:"hash" gencodec:"required"`
		StateRoot    common.Hash             `json:"stateRoot" gencodec:"required"`
	}

	rocksdb_opts_r_default := grocksdb.NewDefaultReadOptions()
	getBlockByNumber := func(block_num types.BlockNum) *BlockInfo {
		block_json, err := blk_db.GetPinned(rocksdb_opts_r_default, bin.BytesView(fmt.Sprintf("%09d", block_num)))
		util.PanicIfNotNil(err)
		ret := new(BlockInfo)
		util.PanicIfNotNil(json.Unmarshal(block_json.Data(), ret))
		block_json.Destroy()
		return ret
	}

	statedb := new(state_db_rocksdb.DB).Init(state_db_rocksdb.Opts{
		Path: files.CreateDirectories(dest_data_dir, "state_db"),
	})
	defer statedb.Close()
	api := new(state.API).Init(
		statedb,
		func(num types.BlockNum) *big.Int {
			return new(big.Int).SetBytes(getBlockByNumber(num).Hash[:])
		},
		&chain_config.ChainConfig{
			EVMChainConfig:  *params.MainnetChainConfig,
			GenesisBalances: core.MainnetGenesisBalances(),
		},
		state.APIOpts{
			ExpectedMaxTrxPerBlock:        desired_num_trx_per_block + 1,
			MainTrieFullNodeLevelsToCache: 4,
		},
	)
	defer api.Close()
	st := api.GetStateTransition()

	last_committed_state_desc := statedb.GetLatestState().GetCommittedDescriptor()
	last_blk_num := last_committed_state_desc.BlockNum
	asserts.EQ(getBlockByNumber(last_blk_num).StateRoot.Hex(), last_committed_state_desc.StateRoot.Hex())
	tps_sum, tps_cnt, tps_min, tps_max := 0.0, 0, math.MaxFloat64, -1.0
	for {
		blk_num_since := last_blk_num + 1
		var block_buf []*BlockInfo
		tx_count := 0
		for {
			last_blk_num++
			last_block := getBlockByNumber(last_blk_num)
			block_buf = append(block_buf, last_block)
			tx_count += len(last_block.Transactions)
			if tx_count >= desired_num_trx_per_block {
				break
			}
		}
		fmt.Println("blocks:", blk_num_since, "-", last_blk_num, "tx_count:", tx_count)
		time_before_execution := time.Now()
		for _, b := range block_buf {
			st.BeginBlock((*vm.BlockInfo)(unsafe.Pointer(&b.VmBlock)))
			for _, trx_and_receipt := range b.Transactions {
				res := st.ExecuteTransaction((*vm.Transaction)(unsafe.Pointer(&trx_and_receipt.Transaction)))
				receipt := trx_and_receipt.Receipt
				asserts.EQ(uint64(receipt.GasUsed), res.GasUsed)
				asserts.EQ(len(receipt.Logs), len(res.Logs))
				for i, log := range receipt.Logs {
					actual_log := res.Logs[i]
					asserts.EQ(log.Address, actual_log.Address)
					asserts.Holds(bytes.Equal(log.Data, actual_log.Data))
					asserts.EQ(len(log.Topics), len(actual_log.Topics))
					for i, topic := range log.Topics {
						asserts.EQ(topic, actual_log.Topics[i])
					}
				}
				if receipt.ContractAddress == nil {
					asserts.EQ(common.ZeroAddress, res.NewContractAddr)
				} else {
					asserts.EQ(*receipt.ContractAddress, res.NewContractAddr)
				}
			}
			st.EndBlock()
		}
		state_root := st.PrepareCommit()
		asserts.EQ(block_buf[len(block_buf)-1].StateRoot.Hex(), state_root.Hex())
		//return
		st.Commit()
		tps := float64(tx_count) / time.Now().Sub(time_before_execution).Seconds()
		tps_sum += tps
		tps_cnt++
		if tps < tps_min {
			tps_min = tps
		}
		if tps_max < tps {
			tps_max = tps
		}
		fmt.Println("TPS current:", tps, "avg:", tps_sum/float64(tps_cnt), "min:", tps_min, "max:", tps_max)
		if profiling {
			pprof.StopCPUProfile()
			fmt.Println("gc...")
			runtime.GC()
			util.PanicIfNotNil(pprof.WriteHeapProfile(new_prof_file(last_profile_snapshot_time, "heap")))
			last_profile_snapshot_time = time.Now()
			util.PanicIfNotNil(pprof.StartCPUProfile(new_prof_file(last_profile_snapshot_time, "cpu")))
		} else {
			fmt.Println("gc...")
			runtime.GC()
		}
		debug.FreeOSMemory()
	}
}

func write_file(fname string, val []byte) {
	util.PanicIfNotNil(ioutil.WriteFile(fname, val, 0644))
}

func read_file(fname string) []byte {
	ret, err := ioutil.ReadFile(fname)
	if os.IsNotExist(err) {
		return nil
	}
	util.PanicIfNotNil(err)
	return ret
}
