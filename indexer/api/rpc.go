package api

import (
	"context"
	"net"
	"net/http"

	"github.com/tendermint/tendermint/libs/log"
	tmpubsub "github.com/tendermint/tendermint/libs/pubsub"
	rpcserver "github.com/tendermint/tendermint/rpc/jsonrpc/server"
	"github.com/tendermint/tendermint/state/txindex"
	ttypes "github.com/tendermint/tendermint/types"
)

func StartRPC(indexer txindex.TxIndexer, listner string) ([]net.Listener, error) {
	logger := log.NewNopLogger()
	eventBus := ttypes.NewEventBus()
	eventBus.SetLogger(logger.With("module", "events"))

	var routes = map[string]*rpcserver.RPCFunc{
		"tx":        rpcserver.NewRPCFunc(NewTx(indexer), "hash"),
		"tx_search": rpcserver.NewRPCFunc(NewTxSearch(indexer), "query,page,per_page,order_by"),
	}

	listenAddrs := []string{listner}

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
