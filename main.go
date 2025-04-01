package main

import (
	"fmt"
	"strconv"

	"github.com/KMulji/GoBlockChain/blockchain"
)

func main() {
	blockChain := blockchain.NewBlockChain()

	blockChain.AddBlock("First")
	blockChain.AddBlock("Second")
	blockChain.AddBlock("Third")

	for _, block := range blockChain.Blocks {
		fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)
		fmt.Println()

		pow := blockchain.NewProofOfWork(block)
		fmt.Printf("pow is %s", strconv.FormatBool(pow.Validate()))
		fmt.Println()

	}
}
