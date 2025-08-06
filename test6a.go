package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/holiman/uint256"
)

func prefixedRlpHash(prefix byte, x interface{}) (h common.Hash) {
	sha := crypto.NewKeccakState()
	sha.Write([]byte{prefix})
	rlp.Encode(sha, x)
	sha.Read(h[:])
	return h
}

func main() {
	inputFile := "tx_data.csv"
	outputFile := "suffix_mining_results.csv"

	// Open CSV
	inFile, err := os.Open(inputFile)
	if err != nil {
		log.Fatalf("Failed to open input file: %v", err)
	}
	defer inFile.Close()

	reader := csv.NewReader(inFile)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("Failed to read CSV: %v", err)
	}

	headers := records[0]
	colIndex := make(map[string]int)
	for i, col := range headers {
		colIndex[col] = i
	}

	// Prepare output CSV
	outFile, err := os.Create(outputFile)
	if err != nil {
		log.Fatalf("Failed to create output file: %v", err)
	}
	defer outFile.Close()
	writer := csv.NewWriter(outFile)
	defer writer.Flush()

	header := []string{"original_hash"}
	for i := 0; i < 16; i++ {
		header = append(header, fmt.Sprintf("%x", i))
	}
	header = append(header, "time_microseconds")
	writer.Write(header)

	totalStart := time.Now()

	// Process each tx
	for i, row := range records {
		if i == 0 {
			continue
		}

		expectedHash := row[colIndex["hash"]]
		chainID := uint256.NewInt(1)
		nonce, _ := strconv.ParseUint(row[colIndex["nonce"]], 10, 64)
		to := common.HexToAddress(row[colIndex["to"]])
		value := uint256.NewInt(0)

		gasFeeCapBig, _ := new(big.Int).SetString(row[colIndex["gas_fee_cap"]], 10)
		gasFeeCap := uint256.MustFromBig(gasFeeCapBig)
		gasLimit := uint64(21000)
		data := []byte{}

		blobFeeCapBig, _ := new(big.Int).SetString(row[colIndex["blob_gas_fee_cap"]], 10)
		blobFeeCap := uint256.MustFromBig(blobFeeCapBig)

		rBig, _ := new(big.Int).SetString(row[colIndex["r"]], 10)
		r := uint256.MustFromBig(rBig)

		sBig, _ := new(big.Int).SetString(row[colIndex["s"]], 10)
		s := uint256.MustFromBig(sBig)

		vBig, _ := new(big.Int).SetString(row[colIndex["v"]], 10)
		v := uint256.MustFromBig(vBig)

		tipCapBig, _ := new(big.Int).SetString(row[colIndex["gas_tip_cap"]], 10)
		tipCapStart := tipCapBig.Int64()

		var blobHashes []common.Hash
		for _, bh := range strings.Split(row[colIndex["blob_versioned_hashes"]], ";") {
			if bh != "" {
				blobHashes = append(blobHashes, common.HexToHash(bh))
			}
		}

		// Verify original hash
		blobTx := &types.BlobTx{
			ChainID:     chainID,
			Nonce:       nonce,
			GasTipCap:   uint256.MustFromBig(tipCapBig),
			GasFeeCap:   gasFeeCap,
			Gas:         gasLimit,
			To:          to,
			Value:       value,
			Data:        data,
			AccessList:  types.AccessList{},
			BlobFeeCap:  blobFeeCap,
			BlobHashes:  blobHashes,
			V:           v,
			R:           r,
			S:           s,
		}

		initialHash := prefixedRlpHash(0x03, blobTx)
		if initialHash.Hex() == expectedHash {
			fmt.Printf("✅ Match for tx %d: %s\n", i, initialHash.Hex())
		} else {
			fmt.Printf("❌ MISMATCH for tx %d\nExpected: %s\nGot:      %s\n\n", i, expectedHash, initialHash.Hex())
		}

		// Suffix mining loop
		found := make(map[byte]string)
		start := time.Now()
		for j := tipCapStart; ; j++ {
			blobTx.GasTipCap = uint256.NewInt(uint64(j))
			h := prefixedRlpHash(0x03, blobTx)
			suffix := h[31] & 0x0F
			if _, ok := found[suffix]; !ok {
				found[suffix] = h.Hex()
			}
			if len(found) == 16 {
				break
			}
		}
		durationMicros := time.Since(start).Microseconds()

		// Write row
		rowOut := []string{expectedHash}
		for k := 0; k < 16; k++ {
			rowOut = append(rowOut, found[byte(k)])
		}
		rowOut = append(rowOut, fmt.Sprintf("%d", time.Since(start).Microseconds()))

		writer.Write(rowOut)

		fmt.Printf("✅ All 16 suffixes found for tx %d in %d μs\n", i, durationMicros)
	}

	fmt.Printf("🎉 Done! Total time: %s\n", time.Since(totalStart))
}
