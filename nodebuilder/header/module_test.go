package header

import (
	"testing"

	"github.com/ipfs/go-datastore"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/p2p/net/conngater"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"

	"github.com/celestiaorg/celestia-node/header"
	"github.com/celestiaorg/celestia-node/header/p2p"
	"github.com/celestiaorg/celestia-node/header/store"
	"github.com/celestiaorg/celestia-node/nodebuilder/node"
	modp2p "github.com/celestiaorg/celestia-node/nodebuilder/p2p"
)

// TestConstructModule_StoreParams ensures that all passed via functional options
// params are set in store correctly.
func TestConstructModule_StoreParams(t *testing.T) {
	cfg := DefaultConfig(node.Light)
	cfg.Store.StoreCacheSize = 15
	cfg.Store.IndexCacheSize = 25
	cfg.Store.WriteBatchSize = 35
	var headerStore *store.Store

	app := fxtest.New(t,
		fx.Provide(func() datastore.Batching {
			return datastore.NewMapDatastore()
		}),
		ConstructModule(node.Light, &cfg),
		fx.Invoke(
			func(s header.Store) {
				ss := s.(*store.Store)
				headerStore = ss
			}),
	)
	require.NoError(t, app.Err())
	require.Equal(t, headerStore.Params.StoreCacheSize, cfg.Store.StoreCacheSize)
	require.Equal(t, headerStore.Params.IndexCacheSize, cfg.Store.IndexCacheSize)
	require.Equal(t, headerStore.Params.WriteBatchSize, cfg.Store.WriteBatchSize)
}

// TestConstructModule_ExchangeParams ensures that all passed via functional options
// params are set in store correctly.
func TestConstructModule_ExchangeParams(t *testing.T) {
	cfg := DefaultConfig(node.Light)
	cfg.Client.MinResponses = 10
	cfg.Client.MaxRequestSize = 200
	cfg.Client.MaxHeadersPerRequest = 15
	var exchange *p2p.Exchange
	var exchangeServer *p2p.ExchangeServer

	app := fxtest.New(t,
		fx.Supply(modp2p.Private),
		fx.Supply(modp2p.Bootstrappers{}),
		fx.Provide(libp2p.New),
		fx.Provide(func() datastore.Batching {
			return datastore.NewMapDatastore()
		}),
		ConstructModule(node.Light, &cfg),
		fx.Provide(func(b datastore.Batching) (*conngater.BasicConnectionGater, error) {
			return conngater.NewBasicConnectionGater(b)
		}),
		fx.Invoke(
			func(e header.Exchange, server *p2p.ExchangeServer) {
				ex := e.(*p2p.Exchange)
				exchange = ex
				exchangeServer = server
			}),
	)
	require.NoError(t, app.Err())
	require.Equal(t, exchange.Params.MinResponses, cfg.Client.MinResponses)
	require.Equal(t, exchange.Params.MaxRequestSize, cfg.Client.MaxRequestSize)
	require.Equal(t, exchange.Params.MaxHeadersPerRequest, cfg.Client.MaxHeadersPerRequest)
	require.Equal(t, exchange.Params.MaxAwaitingTime, cfg.Client.MaxAwaitingTime)
	require.Equal(t, exchange.Params.DefaultScore, cfg.Client.DefaultScore)
	require.Equal(t, exchange.Params.MaxPeerTrackerSize, cfg.Client.MaxPeerTrackerSize)

	require.Equal(t, exchangeServer.Params.WriteDeadline, cfg.Server.WriteDeadline)
	require.Equal(t, exchangeServer.Params.ReadDeadline, cfg.Server.ReadDeadline)
	require.Equal(t, exchangeServer.Params.MaxRequestSize, cfg.Server.MaxRequestSize)
}
