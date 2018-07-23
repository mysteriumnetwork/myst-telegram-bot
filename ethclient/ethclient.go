package ethclient

import (
	"flag"
	"github.com/ethereum/go-ethereum/ethclient"
	"fmt"
	"context"
)

var GethUrl = flag.String("geth.url", "https://ropsten.infura.io/v3/0cf3087cfc4f4c80a349c305aed2d835", "URL value of started geth to connect")

func LookupBackend() (*ethclient.Client, chan bool, error) {
	ethClient, err := ethclient.Dial(*GethUrl)
	if err != nil {
		return nil, nil, err
	}

	block, err := ethClient.BlockByNumber(context.Background(), nil)
	if err != nil {
		return nil, nil, err
	}
	fmt.Println("Latest known block is: ", block.NumberU64())

	progress, err := ethClient.SyncProgress(context.Background())
	if err != nil {
		return nil, nil, err
	}
	completed := make(chan bool)
	if progress != nil {
		fmt.Println("Client is in syncing state - any operations will be delayed until finished")
		// go trackGethProgress(ethClient, progress, completed)
	} else {
		fmt.Println("Geth process fully synced")
		close(completed)
	}

	return ethClient, completed, nil
}
