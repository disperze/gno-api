package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/gnolang/gno/pkgs/bft/rpc/client"

	"github.com/gorilla/mux"
	"github.com/rs/cors"

	"github.com/disperze/gno-api/cmd/handler"

	_ "github.com/gnolang/gno/pkgs/sdk/bank"
	_ "github.com/gnolang/gno/pkgs/sdk/vm"
)

var (
	remotePtr  = flag.String("remote", "http://gno.land:36657", "Remote rpc")
	apiPortPtr = flag.String("port", "8888", "Api port")
	corsPtr    = flag.Bool("cors", false, "Enable CORS")
)

func main() {
	flag.Parse()

	if remotePtr == nil || *remotePtr == "" {
		log.Fatal("remote url is required")
	}

	if apiPortPtr == nil || *apiPortPtr == "" {
		log.Fatal("api port is required")
	}

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{http.MethodGet, http.MethodPost},
		AllowedHeaders: []string{"Content-Type", "Accept"},
	})

	apiPort := *apiPortPtr
	cli := client.NewHTTP(*remotePtr, "/websocket")

	r := mux.NewRouter()
	r.HandleFunc("/gno/render", handler.GnoRenderQueryHandler(cli)).Methods(http.MethodGet)
	r.HandleFunc("/gno/eval", handler.GnoEvalQueryHandler(cli)).Methods(http.MethodGet)
	r.HandleFunc("/gno/funcs", handler.GnoFuncsQueryHandler(cli)).Methods(http.MethodGet)
	r.HandleFunc("/cosmos/auth/v1beta1/accounts/{address}", handler.AuthQueryHandler(cli))
	r.HandleFunc("/cosmos/bank/v1beta1/balances/{address}", handler.BankQueryHandler(cli))
	r.HandleFunc("/cosmos/staking/v1beta1/delegations/{address}", handler.StakingQueryHandler(cli))
	r.HandleFunc("/cosmos/staking/v1beta1/delegators/{address}/unbonding_delegations", handler.StakingUnbondingQueryHandler(cli))
	r.HandleFunc("/cosmos/tx/v1beta1/txs", handler.ProtoTxsHandler(cli)).Methods(http.MethodPost)
	r.HandleFunc("/cosmos/tx/v1beta1/simulate", handler.SimulateTxHandler(cli)).Methods(http.MethodPost)
	r.HandleFunc("/txs/decode", handler.TxDecodeHandler(cli)).Methods(http.MethodGet)
	r.HandleFunc("/txs", handler.TxsHandler(cli)).Methods(http.MethodPost)

	var h http.Handler = r
	if *corsPtr {
		h = c.Handler(r)
	}

	fmt.Println("Running on port", apiPort)
	log.Fatal(http.ListenAndServe(":"+apiPort, h))
}
