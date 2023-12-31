package wallet

import (
	"encoding/hex"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/pkt-cash/pktd/btcutil/hdkeychain"
	"github.com/pkt-cash/pktd/chaincfg"
	"github.com/pkt-cash/pktd/pktwallet/chainiface"
)

// testWallet creates a test wallet and unlocks it.
func testWallet(t *testing.T) (*Wallet, func()) {
	// Set up a wallet.
	dir, errr := ioutil.TempDir("", "test_wallet")
	if errr != nil {
		t.Fatalf("Failed to create db dir: %v", errr)
	}

	cleanup := func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatalf("could not cleanup test: %v", err)
		}
	}

	seed, err := hdkeychain.GenerateSeed(hdkeychain.MinSeedBytes)
	if err != nil {
		t.Fatalf("unable to create seed: %v", err)
	}

	pubPass := []byte("hello")
	privPass := []byte("world")

	loader := NewLoader(
		&chaincfg.TestNet3Params, dir, "wallet.db", true, 250,
	)
	w, err := loader.CreateNewWallet(pubPass, privPass,
		[]byte(hex.EncodeToString(seed)), time.Now(), nil, nil)
	if err != nil {
		t.Fatalf("unable to create wallet: %v", err)
	}
	chainClient := &chainiface.Mock{}
	w.chainClient = chainClient
	if err := w.Unlock(privPass, time.After(10*time.Minute)); err != nil {
		t.Fatalf("unable to unlock wallet: %v", err)
	}

	return w, cleanup
}
