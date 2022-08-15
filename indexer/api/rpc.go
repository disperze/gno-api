package api

import (
	"net"
	"net/http"

	"github.com/tendermint/tendermint/libs/log"
	rpcserver "github.com/tendermint/tendermint/rpc/jsonrpc/server"
	"github.com/tendermint/tendermint/state/txindex"
)

func StartRPC(indexer txindex.TxIndexer, listner string) ([]net.Listener, error) {
	var routes = map[string]*rpcserver.RPCFunc{
		"tx_search": rpcserver.NewRPCFunc(NewTxSearch(indexer), "query,prove,page,per_page,order_by"),
	}

	listenAddrs := []string{listner}

	config := rpcserver.DefaultConfig()
	config.MaxOpenConnections = 100

	listeners := make([]net.Listener, len(listenAddrs))
	for i, listenAddr := range listenAddrs {
		mux := http.NewServeMux()
		rpcLogger := log.NewNopLogger()
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
