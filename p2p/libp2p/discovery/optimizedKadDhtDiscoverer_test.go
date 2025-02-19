package discovery_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/TerraDharitri/drt-go-chain-communication/p2p"
	"github.com/TerraDharitri/drt-go-chain-communication/p2p/libp2p/discovery"
	"github.com/TerraDharitri/drt-go-chain-communication/p2p/mock"
	"github.com/TerraDharitri/drt-go-chain-core/core/check"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
)

func TestNewOptimizedKadDhtDiscoverer(t *testing.T) {
	t.Parallel()

	t.Run("invalid argument should error", func(t *testing.T) {
		t.Parallel()

		arg := createTestArgument()
		arg.Host = nil
		okdd, err := discovery.NewOptimizedKadDhtDiscoverer(arg)
		assert.Equal(t, p2p.ErrNilHost, err)
		assert.True(t, check.IfNil(okdd))

		arg = createTestArgument()
		arg.SeedersReconnectionInterval = 0
		okdd, err = discovery.NewOptimizedKadDhtDiscoverer(arg)
		assert.Equal(t, p2p.ErrInvalidSeedersReconnectionInterval, err)
		assert.True(t, check.IfNil(okdd))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		arg := createTestArgument()
		var cancelFunc func()
		arg.Context, cancelFunc = context.WithCancel(context.Background())
		okdd, err := discovery.NewOptimizedKadDhtDiscoverer(arg)

		assert.Nil(t, err)
		assert.False(t, check.IfNil(okdd))
		cancelFunc()

		assert.Equal(t, discovery.OptimizedKadDhtName, okdd.Name())
	})
}

func TestOptimizedKadDhtDiscoverer_BootstrapWithRealKadDhtFuncShouldNotError(t *testing.T) {
	t.Parallel()

	arg := createTestArgument()
	arg.InitialPeersList = make([]string, 0)
	var cancelFunc func()
	arg.Context, cancelFunc = context.WithCancel(context.Background())
	okdd, _ := discovery.NewOptimizedKadDhtDiscoverer(arg)

	err := okdd.Bootstrap()

	assert.Nil(t, err)
	cancelFunc()
}

func TestOptimizedKadDhtDiscoverer_BootstrapEmptyPeerListShouldStartBootstrap(t *testing.T) {
	t.Parallel()

	arg := createTestArgument()
	arg.InitialPeersList = make([]string, 0)
	var cancelFunc func()
	arg.Context, cancelFunc = context.WithCancel(context.Background())
	bootstrapCalled := uint32(0)
	kadDhtStub := &mock.KadDhtHandlerStub{
		BootstrapCalled: func(ctx context.Context) error {
			atomic.AddUint32(&bootstrapCalled, 1)
			return nil
		},
	}

	okdd, _ := discovery.NewOptimizedKadDhtDiscovererWithInitFunc(
		arg,
		func(ctx context.Context) (discovery.KadDhtHandler, error) {
			return kadDhtStub, nil
		},
	)

	err := okdd.Bootstrap()
	// a little delay as the bootstrap returns immediately after init. The seeder reconnection and bootstrap part
	// are called async
	time.Sleep(time.Second + time.Millisecond*500) // the value is chosen as such as to avoid edgecases on select statement

	assert.Nil(t, err)
	assert.Equal(t, uint32(2), atomic.LoadUint32(&bootstrapCalled))
	cancelFunc()
}

func TestOptimizedKadDhtDiscoverer_BootstrapWithPeerListShouldStartBootstrap(t *testing.T) {
	t.Parallel()

	arg := createTestArgument()
	arg.SeedersReconnectionInterval = time.Second
	bootstrapCalled := uint32(0)
	connectCalled := uint32(0)
	arg.Host = &mock.ConnectableHostStub{
		ConnectCalled: func(ctx context.Context, pi peer.AddrInfo) error {
			atomic.AddUint32(&connectCalled, 1)
			return nil
		},
		AddressToPeerInfoCalled: func(address string) (*peer.AddrInfo, error) {
			return &peer.AddrInfo{}, nil
		},
	}
	var cancelFunc func()
	arg.Context, cancelFunc = context.WithCancel(context.Background())

	kadDhtStub := &mock.KadDhtHandlerStub{
		BootstrapCalled: func(ctx context.Context) error {
			atomic.AddUint32(&bootstrapCalled, 1)
			return nil
		},
	}

	okdd, _ := discovery.NewOptimizedKadDhtDiscovererWithInitFunc(
		arg,
		func(ctx context.Context) (discovery.KadDhtHandler, error) {
			return kadDhtStub, nil
		},
	)

	err := okdd.Bootstrap()
	time.Sleep(time.Second*4 + time.Millisecond*500) // the value is chosen as such as to avoid edgecases on select statement
	cancelFunc()

	assert.Nil(t, err)
	assert.Equal(t, uint32(5), atomic.LoadUint32(&bootstrapCalled))
	assert.Equal(t, uint32(10), atomic.LoadUint32(&connectCalled))
}

func TestOptimizedKadDhtDiscoverer_BootstrapErrorsShouldKeepRetrying(t *testing.T) {
	t.Parallel()

	arg := createTestArgument()
	var cancelFunc func()
	arg.Context, cancelFunc = context.WithCancel(context.Background())
	bootstrapCalled := uint32(0)
	expectedErr := errors.New("expected error")
	kadDhtStub := &mock.KadDhtHandlerStub{
		BootstrapCalled: func(ctx context.Context) error {
			atomic.AddUint32(&bootstrapCalled, 1)
			return expectedErr
		},
	}

	okdd, _ := discovery.NewOptimizedKadDhtDiscovererWithInitFunc(
		arg,
		func(ctx context.Context) (discovery.KadDhtHandler, error) {
			return kadDhtStub, nil
		},
	)

	err := okdd.Bootstrap()
	// a little delay as the bootstrap returns immediately after init. The seeder reconnection and bootstrap part
	// are called async
	time.Sleep(time.Second*4 + time.Millisecond*500) // the value is chosen as such as to avoid edgecases on select statement

	assert.Nil(t, err)
	assert.Equal(t, uint32(5), atomic.LoadUint32(&bootstrapCalled))
	cancelFunc()
}

func TestOptimizedKadDhtDiscoverer_BootstrapErrorsForSeedersShouldRetryFast(t *testing.T) {
	t.Parallel()

	numConnectCalls := uint32(0)
	arg := createTestArgument()
	arg.Host = &mock.ConnectableHostStub{
		ConnectCalled: func(ctx context.Context, pi peer.AddrInfo) error {
			atomic.AddUint32(&numConnectCalls, 1)
			return errors.New("cannot connect")
		},
	}
	arg.InitialPeersList = []string{"/ip4/127.0.0.1/tcp/9999/p2p/16Uiu2HAkw5SNNtSvH1zJiQ6Gc3WoGNSxiyNueRKe6fuAuh57G3Bk"}
	var cancelFunc func()
	arg.Context, cancelFunc = context.WithCancel(context.Background())
	kadDhtStub := &mock.KadDhtHandlerStub{
		BootstrapCalled: func(ctx context.Context) error {
			return nil
		},
	}

	okdd, _ := discovery.NewOptimizedKadDhtDiscovererWithInitFunc(
		arg,
		func(ctx context.Context) (discovery.KadDhtHandler, error) {
			return kadDhtStub, nil
		},
	)

	err := okdd.Bootstrap()
	// a little delay as the bootstrap returns immediately after init. The seeder reconnection and bootstrap part
	// are called async
	time.Sleep(time.Second*4 + time.Millisecond*500) // the value is chosen as such as to avoid edgecases on select statement

	assert.Nil(t, err)
	assert.True(t, atomic.LoadUint32(&numConnectCalls) > 1)
	cancelFunc()
}

func TestOptimizedKadDhtDiscoverer_ReconnectToNetwork(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	t.Parallel()

	arg := createTestArgument()
	var cancelFunc func()
	arg.Context, cancelFunc = context.WithCancel(context.Background())
	bootstrapCalled := uint32(0)
	expectedErr := errors.New("expected error")
	mutConnect := sync.Mutex{}
	connectCalled := 0
	arg.Host = &mock.ConnectableHostStub{
		ConnectCalled: func(ctx context.Context, pi peer.AddrInfo) error {
			mutConnect.Lock()
			defer mutConnect.Unlock()

			connectCalled++

			return nil
		},
		AddressToPeerInfoCalled: func(address string) (*peer.AddrInfo, error) {
			return &peer.AddrInfo{}, nil
		},
	}
	kadDhtStub := &mock.KadDhtHandlerStub{
		BootstrapCalled: func(ctx context.Context) error {
			atomic.AddUint32(&bootstrapCalled, 1)
			return expectedErr
		},
	}

	okdd, _ := discovery.NewOptimizedKadDhtDiscovererWithInitFunc(
		arg,
		func(ctx context.Context) (discovery.KadDhtHandler, error) {
			return kadDhtStub, nil
		},
	)

	err := okdd.Bootstrap()
	time.Sleep(time.Second)
	okdd.ReconnectToNetwork(context.Background())
	time.Sleep(time.Millisecond * 500) // the value is chosen as such as to avoid edge cases on select statement
	cancelFunc()

	assert.Nil(t, err)
	assert.Equal(t, uint32(2), atomic.LoadUint32(&bootstrapCalled))
	mutConnect.Lock()
	assert.True(t, connectCalled > 0)
	mutConnect.Unlock()
}
