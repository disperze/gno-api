package api

import (
	"context"
	"net"
	"net/http"

	tmconfig "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/libs/log"
	tmpubsub "github.com/tendermint/tendermint/libs/pubsub"
	rpccore "github.com/tendermint/tendermint/rpc/core"
	rpcserver "github.com/tendermint/tendermint/rpc/jsonrpc/server"
	"github.com/tendermint/tendermint/state/txindex"
	ttypes "github.com/tendermint/tendermint/types"
)

func StartRPC(listener string, indexer txindex.TxIndexer, eventBus *ttypes.EventBus, logger log.Logger) ([]net.Listener, error) {
	rpccore.SetEnvironment(&rpccore.Environment{
		EventBus: eventBus,
		Logger:   logger.With("module", "rpc"),

		Config: *tmconfig.DefaultRPCConfig(),
	})

	var routes = map[string]*rpcserver.RPCFunc{
		// subscribe/unsubscribe are reserved for websocket events.
		"subscribe":       rpcserver.NewWSRPCFunc(rpccore.Subscribe, "query"),
		"unsubscribe":     rpcserver.NewWSRPCFunc(rpccore.Unsubscribe, "query"),
		"unsubscribe_all": rpcserver.NewWSRPCFunc(rpccore.UnsubscribeAll, ""),

		"tx":        rpcserver.NewRPCFunc(NewTx(indexer), "hash"),
		"tx_search": rpcserver.NewRPCFunc(NewTxSearch(indexer), "query,page,per_page,order_by"),
	}

	listenAddrs := []string{listener}

	config := rpcserver.DefaultConfig()
	config.MaxOpenConnections = 100

	listeners := make([]net.Listener, len(listenAddrs))
	for i, listenAddr := range listenAddrs {
		mux := http.NewServeMux()
		rpcLogger := logger.With("module", "rpc-server")
		wmLogger := rpcLogger.With("protocol", "websocket")
		wm := rpcserver.NewWebsocketManager(routes,
			rpcserver.OnDisconnect(func(remoteAddr string) {
				err := eventBus.UnsubscribeAll(context.Background(), remoteAddr)
				if err != nil && err != tmpubsub.ErrSubscriptionNotFound {
					wmLogger.Error("Failed to unsubscribe addr from events", "addr", remoteAddr, "err", err)
				}
			}),
			rpcserver.ReadLimit(config.MaxBodyBytes),
			// rpcserver.WriteChanCapacity(n.config.RPC.WebSocketWriteBufferSize),
		)
		wm.SetLogger(wmLogger)
		mux.HandleFunc("/websocket", wm.WebsocketHandler)
		rpcserver.RegisterRPCFuncs(mux, routes, rpcLogger)
		listener, err := rpcserver.Listen(
			listenAddr,
			config,
		)
		if err != nil {
			return nil, err
		}

		var rootHandler http.Handler = mux
		go func() {
			if err := rpcserver.Serve(
				listener,
				rootHandler,
				rpcLogger,
				config,
			); err != nil {
				rpcLogger.Error("Error serving server", "err", err)
			}
		}()

		listeners[i] = listener
	}

	return listeners, nil
}
