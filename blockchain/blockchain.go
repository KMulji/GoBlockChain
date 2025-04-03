package blockchain

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"runtime"

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

func dbExists() bool {
	if _, err := os.Stat("blockchain.db"); os.IsNotExist(err) {
		return false
	}
	return true
}

func ContinueBlockChain(address string) *BlockChain {
	if dbExists() == false {
		fmt.Println("No Existing blockchainf found please create one")
		runtime.Goexit()
	}

	var lastHash []byte

	db, err := bbolt.Open("blockchain.db", 0600, nil)

	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("blocks"))

		if b != nil {
			lastHash = b.Get([]byte("l"))
		}
		return nil
	})

	return &BlockChain{lastHash, db}
}

func NewBlockChain(address string) *BlockChain {
	var lastHash []byte

	if dbExists() {
		fmt.Println("Blockchain already exists")
		runtime.Goexit()
	}
	db, err := bbolt.Open("blockchain.db", 0600, nil)

	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("blocks"))
		if b == nil {
			cbtx := CoinBaseTx(address, "my first chain is hot")
			genesis := NewGenesisBlock(cbtx)

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
func (bc *BlockChain) AddBlock(transactions []*Transaction) {
	var lastHash []byte

	err := bc.Database.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("blocks"))
		lastHash = b.Get([]byte("l"))

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	newBlock := NewBlock(transactions, lastHash)

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
	var block *Block = &Block{}

	err := iter.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("blocks"))
		encoded := b.Get(iter.currentHash)
		block.Deserialize(encoded)

		return nil

	})
	if err != nil {
		log.Panic(err)
	}
	iter.currentHash = block.PrevBlockHash

	return block
}

func (chain *BlockChain) FindUnspentTransactions(address string) []Transaction {
	var unspentTxs []Transaction

	spentTXOs := make(map[string][]int)

	iter := chain.Iterator()

	for {
		block := iter.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Vout {
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}
				if out.CanBeUnlocked(address) {
					unspentTxs = append(unspentTxs, *tx)
				}
			}
			if tx.IsCoinBase() == false {
				for _, in := range tx.Vin {
					if in.CanUnlock(address) {
						inTxID := hex.EncodeToString(in.Txid)
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
					}
				}
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
	return unspentTxs
}

func (chain *BlockChain) FindUTXO(address string) []TXOutput {
	var UTXOs []TXOutput
	unspentTransactions := chain.FindUnspentTransactions(address)

	for _, tx := range unspentTransactions {
		for _, out := range tx.Vout {
			if out.CanBeUnlocked(address) {
				UTXOs = append(UTXOs, out)
			}
		}
	}
	return UTXOs
}

func (chain *BlockChain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	unspentOuts := make(map[string][]int)
	unspentTxs := chain.FindUnspentTransactions(address)
	accumulated := 0

Work:
	for _, tx := range unspentTxs {
		txID := hex.EncodeToString(tx.ID)

		for outIdx, out := range tx.Vout {
			if out.CanBeUnlocked(address) && accumulated < amount {
				accumulated += out.Value
				unspentOuts[txID] = append(unspentOuts[txID], outIdx)

				if accumulated >= amount {
					break Work
				}
			}
		}
	}

	return accumulated, unspentOuts
}
