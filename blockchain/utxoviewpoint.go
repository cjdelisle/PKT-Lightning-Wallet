// Copyright (c) 2015-2016 The btcsuite developers
// Copyright (c) 2019 Caleb James DeLisle
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockchain

import (
	"fmt"

	"github.com/pkt-cash/pktd/btcutil/er"

	"github.com/pkt-cash/pktd/btcutil"
	"github.com/pkt-cash/pktd/chaincfg/chainhash"
	"github.com/pkt-cash/pktd/chaincfg/globalcfg"
	"github.com/pkt-cash/pktd/database"
	"github.com/pkt-cash/pktd/txscript"
	"github.com/pkt-cash/pktd/wire"
)

// txoFlags is a bitmask defining additional information and state for a
// transaction output in a utxo view.
type txoFlags uint8

const (
	// tfCoinBase indicates that a txout was contained in a coinbase tx.
	tfCoinBase txoFlags = 1 << iota

	// tfSpent indicates that a txout is spent.
	tfSpent

	// tfModified indicates that a txout has been modified since it was
	// loaded.
	tfModified

	// tfNetworkSteward indicates that a txout has been identified as a payment
	// of the tax to the network steward, and thus will be subject to expiration
	// rules.
	tfNetworkSteward
)

// UtxoEntry houses details about an individual transaction output in a utxo
// view such as whether or not it was contained in a coinbase tx, the height of
// the block that contains the tx, whether or not it is spent, its public key
// script, and how much it pays.
type UtxoEntry struct {
	// NOTE: Additions, deletions, or modifications to the order of the
	// definitions in this struct should not be changed without considering
	// how it affects alignment on 64-bit platforms.  The current order is
	// specifically crafted to result in minimal padding.  There will be a
	// lot of these in memory, so a few extra bytes of padding adds up.

	amount      int64
	pkScript    []byte // The public key script for the output.
	blockHeight int32  // Height of block containing tx.

	// packedFlags contains additional info about output such as whether it
	// is a coinbase, whether it is spent, and whether it has been modified
	// since it was loaded.  This approach is used in order to reduce memory
	// usage since there will be a lot of these in memory.
	packedFlags txoFlags
}

// isNetworkSteward returns whether or not this output has been identified as
// a payment of the Network Steward mining tax.
func (entry *UtxoEntry) isNetworkSteward() bool {
	return entry.packedFlags&tfNetworkSteward == tfNetworkSteward
}

// isModified returns whether or not the output has been modified since it was
// loaded.
func (entry *UtxoEntry) isModified() bool {
	return entry.packedFlags&tfModified == tfModified
}

// IsCoinBase returns whether or not the output was contained in a coinbase
// transaction.
func (entry *UtxoEntry) IsCoinBase() bool {
	return entry.packedFlags&tfCoinBase == tfCoinBase
}

// BlockHeight returns the height of the block containing the output.
func (entry *UtxoEntry) BlockHeight() int32 {
	return entry.blockHeight
}

// IsSpent returns whether or not the output has been spent based upon the
// current state of the unspent transaction output view it was obtained from.
func (entry *UtxoEntry) IsSpent() bool {
	return entry.packedFlags&tfSpent == tfSpent
}

// Spend marks the output as spent.  Spending an output that is already spent
// has no effect.
func (entry *UtxoEntry) Spend() {
	// Nothing to do if the output is already spent.
	if entry.IsSpent() {
		return
	}

	// Mark the output as spent and modified.
	entry.packedFlags |= tfSpent | tfModified
}

// Amount returns the amount of the output.
func (entry *UtxoEntry) Amount() int64 {
	return entry.amount
}

// PkScript returns the public key script for the output.
func (entry *UtxoEntry) PkScript() []byte {
	return entry.pkScript
}

// Clone returns a shallow copy of the utxo entry.
func (entry *UtxoEntry) Clone() *UtxoEntry {
	if entry == nil {
		return nil
	}

	return &UtxoEntry{
		amount:      entry.amount,
		pkScript:    entry.pkScript,
		blockHeight: entry.blockHeight,
		packedFlags: entry.packedFlags,
	}
}

// UtxoViewpoint represents a view into the set of unspent transaction outputs
// from a specific point of view in the chain.  For example, it could be for
// the end of the main chain, some point in the history of the main chain, or
// down a side chain.
//
// The unspent outputs are needed by other transactions for things such as
// script validation and double spend prevention.
type UtxoViewpoint struct {
	entries  map[wire.OutPoint]*UtxoEntry
	bestHash chainhash.Hash
}

// BestHash returns the hash of the best block in the chain the view currently
// respresents.
func (view *UtxoViewpoint) BestHash() *chainhash.Hash {
	return &view.bestHash
}

// SetBestHash sets the hash of the best block in the chain the view currently
// respresents.
func (view *UtxoViewpoint) SetBestHash(hash *chainhash.Hash) {
	view.bestHash = *hash
}

// LookupEntry returns information about a given transaction output according to
// the current state of the view.  It will return nil if the passed output does
// not exist in the view or is otherwise not available such as when it has been
// disconnected during a reorg.
func (view *UtxoViewpoint) LookupEntry(outpoint wire.OutPoint) *UtxoEntry {
	return view.entries[outpoint]
}

// addTxOut adds the specified output to the view if it is not provably
// unspendable.  When the view already has an entry for the output, it will be
// marked unspent.  All fields will be updated for existing entries since it's
// possible it has changed during a reorg.
func (view *UtxoViewpoint) addTxOut(outpoint wire.OutPoint, txOut *wire.TxOut, isCoinBase bool, blockHeight int32) {
	// Don't add provably unspendable outputs.
	if txscript.IsUnspendable(txOut.PkScript) {
		return
	}

	// Update existing entries.  All fields are updated because it's
	// possible (although extremely unlikely) that the existing entry is
	// being replaced by a different transaction with the same hash.  This
	// is allowed so long as the previous transaction is fully spent.
	entry := view.LookupEntry(outpoint)
	if entry == nil {
		entry = new(UtxoEntry)
		view.entries[outpoint] = entry
	}

	entry.amount = txOut.Value
	entry.pkScript = txOut.PkScript
	entry.blockHeight = blockHeight
	entry.packedFlags = tfModified
	if isCoinBase {
		entry.packedFlags |= tfCoinBase
	}
	dirtyGuessIsNetworkStewardPayment(entry)
}

// AddTxOut adds the specified output of the passed transaction to the view if
// it exists and is not provably unspendable.  When the view already has an
// entry for the output, it will be marked unspent.  All fields will be updated
// for existing entries since it's possible it has changed during a reorg.
func (view *UtxoViewpoint) AddTxOut(tx *btcutil.Tx, txOutIdx uint32, blockHeight int32) {
	// Can't add an output for an out of bounds index.
	if txOutIdx >= uint32(len(tx.MsgTx().TxOut)) {
		return
	}

	// Update existing entries.  All fields are updated because it's
	// possible (although extremely unlikely) that the existing entry is
	// being replaced by a different transaction with the same hash.  This
	// is allowed so long as the previous transaction is fully spent.
	prevOut := wire.OutPoint{Hash: *tx.Hash(), Index: txOutIdx}
	txOut := tx.MsgTx().TxOut[txOutIdx]
	view.addTxOut(prevOut, txOut, IsCoinBase(tx), blockHeight)
}

// AddTxOuts adds all outputs in the passed transaction which are not provably
// unspendable to the view.  When the view already has entries for any of the
// outputs, they are simply marked unspent.  All fields will be updated for
// existing entries since it's possible it has changed during a reorg.
func (view *UtxoViewpoint) AddTxOuts(tx *btcutil.Tx, blockHeight int32) {
	// Loop all of the transaction outputs and add those which are not
	// provably unspendable.
	isCoinBase := IsCoinBase(tx)
	prevOut := wire.OutPoint{Hash: *tx.Hash()}
	for txOutIdx, txOut := range tx.MsgTx().TxOut {
		// Update existing entries.  All fields are updated because it's
		// possible (although extremely unlikely) that the existing
		// entry is being replaced by a different transaction with the
		// same hash.  This is allowed so long as the previous
		// transaction is fully spent.
		prevOut.Index = uint32(txOutIdx)
		view.addTxOut(prevOut, txOut, isCoinBase, blockHeight)
	}
}

// connectTransaction updates the view by adding all new utxos created by the
// passed transaction and marking all utxos that the transactions spend as
// spent.  In addition, when the 'stxos' argument is not nil, it will be updated
// to append an entry for each spent txout.  An error will be returned if the
// view does not contain the required utxos.
func (view *UtxoViewpoint) connectTransaction(tx *btcutil.Tx, blockHeight int32, stxos *[]SpentTxOut) er.R {
	// Coinbase transactions don't have any inputs to spend.
	if IsCoinBase(tx) {
		// Add the transaction's outputs as available utxos.
		view.AddTxOuts(tx, blockHeight)
		return nil
	}

	// Spend the referenced utxos by marking them spent in the view and,
	// if a slice was provided for the spent txout details, append an entry
	// to it.
	for _, txIn := range tx.MsgTx().TxIn {
		// Ensure the referenced utxo exists in the view.  This should
		// never happen unless there is a bug is introduced in the code.
		entry := view.entries[txIn.PreviousOutPoint]
		if entry == nil {
			return AssertError(fmt.Sprintf("view missing input %v",
				txIn.PreviousOutPoint))
		}

		// Only create the stxo details if requested.
		if stxos != nil {
			// Populate the stxo details using the utxo entry.
			var stxo = SpentTxOut{
				Amount:     entry.Amount(),
				PkScript:   entry.PkScript(),
				Height:     entry.BlockHeight(),
				IsCoinBase: entry.IsCoinBase(),
			}
			*stxos = append(*stxos, stxo)
		}

		// Mark the entry as spent.  This is not done until after the
		// relevant details have been accessed since spending it might
		// clear the fields from memory in the future.
		entry.Spend()
	}

	// Add the transaction's outputs as available utxos.
	view.AddTxOuts(tx, blockHeight)
	return nil
}

// connectTransactions updates the view by adding all new utxos created by all
// of the transactions in the passed block, marking all utxos the transactions
// spend as spent, and setting the best hash for the view to the passed block.
// In addition, when the 'stxos' argument is not nil, it will be updated to
// append an entry for each spent txout.
func (view *UtxoViewpoint) connectTransactions(block *btcutil.Block, stxos *[]SpentTxOut) er.R {
	for _, tx := range block.Transactions() {
		err := view.connectTransaction(tx, block.Height(), stxos)
		if err != nil {
			return err
		}
	}

	// Update the best hash for view to include this block since all of its
	// transactions have been connected.
	view.SetBestHash(block.Hash())
	return nil
}

// dirtyGuessIsNetworkStewardPayment tries to determine whether this utxo is a
// payment to the network steward. Ideally the UTXO would carry this flag into
// the spent txo database, but in the interest of compactness, that datastructure
// has no extensibility, so rather than rewrite that entire codebase, we'll simply
// treat any coinbase txo which pays exactly the amount of the network steward tax,
// as a network steward tax.
// However, we need to be sure that we apply this same rule all of the time, so that
// the consensus rules do not spontanously change after a transaction is rolled back.
func dirtyGuessIsNetworkStewardPayment(entry *UtxoEntry) {
	if !globalcfg.HasNetworkSteward() {
		return
	}
	if entry.packedFlags|tfCoinBase != tfCoinBase {
		return
	}
	subsidy := pktCalcBlockSubsidy(pktPeriodForBlock(entry.blockHeight))
	tax := PktCalcNetworkStewardPayout(subsidy)
	if entry.amount == tax {
		entry.packedFlags |= tfNetworkSteward
	}
}

// RemoveEntry removes the given transaction output from the current state of
// the view.  It will have no effect if the passed output does not exist in the
// view.
func (view *UtxoViewpoint) RemoveEntry(outpoint wire.OutPoint) {
	delete(view.entries, outpoint)
}

// Entries returns the underlying map that stores of all the utxo entries.
func (view *UtxoViewpoint) Entries() map[wire.OutPoint]*UtxoEntry {
	return view.entries
}

// fetchUtxosMain fetches unspent transaction output data about the provided
// set of outpoints from the point of view of the end of the main chain at the
// time of the call.
//
// Upon completion of this function, the view will contain an entry for each
// requested outpoint.  Spent outputs, or those which otherwise don't exist,
// will result in a nil entry in the view.
func (view *UtxoViewpoint) fetchUtxosMain(db database.DB, outpoints map[wire.OutPoint]struct{}) er.R {
	// Nothing to do if there are no requested outputs.
	if len(outpoints) == 0 {
		return nil
	}

	// Load the requested set of unspent transaction outputs from the point
	// of view of the end of the main chain.
	//
	// NOTE: Missing entries are not considered an error here and instead
	// will result in nil entries in the view.  This is intentionally done
	// so other code can use the presence of an entry in the store as a way
	// to unnecessarily avoid attempting to reload it from the database.
	return db.View(func(dbTx database.Tx) er.R {
		for outpoint := range outpoints {
			entry, err := dbFetchUtxoEntry(dbTx, outpoint)
			if err != nil {
				return err
			}

			view.entries[outpoint] = entry
		}

		return nil
	})
}

// fetchUtxos loads the unspent transaction outputs for the provided set of
// outputs into the view from the database as needed unless they already exist
// in the view in which case they are ignored.
func (view *UtxoViewpoint) fetchUtxos(db database.DB, outpoints map[wire.OutPoint]struct{}) er.R {
	// Nothing to do if there are no requested outputs.
	if len(outpoints) == 0 {
		return nil
	}

	// Filter entries that are already in the view.
	neededSet := make(map[wire.OutPoint]struct{})
	for outpoint := range outpoints {
		// Already loaded into the current view.
		if _, ok := view.entries[outpoint]; ok {
			continue
		}

		neededSet[outpoint] = struct{}{}
	}

	// Request the input utxos from the database.
	return view.fetchUtxosMain(db, neededSet)
}

// fetchInputUtxos loads the unspent transaction outputs for the inputs
// referenced by the transactions in the given block into the view from the
// database as needed.  In particular, referenced entries that are earlier in
// the block are added to the view and entries that are already in the view are
// not modified.
func (view *UtxoViewpoint) fetchInputUtxos(db database.DB, block *btcutil.Block) er.R {
	// Build a map of in-flight transactions because some of the inputs in
	// this block could be referencing other transactions earlier in this
	// block which are not yet in the chain.
	txInFlight := map[chainhash.Hash]int{}
	transactions := block.Transactions()
	for i, tx := range transactions {
		txInFlight[*tx.Hash()] = i
	}

	// Loop through all of the transaction inputs (except for the coinbase
	// which has no inputs) collecting them into sets of what is needed and
	// what is already known (in-flight).
	neededSet := make(map[wire.OutPoint]struct{})
	for i, tx := range transactions[1:] {
		for _, txIn := range tx.MsgTx().TxIn {
			// It is acceptable for a transaction input to reference
			// the output of another transaction in this block only
			// if the referenced transaction comes before the
			// current one in this block.  Add the outputs of the
			// referenced transaction as available utxos when this
			// is the case.  Otherwise, the utxo details are still
			// needed.
			//
			// NOTE: The >= is correct here because i is one less
			// than the actual position of the transaction within
			// the block due to skipping the coinbase.
			originHash := &txIn.PreviousOutPoint.Hash
			if inFlightIndex, ok := txInFlight[*originHash]; ok &&
				i >= inFlightIndex {

				originTx := transactions[inFlightIndex]
				view.AddTxOuts(originTx, block.Height())
				continue
			}

			// Don't request entries that are already in the view
			// from the database.
			if _, ok := view.entries[txIn.PreviousOutPoint]; ok {
				continue
			}

			neededSet[txIn.PreviousOutPoint] = struct{}{}
		}
	}

	// Request the input utxos from the database.
	return view.fetchUtxosMain(db, neededSet)
}

// NewUtxoViewpoint returns a new empty unspent transaction output view.
func NewUtxoViewpoint() *UtxoViewpoint {
	return &UtxoViewpoint{
		entries: make(map[wire.OutPoint]*UtxoEntry),
	}
}

// FetchUtxoView loads unspent transaction outputs for the inputs referenced by
// the passed transaction from the point of view of the end of the main chain.
// It also attempts to fetch the utxos for the outputs of the transaction itself
// so the returned view can be examined for duplicate transactions.
//
// This function is safe for concurrent access however the returned view is NOT.
func (b *BlockChain) FetchUtxoView(tx *btcutil.Tx) (*UtxoViewpoint, er.R) {
	// Create a set of needed outputs based on those referenced by the
	// inputs of the passed transaction and the outputs of the transaction
	// itself.
	neededSet := make(map[wire.OutPoint]struct{})
	prevOut := wire.OutPoint{Hash: *tx.Hash()}
	for txOutIdx := range tx.MsgTx().TxOut {
		prevOut.Index = uint32(txOutIdx)
		neededSet[prevOut] = struct{}{}
	}
	if !IsCoinBase(tx) {
		for _, txIn := range tx.MsgTx().TxIn {
			neededSet[txIn.PreviousOutPoint] = struct{}{}
		}
	}

	// Request the utxos from the point of view of the end of the main
	// chain.
	view := NewUtxoViewpoint()
	b.chainLock.RLock()
	err := view.fetchUtxosMain(b.db, neededSet)
	b.chainLock.RUnlock()
	return view, err
}

// FetchUtxoEntry loads and returns the requested unspent transaction output
// from the point of view of the end of the main chain.
//
// NOTE: Requesting an output for which there is no data will NOT return an
// error.  Instead both the entry and the error will be nil.  This is done to
// allow pruning of spent transaction outputs.  In practice this means the
// caller must check if the returned entry is nil before invoking methods on it.
//
// This function is safe for concurrent access however the returned entry (if
// any) is NOT.
func (b *BlockChain) FetchUtxoEntry(outpoint wire.OutPoint) (*UtxoEntry, er.R) {
	b.chainLock.RLock()
	defer b.chainLock.RUnlock()

	var entry *UtxoEntry
	err := b.db.View(func(dbTx database.Tx) er.R {
		var err er.R
		entry, err = dbFetchUtxoEntry(dbTx, outpoint)
		return err
	})
	if err != nil {
		return nil, err
	}

	return entry, nil
}
