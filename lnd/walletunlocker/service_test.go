package walletunlocker_test

import (
	"bytes"
	"context"
	"encoding/hex"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/pkt-cash/pktd/btcutil/er"
	"github.com/pkt-cash/pktd/btcutil/util"
	"github.com/pkt-cash/pktd/chaincfg"
	"github.com/pkt-cash/pktd/lnd/channeldb/kvdb"
	"github.com/pkt-cash/pktd/lnd/lnrpc"
	"github.com/pkt-cash/pktd/lnd/lnwallet"
	"github.com/pkt-cash/pktd/lnd/lnwallet/btcwallet"
	"github.com/pkt-cash/pktd/lnd/macaroons"
	"github.com/pkt-cash/pktd/lnd/walletunlocker"
	"github.com/pkt-cash/pktd/pktlog/log"
	"github.com/pkt-cash/pktd/pktwallet/snacl"
	"github.com/pkt-cash/pktd/pktwallet/waddrmgr"
	"github.com/pkt-cash/pktd/pktwallet/wallet"
	"github.com/pkt-cash/pktd/pktwallet/wallet/seedwords"
	"github.com/stretchr/testify/require"
)

const (
	testWalletFilename string = "wallet.db"
)

var (
	testPassword = []byte("test-password")
	testSeed     = []byte("test-seed-123456789")
	testMac      = []byte("fakemacaroon")

	testNetParams = &chaincfg.MainNetParams

	testRecoveryWindow uint32 = 150

	defaultTestTimeout = 3 * time.Second

	defaultRootKeyIDContext = macaroons.ContextWithRootKeyID(
		context.Background(), macaroons.DefaultRootKeyID,
	)
)

func createTestWallet(t *testing.T, dir string, netParams *chaincfg.Params) {
	createTestWalletWithPw(t, testPassword, testPassword, dir, netParams)
}

func createTestWalletWithPw(t *testing.T, pubPw, privPw []byte, dir string,
	netParams *chaincfg.Params) {

	// Instruct waddrmgr to use the cranked down scrypt parameters when
	// creating new wallet encryption keys.
	fastScrypt := waddrmgr.FastScryptOptions
	keyGen := func(passphrase *[]byte, config *waddrmgr.ScryptOptions) (
		*snacl.SecretKey, er.R) {

		return snacl.NewSecretKey(
			passphrase, fastScrypt.N, fastScrypt.R, fastScrypt.P,
		)
	}
	waddrmgr.SetSecretKeyGen(keyGen)

	// Create a new test wallet that uses fast scrypt as KDF.
	netDir := btcwallet.NetworkDir(dir, netParams)
	loader := wallet.NewLoader(netParams, netDir, testWalletFilename, true, 0)
	_, err := loader.CreateNewWallet(
		pubPw, privPw, []byte(hex.EncodeToString(testSeed)), time.Time{}, nil,
	)
	util.RequireNoErr(t, err)

	realWalletPathname := wallet.WalletDbPath(netDir, testWalletFilename)
	log.Debugf(">>>>> [1] wallet path: %s", realWalletPathname)
	walletFileExists := true
	_, errr := os.Stat(realWalletPathname)
	if err != nil {
		if os.IsNotExist(errr) {
			walletFileExists = false
		} else {
			require.NoError(t, errr)
		}
	}
	log.Debugf(">>>>> [2] after loader.CreateNewWallet() the wallet file exists: %t", walletFileExists)

	err = loader.UnloadWallet()
	util.RequireNoErr(t, err)
}

func createSeedAndMnemonic(t *testing.T, pass []byte) (*seedwords.Seed, string) {

	cipherSeed, err := seedwords.RandomSeed()
	util.RequireNoErr(t, err)

	encipheredSeed := cipherSeed.Encrypt(pass)

	// With the new seed created, we'll convert it into a mnemonic phrase
	// that we'll send over to initialize the wallet.
	mnemonic, err := encipheredSeed.Words("english")
	util.RequireNoErr(t, err)

	return cipherSeed, mnemonic
}

// openOrCreateTestMacStore opens or creates a bbolt DB and then initializes a
// root key storage for that DB and then unlocks it, creating a root key in the
// process.
func openOrCreateTestMacStore(tempDir string, pw *[]byte,
	netParams *chaincfg.Params) (*macaroons.RootKeyStorage, er.R) {

	netDir := btcwallet.NetworkDir(tempDir, netParams)
	errr := os.MkdirAll(netDir, 0700)
	if errr != nil {
		return nil, er.E(errr)
	}
	db, err := kvdb.Create(
		kvdb.BoltBackendName, path.Join(netDir, macaroons.DBFilename),
		true,
	)
	if err != nil {
		return nil, err
	}

	store, err := macaroons.NewRootKeyStorage(db)
	if err != nil {
		_ = db.Close()
		return nil, err
	}

	err = store.CreateUnlock(pw)
	if err != nil {
		_ = store.Close()
		return nil, err
	}
	_, _, errr = store.RootKey(defaultRootKeyIDContext)
	if errr != nil {
		_ = store.Close()
		return nil, er.E(errr)
	}

	return store, nil
}

// TestGenSeedUserEntropy tests that the gen seed method generates a valid
// cipher seed mnemonic phrase and user provided source of entropy.
func TestGenSeed(t *testing.T) {
	t.Parallel()

	log.Debugf(">>>>> running TestGenSeed()")
	// First, we'll create a new test directory and unlocker service for
	// that directory.
	testDir, errr := ioutil.TempDir("", "testcreate")
	require.NoError(t, errr)
	defer func() {
		_ = os.RemoveAll(testDir)
	}()

	service := walletunlocker.New(testDir, testNetParams, true, nil, "", testWalletFilename)

	// Now that the service has been created, we'll ask it to generate a
	// new seed for us given a test passphrase.
	aezeedPass := []byte("kek")
	genSeedReq := &lnrpc.GenSeedRequest{
		AezeedPassphrase: aezeedPass,
		SeedEntropy:      make([]byte, 0),
	}

	ctx := context.Background()
	seedResp, errr := service.GenSeed(ctx, genSeedReq)
	require.NoError(t, errr)

	// We should then be able to take the generated mnemonic, and properly
	// decipher both it.
	mnemonic := strings.Join(seedResp.CipherSeedMnemonic, " ")
	_, err := seedwords.SeedFromWords(mnemonic)
	util.RequireNoErr(t, err)
}

// TestGenSeedInvalidEntropy tests that the gen seed method generates a valid
// cipher seed mnemonic pass phrase even when the user doesn't provide its own
// source of entropy.
//	the following test makes no sense anymore, since the seedwords package doesn't support entropy
/*
func TestGenSeedGenerateEntropy(t *testing.T) {
	t.Parallel()

	log.Debugf(">>>>> running TestGenSeedGenerateEntropy()")
	// First, we'll create a new test directory and unlocker service for
	// that directory.
	testDir, errr := ioutil.TempDir("", "testcreate")
	require.NoError(t, errr)
	defer func() {
		_ = os.RemoveAll(testDir)
	}()
	service := walletunlocker.New(testDir, testNetParams, true, nil, testDir, testWalletFilename)

	// Now that the service has been created, we'll ask it to generate a
	// new seed for us given a test passphrase. Note that we don't actually
	aezeedPass := []byte("kek")
	genSeedReq := &lnrpc.GenSeedRequest{
		AezeedPassphrase: aezeedPass,
	}

	ctx := context.Background()
	seedResp, errr := service.GenSeed(ctx, genSeedReq)
	require.NoError(t, errr)

	// We should then be able to take the generated mnemonic, and properly
	// decipher both it.
	mnemonic := strings.Join(seedResp.CipherSeedMnemonic, " ")
	_, err := seedwords.SeedFromWords(mnemonic)
	util.RequireNoErr(t, err)
}
*/

/*
// TestGenSeedInvalidEntropy tests that if a user attempt to create a seed with
// the wrong number of bytes for the initial entropy, then the proper error is
// returned.
*/
// TestGenSeedInvalidEntropy tests that if a user attempt to create a seed with
// a non empty initial entropy, then the proper error is returned.
func TestGenSeedInvalidEntropy(t *testing.T) {
	t.Parallel()

	log.Debugf(">>>>> running TestGenSeedInvalidEntropy()")
	// First, we'll create a new test directory and unlocker service for
	// that directory.
	testDir, errr := ioutil.TempDir("", "testcreate")
	require.NoError(t, errr)
	defer func() {
		_ = os.RemoveAll(testDir)
	}()
	service := walletunlocker.New(testDir, testNetParams, true, nil, "", testWalletFilename)

	// Now that the service has been created, we'll ask it to generate a
	// new seed for us given a test passphrase. However, we'll be using an
	// invalid set of entropy that's 55 bytes, instead of 15 bytes.
	aezeedPass := []byte("kek")
	genSeedReq := &lnrpc.GenSeedRequest{
		AezeedPassphrase: aezeedPass,
		SeedEntropy:      bytes.Repeat([]byte("a"), 55),
	}

	// We should get an error now since the entropy source was invalid.
	ctx := context.Background()
	_, errr = service.GenSeed(ctx, genSeedReq)
	require.Error(t, errr)
	require.Contains(t, errr.Error(), "custom seed input entropy is not supported")
}

// TestInitWallet tests that the user is able to properly initialize the wallet
// given an existing cipher seed passphrase.
func TestInitWallet(t *testing.T) {
	t.Parallel()

	log.Debugf(">>>>> running TestInitWallet()")
	// testDir is empty, meaning wallet was not created from before.
	testDir, errr := ioutil.TempDir("", "testcreate")
	require.NoError(t, errr)
	defer func() {
		_ = os.RemoveAll(testDir)
	}()

	// Create new UnlockerService.
	service := walletunlocker.New(testDir, testNetParams, true, nil, "", testWalletFilename)

	// Once we have the unlocker service created, we'll now instantiate a
	// new cipher seed and its mnemonic.
	pass := []byte("test")
	cipherSeed, mnemonic := createSeedAndMnemonic(t, pass)

	// Now that we have all the necessary items, we'll now issue the Init
	// command to the wallet. This should check the validity of the cipher
	// seed, then send over the initialization information over the init
	// channel.
	ctx := context.Background()
	req := &lnrpc.InitWalletRequest{
		WalletPassword:     testPassword,
		CipherSeedMnemonic: strings.Split(mnemonic, " "),
		AezeedPassphrase:   pass,
		RecoveryWindow:     int32(testRecoveryWindow),
		StatelessInit:      true,
	}

	errChan := make(chan er.R, 1)

	go func() {
		response, err := service.InitWallet(ctx, req)
		if err != nil {
			errChan <- er.E(err)
			return
		}

		if !bytes.Equal(response.AdminMacaroon, testMac) {
			errChan <- er.Errorf("mismatched macaroon: "+
				"expected %x, got %x", testMac,
				response.AdminMacaroon)
		}
	}()

	// The same user passphrase, and also the plaintext cipher seed
	// should be sent over and match exactly.
	select {
	case err := <-errChan:
		t.Fatalf("InitWallet call failed: %v", err)

	case msg := <-service.InitMsgs:
		msgSeed := msg.Seed
		require.Equal(t, testPassword, msg.Passphrase)
		require.Equal(
			t, cipherSeed.Version(), msgSeed.Version(),
		)
		require.Equal(t, cipherSeed.Birthday(), msgSeed.Birthday())
		require.Equal(t, testRecoveryWindow, msg.RecoveryWindow)
		require.Equal(t, true, msg.StatelessInit)

		// Send a fake macaroon that should be returned in the response
		// in the async code above.
		service.MacResponseChan <- testMac

	case <-time.After(defaultTestTimeout):
		t.Fatalf("password not received")
	}

	// Create a wallet in testDir.
	createTestWallet(t, testDir, testNetParams)

	// Now calling InitWallet should fail, since a wallet already exists in
	// the directory.
	_, errr = service.InitWallet(ctx, req)
	require.Error(t, errr)
	require.Contains(t, errr.Error(), "wallet already exists")

	// Similarly, if we try to do GenSeed again, we should get an error as
	// the wallet already exists.
	_, errr = service.GenSeed(ctx, &lnrpc.GenSeedRequest{})
	require.Error(t, errr)
	require.Contains(t, errr.Error(), "wallet already exists")
}

// TestInitWalletInvalidCipherSeed tests that if we attempt to create a wallet
// with an invalid cipher seed, then we'll receive an error.
func TestCreateWalletInvalidEntropy(t *testing.T) {
	t.Parallel()

	log.Debugf(">>>>> running TestCreateWalletInvalidEntropy()")
	// testDir is empty, meaning wallet was not created from before.
	testDir, errr := ioutil.TempDir("", "testcreate")
	require.NoError(t, errr)
	defer func() {
		_ = os.RemoveAll(testDir)
	}()

	// Create new UnlockerService.
	service := walletunlocker.New(testDir, testNetParams, true, nil, "", testWalletFilename)

	// We'll attempt to init the wallet with an invalid cipher seed and
	// passphrase.
	req := &lnrpc.InitWalletRequest{
		WalletPassword:     testPassword,
		CipherSeedMnemonic: []string{"invalid", "seed"},
		AezeedPassphrase:   []byte("fake pass"),
	}

	ctx := context.Background()
	_, errr = service.InitWallet(ctx, req)
	require.Error(t, errr)
	require.Contains(t, errr.Error(), "Expected a 15 word seed")
}

// TestUnlockWallet checks that trying to unlock non-existing wallet fail, that
// unlocking existing wallet with wrong passphrase fails, and that unlocking
// existing wallet with correct passphrase succeeds.
func TestUnlockWallet(t *testing.T) {
	t.Parallel()

	log.Debugf(">>>>> running TestUnlockWallet()")
	// testDir is empty, meaning wallet was not created from before.
	testDir, errr := ioutil.TempDir("", "testunlock")
	require.NoError(t, errr)
	defer func() {
		_ = os.RemoveAll(testDir)
	}()

	// Create new UnlockerService.
	service := walletunlocker.New(testDir, testNetParams, true, nil, "", testWalletFilename)

	ctx := context.Background()
	req := &lnrpc.UnlockWalletRequest{
		WalletPassword:    testPassword,
		WalletPubPassword: testPassword,
		RecoveryWindow:    int32(testRecoveryWindow),
		StatelessInit:     true,
	}

	// Should fail to unlock non-existing wallet.
	_, err := service.UnlockWallet(ctx, req)
	require.Error(t, err)
	require.Contains(t, err.Error(), "wallet not found")

	// Create a wallet we can try to unlock.
	createTestWallet(t, testDir, testNetParams)

	// Try unlocking this wallet with the wrong passphrase.
	wrongReq := &lnrpc.UnlockWalletRequest{
		WalletPassword: []byte("wrong-ofc"),
	}
	_, err = service.UnlockWallet(ctx, wrongReq)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid passphrase for master public key")

	// With the correct password, we should be able to unlock the wallet.
	errChan := make(chan er.R, 1)
	go func() {
		_, err := service.UnlockWallet(ctx, req)
		if err != nil {
			errChan <- er.E(err)
		}
	}()

	// Password and recovery window should be sent over the channel.
	select {
	case err := <-errChan:
		t.Fatalf("UnlockWallet call failed: %v", err)

	case unlockMsg := <-service.UnlockMsgs:
		require.Equal(t, testPassword, unlockMsg.Passphrase)
		require.Equal(t, testRecoveryWindow, unlockMsg.RecoveryWindow)
		require.Equal(t, true, unlockMsg.StatelessInit)

		// Send a fake macaroon that should be returned in the response
		// in the async code above.
		service.MacResponseChan <- testMac

	case <-time.After(defaultTestTimeout):
		t.Fatalf("password not received")
	}
}

// TestChangeWalletPasswordNewRootkey tests that we can successfully change the
// wallet's password needed to unlock it and rotate the root key for the
// macaroons in the same process.
func TestChangeWalletPasswordNewRootkey(t *testing.T) {
	t.Parallel()

	log.Debugf(">>>>> running TestChangeWalletPasswordNewRootkey()")
	// testDir is empty, meaning wallet was not created from before.
	testDir, errr := ioutil.TempDir("", "testchangepassword")
	require.NoError(t, errr)
	defer func() {
		_ = os.RemoveAll(testDir)
	}()

	// Changing the password of the wallet will also try to change the
	// password of the macaroon DB. We create a default DB here but close it
	// immediately so the service does not fail when trying to open it.
	store, err := openOrCreateTestMacStore(
		testDir, &testPassword, testNetParams,
	)
	util.RequireNoErr(t, err)

	err = store.Close()
	util.RequireNoErr(t, err)

	// Create some files that will act as macaroon files that should be
	// deleted after a password change is successful with a new root key
	// requested.
	var tempFiles []string
	for i := 0; i < 3; i++ {
		file, err := ioutil.TempFile(testDir, "")
		if err != nil {
			t.Fatalf("unable to create temp file: %v", err)
		}
		tempFiles = append(tempFiles, file.Name())
		err = file.Close()
		require.NoError(t, err)

		log.Debugf(">>>>> [3] macaroon file created: %s", file.Name())
	}

	// Create a new UnlockerService with our temp files.
	service := walletunlocker.New(testDir, testNetParams, true, nil, "", testWalletFilename)

	ctx := context.Background()
	newPassword := []byte("hunter2???")

	req := &lnrpc.ChangePasswordRequest{
		CurrentPassword:    testPassword,
		CurrentPubPassword: testPassword,
		NewPassword:        newPassword,
		NewMacaroonRootKey: true,
	}

	// Changing the password to a non-existing wallet should fail.
	_, errr = service.ChangePassword(ctx, req)
	require.Error(t, errr)
	require.Contains(t, errr.Error(), "wallet not found")

	// Create a wallet to test changing the password.
	createTestWallet(t, testDir, testNetParams)

	// Attempting to change the wallet's password using an incorrect
	// current password should fail.
	wrongReq := &lnrpc.ChangePasswordRequest{
		CurrentPassword: []byte("wrong-ofc"),
		NewPassword:     newPassword,
	}
	_, errr = service.ChangePassword(ctx, wrongReq)
	require.Error(t, errr)
	require.Contains(t, errr.Error(), "invalid passphrase for master public key")

	// The files should still exist after an unsuccessful attempt to change
	// the wallet's password.
	for _, tempFile := range tempFiles {
		if _, err := os.Stat(tempFile); os.IsNotExist(err) {
			t.Fatal("file does not exist but it should")
		}
		log.Debugf(">>>>> [4] macaroon file still exists: %s", tempFile)
	}

	// Attempting to change the wallet's password using an invalid
	// new password should fail.
	wrongReq.NewPassword = []byte("8")
	_, errr = service.ChangePassword(ctx, wrongReq)
	require.Error(t, errr)
	require.Contains(t, errr.Error(), "custom password must have at least 8 characters")

	// When providing the correct wallet's current password and a new
	// password that meets the length requirement, the password change
	// should succeed.
	errChan := make(chan er.R, 1)
	go doChangePassword(service, testDir, req, errChan)

	// The new password should be sent over the channel.
	select {
	case err := <-errChan:
		t.Fatalf("ChangePassword call failed: %v", err)

	case unlockMsg := <-service.UnlockMsgs:
		require.Equal(t, newPassword, unlockMsg.Passphrase)

		// Send a fake macaroon that should be returned in the response
		// in the async code above.
		service.MacResponseChan <- testMac

	case <-time.After(defaultTestTimeout):
		t.Fatalf("password not received")
	}

	// The files should no longer exist.
	for _, tempFile := range tempFiles {
		if _, err := os.Stat(tempFile); os.IsNotExist(err) {
			t.Fatal("file exists but it shouldn't: " + tempFile)
		}
	}
}

// TestChangeWalletPasswordStateless checks that trying to change the password
// of an existing wallet that was initialized stateless works when when the
// --stateless_init flat is set. Also checks that if no password is given,
// the default password is used.
func TestChangeWalletPasswordStateless(t *testing.T) {
	t.Parallel()

	log.Debugf(">>>>> running TestChangeWalletPasswordStateless()")
	// testDir is empty, meaning wallet was not created from before.
	testDir, errr := ioutil.TempDir("", "testchangepasswordstateless")
	require.NoError(t, errr)
	defer func() {
		_ = os.RemoveAll(testDir)
	}()

	// Changing the password of the wallet will also try to change the
	// password of the macaroon DB. We create a default DB here but close it
	// immediately so the service does not fail when trying to open it.
	store, err := openOrCreateTestMacStore(
		testDir, &lnwallet.DefaultPrivatePassphrase, testNetParams,
	)
	util.RequireNoErr(t, err)
	util.RequireNoErr(t, store.Close())

	// Create a temp file that will act as the macaroon DB file that will
	// be deleted by changing the password.
	tmpFile, errr := ioutil.TempFile(testDir, "")
	require.NoError(t, errr)
	tempMacFile := tmpFile.Name()
	errr = tmpFile.Close()
	require.NoError(t, errr)

	// Create a file name that does not exist that will be used as a
	// macaroon file reference. The fact that the file does not exist should
	// not throw an error when --stateless_init is used.
	nonExistingFile := path.Join(testDir, "does-not-exist")

	// Create a new UnlockerService with our temp files.
	service := walletunlocker.New(testDir, testNetParams, true, []string{
		tempMacFile, nonExistingFile,
	}, "", testWalletFilename)

	// Create a wallet we can try to unlock. We use the default password
	// so we can check that the unlocker service defaults to this when
	// we give it an empty CurrentPassword to indicate we come from a
	// --noencryptwallet state.
	createTestWalletWithPw(
		t, lnwallet.DefaultPublicPassphrase,
		lnwallet.DefaultPrivatePassphrase, testDir, testNetParams,
	)

	// We make sure that we get a proper error message if we forget to
	// add the --stateless_init flag but the macaroon files don't exist.
	badReq := &lnrpc.ChangePasswordRequest{
		NewPassword:        testPassword,
		NewMacaroonRootKey: true,
	}
	ctx := context.Background()
	_, errr = service.ChangePassword(ctx, badReq)
	require.Error(t, errr)
	require.Contains(t, errr.Error(), "if the wallet was initialized stateless")

	// Prepare the correct request we are going to send to the unlocker
	// service. We don't provide a current password to indicate there
	// was none set before.
	req := &lnrpc.ChangePasswordRequest{
		NewPassword:        testPassword,
		StatelessInit:      true,
		NewMacaroonRootKey: true,
	}

	// Since we indicated the wallet was initialized stateless, the service
	// will block until it receives the macaroon through the channel
	// provided in the message in UnlockMsgs. So we need to call the service
	// async and then wait for the unlock message to arrive so we can send
	// back a fake macaroon.
	errChan := make(chan er.R, 1)
	go doChangePassword(service, testDir, req, errChan)

	// Password and recovery window should be sent over the channel.
	select {
	case err := <-errChan:
		t.Fatalf("ChangePassword call failed: %v", err)

	case unlockMsg := <-service.UnlockMsgs:
		require.Equal(t, testPassword, unlockMsg.Passphrase)

		// Send a fake macaroon that should be returned in the response
		// in the async code above.
		service.MacResponseChan <- testMac

	case <-time.After(defaultTestTimeout):
		t.Fatalf("password not received")
	}
}

func doChangePassword(service *walletunlocker.UnlockerService, testDir string,
	req *lnrpc.ChangePasswordRequest, errChan chan er.R) {

	// When providing the correct wallet's current password and a
	// new password that meets the length requirement, the password
	// change should succeed.
	ctx := context.Background()
	response, errr := service.ChangePassword(ctx, req)
	if errr != nil {
		errChan <- er.Errorf("could not change password: %v", errr)
		return
	}

	if !bytes.Equal(response.AdminMacaroon, testMac) {
		errChan <- er.Errorf("mismatched macaroon: expected "+
			"%x, got %x", testMac, response.AdminMacaroon)
	}

	// Close the macaroon DB and try to open it and read the root
	// key with the new password.
	store, err := openOrCreateTestMacStore(
		testDir, &testPassword, testNetParams,
	)
	if err != nil {
		errChan <- er.Errorf("could not create test store: %v", err)
		return
	}
	_, _, errr = store.RootKey(defaultRootKeyIDContext)
	if errr != nil {
		errChan <- er.Errorf("could not get root key: %v", errr)
		return
	}

	// Do cleanup now. Since we are in a go func, the defer at the
	// top of the outer would not work, because it would delete
	// the directory before we could check the content in here.
	err = store.Close()
	if err != nil {
		errChan <- er.Errorf("could not close store: %v", err)
		return
	}
	errr = os.RemoveAll(testDir)
	if errr != nil {
		errChan <- er.E(errr)
		return
	}
}
