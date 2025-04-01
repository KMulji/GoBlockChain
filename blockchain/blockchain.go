package blockchain

import (
	"log"

	"go.etcd.io/bbolt"
)

type BlockChain struct {
	LastHash []byte
	Database *bbolt.DB
}
type BlockChainIterator struct {
	currentHash []byte
	db          *bbolt.DB
}

func NewBlockChain() *BlockChain {
	var lastHash []byte

	db, err := bbolt.Open("blockchain.db", 0600, nil)

	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("blocks"))
		if b == nil {
			genesis := NewGenesisBlock()

			b, err := tx.CreateBucket([]byte("blocks"))
			if err != nil {
				log.Panic(err)
			}
			err = b.Put(genesis.Hash, genesis.Serialize())
			if err != nil {
				log.Panic(err)
			}
			err = b.Put([]byte("l"), genesis.Hash)
			if err != nil {
				log.Panic(err)
			}
			lastHash = genesis.Hash
		} else {
			lastHash = b.Get([]byte("l"))
		}
		if err != nil {
			return err
		}
		return nil
	})
	return &BlockChain{lastHash, db}
}
func (bc *BlockChain) AddBlock(data string) {
	var lastHash []byte

	err := bc.Database.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("blocks"))
		lastHash = b.Get([]byte("l"))

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	newBlock := NewBlock(data, lastHash)

	err = bc.Database.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("blocks"))
		err := b.Put(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			log.Panic(err)
		}
		err = b.Put([]byte("l"), newBlock.Hash)

		bc.LastHash = newBlock.Hash

		return err
	})
	if err != nil {
		log.Panic(err)
	}

}

func (bc *BlockChain) Iterator() *BlockChainIterator {
	return &BlockChainIterator{bc.LastHash, bc.Database}
}

func (iter *BlockChainIterator) Next() *Block {
	var block *Block

	err := iter.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("blocks"))
		encoded := b.Get(iter.currentHash)
		block = Deserialize(encoded)

		return nil

	})
	if err != nil {
		log.Panic(err)
	}
	iter.currentHash = block.PrevBlockHash

	return block
}
