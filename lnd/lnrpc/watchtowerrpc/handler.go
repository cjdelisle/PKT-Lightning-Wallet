package watchtowerrpc

import (
	"context"
	fmt "fmt"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/pkt-cash/pktd/btcutil/er"
	"github.com/pkt-cash/pktd/generated/proto/watchtower_pb"
	"github.com/pkt-cash/pktd/pktlog/log"
	"google.golang.org/grpc"
)

const (
	// subServerName is the name of the sub rpc server. We'll use this name
	// to register ourselves, and we also require that the main
	// SubServerConfigDispatcher instance recognizes it as the name of our
	// RPC service.
	subServerName = "WatchtowerRPC"
)

var (
	// ErrTowerNotActive signals that RPC calls cannot be processed because
	// the watchtower is not active.
	ErrTowerNotActive = er.GenericErrorType.CodeWithDetail("ErrTowerNotActive", "watchtower not active")
)

// Handler is the RPC server we'll use to interact with the backing active
// watchtower.
type Handler struct {
	cfg Config
	watchtower_pb.UnimplementedWatchtowerServer
}

// A compile time check to ensure that Handler fully implements the Handler gRPC
// service.
var _ watchtower_pb.WatchtowerServer = (*Handler)(nil)

// New returns a new instance of the Watchtower sub-server. We also return the
// set of permissions for the macaroons that we may create within this method.
// If the macaroons we need aren't found in the filepath, then we'll create them
// on start up. If we're unable to locate, or create the macaroons we need, then
// we'll return with an error.
func New(cfg *Config) (*Handler, er.R) {
	return &Handler{cfg: *cfg}, nil
}

// Start launches any helper goroutines required for the Handler to function.
//
// NOTE: This is part of the lnrpc.SubServer interface.
func (c *Handler) Start() er.R {
	return nil
}

// Stop signals any active goroutines for a graceful closure.
//
// NOTE: This is part of the lnrpc.SubServer interface.
func (c *Handler) Stop() er.R {
	return nil
}

// Name returns a unique string representation of the sub-server. This can be
// used to identify the sub-server and also de-duplicate them.
//
// NOTE: This is part of the lnrpc.SubServer interface.
func (c *Handler) Name() string {
	return subServerName
}

// RegisterWithRootServer will be called by the root gRPC server to direct a sub
// RPC server to register itself with the main gRPC root server. Until this is
// called, each sub-server won't be able to have requests routed towards it.
//
// NOTE: This is part of the lnrpc.SubServer interface.
func (c *Handler) RegisterWithRootServer(grpcServer *grpc.Server) er.R {
	// We make sure that we register it with the main gRPC server to ensure
	// all our methods are routed properly.
	watchtower_pb.RegisterWatchtowerServer(grpcServer, c)

	log.Debugf("Watchtower RPC server successfully register with root " +
		"gRPC server")

	return nil
}

// RegisterWithRestServer will be called by the root REST mux to direct a sub
// RPC server to register itself with the main REST mux server. Until this is
// called, each sub-server won't be able to have requests routed towards it.
//
// NOTE: This is part of the lnrpc.SubServer interface.
func (c *Handler) RegisterWithRestServer(ctx context.Context,
	mux *runtime.ServeMux, dest string, opts []grpc.DialOption) er.R {

	// We make sure that we register it with the main REST server to ensure
	// all our methods are routed properly.
	// err := RegisterWatchtowerHandlerFromEndpoint(ctx, mux, dest, opts)
	// if err != nil {
	// 	log.Errorf("Could not register Watchtower REST server "+
	// 		"with root REST server: %v", err)
	// 	return er.E(err)
	// }

	log.Debugf("Watchtower REST server successfully registered with " +
		"root REST server")
	return nil
}

// AddTower adds a new watchtower reachable at the given address and considers
// it for new sessions. If the watchtower already exists, then any new addresses
// included will be considered when dialing it for session negotiations and
// backups.
func (c *Handler) GetInfo(ctx context.Context,
	req *watchtower_pb.GetInfoRequest) (*watchtower_pb.GetInfoResponse, error) {
	out, err := c.GetInfo0(ctx, req)
	return out, er.Native(err)
}
func (c *Handler) GetInfo0(ctx context.Context,
	req *watchtower_pb.GetInfoRequest) (*watchtower_pb.GetInfoResponse, er.R) {

	if err := c.isActive(); err != nil {
		return nil, err
	}

	pubkey := c.cfg.Tower.PubKey().SerializeCompressed()

	var listeners []string
	for _, addr := range c.cfg.Tower.ListeningAddrs() {
		listeners = append(listeners, addr.String())
	}

	var uris []string
	for _, addr := range c.cfg.Tower.ExternalIPs() {
		uris = append(uris, fmt.Sprintf("%x@%v", pubkey, addr))
	}

	return &watchtower_pb.GetInfoResponse{
		Pubkey:    pubkey,
		Listeners: listeners,
		Uris:      uris,
	}, nil
}

// isActive returns nil if the tower backend is initialized, and the Handler can
// proccess RPC requests.
func (c *Handler) isActive() er.R {
	if c.cfg.Active {
		return nil
	}
	return ErrTowerNotActive.Default()
}
