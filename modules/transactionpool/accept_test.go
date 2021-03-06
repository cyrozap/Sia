package transactionpool

import (
	"testing"

	"github.com/NebulousLabs/Sia/modules"
	"github.com/NebulousLabs/Sia/types"
)

// addSiacoinTransactionToPool creates a transaction with siacoin outputs and
// adds them to the pool, returning the transaction.
func (tpt *tpoolTester) addSiacoinTransactionToPool() (txn types.Transaction) {
	// spendCoins will automatically add transaction(s) to the transaction pool.
	// They will contain siacoin output(s).
	txn, err := tpt.spendCoins(types.NewCurrency64(1), types.ZeroUnlockHash)
	if err != nil {
		tpt.t.Fatal(err)
	}

	return
}

// addDependentSiacoinTransactionToPool adds a transaction to the pool with a
// siacoin output, and then adds a second transaction to the pool that requires
// the unconfirmed siacoin output.
func (tpt *tpoolTester) addDependentSiacoinTransactionToPool() (firstTxn, dependentTxn types.Transaction) {
	// Get an address to receive coins.
	addr, _, err := tpt.wallet.CoinAddress(false) // false means hide the address from the user; doesn't matter for test.
	if err != nil {
		tpt.t.Fatal(err)
	}

	// spendCoins will automatically add transaction(s) to the transaction
	// pool. They will contain siacoin output(s). We send all of our coins to
	// ourself to guarantee that the next transaction will depend on an
	// existing unconfirmed transaction.
	balance := tpt.wallet.Balance(false)
	firstTxn, err = tpt.spendCoins(balance, addr)
	if err != nil {
		tpt.t.Fatal(err)
	}

	// Send the full balance to ourselves again. The second transaction will
	// necesarily require the first transaction as a dependency, since we're
	// sending all of the coins again.
	dependentTxn, err = tpt.spendCoins(balance, addr)
	if err != nil {
		tpt.t.Fatal(err)
	}

	return
}

// TestAddSiacoinTransactionToPool creates a tpoolTester and uses it to call
// addSiacoinTransactionToPool.
func TestAddSiacoinTransactionToPool(t *testing.T) {
	tpt := newTpoolTester("TestAddSiacoinTransactionToPool", t)
	tpt.addSiacoinTransactionToPool()
}

// TestAddDependentSiacoinTransactionToPool creates a tpoolTester and uses it
// to cal addDependentSiacoinTransactionToPool.
func TestAddDependentSiacoinTransactionToPool(t *testing.T) {
	tpt := newTpoolTester("TestAddDependentSiacoinTransactionToPool", t)
	tpt.addDependentSiacoinTransactionToPool()
}

// TestDuplicateTransaction checks that a duplicate transaction error is
// triggered when duplicate transactions are added to the transaction pool.
// This test won't be needed after the duplication prevention mechanism is
// removed, and that will be removed after fees are required in all
// transactions submitted to the pool.
func TestDuplicateTransaction(t *testing.T) {
	tpt := newTpoolTester("TestDuplicateTransaction", t)
	txn := tpt.addSiacoinTransactionToPool()
	err := tpt.tpool.AcceptTransaction(txn)
	if err != modules.ErrTransactionPoolDuplicate {
		t.Fatal("expecting ErrDuplicate got:", err)
	}
}
