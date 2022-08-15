package api

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"

	abci "github.com/tendermint/tendermint/abci/types"
	tmquery "github.com/tendermint/tendermint/libs/pubsub/query"
	"github.com/tendermint/tendermint/state/txindex/kv"
)

type TxsResult struct {
	Txs        []TxResult `json:"txs"`
	TotalCount string     `json:"total_count"`
}

type TxResult struct {
	Hash     string                 `json:"hash"`
	Height   string                 `json:"height"`
	Index    uint32                 `json:"index"`
	TxResult abci.ResponseDeliverTx `json:"tx_result"`
	Tx       []byte                 `json:"tx"`
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

		txs := make([]TxResult, len(res))
		for i, tx := range res {
			txs[i] = TxResult{
				Hash:     fmt.Sprintf("%X", sha256.Sum256(tx.Tx)),
				Height:   fmt.Sprintf("%d", tx.Height),
				Index:    tx.Index,
				TxResult: tx.Result,
				Tx:       tx.Tx,
			}
		}
		output := TxsResult{
			Txs:        txs,
			TotalCount: fmt.Sprintf("%d", len(txs)),
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
