package host

import (
	"crypto/rand"
	"testing"

	"github.com/NebulousLabs/Sia/consensus"
	"github.com/NebulousLabs/Sia/crypto"
)

// AddObligation put a file contract into the blockchain and a corresponding
// file obligation into the host strucure. The payout will go to the host, and
// the host will automatically submit a proof of storage at 'start'.
func (ht *HostTester) AddObligation(data []byte, start, expiration consensus.BlockHeight) (fcid consensus.FileContractID) {
	// Allocate the file that the host is required to store.
	filesize := uint64(len(data))
	file, path, err := ht.allocate(filesize)
	if err != nil {
		ht.Fatal(err)
	}
	_, err = file.Write(data)
	if err != nil {
		ht.Fatal(err)
	}
	file.Seek(0, 0)

	// Create, finance, and mine a transaction with a file contract in it using
	// the data's merkle root.
	input, value := ht.FindSpendableSiacoinInput()
	txn := ht.AddSiacoinInputToTransaction(consensus.Transaction{}, input)
	merkleRoot, err := crypto.ReaderMerkleRoot(file, filesize)
	if err != nil {
		ht.Fatal(err)
	}
	file.Close()
	fc := consensus.FileContract{
		FileSize:       filesize,
		FileMerkleRoot: merkleRoot,
		Start:          start,
		Expiration:     expiration,
		Payout:         value,
		ValidProofOutputs: []consensus.SiacoinOutput{
			consensus.SiacoinOutput{Value: value, UnlockHash: ht.Host.UnlockHash},
		},
		MissedProofOutputs: []consensus.SiacoinOutput{
			consensus.SiacoinOutput{Value: value, UnlockHash: consensus.ZeroUnlockHash},
		},
	}
	fc.ValidProofOutputs[0].Value = fc.ValidProofOutputs[0].Value.Sub(fc.Tax())
	txn.FileContracts = append(txn.FileContracts, fc)
	ht.MineAndSubmitCurrentBlock([]consensus.Transaction{txn})

	// Add the obligation for the file to the host.
	fcid = txn.FileContractID(0)
	co := contractObligation{
		id:           fcid,
		fileContract: fc,
		path:         path,
	}
	ht.mu.Lock()
	ht.obligationsByHeight[ht.Height()+1] = append(ht.obligationsByHeight[ht.Height()+1], co)
	ht.obligationsByID[fcid] = co
	ht.mu.Unlock()
	return
}

// testObligation adds a file obligation to the host's set of obligations, then
// mies blocks and updates the host, causing the host to submit a storage
// proof. Then the storage proof is mined and a check is made to see that the
// host gets the payout.
func (ht *HostTester) testObligation() {
	// Add an obligation to the host.
	data := make([]byte, 4e3)
	rand.Read(data)
	ht.AddObligation(data, ht.State.Height()+2, ht.State.Height()+3)

	// Get the balance before the proof is submitted, mine enough blocks to
	// have the proof submitted, then check that the balance experienced an
	// increase.
	startingBalance := ht.wallet.Balance(true)
	for i := 0; i < 3+consensus.MaturityDelay; i++ {
		ht.Host.update()
		tSet, err := ht.Host.tpool.TransactionSet()
		if err != nil {
			ht.Error(err)
		}
		ht.MineAndSubmitCurrentBlock(tSet)
	}
	if startingBalance.Cmp(ht.wallet.Balance(true)) >= 0 {
		ht.Error("balance did not increase after submitting and maturing a storage proof")
	}
}

// TestObligation creates a host tester and calls testObligation.
func TestObligation(t *testing.T) {
	ht := CreateHostTester(t)
	ht.testObligation()
}
