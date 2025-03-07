package factory_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/TerraDharitri/drt-go-chain-communication/p2p"
	"github.com/TerraDharitri/drt-go-chain-communication/p2p/config"
	"github.com/TerraDharitri/drt-go-chain-communication/p2p/libp2p/discovery"
	"github.com/TerraDharitri/drt-go-chain-communication/p2p/libp2p/discovery/factory"
	"github.com/TerraDharitri/drt-go-chain-communication/p2p/mock"
	"github.com/TerraDharitri/drt-go-chain-communication/testscommon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPeerDiscoverers_NilLoggerShouldError(t *testing.T) {
	t.Parallel()

	args := factory.ArgsPeerDiscoverer{
		Context:            context.Background(),
		Host:               &mock.ConnectableHostStub{},
		Sharder:            &mock.SharderStub{},
		P2pConfig:          config.P2PConfig{},
		ConnectionsWatcher: &mock.ConnectionsWatcherStub{},
		Logger:             nil,
	}
	pDiscoverer, err := factory.NewPeerDiscoverers(args)
	assert.Equal(t, p2p.ErrNilLogger, err)
	assert.Nil(t, pDiscoverer)
}
func TestNewPeerDiscoverers_NoDiscoveryEnabledShouldRetNullDiscoverer(t *testing.T) {
	t.Parallel()

	args := factory.ArgsPeerDiscoverer{
		Context: context.Background(),
		Host:    &mock.ConnectableHostStub{},
		Sharder: &mock.SharderStub{},
		P2pConfig: config.P2PConfig{
			KadDhtPeerDiscovery: config.KadDhtPeerDiscoveryConfig{
				Enabled: false,
			},
		},
		ConnectionsWatcher: &mock.ConnectionsWatcherStub{},
		Logger:             &testscommon.LoggerStub{},
	}
	pDiscoverers, err := factory.NewPeerDiscoverers(args)
	require.Equal(t, 1, len(pDiscoverers))
	_, ok := pDiscoverers[0].(*discovery.NilDiscoverer)

	assert.True(t, ok)
	assert.Nil(t, err)
}

func TestNewPeerDiscoverer_ListsSharderShouldWork(t *testing.T) {
	t.Parallel()

	args := factory.ArgsPeerDiscoverer{
		Context: context.Background(),
		Host:    &mock.ConnectableHostStub{},
		Sharder: &mock.KadSharderStub{},
		P2pConfig: config.P2PConfig{
			KadDhtPeerDiscovery: config.KadDhtPeerDiscoveryConfig{
				Enabled:                          true,
				RefreshIntervalInSec:             1,
				RoutingTableRefreshIntervalInSec: 300,
				Type:                             "legacy",
				ProtocolIDs:                      []string{"protocol1", "protocol2"},
			},
			Sharding: config.ShardingConfig{
				Type: p2p.ListsSharder,
			},
		},
		ConnectionsWatcher: &mock.ConnectionsWatcherStub{},
		Logger:             &testscommon.LoggerStub{},
	}

	pDiscoverers, err := factory.NewPeerDiscoverers(args)
	assert.Equal(t, 2, len(pDiscoverers))
	assert.NotNil(t, pDiscoverers[0])
	assert.NotNil(t, pDiscoverers[1])
	assert.Nil(t, err)

	assert.Equal(t, "*discovery.continuousKadDhtDiscoverer", fmt.Sprintf("%T", pDiscoverers[0]))
	assert.Equal(t, "*discovery.continuousKadDhtDiscoverer", fmt.Sprintf("%T", pDiscoverers[1]))
}

func TestNewPeerDiscoverer_OptimizedKadDhtShouldWork(t *testing.T) {
	t.Parallel()

	args := factory.ArgsPeerDiscoverer{
		Context: context.Background(),
		Host:    &mock.ConnectableHostStub{},
		Sharder: &mock.KadSharderStub{},
		P2pConfig: config.P2PConfig{
			KadDhtPeerDiscovery: config.KadDhtPeerDiscoveryConfig{
				Enabled:                          true,
				RefreshIntervalInSec:             1,
				RoutingTableRefreshIntervalInSec: 300,
				Type:                             "optimized",
				ProtocolIDs:                      []string{"protocol1", "protocol2"},
			},
			Sharding: config.ShardingConfig{
				Type: p2p.ListsSharder,
			},
		},
		ConnectionsWatcher: &mock.ConnectionsWatcherStub{},
		Logger:             &testscommon.LoggerStub{},
	}
	pDiscoverers, err := factory.NewPeerDiscoverers(args)
	assert.Equal(t, 2, len(pDiscoverers))
	assert.NotNil(t, pDiscoverers[0])
	assert.NotNil(t, pDiscoverers[1])
	assert.Nil(t, err)

	assert.Equal(t, "*discovery.optimizedKadDhtDiscoverer", fmt.Sprintf("%T", pDiscoverers[0]))
	assert.Equal(t, "*discovery.optimizedKadDhtDiscoverer", fmt.Sprintf("%T", pDiscoverers[1]))
}

func TestNewPeerDiscoverer_OptimizedKadDhtWithoutProtocolIDsShouldError(t *testing.T) {
	t.Parallel()

	args := factory.ArgsPeerDiscoverer{
		Context: context.Background(),
		Host:    &mock.ConnectableHostStub{},
		Sharder: &mock.KadSharderStub{},
		P2pConfig: config.P2PConfig{
			KadDhtPeerDiscovery: config.KadDhtPeerDiscoveryConfig{
				Enabled:                          true,
				RefreshIntervalInSec:             1,
				RoutingTableRefreshIntervalInSec: 300,
				Type:                             "optimized",
				ProtocolIDs:                      make([]string, 0),
			},
			Sharding: config.ShardingConfig{
				Type: p2p.ListsSharder,
			},
		},
		ConnectionsWatcher: &mock.ConnectionsWatcherStub{},
		Logger:             &testscommon.LoggerStub{},
	}
	pDiscoverers, err := factory.NewPeerDiscoverers(args)
	assert.Nil(t, pDiscoverers)
	assert.ErrorIs(t, err, p2p.ErrInvalidConfig)
	assert.Contains(t, err.Error(), "KadDhtPeerDiscovery.Enabled is enabled but no protocol ID was provided")
}

func TestNewPeerDiscoverer_UnknownSharderShouldErr(t *testing.T) {
	t.Parallel()

	args := factory.ArgsPeerDiscoverer{
		Context: context.Background(),
		Host:    &mock.ConnectableHostStub{},
		Sharder: &mock.SharderStub{},
		P2pConfig: config.P2PConfig{
			KadDhtPeerDiscovery: config.KadDhtPeerDiscoveryConfig{
				Enabled:              true,
				RefreshIntervalInSec: 1,
				ProtocolIDs:          []string{"protocol1"},
			},
			Sharding: config.ShardingConfig{
				Type: "unknown",
			},
		},
		ConnectionsWatcher: &mock.ConnectionsWatcherStub{},
		Logger:             &testscommon.LoggerStub{},
	}

	pDiscoverers, err := factory.NewPeerDiscoverers(args)

	assert.Nil(t, pDiscoverers)
	assert.ErrorIs(t, err, p2p.ErrInvalidValue)
}

func TestNewPeerDiscoverer_UnknownKadDhtShouldErr(t *testing.T) {
	t.Parallel()

	args := factory.ArgsPeerDiscoverer{
		Context: context.Background(),
		Host:    &mock.ConnectableHostStub{},
		Sharder: &mock.SharderStub{},
		P2pConfig: config.P2PConfig{
			KadDhtPeerDiscovery: config.KadDhtPeerDiscoveryConfig{
				Enabled:                          true,
				RefreshIntervalInSec:             1,
				RoutingTableRefreshIntervalInSec: 300,
				Type:                             "unknown",
				ProtocolIDs:                      []string{"protocol"},
			},
			Sharding: config.ShardingConfig{
				Type: p2p.ListsSharder,
			},
		},
		ConnectionsWatcher: &mock.ConnectionsWatcherStub{},
		Logger:             &testscommon.LoggerStub{},
	}

	pDiscoverers, err := factory.NewPeerDiscoverers(args)

	assert.ErrorIs(t, err, p2p.ErrInvalidValue)
	assert.Nil(t, pDiscoverers)
}
