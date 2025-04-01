package blockchain

import (
	"bytes"
	"encoding/gob"
	"log"
	"time"
)

type Block struct {
	Timestamp     int64
	Data          []byte
	PrevBlockHash []byte
	Hash          []byte
	nonce         int
}

func NewBlock(data string, prevBlockHash []byte) *Block {
	block := &Block{time.Now().Unix(), []byte(data), prevBlockHash, []byte{}, 0}
	pow := NewProofOfWork(block)

	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.nonce = nonce

	return block
}
func NewGenesisBlock() *Block {
	return NewBlock("This is Genesis", []byte{})
}

func (b *Block) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(b)
	if err != nil {
		log.Panic(err)
	}
	return result.Bytes()
}

func Deserialize(arr []byte) *Block {
	var block Block
	decoder := gob.NewDecoder(bytes.NewReader(arr))
	err := decoder.Decode(block)
	if err != nil {
		log.Panic(err)
	}

	return &block
}
