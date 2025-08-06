package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/core/types"
)

func main() {
	// 1. Connect to Ethereum
	client, err := ethclient.Dial("https://mainnet.infura.io/v3/<PROJECT_ID>")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// 2. SET TX HASH HERE WHODE DATA IS NEEDED
	txHash := common.HexToHash("TX_HASH_HERE")

	// 3. Get tx object
	tx, _, err := client.TransactionByHash(context.Background(), txHash)
	if err != nil {
		log.Fatal("tx fetch:", err)
	}

	// 4. Extract standard fieldss
	fmt.Println("Transaction hash:", tx.Hash())
	fmt.Println("Type:", tx.Type())
	fmt.Println("ChainID:", tx.ChainId())
	fmt.Println("Nonce:", tx.Nonce())
	fmt.Println("GasTipCap:", tx.GasTipCap())
	fmt.Println("GasFeeCap:", tx.GasFeeCap())
	fmt.Println("Gas:", tx.Gas())
	fmt.Println("To:", tx.To())
	fmt.Println("Value:", tx.Value())
	fmt.Println("Data: ", tx.Data())
	fmt.Println("AccessList:", tx.AccessList())
	fmt.Println("BlobGasFeeCap: ", tx.BlobGasFeeCap())
	fmt.Println("BlobVersionedHashes:", tx.BlobHashes())
	
	// 5. Signature
	v, r, s := tx.RawSignatureValues()
	fmt.Println("v:", v)
	fmt.Println("r:", r)
	fmt.Println("s:", s)

	// 6. Optional: Print calldata
	fmt.Printf("Data: 0x%x\n", tx.Data())

	// Add this line to get the hash to be signed (sighash)
	signer := types.NewCancunSigner(tx.ChainId())
	sighash := signer.Hash(tx)
	fmt.Println("SigHash (to be signed):", sighash.Hex())

}
