package walletrpc

import (
	"math"
	"time"

	"github.com/pkt-cash/pktd/btcutil/er"
	"github.com/pkt-cash/pktd/btcutil/psbt"
	"github.com/pkt-cash/pktd/lnd/lnwallet"
	"github.com/pkt-cash/pktd/pktlog/log"
	"github.com/pkt-cash/pktd/pktwallet/wtxmgr"
	"github.com/pkt-cash/pktd/wire"
)

const (
	defaultMinConf = 1
	defaultMaxConf = math.MaxInt32
)

// utxoLock is a type that contains an outpoint of an UTXO and its lock lease
// information.
type utxoLock struct {
	lockID     wtxmgr.LockID
	outpoint   wire.OutPoint
	expiration time.Time
}

// verifyInputsUnspent checks that all inputs are contained in the list of
// known, non-locked UTXOs given.
func verifyInputsUnspent(inputs []*wire.TxIn, utxos []*lnwallet.Utxo) er.R {
	// TODO(guggero): Pass in UTXOs as a map to make lookup more efficient.
	for idx, txIn := range inputs {
		found := false
		for _, u := range utxos {
			if u.OutPoint == txIn.PreviousOutPoint {
				found = true
				break
			}
		}

		if !found {
			return er.Errorf("input %d not found in list of non-"+
				"locked UTXO", idx)
		}
	}

	return nil
}

// lockInputs requests a lock lease for all inputs specified in a PSBT packet
// by using the internal, static lock ID of lnd's wallet.
func lockInputs(w lnwallet.WalletController, packet *psbt.Packet) ([]*utxoLock, er.R) {

	locks := make([]*utxoLock, len(packet.UnsignedTx.TxIn))
	for idx, rawInput := range packet.UnsignedTx.TxIn {
		lock := &utxoLock{
			lockID:   LndInternalLockID,
			outpoint: rawInput.PreviousOutPoint,
		}

		expiration, err := w.LeaseOutput(lock.lockID, lock.outpoint)
		if err != nil {
			// If we run into a problem with locking one output, we
			// should try to unlock those that we successfully
			// locked so far. If that fails as well, there's not
			// much we can do.
			for i := 0; i < idx; i++ {
				op := locks[i].outpoint
				if err := w.ReleaseOutput(
					LndInternalLockID, op,
				); err != nil {

					log.Errorf("could not release the "+
						"lock on %v: %v", op, err)
				}
			}

			return nil, er.Errorf("could not lease a lock on "+
				"UTXO: %v", err)
		}

		lock.expiration = expiration
		locks[idx] = lock
	}

	return locks, nil
}
