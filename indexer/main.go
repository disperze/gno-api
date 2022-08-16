package main

import (
	"flag"

	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/state/txindex/kv"
	ttypes "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/disperze/gno-api/indexer/api"
	_ "github.com/gnolang/gno/pkgs/sdk/auth"
)

var (
	remotePtr      = flag.String("remote", "http://localhost:26657", "Remote rpc")
	rpcPtr         = flag.String("port", "tcp://127.0.0.1:26657", "RPC addr")
	startHeightPtr = flag.Int64("start", 1, "Start height")
)

func main() {
	flag.Parse()

	store, err := dbm.NewDB("tx_index", "goleveldb", "data")
	if err != nil {
		panic(err)
	}
	defer store.Close()

	indexer := kv.NewTxIndex(store)
	logger := log.NewNopLogger()

	eventBus := ttypes.NewEventBus()
	eventBus.SetLogger(logger.With("module", "events"))
	go api.StartRPC(*rpcPtr, indexer, eventBus, logger)

	err = StartIndexer(*remotePtr, store, *startHeightPtr)
	if err != nil {
		panic(err)
	}
}
