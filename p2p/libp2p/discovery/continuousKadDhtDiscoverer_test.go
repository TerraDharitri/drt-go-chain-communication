package discovery_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/TerraDharitri/drt-go-chain-communication/p2p"
	"github.com/TerraDharitri/drt-go-chain-communication/p2p/libp2p/discovery"
	"github.com/TerraDharitri/drt-go-chain-communication/p2p/mock"
	"github.com/TerraDharitri/drt-go-chain-communication/testscommon"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/stretchr/testify/assert"
)

var timeoutWaitResponses = 2 * time.Second

func createTestArgument() discovery.ArgKadDht {
	return discovery.ArgKadDht{
		Context:                     context.Background(),
		Host:                        &mock.ConnectableHostStub{},
		KddSharder:                  &mock.KadSharderStub{},
		PeersRefreshInterval:        time.Second,
		ProtocolID:                  "/drt/test/0.0.0",
		InitialPeersList:            []string{"peer1", "peer2"},
		BucketSize:                  100,
		RoutingTableRefresh:         5 * time.Second,
		SeedersReconnectionInterval: time.Second * 5,
		ConnectionWatcher:           &mock.ConnectionsWatcherStub{},
		Logger:                      &testscommon.LoggerStub{},
	}
}

func TestNewContinuousKadDhtDiscoverer(t *testing.T) {
	t.Parallel()

	t.Run("nil context should error", func(t *testing.T) {
		t.Parallel()

		arg := createTestArgument()
		arg.Context = nil

		kdd, err := discovery.NewContinuousKadDhtDiscoverer(arg)

		assert.Nil(t, kdd)
		assert.True(t, errors.Is(err, p2p.ErrNilContext))
	})
	t.Run("nil host should error", func(t *testing.T) {
		t.Parallel()

		arg := createTestArgument()
		arg.Host = nil

		kdd, err := discovery.NewContinuousKadDhtDiscoverer(arg)

		assert.Nil(t, kdd)
		assert.True(t, errors.Is(err, p2p.ErrNilHost))
	})
	t.Run("nil sharder should error", func(t *testing.T) {
		t.Parallel()

		arg := createTestArgument()
		arg.KddSharder = nil

		kdd, err := discovery.NewContinuousKadDhtDiscoverer(arg)

		assert.Nil(t, kdd)
		assert.True(t, errors.Is(err, p2p.ErrNilSharder))
	})
	t.Run("wrong sharder should error", func(t *testing.T) {
		t.Parallel()

		arg := createTestArgument()
		arg.KddSharder = &mock.SharderStub{}

		kdd, err := discovery.NewContinuousKadDhtDiscoverer(arg)

		assert.Nil(t, kdd)
		assert.True(t, errors.Is(err, p2p.ErrWrongTypeAssertion))
	})
	t.Run("invalid peers refresh interval should error", func(t *testing.T) {
		t.Parallel()

		arg := createTestArgument()
		arg.PeersRefreshInterval = time.Second - time.Microsecond

		kdd, err := discovery.NewContinuousKadDhtDiscoverer(arg)

		assert.Nil(t, kdd)
		assert.True(t, errors.Is(err, p2p.ErrInvalidValue))
	})
	t.Run("invalid routing table refresh interval should error", func(t *testing.T) {
		t.Parallel()

		arg := createTestArgument()
		arg.RoutingTableRefresh = time.Second - time.Microsecond

		kdd, err := discovery.NewContinuousKadDhtDiscoverer(arg)

		assert.Nil(t, kdd)
		assert.True(t, errors.Is(err, p2p.ErrInvalidValue))
	})
	t.Run("nil connections watcher should error", func(t *testing.T) {
		t.Parallel()

		arg := createTestArgument()
		arg.ConnectionWatcher = nil

		kdd, err := discovery.NewContinuousKadDhtDiscoverer(arg)

		assert.Nil(t, kdd)
		assert.True(t, errors.Is(err, p2p.ErrNilConnectionsWatcher))
	})
	t.Run("nil logger should error", func(t *testing.T) {
		t.Parallel()

		arg := createTestArgument()
		arg.Logger = nil

		kdd, err := discovery.NewContinuousKadDhtDiscoverer(arg)

		assert.Nil(t, kdd)
		assert.True(t, errors.Is(err, p2p.ErrNilLogger))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		arg := createTestArgument()

		kdd, err := discovery.NewContinuousKadDhtDiscoverer(arg)

		assert.NotNil(t, kdd)
		assert.Nil(t, err)
	})
}

func TestNewContinuousKadDhtDiscoverer_EmptyInitialPeersShouldWork(t *testing.T) {
	t.Parallel()

	arg := createTestArgument()
	arg.InitialPeersList = nil

	kdd, err := discovery.NewContinuousKadDhtDiscoverer(arg)

	assert.NotNil(t, kdd)
	assert.Nil(t, err)
}

// ------- Bootstrap

func TestContinuousKadDhtDiscoverer_BootstrapCalledOnceShouldWork(t *testing.T) {
	t.Parallel()

	arg := createTestArgument()
	ckdd, _ := discovery.NewContinuousKadDhtDiscoverer(arg)

	err := ckdd.Bootstrap()

	assert.Nil(t, err)
	time.Sleep(arg.PeersRefreshInterval * 2)
}

func TestContinuousKadDhtDiscoverer_BootstrapCalledTwiceShouldErr(t *testing.T) {
	t.Parallel()

	arg := createTestArgument()
	ckdd, _ := discovery.NewContinuousKadDhtDiscoverer(arg)

	_ = ckdd.Bootstrap()
	err := ckdd.Bootstrap()

	assert.Equal(t, p2p.ErrPeerDiscoveryProcessAlreadyStarted, err)
}

// ------- connectToOnePeerFromInitialPeersList

func TestContinuousKadDhtDiscoverer_ConnectToOnePeerFromInitialPeersListNilListShouldRetWithChanFull(t *testing.T) {
	t.Parallel()

	arg := createTestArgument()
	ckdd, _ := discovery.NewContinuousKadDhtDiscoverer(arg)

	chanDone := ckdd.ConnectToOnePeerFromInitialPeersList(time.Second, nil)

	assert.Equal(t, 1, len(chanDone))
}

func TestContinuousKadDhtDiscoverer_ConnectToOnePeerFromInitialPeersListEmptyListShouldRetWithChanFull(t *testing.T) {
	t.Parallel()

	arg := createTestArgument()
	ckdd, _ := discovery.NewContinuousKadDhtDiscoverer(arg)

	chanDone := ckdd.ConnectToOnePeerFromInitialPeersList(time.Second, make([]string, 0))

	assert.Equal(t, 1, len(chanDone))
}

func TestContinuousKadDhtDiscoverer_ConnectToOnePeerFromInitialPeersOnePeerShouldTryToConnect(t *testing.T) {
	t.Parallel()

	arg := createTestArgument()
	peerID := "peer"
	wasConnectCalled := int32(0)

	arg.Host = &mock.ConnectableHostStub{
		ConnectToPeerCalled: func(ctx context.Context, address string) error {
			if peerID == address {
				atomic.AddInt32(&wasConnectCalled, 1)
			}

			return nil
		},
		EventBusCalled: func() event.Bus {
			return &mock.EventBusStub{
				SubscribeCalled: func(eventType interface{}, opts ...event.SubscriptionOpt) (event.Subscription, error) {
					return &mock.EventSubscriptionStub{}, nil
				},
			}
		},
	}
	ckdd, _ := discovery.NewContinuousKadDhtDiscoverer(arg)
	chanDone := ckdd.ConnectToOnePeerFromInitialPeersList(time.Second, []string{peerID})

	select {
	case <-chanDone:
		assert.Equal(t, int32(1), atomic.LoadInt32(&wasConnectCalled))
	case <-time.After(timeoutWaitResponses):
		assert.Fail(t, "timeout")
	}
}

func TestContinuousKadDhtDiscoverer_ConnectToOnePeerFromInitialPeersOnePeerShouldTryToConnectContinously(t *testing.T) {
	t.Parallel()

	arg := createTestArgument()
	peerID := "peer"
	wasConnectCalled := int32(0)

	errDidNotConnect := errors.New("did not connect")
	noOfTimesToRefuseConnection := 5
	arg.Host = &mock.ConnectableHostStub{
		ConnectToPeerCalled: func(ctx context.Context, address string) error {
			if peerID != address {
				assert.Fail(t, "should have tried to connect to the same ID")
			}

			atomic.AddInt32(&wasConnectCalled, 1)

			if atomic.LoadInt32(&wasConnectCalled) < int32(noOfTimesToRefuseConnection) {
				return errDidNotConnect
			}

			return nil
		},
		EventBusCalled: func() event.Bus {
			return &mock.EventBusStub{
				SubscribeCalled: func(eventType interface{}, opts ...event.SubscriptionOpt) (event.Subscription, error) {
					return &mock.EventSubscriptionStub{}, nil
				},
			}
		},
	}
	ckdd, _ := discovery.NewContinuousKadDhtDiscoverer(arg)

	chanDone := ckdd.ConnectToOnePeerFromInitialPeersList(time.Millisecond*10, []string{peerID})

	select {
	case <-chanDone:
		assert.Equal(t, int32(noOfTimesToRefuseConnection), atomic.LoadInt32(&wasConnectCalled))
	case <-time.After(timeoutWaitResponses):
		assert.Fail(t, "timeout")
	}
}

func TestContinuousKadDhtDiscoverer_ConnectToOnePeerFromInitialPeersTwoPeersShouldAlternate(t *testing.T) {
	t.Parallel()

	arg := createTestArgument()
	peerID1 := "peer1"
	peerID2 := "peer2"
	wasConnectCalled := int32(0)
	errDidNotConnect := errors.New("did not connect")
	noOfTimesToRefuseConnection := 5
	arg.Host = &mock.ConnectableHostStub{
		ConnectToPeerCalled: func(ctx context.Context, address string) error {
			connCalled := atomic.LoadInt32(&wasConnectCalled)

			atomic.AddInt32(&wasConnectCalled, 1)

			if connCalled >= int32(noOfTimesToRefuseConnection) {
				return nil
			}

			connCalled = connCalled % 2
			if connCalled == 0 {
				if peerID1 != address {
					assert.Fail(t, "should have tried to connect to "+peerID1)
				}
			}

			if connCalled == 1 {
				if peerID2 != address {
					assert.Fail(t, "should have tried to connect to "+peerID2)
				}
			}

			return errDidNotConnect
		},
		EventBusCalled: func() event.Bus {
			return &mock.EventBusStub{
				SubscribeCalled: func(eventType interface{}, opts ...event.SubscriptionOpt) (event.Subscription, error) {
					return &mock.EventSubscriptionStub{}, nil
				},
			}
		},
	}
	ckdd, _ := discovery.NewContinuousKadDhtDiscoverer(arg)

	chanDone := ckdd.ConnectToOnePeerFromInitialPeersList(time.Millisecond*10, []string{peerID1, peerID2})

	select {
	case <-chanDone:
	case <-time.After(timeoutWaitResponses):
		assert.Fail(t, "timeout")
	}
}

func TestContinuousKadDhtDiscoverer_Name(t *testing.T) {
	t.Parallel()

	arg := createTestArgument()
	kdd, _ := discovery.NewContinuousKadDhtDiscoverer(arg)

	assert.Equal(t, discovery.KadDhtName, kdd.Name())
}

func TestContinuousKadDhtDiscoverer_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	arg := createTestArgument()
	arg.Logger = nil
	kdd, _ := discovery.NewContinuousKadDhtDiscoverer(arg)
	assert.True(t, kdd.IsInterfaceNil())

	kdd, _ = discovery.NewContinuousKadDhtDiscoverer(createTestArgument())
	assert.False(t, kdd.IsInterfaceNil())
}
