package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	abci "github.com/tendermint/tendermint/abci/types"
	tmquery "github.com/tendermint/tendermint/libs/pubsub/query"
	"github.com/tendermint/tendermint/state/txindex/kv"
)

type QueryResult struct {
	Jsonrpc string   `json:"jsonrpc"`
	ID      int      `json:"id"`
	Result  TxResult `json:"result"`
}
type TxResult struct {
	Txs []*abci.TxResult `json:"txs"`
}

func TxSearch(indexer *kv.TxIndex) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()

		query, err := tmquery.New(params.Get("query"))
		if err != nil {
			writeError(w, err)
			return
		}
		res, err := indexer.Search(r.Context(), query)
		if err != nil {
			writeError(w, err)
			return
		}

		output := QueryResult{
			Jsonrpc: "2.0",
			ID:      -1,
			Result:  TxResult{Txs: res},
		}

		jsonResponse, err := json.Marshal(output)
		if err != nil {
			writeError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonResponse)
	}
}

func writeError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprintf(w, "Error: %s", err.Error())
}
