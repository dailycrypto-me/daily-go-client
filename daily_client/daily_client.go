package daily_client

import (
	"errors"
	"math/big"

	"github.com/dailycrypto-me/daily-go-client/daily_client/dpos_contract_client"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Network uint64

const (
	Undefined Network = iota
	Mainnet
	Testnet
	Devnet
)

// DailyClient contains variables and methods needed for communication with daily dpos smart contract and RPC
type DailyClient struct {
	chainID            *big.Int
	DposContractClient *dpos_contract_client.DposContractClient
	EthRpcClient       *ethclient.Client
	// TODO: add dailyRpcClient
}

func NewDailyClient(chainID *big.Int, url string, dposContractAddress common.Address) (*DailyClient, error) {
	dailyClient := new(DailyClient)
	dailyClient.chainID = chainID

	var err error
	dailyClient.EthRpcClient, err = ethclient.Dial(url)
	if err != nil {
		return nil, err
	}

	dailyClient.DposContractClient, err = dpos_contract_client.NewDposContractClient(dailyClient.EthRpcClient, dposContractAddress, dailyClient.chainID)
	if err != nil {
		return nil, err
	}

	return dailyClient, nil
}

// Creates new daily client based on default values set for different daily networks
func NewDefaultDailyClient(network Network) (*DailyClient, error) {
	var networkRpcUrl string
	var chainID *big.Int
	switch network {
	case Mainnet:
		networkRpcUrl = "https://rpc.mainnet.daily.io"
		chainID = big.NewInt(841)
		break
	case Testnet:
		networkRpcUrl = "https://rpc.testnet.daily.io"
		chainID = big.NewInt(842)
		break
	case Devnet:
		networkRpcUrl = "https://rpc.devnet.daily.io"
		chainID = big.NewInt(843)
		break
	default:
		return nil, errors.New("Invalid network argument")
	}

	// Precompiled dpos contract has the same address in all daily networks
	dposContractAddress := common.HexToAddress("0x00000000000000000000000000000000000000FE")

	return NewDailyClient(chainID, networkRpcUrl, dposContractAddress)
}
