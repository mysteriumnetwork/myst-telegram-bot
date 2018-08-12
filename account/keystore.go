package account

import (
	"context"
	"errors"
	"flag"
	"log"
	"regexp"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

var keyStoreDir = flag.String("keystore.directory", "keystore", "specify runtime dir for keystore keys")
var passphrase = flag.String("keystore.passphrase", "", "passphrase to unlock specified key from keystore")
var address = flag.String("ether.address", "", "Ethereum acc to use for deployment")
var newAccount = flag.Bool("create.account", false, "Creates a new Ethereum address")

type FaucetAccount struct {
	KS  *keystore.KeyStore
	Acc *accounts.Account
}

func CreateFaucetAccount() (*FaucetAccount, error) {
	log.Println("Faucet newAccount: ", *newAccount)

	faucetAccount, err := createNewAccount()
	if err != nil {
		return nil, err
	}

	if faucetAccount != nil {
		*address = faucetAccount.Address.String()
	}

	var ks *keystore.KeyStore
	var account *accounts.Account

	log.Println("Trying to use account: ", *address)

	if *address != "" {
		ks = GetKeystore()
		account, err = getUnlockedAcc(*address, ks)
	} else {
		log.Println("no address specified, generate new or choose from: ")
		listAccounts()
		return nil, errors.New("no account specified")
	}

	return &FaucetAccount{ks, account}, err
}

func GetKeystore() *keystore.KeyStore {
	return keystore.NewKeyStore(*keyStoreDir, keystore.StandardScryptN, keystore.StandardScryptP)
}

func listAccounts() error {
	ks := GetKeystore()
	for i, acc := range ks.Accounts() {
		log.Printf("%d: Address: %s\n", i, acc.Address.String())
	}
	return nil
}

func createNewAccount() (*accounts.Account, error) {
	if *newAccount {
		ks := GetKeystore()
		a, err := ks.NewAccount(*passphrase)
		return &a, err
	}
	return nil, nil
}

func getUnlockedAcc(address string, ks *keystore.KeyStore) (*accounts.Account, error) {
	searchAcc := accounts.Account{Address: common.HexToAddress(address)}
	foundAcc, err := ks.Find(searchAcc)
	if err != nil {
		return nil, err
	}
	err = ks.Unlock(foundAcc, *passphrase)
	if err != nil {
		return nil, err
	}
	return &foundAcc, nil
}

func (aa *FaucetAccount) CreateNewKeystoreTransactor() *bind.TransactOpts {
	return &bind.TransactOpts{
		From: aa.Acc.Address,
		Signer: func(signer types.Signer, address common.Address, tx *types.Transaction) (*types.Transaction, error) {
			if address != aa.Acc.Address {
				return nil, errors.New("not authorized to sign this account")
			}
			signature, err := aa.KS.SignHash(*aa.Acc, signer.Hash(tx).Bytes())
			if err != nil {
				return nil, err
			}
			return tx.WithSignature(signer, signature)

		},
	}
}

func (aa *FaucetAccount) CreateNewFilterer() *bind.FilterOpts {
	return &bind.FilterOpts{
		Start:   3690000,
		End:     nil,
		Context: context.Background(),
	}
}

func IsAddressValid(address string) bool {
	var validID = regexp.MustCompile(`^0x[a-fA-F0-9]{40}$`)
	return validID.MatchString(address)
}
