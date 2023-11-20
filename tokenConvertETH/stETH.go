package main

import (
	"context"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	nodeURL             = "wss://ethereum-goerli.publicnode.com" // Замените на вашу ноду
	targetAddressString = "0x1643E812aE58766192Cf7D2Cf9567dF2C37e9B7F"
	chainID             = 5 // ID сети Goerli
	ethToSend           = 1 // количество ETH для отправки
)

func main() {
	client, err := ethclient.Dial(nodeURL)
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}

	wallets := []struct {
		Address    string
		PrivateKey string
		Delay      time.Duration // задержка перед отправкой транзакции
	}{
		//кошельки с приватниками
		{"0x1", "приватник", 10 * time.Second},
		{"0x05", "приватник", 10 * time.Second},
		{"0xc4", "приватник", 10 * time.Second},
	}

	targetAddress := common.HexToAddress(targetAddressString)

	for _, wallet := range wallets {
		go sendEthereum(client, wallet.Address, wallet.PrivateKey, targetAddress, chainID, wallet.Delay)
	}

	select {}
}

func sendEthereum(client *ethclient.Client, fromAddressStr, privateKeyStr string, targetAddress common.Address, chainID int64, delay time.Duration) {
	privateKey, err := crypto.HexToECDSA(privateKeyStr)
	if err != nil {
		log.Fatalf("Failed to parse private key: %v", err)
	}

	fromAddress := common.HexToAddress(fromAddressStr)

	for {
		// Задержка перед отправкой каждой транзакции
		time.Sleep(delay)

		balance, err := client.BalanceAt(context.Background(), fromAddress, nil)
		if err != nil {
			log.Printf("Failed to retrieve balance: %v", err)
			continue
		}

		sendAmount := big.NewInt(0).Mul(big.NewInt(ethToSend), big.NewInt(1e18)) // 1 ETH в wei
		if balance.Cmp(sendAmount) < 0 {
			log.Printf("Insufficient funds for address %s", fromAddressStr)
			break // Прерывание цикла, если недостаточно средств
		}

		nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
		if err != nil {
			log.Printf("Failed to retrieve nonce: %v", err)
			continue
		}

		gasPrice, err := client.SuggestGasPrice(context.Background())
		if err != nil {
			log.Printf("Failed to suggest gas price: %v", err)
			continue
		}

		auth, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(chainID))
		if err != nil {
			log.Printf("Failed to create authorized transactor: %v", err)
			continue
		}
		auth.Nonce = big.NewInt(int64(nonce))
		auth.Value = sendAmount         // установка суммы отправки
		auth.GasLimit = uint64(2100000) // вы можете настроить это значение
		auth.GasPrice = gasPrice

		tx := types.NewTransaction(nonce, targetAddress, sendAmount, auth.GasLimit, auth.GasPrice, nil)
		signedTx, err := types.SignTx(tx, types.NewEIP155Signer(big.NewInt(chainID)), privateKey)
		if err != nil {
			log.Printf("Failed to sign transaction: %v", err)
			continue
		}

		err = client.SendTransaction(context.Background(), signedTx)
		if err != nil {
			log.Printf("Failed to send transaction: %v", err)
			continue
		}

		log.Printf("Transaction sent: %s", signedTx.Hash().Hex())
	}
}
