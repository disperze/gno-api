package api

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"

	"github.com/tendermint/tendermint/libs/bytes"
	tmmath "github.com/tendermint/tendermint/libs/math"
	tmquery "github.com/tendermint/tendermint/libs/pubsub/query"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	"github.com/tendermint/tendermint/state/txindex/kv"
)

func TxSearch(indexer *kv.TxIndex) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()
		orderBy := params.Get("order_by")
		var pagePtr, perPagePtr *int

		q, err := tmquery.New(params.Get("query"))
		if err != nil {
			writeError(w, err)
			return
		}
		results, err := indexer.Search(r.Context(), q)
		if err != nil {
			writeError(w, err)
			return
		}

		// sort results (must be done before pagination)
		switch orderBy {
		case "desc":
			sort.Slice(results, func(i, j int) bool {
				if results[i].Height == results[j].Height {
					return results[i].Index > results[j].Index
				}
				return results[i].Height > results[j].Height
			})
		case "asc", "":
			sort.Slice(results, func(i, j int) bool {
				if results[i].Height == results[j].Height {
					return results[i].Index < results[j].Index
				}
				return results[i].Height < results[j].Height
			})
		default:
			// return errors.New("expected order_by to be either `asc` or `desc` or empty")
		}

		// paginate results
		totalCount := len(results)
		perPage := validatePerPage(perPagePtr)

		page, err := validatePage(pagePtr, perPage, totalCount)
		if err != nil {
			writeError(w, err)
			return
		}

		skipCount := validateSkipCount(page, perPage)
		pageSize := tmmath.MinInt(perPage, totalCount-skipCount)

		apiResults := make([]*ctypes.ResultTx, 0, pageSize)
		for i := skipCount; i < skipCount+pageSize; i++ {
			r := results[i]

			apiResults = append(apiResults, &ctypes.ResultTx{
				Hash:     bytes.HexBytes(NewSHA256(r.Tx)),
				Height:   r.Height,
				Index:    r.Index,
				TxResult: r.Result,
				Tx:       r.Tx,
			})
		}

		output := &ctypes.ResultTxSearch{Txs: apiResults, TotalCount: totalCount}

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

func NewSHA256(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}
