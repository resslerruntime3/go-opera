package genesis

import (
	"math/big"

	"github.com/Fantom-foundation/go-lachesis/crypto"
	"github.com/Fantom-foundation/go-lachesis/inter/pos"
)

// FakeAccounts returns accounts and validators for fakenet
func FakeAccounts(from, count int, balance *big.Int, stake pos.Stake) VAccounts {
	accs := make(Accounts, count)
	validators := pos.NewValidators()

	for i := from; i < from+count; i++ {
		key := crypto.FakeKey(i)
		addr := crypto.PubkeyToAddress(key.PublicKey)
		accs[addr] = Account{
			Balance:    balance,
			PrivateKey: key,
		}
		validators.Set(addr, stake)
	}

	return VAccounts{accs, *validators}
}