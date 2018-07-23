package keystore

import (
	"flag"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"errors"
	"log"
	"github.com/ethereum/go-ethereum/core/types"
)

var KeyStoreDir = flag.String("keystore.directory", "testnet", "specify runtime dir for keystore keys")
var Passphrase = flag.String("keystore.passphrase", "***REMOVED***", "Pashprase to unlock specified key from keystore")
var Address = *flag.String("ether.address", "0xCf16489612B1D8407Fd66960eCB21941718CD8FD", "Ethereum acc to use for deployment")
var newAccount = *flag.Bool("create.account",  false, "Creates a new Ethereum address")

type ActiveAccount struct {
	Keystore *keystore.KeyStore
	Account *accounts.Account
}

func getKeystore() *keystore.KeyStore {
	return keystore.NewKeyStore(*KeyStoreDir, keystore.StandardScryptN, keystore.StandardScryptP)
}

func listAccounts() error {
	ks := getKeystore()
	for i, acc := range ks.Accounts() {
		log.Printf("%d: Address: %s\n", i, acc.Address.String())
	}
	return nil
}

func createNewAccount() (err error) {
	if newAccount {
		ks := getKeystore()
		_, err = ks.NewAccount(*Passphrase)
	}
	return
}

func (aa *ActiveAccount) GetAccount() (*accounts.Account, error) {
	err := createNewAccount()
	if err != nil {
		return nil, err
	}

	if Address != "" {
		aa.Keystore = getKeystore()
		aa.Account, err = getUnlockedAcc(Address, aa.Keystore)
		return aa.Account, err
	}

	log.Println("no address specified, generate new or choose from: ")
	listAccounts()
	return nil, errors.New("no account specified")
}

func getUnlockedAcc(address string, ks *keystore.KeyStore) (*accounts.Account, error) {
	searchAcc := accounts.Account{Address: common.HexToAddress(address)}
	foundAcc, err := ks.Find(searchAcc)
	if err != nil {
		return nil, err
	}
	err = ks.Unlock(foundAcc, *Passphrase)
	if err != nil {
		return nil, err
	}
	return &foundAcc, nil
}

func (aa *ActiveAccount) CreateNewKeystoreTransactor() *bind.TransactOpts {
	return &bind.TransactOpts{
		From: aa.Account.Address,
		Signer: func(signer types.Signer, address common.Address, tx *types.Transaction) (*types.Transaction, error) {
			if address != aa.Account.Address {
				return nil, errors.New("not authorized to sign this account")
			}
			signature, err := aa.Keystore.SignHash(*aa.Account, signer.Hash(tx).Bytes())
			if err != nil {
				return nil, err
			}
			return tx.WithSignature(signer, signature)

		},
	}
}
