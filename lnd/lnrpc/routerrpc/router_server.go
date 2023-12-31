package routerrpc

import (
	"context"
	"path/filepath"
	"sync/atomic"

	"github.com/pkt-cash/pktd/btcutil"
	"github.com/pkt-cash/pktd/btcutil/er"
	"github.com/pkt-cash/pktd/generated/proto/routerrpc_pb"
	"github.com/pkt-cash/pktd/generated/proto/rpc_pb"
	"github.com/pkt-cash/pktd/lnd/channeldb"
	"github.com/pkt-cash/pktd/lnd/lntypes"
	"github.com/pkt-cash/pktd/lnd/lnwire"
	"github.com/pkt-cash/pktd/lnd/routing"
	"github.com/pkt-cash/pktd/lnd/routing/route"
	"github.com/pkt-cash/pktd/pktlog/log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	errServerShuttingDown = er.GenericErrorType.CodeWithDetail("errServerShuttingDown",
		"routerrpc server shutting down")

	// ErrInterceptorAlreadyExists is an error returned when the a new stream
	// is opened and there is already one active interceptor.
	// The user must disconnect prior to open another stream.
	ErrInterceptorAlreadyExists = er.GenericErrorType.CodeWithDetail("ErrInterceptorAlreadyExists",
		"interceptor already exists")

	// DefaultRouterMacFilename is the default name of the router macaroon
	// that we expect to find via a file handle within the main
	// configuration file in this package.
	DefaultRouterMacFilename = "router.macaroon"
)

// Server is a stand alone sub RPC server which exposes functionality that
// allows clients to route arbitrary payment through the Lightning Network.
type Server struct {
	forwardInterceptorActive int32 // To be used atomically.

	cfg *Config

	quit chan struct{}
}

// New creates a new instance of the RouterServer given a configuration struct
// that contains all external dependencies. If the target macaroon exists, and
// we're unable to create it, then an error will be returned. We also return
// the set of permissions that we require as a server. At the time of writing
// of this documentation, this is the same macaroon as as the admin macaroon.
func New(cfg *Config) (*Server, er.R) {
	// If the path of the router macaroon wasn't generated, then we'll
	// assume that it's found at the default network directory.
	if cfg.RouterMacPath == "" {
		cfg.RouterMacPath = filepath.Join(
			cfg.NetworkDir, DefaultRouterMacFilename,
		)
	}

	routerServer := &Server{
		cfg:  cfg,
		quit: make(chan struct{}),
	}

	return routerServer, nil
}

// SendPaymentV2 attempts to route a payment described by the passed
// PaymentRequest to the final destination. If we are unable to route the
// payment, or cannot find a route that satisfies the constraints in the
// PaymentRequest, then an error will be returned. Otherwise, the payment
// pre-image, along with the final route will be returned.
func (s *Server) SendPaymentV2(req *routerrpc_pb.SendPaymentRequest,
	stream routerrpc_pb.Router_SendPaymentV2Server) error {

	payment, err := s.cfg.RouterBackend.extractIntentFromSendRequest(req)
	if err != nil {
		return er.Native(err)
	}

	err = s.cfg.Router.SendPaymentAsync(payment)
	if err != nil {
		// Transform user errors to grpc code.
		if channeldb.ErrPaymentInFlight.Is(err) ||
			channeldb.ErrAlreadyPaid.Is(err) {

			log.Debugf("SendPayment async result for hash %x: %v",
				payment.PaymentHash, err)

			return status.Error(
				codes.AlreadyExists, err.String(),
			)
		}

		log.Errorf("SendPayment async error for hash %x: %v",
			payment.PaymentHash, err)

		return er.Native(err)
	}

	return s.trackPayment(payment.PaymentHash, stream, req.NoInflightUpdates)
}

// EstimateRouteFee allows callers to obtain a lower bound w.r.t how much it
// may cost to send an HTLC to the target end destination.
func (s *Server) EstimateRouteFee(ctx context.Context,
	req *routerrpc_pb.RouteFeeRequest) (*routerrpc_pb.RouteFeeResponse, error) {

	if len(req.Dest) != 33 {
		return nil, er.Native(er.New("invalid length destination key"))
	}
	var destNode route.Vertex
	copy(destNode[:], req.Dest)

	// Next, we'll convert the amount in satoshis to mSAT, which are the
	// native unit of LN.
	amtMsat := lnwire.NewMSatFromSatoshis(btcutil.Amount(req.AmtSat))

	// Pick a fee limit
	//
	// TODO: Change this into behaviour that makes more sense.
	feeLimit := lnwire.NewMSatFromSatoshis(btcutil.UnitsPerCoin())

	// Finally, we'll query for a route to the destination that can carry
	// that target amount, we'll only request a single route. Set a
	// restriction for the default CLTV limit, otherwise we can find a route
	// that exceeds it and is useless to us.
	mc := s.cfg.RouterBackend.MissionControl
	route, err := s.cfg.Router.FindRoute(
		s.cfg.RouterBackend.SelfNode, destNode, amtMsat,
		&routing.RestrictParams{
			FeeLimit:          feeLimit,
			CltvLimit:         s.cfg.RouterBackend.MaxTotalTimelock,
			ProbabilitySource: mc.GetProbability,
		}, nil, nil, s.cfg.RouterBackend.DefaultFinalCltvDelta,
	)
	if err != nil {
		return nil, er.Native(err)
	}

	return &routerrpc_pb.RouteFeeResponse{
		RoutingFeeMsat: int64(route.TotalFees()),
		TimeLockDelay:  int64(route.TotalTimeLock),
	}, nil
}

// SendToRouteV2 sends a payment through a predefined route. The response of this
// call contains structured error information.
func (s *Server) SendToRouteV2(ctx context.Context,
	req *routerrpc_pb.SendToRouteRequest) (*rpc_pb.HTLCAttempt, er.R) {

	if req.Route == nil {
		return nil, er.Errorf("unable to send, no routes provided")
	}

	route, err := s.cfg.RouterBackend.UnmarshallRoute(req.Route)
	if err != nil {
		return nil, err
	}

	hash, err := lntypes.MakeHash(req.PaymentHash)
	if err != nil {
		return nil, err
	}

	// Pass route to the router. This call returns the full htlc attempt
	// information as it is stored in the database. It is possible that both
	// the attempt return value and err are non-nil. This can happen when
	// the attempt was already initiated before the error happened. In that
	// case, we give precedence to the attempt information as stored in the
	// db.
	attempt, err := s.cfg.Router.SendToRoute(hash, route)
	if attempt != nil {
		rpcAttempt, err := s.cfg.RouterBackend.MarshalHTLCAttempt(
			*attempt,
		)
		if err != nil {
			return nil, err
		}
		return rpcAttempt, nil
	}
	return nil, err
}

// ResetMissionControl clears all mission control state and starts with a clean
// slate.
func (s *Server) ResetMissionControl(ctx context.Context,
	_ *rpc_pb.Null) (*rpc_pb.Null, er.R) {
	return nil, s.cfg.RouterBackend.MissionControl.ResetHistory()
}

// QueryMissionControl exposes the internal mission control state to callers. It
// is a development feature.
func (s *Server) QueryMissionControl(ctx context.Context,
	_ *rpc_pb.Null) (*routerrpc_pb.QueryMissionControlResponse, er.R) {

	snapshot := s.cfg.RouterBackend.MissionControl.GetHistorySnapshot()

	rpcPairs := make([]*routerrpc_pb.PairHistory, 0, len(snapshot.Pairs))
	for _, p := range snapshot.Pairs {
		// Prevent binding to loop variable.
		pair := p

		rpcPair := routerrpc_pb.PairHistory{
			NodeFrom: pair.Pair.From[:],
			NodeTo:   pair.Pair.To[:],
			History:  toRPCPairData(&pair.TimedPairResult),
		}

		rpcPairs = append(rpcPairs, &rpcPair)
	}

	response := routerrpc_pb.QueryMissionControlResponse{
		Pairs: rpcPairs,
	}

	return &response, nil
}

// toRPCPairData marshalls mission control pair data to the rpc struct.
func toRPCPairData(data *routing.TimedPairResult) *routerrpc_pb.PairData {
	rpcData := routerrpc_pb.PairData{
		FailAmtSat:     int64(data.FailAmt.ToSatoshis()),
		FailAmtMsat:    int64(data.FailAmt),
		SuccessAmtSat:  int64(data.SuccessAmt.ToSatoshis()),
		SuccessAmtMsat: int64(data.SuccessAmt),
	}

	if !data.FailTime.IsZero() {
		rpcData.FailTime = data.FailTime.Unix()
	}

	if !data.SuccessTime.IsZero() {
		rpcData.SuccessTime = data.SuccessTime.Unix()
	}

	return &rpcData
}

// QueryProbability returns the current success probability estimate for a
// given node pair and amount.
func (s *Server) QueryProbability(ctx context.Context,
	req *routerrpc_pb.QueryProbabilityRequest) (*routerrpc_pb.QueryProbabilityResponse, er.R) {

	fromNode, err := route.NewVertexFromBytes(req.FromNode)
	if err != nil {
		return nil, err
	}

	toNode, err := route.NewVertexFromBytes(req.ToNode)
	if err != nil {
		return nil, err
	}

	amt := lnwire.MilliSatoshi(req.AmtMsat)

	mc := s.cfg.RouterBackend.MissionControl
	prob := mc.GetProbability(fromNode, toNode, amt)
	history := mc.GetPairHistorySnapshot(fromNode, toNode)

	return &routerrpc_pb.QueryProbabilityResponse{
		Probability: prob,
		History:     toRPCPairData(&history),
	}, nil
}

// TrackPaymentV2 returns a stream of payment state updates. The stream is
// closed when the payment completes.
func (s *Server) TrackPaymentV2(request *routerrpc_pb.TrackPaymentRequest,
	stream routerrpc_pb.Router_TrackPaymentV2Server) error {

	paymentHash, err := lntypes.MakeHash(request.PaymentHash)
	if err != nil {
		return er.Native(err)
	}

	log.Debugf("TrackPayment called for payment %v", paymentHash)

	return s.trackPayment(paymentHash, stream, request.NoInflightUpdates)
}

// trackPayment writes payment status updates to the provided stream.
func (s *Server) trackPayment(paymentHash lntypes.Hash,
	stream routerrpc_pb.Router_TrackPaymentV2Server, noInflightUpdates bool) error {

	router := s.cfg.RouterBackend

	// Subscribe to the outcome of this payment.
	subscription, err := router.Tower.SubscribePayment(
		paymentHash,
	)
	switch {
	case channeldb.ErrPaymentNotInitiated.Is(err):
		return status.Error(codes.NotFound, err.String())
	case err != nil:
		return er.Native(err)
	}
	defer subscription.Close()

	// Stream updates back to the client. The first update is always the
	// current state of the payment.
	for {
		select {
		case item, ok := <-subscription.Updates:
			if !ok {
				// No more payment updates.
				return nil
			}
			result := item.(*channeldb.MPPayment)

			// Skip in-flight updates unless requested.
			if noInflightUpdates &&
				result.Status == channeldb.StatusInFlight {

				continue
			}

			rpcPayment, err := router.MarshallPayment(result)
			if err != nil {
				return er.Native(err)
			}

			// Send event to the client.
			errr := stream.Send(rpcPayment)
			if errr != nil {
				return errr
			}

		case <-s.quit:
			return er.Native(errServerShuttingDown.Default())

		case <-stream.Context().Done():
			log.Debugf("Payment status stream %v canceled", paymentHash)
			return stream.Context().Err()
		}
	}
}

// BuildRoute builds a route from a list of hop addresses.
func (s *Server) BuildRoute(ctx context.Context,
	req *routerrpc_pb.BuildRouteRequest) (*routerrpc_pb.BuildRouteResponse, er.R) {

	// Unmarshall hop list.
	hops := make([]route.Vertex, len(req.HopPubkeys))
	for i, pubkeyBytes := range req.HopPubkeys {
		pubkey, err := route.NewVertexFromBytes(pubkeyBytes)
		if err != nil {
			return nil, err
		}
		hops[i] = pubkey
	}

	// Prepare BuildRoute call parameters from rpc request.
	var amt *lnwire.MilliSatoshi
	if req.AmtMsat != 0 {
		rpcAmt := lnwire.MilliSatoshi(req.AmtMsat)
		amt = &rpcAmt
	}

	var outgoingChan *uint64
	if req.OutgoingChanId != 0 {
		outgoingChan = &req.OutgoingChanId
	}

	// Build the route and return it to the caller.
	route, err := s.cfg.Router.BuildRoute(
		amt, hops, outgoingChan, req.FinalCltvDelta,
	)
	if err != nil {
		return nil, err
	}

	rpcRoute, err := s.cfg.RouterBackend.MarshallRoute(route)
	if err != nil {
		return nil, err
	}

	routeResp := &routerrpc_pb.BuildRouteResponse{
		Route: rpcRoute,
	}

	return routeResp, nil
}

// SubscribeHtlcEvents creates a uni-directional stream from the server to
// the client which delivers a stream of htlc events.
func (s *Server) SubscribeHtlcEvents(req *routerrpc_pb.SubscribeHtlcEventsRequest,
	stream routerrpc_pb.Router_SubscribeHtlcEventsServer) error {

	htlcClient, err := s.cfg.RouterBackend.SubscribeHtlcEvents()
	if err != nil {
		return er.Native(err)
	}
	defer htlcClient.Cancel()

	for {
		select {
		case event := <-htlcClient.Updates():
			rpcEvent, err := rpcHtlcEvent(event)
			if err != nil {
				return er.Native(err)
			}

			if err := stream.Send(rpcEvent); err != nil {
				return err
			}

		// If the stream's context is cancelled, return an error.
		case <-stream.Context().Done():
			log.Debugf("htlc event stream cancelled")
			return stream.Context().Err()

		// If the subscribe client terminates, exit with an error.
		case <-htlcClient.Quit():
			return er.Native(er.New("htlc event subscription terminated"))

		// If the server has been signalled to shut down, exit.
		case <-s.quit:
			return er.Native(errServerShuttingDown.Default())
		}
	}
}

// HtlcInterceptor is a bidirectional stream for streaming interception
// requests to the caller.
// Upon connection it does the following:
// 1. Check if there is already a live stream, if yes it rejects the request.
// 2. Regsitered a ForwardInterceptor
// 3. Delivers to the caller every √√ and detect his answer.
// It uses a local implementation of holdForwardsStore to keep all the hold
// forwards and find them when manual resolution is later needed.
func (s *Server) HtlcInterceptor(stream routerrpc_pb.Router_HtlcInterceptorServer) error {
	// We ensure there is only one interceptor at a time.
	if !atomic.CompareAndSwapInt32(&s.forwardInterceptorActive, 0, 1) {
		return er.Native(ErrInterceptorAlreadyExists.Default())
	}
	defer atomic.CompareAndSwapInt32(&s.forwardInterceptorActive, 1, 0)

	// run the forward interceptor.
	return er.Native(newForwardInterceptor(s, stream).run())
}
