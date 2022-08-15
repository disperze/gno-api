package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/rpc"
	"github.com/gorilla/rpc/json"
	"github.com/tendermint/tendermint/state/txindex/kv"
	dbm "github.com/tendermint/tm-db"

	"github.com/disperze/gno-api/indexer/api"
	_ "github.com/gnolang/gno/pkgs/sdk/auth"
)

var (
	remotePtr      = flag.String("remote", "http://localhost:26657", "Remote rpc")
	apiPortPtr     = flag.String("port", "8094", "Api port")
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
	txSrv := api.NewTxService(indexer)

	s := rpc.NewServer()
	s.RegisterCodec(json.NewCodec(), "application/json")
	s.RegisterCodec(json.NewCodec(), "application/json;charset=UTF-8")
	s.RegisterService(txSrv, "")

	r := mux.NewRouter()
	r.Handle("/", s)

	fmt.Println("Running on port", *apiPortPtr)
	log.Fatal(http.ListenAndServe(":"+*apiPortPtr, r))
}
