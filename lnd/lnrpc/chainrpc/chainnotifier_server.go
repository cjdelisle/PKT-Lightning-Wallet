package chainrpc

import (
	"bytes"
	"context"
	"path/filepath"
	"sync"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/pkt-cash/pktd/btcutil/er"
	"github.com/pkt-cash/pktd/chaincfg/chainhash"
	"github.com/pkt-cash/pktd/generated/proto/chainrpc_pb"
	"github.com/pkt-cash/pktd/lnd/chainntnfs"
	"github.com/pkt-cash/pktd/lnd/lnrpc"
	"github.com/pkt-cash/pktd/pktlog/log"
	"github.com/pkt-cash/pktd/wire"
	"google.golang.org/grpc"
)

const (
	// subServerName is the name of the RPC sub-server. We'll use this name
	// to register ourselves, and we also require that the main
	// SubServerConfigDispatcher instance recognize this as the name of the
	// config file that we need.
	subServerName = "ChainRPC"
)

var Err = er.NewErrorType("chainrpc")

var (
	// DefaultChainNotifierMacFilename is the default name of the chain
	// notifier macaroon that we expect to find via a file handle within the
	// main configuration file in this package.
	DefaultChainNotifierMacFilename = "chainnotifier.macaroon"

	// ErrChainNotifierServerShuttingDown is an error returned when we are
	// waiting for a notification to arrive but the chain notifier server
	// has been shut down.
	ErrChainNotifierServerShuttingDown = Err.CodeWithDetail("ErrChainNotifierServerShuttingDown", "chain notifier RPC "+
		"subserver shutting down")

	// ErrChainNotifierServerNotActive indicates that the chain notifier hasn't
	// finished the startup process.
	ErrChainNotifierServerNotActive = Err.CodeWithDetail("ErrChainNotifierServerNotActive", "chain notifier RPC is "+
		"still in the process of starting")
)

// Server is a sub-server of the main RPC server: the chain notifier RPC. This
// RPC sub-server allows external callers to access the full chain notifier
// capabilities of lnd. This allows callers to create custom protocols, external
// to lnd, even backed by multiple distinct lnd across independent failure
// domains.
type Server struct {
	started sync.Once
	stopped sync.Once

	cfg Config

	quit chan struct{}

	chainrpc_pb.UnimplementedChainNotifierServer
}

// New returns a new instance of the chainrpc ChainNotifier sub-server. We also
// return the set of permissions for the macaroons that we may create within
// this method. If the macaroons we need aren't found in the filepath, then
// we'll create them on start up. If we're unable to locate, or create the
// macaroons we need, then we'll return with an error.
func New(cfg *Config) (*Server, er.R) {
	// If the path of the chain notifier macaroon wasn't generated, then
	// we'll assume that it's found at the default network directory.
	if cfg.ChainNotifierMacPath == "" {
		cfg.ChainNotifierMacPath = filepath.Join(
			cfg.NetworkDir, DefaultChainNotifierMacFilename,
		)
	}

	return &Server{
		cfg:  *cfg,
		quit: make(chan struct{}),
	}, nil
}

// Compile-time checks to ensure that Server fully implements the
// ChainNotifierServer gRPC service and lnrpc.SubServer interface.
var _ chainrpc_pb.ChainNotifierServer = (*Server)(nil)
var _ lnrpc.SubServer = (*Server)(nil)

// Start launches any helper goroutines required for the server to function.
//
// NOTE: This is part of the lnrpc.SubServer interface.
func (s *Server) Start() er.R {
	s.started.Do(func() {})
	return nil
}

// Stop signals any active goroutines for a graceful closure.
//
// NOTE: This is part of the lnrpc.SubServer interface.
func (s *Server) Stop() er.R {
	s.stopped.Do(func() {
		close(s.quit)
	})
	return nil
}

// Name returns a unique string representation of the sub-server. This can be
// used to identify the sub-server and also de-duplicate them.
//
// NOTE: This is part of the lnrpc.SubServer interface.
func (s *Server) Name() string {
	return subServerName
}

// RegisterWithRootServer will be called by the root gRPC server to direct a RPC
// sub-server to register itself with the main gRPC root server. Until this is
// called, each sub-server won't be able to have requests routed towards it.
//
// NOTE: This is part of the lnrpc.SubServer interface.
func (s *Server) RegisterWithRootServer(grpcServer *grpc.Server) er.R {
	// We make sure that we register it with the main gRPC server to ensure
	// all our methods are routed properly.
	chainrpc_pb.RegisterChainNotifierServer(grpcServer, s)

	log.Debug("ChainNotifier RPC server successfully register with root " +
		"gRPC server")

	return nil
}

// RegisterWithRestServer will be called by the root REST mux to direct a sub
// RPC server to register itself with the main REST mux server. Until this is
// called, each sub-server won't be able to have requests routed towards it.
//
// NOTE: This is part of the lnrpc.SubServer interface.
func (s *Server) RegisterWithRestServer(ctx context.Context,
	mux *runtime.ServeMux, dest string, opts []grpc.DialOption) er.R {

	// We make sure that we register it with the main REST server to ensure
	// all our methods are routed properly.
	// err := RegisterChainNotifierHandlerFromEndpoint(ctx, mux, dest, opts)
	// if err != nil {
	// 	log.Errorf("Could not register ChainNotifier REST server "+
	// 		"with root REST server: %v", err)
	// 	return er.E(err)
	// }

	log.Debugf("ChainNotifier REST server successfully registered with " +
		"root REST server")
	return nil
}

// RegisterConfirmationsNtfn is a synchronous response-streaming RPC that
// registers an intent for a client to be notified once a confirmation request
// has reached its required number of confirmations on-chain.
//
// A client can specify whether the confirmation request should be for a
// particular transaction by its hash or for an output script by specifying a
// zero hash.
//
// NOTE: This is part of the chainrpc.ChainNotifierService interface.
func (s *Server) RegisterConfirmationsNtfn(
	in *chainrpc_pb.ConfRequest,
	confStream chainrpc_pb.ChainNotifier_RegisterConfirmationsNtfnServer,
) error {
	return er.Native(s.RegisterConfirmationsNtfn0(in, confStream))
}
func (s *Server) RegisterConfirmationsNtfn0(in *chainrpc_pb.ConfRequest,
	confStream chainrpc_pb.ChainNotifier_RegisterConfirmationsNtfnServer) er.R {

	if !s.cfg.ChainNotifier.Started() {
		return ErrChainNotifierServerNotActive.Default()
	}

	// We'll start by reconstructing the RPC request into what the
	// underlying ChainNotifier expects.
	var txid chainhash.Hash
	copy(txid[:], in.Txid)

	// We'll then register for the spend notification of the request.
	confEvent, err := s.cfg.ChainNotifier.RegisterConfirmationsNtfn(
		&txid, in.Script, in.NumConfs, in.HeightHint,
	)
	if err != nil {
		return err
	}
	defer confEvent.Cancel()

	// With the request registered, we'll wait for its spend notification to
	// be dispatched.
	for {
		select {
		// The transaction satisfying the request has confirmed on-chain
		// and reached its required number of confirmations. We'll
		// dispatch an event to the caller indicating so.
		case details, ok := <-confEvent.Confirmed:
			if !ok {
				return chainntnfs.ErrChainNotifierShuttingDown.Default()
			}

			var rawTxBuf bytes.Buffer
			err := details.Tx.Serialize(&rawTxBuf)
			if err != nil {
				return err
			}

			rpcConfDetails := &chainrpc_pb.ConfDetails{
				RawTx:       rawTxBuf.Bytes(),
				BlockHash:   details.BlockHash[:],
				BlockHeight: details.BlockHeight,
				TxIndex:     details.TxIndex,
			}

			conf := &chainrpc_pb.ConfEvent{
				Event: &chainrpc_pb.ConfEvent_Conf{
					Conf: rpcConfDetails,
				},
			}
			if err := confStream.Send(conf); err != nil {
				return er.E(err)
			}

		// The transaction satisfying the request has been reorged out
		// of the chain, so we'll send an event describing so.
		case _, ok := <-confEvent.NegativeConf:
			if !ok {
				return chainntnfs.ErrChainNotifierShuttingDown.Default()
			}

			reorg := &chainrpc_pb.ConfEvent{
				Event: &chainrpc_pb.ConfEvent_Reorg{Reorg: &chainrpc_pb.Reorg{}},
			}
			if err := confStream.Send(reorg); err != nil {
				return er.E(err)
			}

		// The transaction satisfying the request has confirmed and is
		// no longer under the risk of being reorged out of the chain,
		// so we can safely exit.
		case _, ok := <-confEvent.Done:
			if !ok {
				return chainntnfs.ErrChainNotifierShuttingDown.Default()
			}

			return nil

		// The response stream's context for whatever reason has been
		// closed. We'll return the error indicated by the context
		// itself to the caller.
		case <-confStream.Context().Done():
			return er.E(confStream.Context().Err())

		// The server has been requested to shut down.
		case <-s.quit:
			return ErrChainNotifierServerShuttingDown.Default()
		}
	}
}

// RegisterSpendNtfn is a synchronous response-streaming RPC that registers an
// intent for a client to be notification once a spend request has been spent by
// a transaction that has confirmed on-chain.
//
// A client can specify whether the spend request should be for a particular
// outpoint  or for an output script by specifying a zero outpoint.
//
// NOTE: This is part of the chainrpc.ChainNotifierService interface.
func (s *Server) RegisterSpendNtfn(
	in *chainrpc_pb.SpendRequest,
	spendStream chainrpc_pb.ChainNotifier_RegisterSpendNtfnServer,
) error {
	return er.Native(s.RegisterSpendNtfn0(in, spendStream))
}
func (s *Server) RegisterSpendNtfn0(in *chainrpc_pb.SpendRequest,
	spendStream chainrpc_pb.ChainNotifier_RegisterSpendNtfnServer) er.R {

	if !s.cfg.ChainNotifier.Started() {
		return ErrChainNotifierServerNotActive.Default()
	}

	// We'll start by reconstructing the RPC request into what the
	// underlying ChainNotifier expects.
	var op *wire.OutPoint
	if in.Outpoint != nil {
		var txid chainhash.Hash
		copy(txid[:], in.Outpoint.Hash)
		op = &wire.OutPoint{Hash: txid, Index: in.Outpoint.Index}
	}

	// We'll then register for the spend notification of the request.
	spendEvent, err := s.cfg.ChainNotifier.RegisterSpendNtfn(
		op, in.Script, in.HeightHint,
	)
	if err != nil {
		return err
	}
	defer spendEvent.Cancel()

	// With the request registered, we'll wait for its spend notification to
	// be dispatched.
	for {
		select {
		// A transaction that spends the given has confirmed on-chain.
		// We'll return an event to the caller indicating so that
		// includes the details of the spending transaction.
		case details, ok := <-spendEvent.Spend:
			if !ok {
				return chainntnfs.ErrChainNotifierShuttingDown.Default()
			}

			var rawSpendingTxBuf bytes.Buffer
			err := details.SpendingTx.Serialize(&rawSpendingTxBuf)
			if err != nil {
				return err
			}

			rpcSpendDetails := &chainrpc_pb.SpendDetails{
				SpendingOutpoint: &chainrpc_pb.Outpoint{
					Hash:  details.SpentOutPoint.Hash[:],
					Index: details.SpentOutPoint.Index,
				},
				RawSpendingTx:      rawSpendingTxBuf.Bytes(),
				SpendingTxHash:     details.SpenderTxHash[:],
				SpendingInputIndex: details.SpenderInputIndex,
				SpendingHeight:     uint32(details.SpendingHeight),
			}

			spend := &chainrpc_pb.SpendEvent{
				Event: &chainrpc_pb.SpendEvent_Spend{
					Spend: rpcSpendDetails,
				},
			}
			if err := spendStream.Send(spend); err != nil {
				return er.E(err)
			}

		// The spending transaction of the request has been reorged of
		// the chain. We'll return an event to the caller indicating so.
		case _, ok := <-spendEvent.Reorg:
			if !ok {
				return chainntnfs.ErrChainNotifierShuttingDown.Default()
			}

			reorg := &chainrpc_pb.SpendEvent{
				Event: &chainrpc_pb.SpendEvent_Reorg{Reorg: &chainrpc_pb.Reorg{}},
			}
			if err := spendStream.Send(reorg); err != nil {
				return er.E(err)
			}

		// The spending transaction of the requests has confirmed
		// on-chain and is no longer under the risk of being reorged out
		// of the chain, so we can safely exit.
		case _, ok := <-spendEvent.Done:
			if !ok {
				return chainntnfs.ErrChainNotifierShuttingDown.Default()
			}

			return nil

		// The response stream's context for whatever reason has been
		// closed. We'll return the error indicated by the context
		// itself to the caller.
		case <-spendStream.Context().Done():
			return er.E(spendStream.Context().Err())

		// The server has been requested to shut down.
		case <-s.quit:
			return ErrChainNotifierServerShuttingDown.Default()
		}
	}
}

// RegisterBlockEpochNtfn is a synchronous response-streaming RPC that registers
// an intent for a client to be notified of blocks in the chain. The stream will
// return a hash and height tuple of a block for each new/stale block in the
// chain. It is the client's responsibility to determine whether the tuple
// returned is for a new or stale block in the chain.
//
// A client can also request a historical backlog of blocks from a particular
// point. This allows clients to be idempotent by ensuring that they do not
// missing processing a single block within the chain.
//
// NOTE: This is part of the chainrpc.ChainNotifierService interface.
func (s *Server) RegisterBlockEpochNtfn(
	in *chainrpc_pb.BlockEpoch,
	epochStream chainrpc_pb.ChainNotifier_RegisterBlockEpochNtfnServer,
) error {
	return er.Native(s.RegisterBlockEpochNtfn0(in, epochStream))
}
func (s *Server) RegisterBlockEpochNtfn0(in *chainrpc_pb.BlockEpoch,
	epochStream chainrpc_pb.ChainNotifier_RegisterBlockEpochNtfnServer) er.R {

	if !s.cfg.ChainNotifier.Started() {
		return ErrChainNotifierServerNotActive.Default()
	}

	// We'll start by reconstructing the RPC request into what the
	// underlying ChainNotifier expects.
	var hash chainhash.Hash
	copy(hash[:], in.Hash)

	// If the request isn't for a zero hash and a zero height, then we
	// should deliver a backlog of notifications from the given block
	// (hash/height tuple) until tip, and continue delivering epochs for
	// new blocks.
	var blockEpoch *chainntnfs.BlockEpoch
	if hash != chainntnfs.ZeroHash && in.Height != 0 {
		blockEpoch = &chainntnfs.BlockEpoch{
			Hash:   &hash,
			Height: int32(in.Height),
		}
	}

	epochEvent, err := s.cfg.ChainNotifier.RegisterBlockEpochNtfn(blockEpoch)
	if err != nil {
		return err
	}
	defer epochEvent.Cancel()

	for {
		select {
		// A notification for a block has been received. This block can
		// either be a new block or stale.
		case blockEpoch, ok := <-epochEvent.Epochs:
			if !ok {
				return chainntnfs.ErrChainNotifierShuttingDown.Default()
			}

			epoch := &chainrpc_pb.BlockEpoch{
				Hash:   blockEpoch.Hash[:],
				Height: uint32(blockEpoch.Height),
			}
			if err := epochStream.Send(epoch); err != nil {
				return er.E(err)
			}

		// The response stream's context for whatever reason has been
		// closed. We'll return the error indicated by the context
		// itself to the caller.
		case <-epochStream.Context().Done():
			return er.E(epochStream.Context().Err())

		// The server has been requested to shut down.
		case <-s.quit:
			return ErrChainNotifierServerShuttingDown.Default()
		}
	}
}
