package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/tendermint/tendermint/state/txindex/kv"
	dbm "github.com/tendermint/tm-db"

	"github.com/disperze/gno-api/indexer/api"
	_ "github.com/gnolang/gno/pkgs/sdk/auth"
)

var (
	remotePtr      = flag.String("remote", "http://localhost:26657", "Remote rpc")
	startHeightPtr = flag.Int64("start", 1, "Start height")
)

func main() {
	flag.Parse()

	store, err := dbm.NewDB("tx_index", "goleveldb", "data")
	if err != nil {
		panic(err)
	}
	defer store.Close()

	go func() {
		err = StartIndexer(*remotePtr, store, *startHeightPtr)
		if err != nil {
			panic(err)
		}
	}()

	indexer := kv.NewTxIndex(store)
	r := mux.NewRouter()
	r.HandleFunc("/tx_search", api.TxSearch(indexer)).Methods(http.MethodGet)

	fmt.Println("Running on port", 8092)
	log.Fatal(http.ListenAndServe(":8092", r))
}
