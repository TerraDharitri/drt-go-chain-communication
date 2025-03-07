package connectionMonitor

import (
	"context"
	"sync"
	"time"

	"github.com/TerraDharitri/drt-go-chain-communication/p2p"
	"github.com/TerraDharitri/drt-go-chain-communication/p2p/libp2p/disabled"
	"github.com/TerraDharitri/drt-go-chain-core/core"
	"github.com/TerraDharitri/drt-go-chain-core/core/atomic"
	"github.com/TerraDharitri/drt-go-chain-core/core/check"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/multiformats/go-multiaddr"
)

const (
	durationBetweenReconnectAttempts = time.Second * 5
	durationCheckConnections         = time.Second
)

type libp2pConnectionMonitorSimple struct {
	chDoReconnect              chan struct{}
	reconnecters               []p2p.Reconnecter
	thresholdMinConnectedPeers int
	sharder                    Sharder
	preferredPeersHolder       p2p.PreferredPeersHolderHandler
	cancelFunc                 context.CancelFunc
	connectionsWatcher         p2p.ConnectionsWatcher
	network                    network.Network
	mutPeerDenialEvaluator     sync.RWMutex
	peerDenialEvaluator        p2p.PeerDenialEvaluator
	log                        p2p.Logger
}

// ArgsConnectionMonitorSimple is the DTO used in the NewLibp2pConnectionMonitorSimple constructor function
type ArgsConnectionMonitorSimple struct {
	Reconnecters               []p2p.Reconnecter
	ThresholdMinConnectedPeers uint32
	Sharder                    Sharder
	PreferredPeersHolder       p2p.PreferredPeersHolderHandler
	ConnectionsWatcher         p2p.ConnectionsWatcher
	Network                    network.Network
	Logger                     p2p.Logger
}

// NewLibp2pConnectionMonitorSimple creates a new connection monitor (version 2 that is more streamlined and does not care
// about pausing and resuming the discovery process)
// it also handles black listed peers
func NewLibp2pConnectionMonitorSimple(args ArgsConnectionMonitorSimple) (*libp2pConnectionMonitorSimple, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	ctx, cancelFunc := context.WithCancel(context.Background())

	cm := &libp2pConnectionMonitorSimple{
		reconnecters:               args.Reconnecters,
		chDoReconnect:              make(chan struct{}),
		thresholdMinConnectedPeers: int(args.ThresholdMinConnectedPeers),
		sharder:                    args.Sharder,
		cancelFunc:                 cancelFunc,
		preferredPeersHolder:       args.PreferredPeersHolder,
		connectionsWatcher:         args.ConnectionsWatcher,
		network:                    args.Network,
		peerDenialEvaluator:        &disabled.PeerDenialEvaluator{},
		log:                        args.Logger,
	}

	cm.network.Notify(cm)

	go cm.processLoop(ctx)

	return cm, nil
}

func checkArgs(args ArgsConnectionMonitorSimple) error {
	for _, reconnecter := range args.Reconnecters {
		if check.IfNil(reconnecter) {
			return p2p.ErrNilReconnecter
		}
	}

	if check.IfNil(args.Sharder) {
		return p2p.ErrNilSharder
	}
	if check.IfNil(args.PreferredPeersHolder) {
		return p2p.ErrNilPreferredPeersHolder
	}
	if check.IfNil(args.ConnectionsWatcher) {
		return p2p.ErrNilConnectionsWatcher
	}
	if check.IfNilReflect(args.Network) {
		return p2p.ErrNilNetwork
	}
	if check.IfNilReflect(args.Logger) {
		return p2p.ErrNilLogger
	}

	return nil
}

// Listen is called when network starts listening on an addr
func (lcms *libp2pConnectionMonitorSimple) Listen(network.Network, multiaddr.Multiaddr) {}

// ListenClose is called when network stops listening on an addr
func (lcms *libp2pConnectionMonitorSimple) ListenClose(network.Network, multiaddr.Multiaddr) {}

// Request a reconnect to initial list
func (lcms *libp2pConnectionMonitorSimple) doReconn() {
	select {
	case lcms.chDoReconnect <- struct{}{}:
	default:
	}
}

// Connected is called when a connection opened
func (lcms *libp2pConnectionMonitorSimple) Connected(netw network.Network, conn network.Conn) {
	lcms.mutPeerDenialEvaluator.RLock()
	peerDenialEvaluator := lcms.peerDenialEvaluator
	lcms.mutPeerDenialEvaluator.RUnlock()

	pid := conn.RemotePeer()
	if peerDenialEvaluator.IsDenied(core.PeerID(pid)) {
		lcms.log.Trace("dropping connection to blacklisted peer",
			"pid", pid.String(),
		)
		_ = conn.Close()

		return
	}

	allPeers := netw.Peers()

	peerId := core.PeerID(conn.RemotePeer())
	connectionStr := conn.RemoteMultiaddr().String()
	lcms.connectionsWatcher.NewKnownConnection(peerId, connectionStr)
	lcms.preferredPeersHolder.PutConnectionAddress(peerId, connectionStr)

	evictedList := lcms.sharder.ComputeEvictionList(allPeers)
	for _, evictedPID := range evictedList {
		_ = netw.ClosePeer(evictedPID)
	}
}

// Disconnected is called when a connection closed
func (lcms *libp2pConnectionMonitorSimple) Disconnected(netw network.Network, conn network.Conn) {
	if conn != nil {
		lcms.preferredPeersHolder.Remove(core.PeerID(conn.ID()))
	}

	lcms.doReconnectionIfNeeded(netw)
}

func (lcms *libp2pConnectionMonitorSimple) doReconnectionIfNeeded(netw network.Network) {
	if !lcms.IsConnectedToTheNetwork(netw) {
		lcms.doReconn()
	}
}

func (lcms *libp2pConnectionMonitorSimple) processLoop(ctx context.Context) {
	timerCheckConnections := time.NewTimer(durationCheckConnections)
	timerBetweenReconnectAttempts := time.NewTimer(durationBetweenReconnectAttempts)
	defer func() {
		lcms.log.Debug("closing the connection monitor main loop")
		timerCheckConnections.Stop()
		timerBetweenReconnectAttempts.Stop()
	}()

	canReconnect := atomic.Flag{}
	canReconnect.SetValue(true)
	for {
		select {
		case <-timerCheckConnections.C:
			lcms.checkConnectionsBlocking()
			timerCheckConnections.Reset(durationCheckConnections)
		case <-lcms.chDoReconnect:
			if !canReconnect.IsSet() {
				lcms.log.Debug("too early for a new reconnect to network attempt")
				continue
			}

			lcms.log.Debug("reconnecting to network...")
			lcms.reconnectToNetwork(ctx)
			timerBetweenReconnectAttempts.Reset(durationBetweenReconnectAttempts)
			canReconnect.SetValue(false)
		case <-timerBetweenReconnectAttempts.C:
			canReconnect.SetValue(true)
		case <-ctx.Done():
			return
		}
	}
}

func (lcms *libp2pConnectionMonitorSimple) reconnectToNetwork(ctx context.Context) {
	for _, reconnecter := range lcms.reconnecters {
		reconnecter.ReconnectToNetwork(ctx)
	}
}

// IsConnectedToTheNetwork returns true if the number of connected peer is at least equal with thresholdMinConnectedPeers
func (lcms *libp2pConnectionMonitorSimple) IsConnectedToTheNetwork(netw network.Network) bool {
	return len(netw.Peers()) >= lcms.thresholdMinConnectedPeers
}

// SetThresholdMinConnectedPeers sets the minimum connected peers number when the node is considered connected on the network
func (lcms *libp2pConnectionMonitorSimple) SetThresholdMinConnectedPeers(thresholdMinConnectedPeers int, netw network.Network) {
	if check.IfNilReflect(netw) {
		return
	}
	lcms.thresholdMinConnectedPeers = thresholdMinConnectedPeers
	lcms.doReconnectionIfNeeded(netw)
}

// ThresholdMinConnectedPeers returns the minimum connected peers number when the node is considered connected on the network
func (lcms *libp2pConnectionMonitorSimple) ThresholdMinConnectedPeers() int {
	return lcms.thresholdMinConnectedPeers
}

// SetPeerDenialEvaluator sets the handler that is able to tell if a peer can connect to self or not (is or not blacklisted)
func (lcms *libp2pConnectionMonitorSimple) SetPeerDenialEvaluator(handler p2p.PeerDenialEvaluator) error {
	if check.IfNil(handler) {
		return p2p.ErrNilPeerDenialEvaluator
	}

	lcms.mutPeerDenialEvaluator.Lock()
	lcms.peerDenialEvaluator = handler
	lcms.mutPeerDenialEvaluator.Unlock()

	return nil
}

// PeerDenialEvaluator gets the peer denial evaluator
func (lcms *libp2pConnectionMonitorSimple) PeerDenialEvaluator() p2p.PeerDenialEvaluator {
	lcms.mutPeerDenialEvaluator.RLock()
	defer lcms.mutPeerDenialEvaluator.RUnlock()

	return lcms.peerDenialEvaluator
}

// Close closes all underlying components
func (lcms *libp2pConnectionMonitorSimple) Close() error {
	lcms.cancelFunc()
	return nil
}

// checkConnectionsBlocking does a peer sweep, calling Close on those peers that are black listed
func (lcms *libp2pConnectionMonitorSimple) checkConnectionsBlocking() {
	peers := lcms.network.Peers()
	lcms.mutPeerDenialEvaluator.RLock()
	peerDenialEvaluator := lcms.peerDenialEvaluator
	lcms.mutPeerDenialEvaluator.RUnlock()

	for _, pid := range peers {
		if peerDenialEvaluator.IsDenied(core.PeerID(pid)) {
			lcms.log.Trace("dropping connection to blacklisted peer",
				"pid", pid.String(),
			)
			_ = lcms.network.ClosePeer(pid)
		}
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (lcms *libp2pConnectionMonitorSimple) IsInterfaceNil() bool {
	return lcms == nil
}
