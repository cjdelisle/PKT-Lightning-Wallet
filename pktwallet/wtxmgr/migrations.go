package wtxmgr

import (
	"github.com/pkt-cash/pktd/blockchain"
	"github.com/pkt-cash/pktd/btcutil/er"
	"github.com/pkt-cash/pktd/chaincfg"
	"github.com/pkt-cash/pktd/pktlog/log"
	"github.com/pkt-cash/pktd/pktwallet/walletdb"
	"github.com/pkt-cash/pktd/pktwallet/walletdb/migration"
	"github.com/pkt-cash/pktd/pktwallet/wtxmgr/dbstructs"
	"github.com/pkt-cash/pktd/pktwallet/wtxmgr/unspent"
	"github.com/pkt-cash/pktd/txscript"
)

// versions is a list of the different database versions. The last entry should
// reflect the latest database state. If the database happens to be at a version
// number lower than the latest, migrations will be performed in order to catch
// it up.
var versions = []migration.Version{
	{
		Number:    1,
		Migration: nil,
	},
	{
		Number:    2,
		Migration: DropTransactionHistory,
	},
	{
		Number:    3,
		Migration: ExtendUnspent,
	},
}

// getLatestVersion returns the version number of the latest database version.
func getLatestVersion() uint32 {
	return versions[len(versions)-1].Number
}

// MigrationManager is an implementation of the migration.Manager interface that
// will be used to handle migrations for the address manager. It exposes the
// necessary parameters required to successfully perform migrations.
type MigrationManager struct {
	ns walletdb.ReadWriteBucket
}

// A compile-time assertion to ensure that MigrationManager implements the
// migration.Manager interface.
var _ migration.Manager = (*MigrationManager)(nil)

// NewMigrationManager creates a new migration manager for the transaction
// manager. The given bucket should reflect the top-level bucket in which all
// of the transaction manager's data is contained within.
func NewMigrationManager(ns walletdb.ReadWriteBucket) *MigrationManager {
	return &MigrationManager{ns: ns}
}

// Name returns the name of the service we'll be attempting to upgrade.
//
// NOTE: This method is part of the migration.Manager interface.
func (m *MigrationManager) Name() string {
	return "wallet transaction manager"
}

// Namespace returns the top-level bucket of the service.
//
// NOTE: This method is part of the migration.Manager interface.
func (m *MigrationManager) Namespace() walletdb.ReadWriteBucket {
	return m.ns
}

// CurrentVersion returns the current version of the service's database.
//
// NOTE: This method is part of the migration.Manager interface.
func (m *MigrationManager) CurrentVersion(ns walletdb.ReadBucket) (uint32, er.R) {
	if ns == nil {
		ns = m.ns
	}
	return fetchVersion(ns)
}

// SetVersion sets the version of the service's database.
//
// NOTE: This method is part of the migration.Manager interface.
func (m *MigrationManager) SetVersion(ns walletdb.ReadWriteBucket,
	version uint32) er.R {

	if ns == nil {
		ns = m.ns
	}
	return putVersion(ns, version)
}

// Versions returns all of the available database versions of the service.
//
// NOTE: This method is part of the migration.Manager interface.
func (m *MigrationManager) Versions() []migration.Version {
	return versions
}

// DropTransactionHistory is a migration that attempts to recreate the
// transaction store with a clean state.
func DropTransactionHistory(ns walletdb.ReadWriteBucket) er.R {
	log.Info("Dropping wallet transaction history")

	// To drop the store's transaction history, we'll need to remove all of
	// the relevant descendant buckets and key/value pairs.
	if err := deleteBuckets(ns); err != nil {
		return err
	}

	// With everything removed, we'll now recreate our buckets.
	if err := createBuckets(ns); err != nil {
		return err
	}

	return nil
}

func ExtendUnspent(ns walletdb.ReadWriteBucket) er.R {
	log.Info("Adding extended data to unspent table")
	count := 0
	err := unspent.ExtendUnspents(ns, func(unspent *dbstructs.Unspent) er.R {
		txr, err := fetchTxRecord(ns, &unspent.OutPoint.Hash, &unspent.Block)
		if err != nil {
			return err
		}
		if int(unspent.OutPoint.Index) >= len(txr.MsgTx.TxOut) {
			return er.New("unspent.OutPoint.Index not in MsgTx")
		}
		op := txr.MsgTx.TxOut[unspent.OutPoint.Index]
		// TODO(cjd): Incompatible w/ BTC
		unspent.Address = txscript.PkScriptToAddress(op.PkScript, &chaincfg.PktMainNetParams).String()
		unspent.Value = op.Value
		unspent.FromCoinBase = blockchain.IsCoinBaseTx(&txr.MsgTx)
		unspent.PkScript = op.PkScript
		count += 1
		return nil
	})
	if err != nil {
		return err
	}
	log.Infof("Adding extended data to unspent table - done, [%d] entries", count)
	return nil
}
