package main

import (
	"flag"
	"os"

	"github.com/tendermint/tendermint/state/txindex/kv"
	dbm "github.com/tendermint/tm-db"

	"github.com/disperze/gno-api/indexer/api"
	_ "github.com/gnolang/gno/pkgs/sdk/auth"
)

var (
	remotePtr      = flag.String("remote", "http://localhost:26657", "Remote rpc")
	rpcPtr         = flag.String("addr", "tcp://127.0.0.1:26657", "RPC addr")
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
	logger := NewLogger(os.Stdout)
	eventBus, err := createAndStartEventBus(logger)
	if err != nil {
		panic(err)
	}

	go api.StartRPC(*rpcPtr, indexer, eventBus, logger)

	err = StartIndexer(*remotePtr, *startHeightPtr, indexer, store, eventBus, logger)
	if err != nil {
		panic(err)
	}
}
