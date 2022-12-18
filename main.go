package main

import (
	"context"
	"flag"
	"fmt"
	"math/big"
	"strings"
	"sync"

	"github.com/dontpanicdao/caigo"
	"github.com/dontpanicdao/caigo/gateway"
	"github.com/dontpanicdao/caigo/types"
)

var (
	chainId          string
	contractAddress  string
	address          string
	privateKey       string
	passAmount       bool
	approve          bool
	pollIntervalFlag int
)

const (
	pollInterval  int    = 5
	feeMargin     int    = 175
	etherContract string = "0x049d36570d4e46f48e99674bd3fcc84644ddd6b96f7c741b1562b82f9e004dc7"
)

func main() {
	flag.StringVar(&chainId, "chainid", "mainnet", "id of the chain")
	flag.StringVar(&contractAddress, "contractaddress", "0x012f8e318fe04a1fe8bffe005ea4bbd19cb77a656b4f42682aab8a0ed20702f0", "address of the smart contract")
	flag.StringVar(&address, "addresses", "", "wallet addresses")
	flag.StringVar(&privateKey, "privatekeys", "", "wallet private keys")
	flag.Parse()
	fmt.Println(chainId, contractAddress, "add: ", address, "privkeys: ", privateKey)
	if address == "" || privateKey == "" {
		fmt.Println("no wallet or private key was found")
		return
	}

	addresses := strings.Split(address, ",")
	privateKeys := strings.Split(privateKey, ",")
	fmt.Println("addresses: ", addresses, " privkeys: ", privateKeys, " contractaddress: ", contractAddress)
	var wg sync.WaitGroup
	for i := range privateKeys {
		wg.Add(1)
		go func(index int) {
			functionCall := []types.FunctionCall{
				{
					ContractAddress:    types.HexToHash(etherContract),
					EntryPointSelector: "approve",
					Calldata: []string{
						types.HexToBN(contractAddress).String(),
						"50000000000000000",
						"0",
					},
				},
				{
					ContractAddress:    types.HexToHash(contractAddress),
					EntryPointSelector: "publicMint",
				},
			}

			indexConst := fmt.Sprintf("index %d ", index)
			defer wg.Done()

			gw := gateway.NewProvider(gateway.WithBaseURL(chainId))
			fmt.Println(gw.ChainId)
			fmt.Println(indexConst, "creating gateway account")
			acc, err := caigo.NewGatewayAccount(privateKeys[index], addresses[index], gw, caigo.AccountVersion1)
			if err != nil {
				fmt.Println(indexConst, "err:", err)
				return
			}
			fmt.Println(indexConst, acc)
			/*	fmt.Println(indexConst, "Estimating fee")
				feeEstimate, err := acc.EstimateFee(context.Background(), functionCall, types.ExecuteDetails{})
				if err != nil {
					fmt.Println(indexConst, "err:", err)
					return
				}

				fmt.Println(indexConst, "Executing tx")
				fee, _ := big.NewInt(0).SetString(string(feeEstimate.OverallFee), 0)
				expandedFee := big.NewInt(0).Mul(fee, big.NewInt(int64(feeMargin)))
				maxFee := big.NewInt(0).Div(expandedFee, big.NewInt(100))*/
			resp, err := acc.Execute(context.TODO(), functionCall, types.ExecuteDetails{
				MaxFee: big.NewInt(10000000000000000),
			})
			if err != nil {
				fmt.Println(indexConst, "err:", err)
				return
			}
			fmt.Println(indexConst, "polling for tx hash", resp.TransactionHash)
			_, receipt, err := gw.WaitForTransaction(context.TODO(), resp.TransactionHash, pollInterval, 12)
			if err != nil {
				fmt.Println(indexConst, "err:", err)
				return
			}
			fmt.Println(indexConst, "transaction succeed hash:", resp.TransactionHash, "receipt:", receipt.BlockHash)
		}(i)
	}
	wg.Wait()
	fmt.Println("finished")
	return
}
