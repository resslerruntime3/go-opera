package gossip

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/lachesis/genesis"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"math/big"
	"sync"
	"testing"
)

func TestGetGenesisBlock(t *testing.T) {
	assertar := assert.New(t)

	store := NewMemStore()

	g := genesis.FakeGenesis(5)
	addrWithCode := g.Alloc.Addresses()[0]
	accountWithCode := g.Alloc[addrWithCode]
	accountWithCode.Code = []byte{1, 2, 3}
	accountWithCode.Storage = make(map[common.Hash]common.Hash)
	accountWithCode.Storage[common.Hash{}] = common.BytesToHash(common.Big1.Bytes())
	g.Alloc[addrWithCode] = accountWithCode

	genesisHash, stateHash, err := store.ApplyGenesis(&g)
	assertar.NoError(err)

	assertar.NotEqual(common.Hash{}, genesisHash)
	assertar.NotEqual(common.Hash{}, stateHash)

	reader := EvmStateReader{
		store:    store,
		engineMu: new(sync.RWMutex),
	}
	genesisBlock := reader.GetBlock(common.Hash(genesisHash), 0)

	assertar.Equal(common.Hash(genesisHash), genesisBlock.Hash)
	assertar.Equal(stateHash, genesisBlock.Root)
	assertar.Equal(g.Time, genesisBlock.Time)
	assertar.Empty(genesisBlock.Transactions)

	statedb, err := reader.StateAt(stateHash)
	assertar.NoError(err)
	for addr, account := range g.Alloc {
		assertar.Equal(account.Balance.String(), statedb.GetBalance(addr).String())
		if addr == addrWithCode {
			assertar.Equal(account.Code, statedb.GetCode(addr))
			assertar.Equal(account.Storage[common.Hash{}], statedb.GetState(addr, common.Hash{}))
		} else {
			assertar.Empty(statedb.GetCode(addr))
			assertar.Equal(common.Hash{}, statedb.GetState(addr, common.Hash{}))
		}
	}
}

func TestGetBlock(t *testing.T) {
	assertar := assert.New(t)

	store := NewMemStore()

	g := genesis.FakeGenesis(5)
	genesisHash, _, err := store.ApplyGenesis(&g)
	assertar.NoError(err)

	txs := types.Transactions{}
	key, err := crypto.GenerateKey()
	assertar.NoError(err)
	for i := 0; i < 6; i++ {
		tx, err := types.SignTx(types.NewTransaction(uint64(i), common.Address{}, big.NewInt(100), 0, big.NewInt(1), nil), types.HomesteadSigner{}, key)
		assertar.NoError(err)
		txs = append(txs, tx)
	}

	event1 := inter.NewEvent()
	event2 := inter.NewEvent()
	event1.Transactions = txs[:1]
	event1.Seq = 1
	event2.Transactions = txs[1:]
	event1.Seq = 2
	block := inter.NewBlock(1, 123, hash.Events{event1.Hash(), event2.Hash()}, genesisHash)
	block.SkippedTxs = []uint{0, 2, 4}

	store.SetEvent(event1)
	store.SetEvent(event2)
	store.SetBlock(block)

	reader := EvmStateReader{
		store:    store,
		engineMu: new(sync.RWMutex),
	}
	evm_block := reader.GetDagBlock(block.Hash(), block.Index)

	assertar.Equal(uint64(block.Index), evm_block.Number.Uint64())
	assertar.Equal(common.Hash(block.Hash()), evm_block.Hash)
	assertar.Equal(common.Hash(genesisHash), evm_block.ParentHash)
	assertar.Equal(block.Time, evm_block.Time)
	assertar.Equal(len(txs)-len(block.SkippedTxs), evm_block.Transactions.Len())
	assertar.Equal(txs[1].Hash(), evm_block.Transactions[0].Hash())
	assertar.Equal(txs[3].Hash(), evm_block.Transactions[1].Hash())
	assertar.Equal(txs[5].Hash(), evm_block.Transactions[2].Hash())
}
