package dpos_contract_client

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"math/big"

	dpos_interface "github.com/dailycrypto-me/daily-go-client/daily_client/dpos_contract_client/dpos_interface"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// DposContractClient contains variables needed for communication with daily dpos smart contract
type DposContractClient struct {
	dposInterface *dpos_interface.DposInterface
	ethClient     *ethclient.Client
	chainID       *big.Int
}

type Transactor struct {
	TransactOpts *bind.TransactOpts
	Address      common.Address
	Nonce        uint64
}

func NewDposContractClient(ethClient *ethclient.Client, dposContractAddress common.Address, chainID *big.Int) (*DposContractClient, error) {
	dposContractClient := new(DposContractClient)
	var err error

	dposContractClient.dposInterface, err = dpos_interface.NewDposInterface(dposContractAddress, ethClient)
	dposContractClient.ethClient = ethClient
	dposContractClient.chainID = chainID
	if err != nil {
		return nil, err
	}

	return dposContractClient, nil
}

func (DposContractClient *DposContractClient) GetTotalEligibleVotesCount() (uint64, error) {
	return DposContractClient.dposInterface.GetTotalEligibleVotesCount(&bind.CallOpts{})
}

func (DposContractClient *DposContractClient) GetValidator(validator common.Address) (dpos_interface.DposInterfaceValidatorBasicInfo, error) {
	return DposContractClient.dposInterface.GetValidator(&bind.CallOpts{}, validator)
}

func (DposContractClient *DposContractClient) GetValidatorEligibleVotesCount(validator common.Address) (uint64, error) {
	return DposContractClient.dposInterface.GetValidatorEligibleVotesCount(&bind.CallOpts{}, validator)
}

func (DposContractClient *DposContractClient) IsValidatorEligible(validator common.Address) (bool, error) {
	return DposContractClient.dposInterface.IsValidatorEligible(&bind.CallOpts{}, validator)
}

func (DposContractClient *DposContractClient) GetValidators() ([]dpos_interface.DposInterfaceValidatorData, error) {
	return DposContractClient.getValidators(&bind.CallOpts{})
}

func (DposContractClient *DposContractClient) GetValidatorsAtBlock(block_num *big.Int) ([]dpos_interface.DposInterfaceValidatorData, error) {
	return DposContractClient.getValidators(&bind.CallOpts{BlockNumber: block_num})
}

func (DposContractClient *DposContractClient) getValidators(opts *bind.CallOpts) ([]dpos_interface.DposInterfaceValidatorData, error) {
	var validators []dpos_interface.DposInterfaceValidatorData

	for batch := uint32(0); ; batch++ {
		validatorsBatch, err := DposContractClient.dposInterface.GetValidators(opts, batch)
		if err != nil {
			return nil, err
		}

		if len(validatorsBatch.Validators) != 0 {
			validators = append(validators, validatorsBatch.Validators...)
		}

		if validatorsBatch.End == true {
			break
		}
	}

	return validators, nil
}

func (DposContractClient *DposContractClient) GetOwnerValidators(owner common.Address) ([]dpos_interface.DposInterfaceValidatorData, error) {
	var validators []dpos_interface.DposInterfaceValidatorData

	for batch := uint32(0); ; batch++ {
		validatorsBatch, err := DposContractClient.dposInterface.GetValidatorsFor(&bind.CallOpts{}, owner, batch)
		if err != nil {
			return nil, err
		}

		if len(validatorsBatch.Validators) != 0 {
			validators = append(validators, validatorsBatch.Validators...)
		}

		if validatorsBatch.End == true {
			break
		}
	}

	return validators, nil
}

func (DposContractClient *DposContractClient) GetDelegations(delegator common.Address) ([]dpos_interface.DposInterfaceDelegationData, error) {
	var delegations []dpos_interface.DposInterfaceDelegationData

	for batch := uint32(0); ; batch++ {
		delegationsBatch, err := DposContractClient.dposInterface.GetDelegations(&bind.CallOpts{}, delegator, batch)
		if err != nil {
			return nil, err
		}

		if len(delegationsBatch.Delegations) != 0 {
			delegations = append(delegations, delegationsBatch.Delegations...)
		}

		if delegationsBatch.End == true {
			break
		}
	}

	return delegations, nil
}

func (DposContractClient *DposContractClient) GetUndelegations(delegator common.Address) ([]dpos_interface.DposInterfaceUndelegationData, error) {
	var undelegations []dpos_interface.DposInterfaceUndelegationData

	for batch := uint32(0); ; batch++ {
		undelegationsBatch, err := DposContractClient.dposInterface.GetUndelegations(&bind.CallOpts{}, delegator, batch)
		if err != nil {
			return nil, err
		}

		if len(undelegationsBatch.Undelegations) != 0 {
			undelegations = append(undelegations, undelegationsBatch.Undelegations...)
		}

		if undelegationsBatch.End == true {
			break
		}
	}

	return undelegations, nil
}

func (DposContractClient *DposContractClient) NewTransactor(privateKeyStr string) (*Transactor, error) {
	privateKey, err := crypto.HexToECDSA(privateKeyStr)
	if err != nil {
		return nil, err
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	address := crypto.PubkeyToAddress(*publicKeyECDSA)
	if !ok {
		return nil, errors.New("error casting public key to ECDSA")
	}

	transactOpts, err := bind.NewKeyedTransactorWithChainID(privateKey, DposContractClient.chainID)
	if err != nil {
		return nil, err
	}

	nonce, err := DposContractClient.ethClient.PendingNonceAt(context.Background(), address)
	if err != nil {
		return nil, err
	}

	transactor := new(Transactor)
	transactor.Address = address
	transactor.Nonce = nonce
	transactor.TransactOpts = new(bind.TransactOpts)
	*transactor.TransactOpts = *transactOpts

	return transactor, nil
}

func (DposContractClient *DposContractClient) createNewTransactOpts(transactor *Transactor) (*bind.TransactOpts, error) {
	nonce, err := DposContractClient.ethClient.PendingNonceAt(context.Background(), transactor.Address)
	if err != nil {
		return nil, err
	}

	gasPrice, err := DposContractClient.ethClient.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, err
	}

	maxNonce := transactor.Nonce
	if nonce > maxNonce {
		maxNonce = nonce
	}

	transactOpts := new(bind.TransactOpts)
	*transactOpts = *transactor.TransactOpts

	transactOpts.Nonce = big.NewInt(int64(maxNonce))
	transactOpts.GasLimit = uint64(300000) // in units
	transactOpts.GasPrice = gasPrice

	// Increment transactos nonce for the next tx
	transactor.Nonce++

	return transactOpts, nil
}

func (DposContractClient *DposContractClient) Delegate(transactor *Transactor, amount *big.Int, validator common.Address) (*types.Transaction, error) {
	transactOpts, err := DposContractClient.createNewTransactOpts(transactor)
	if err != nil {
		return nil, err
	}

	transactOpts.Value = amount // in wei
	return DposContractClient.dposInterface.Delegate(transactOpts, validator)
}

func (DposContractClient *DposContractClient) Undelegate(transactor *Transactor, amount *big.Int, validator common.Address) (*types.Transaction, error) {
	transactOpts, err := DposContractClient.createNewTransactOpts(transactor)
	if err != nil {
		return nil, err
	}

	return DposContractClient.dposInterface.Undelegate(transactOpts, validator, amount)
}

func (DposContractClient *DposContractClient) ConfirmUndelegate(transactor *Transactor, validator common.Address) (*types.Transaction, error) {
	transactOpts, err := DposContractClient.createNewTransactOpts(transactor)
	if err != nil {
		return nil, err
	}

	return DposContractClient.dposInterface.ConfirmUndelegate(transactOpts, validator)
}

func (DposContractClient *DposContractClient) CancelUndelegate(transactor *Transactor, validator common.Address) (*types.Transaction, error) {
	transactOpts, err := DposContractClient.createNewTransactOpts(transactor)
	if err != nil {
		return nil, err
	}

	return DposContractClient.dposInterface.CancelUndelegate(transactOpts, validator)
}

func (DposContractClient *DposContractClient) RedelegateUndelegate(transactor *Transactor, amount *big.Int, validatorFrom common.Address, validatorTo common.Address) (*types.Transaction, error) {
	transactOpts, err := DposContractClient.createNewTransactOpts(transactor)
	if err != nil {
		return nil, err
	}

	return DposContractClient.dposInterface.ReDelegate(transactOpts, validatorFrom, validatorTo, amount)
}

func (DposContractClient *DposContractClient) ClaimRewards(transactor *Transactor, validator common.Address) (*types.Transaction, error) {
	transactOpts, err := DposContractClient.createNewTransactOpts(transactor)
	if err != nil {
		return nil, err
	}

	return DposContractClient.dposInterface.ClaimRewards(transactOpts, validator)
}

func (DposContractClient *DposContractClient) ClaimCommissionRewards(transactor *Transactor, validator common.Address) (*types.Transaction, error) {
	transactOpts, err := DposContractClient.createNewTransactOpts(transactor)
	if err != nil {
		return nil, err
	}

	return DposContractClient.dposInterface.ClaimCommissionRewards(transactOpts, validator)
}

func (DposContractClient *DposContractClient) RegisterValidator(transactor *Transactor, validator common.Address, proof []byte, vrf_key []byte, commission uint16, description string, endpoint string) (*types.Transaction, error) {
	transactOpts, err := DposContractClient.createNewTransactOpts(transactor)
	if err != nil {
		return nil, err
	}

	return DposContractClient.dposInterface.RegisterValidator(transactOpts, validator, proof, vrf_key, commission, description, endpoint)
}

func (DposContractClient *DposContractClient) SetValidatorInfo(transactor *Transactor, validator common.Address, description string, endpoint string) (*types.Transaction, error) {
	transactOpts, err := DposContractClient.createNewTransactOpts(transactor)
	if err != nil {
		return nil, err
	}

	return DposContractClient.dposInterface.SetValidatorInfo(transactOpts, validator, description, endpoint)
}

func (DposContractClient *DposContractClient) SetCommission(transactor *Transactor, validator common.Address, commission uint16) (*types.Transaction, error) {
	transactOpts, err := DposContractClient.createNewTransactOpts(transactor)
	if err != nil {
		return nil, err
	}

	return DposContractClient.dposInterface.SetCommission(transactOpts, validator, commission)
}
