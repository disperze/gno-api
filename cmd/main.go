package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/gnolang/gno/pkgs/bft/rpc/client"

	"github.com/gorilla/mux"

	"github.com/disperze/gno-api/cmd/handler"

	_ "github.com/gnolang/gno/pkgs/sdk/bank"
	_ "github.com/gnolang/gno/pkgs/sdk/vm"
)

var (
	remotePtr  = flag.String("remote", "http://gno.land:36657", "Remote rpc")
	apiPortPtr = flag.String("port", "8888", "Api port")
)

func main() {
	flag.Parse()

	if remotePtr == nil || *remotePtr == "" {
		log.Fatal("remote url is required")
	}

	if apiPortPtr == nil || *apiPortPtr == "" {
		log.Fatal("api port is required")
	}

	apiPort := *apiPortPtr
	cli := client.NewHTTP(*remotePtr, "/websocket")

	r := mux.NewRouter()
	r.HandleFunc("/gno/render", handler.GnoRenderQueryHandler(cli)).Methods("GET")
	r.HandleFunc("/cosmos/auth/v1beta1/accounts/{address}", handler.AuthQueryHandler(cli))
	r.HandleFunc("/cosmos/bank/v1beta1/balances/{address}", handler.BankQueryHandler(cli))
	r.HandleFunc("/cosmos/staking/v1beta1/delegations/{address}", handler.StakingQueryHandler(cli))
	r.HandleFunc("/cosmos/staking/v1beta1/delegators/{address}/unbonding_delegations", handler.StakingUnbondingQueryHandler(cli))
	r.HandleFunc("/txs", handler.TxsHandler(cli)).Methods(http.MethodPost)

	fmt.Println("Running on port", apiPort)
	log.Fatal(http.ListenAndServe(":"+apiPort, r))
}
