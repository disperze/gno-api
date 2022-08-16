package api

import (
	"errors"
	"fmt"
	"log"
	"sort"

	tmmath "github.com/tendermint/tendermint/libs/math"
	tmquery "github.com/tendermint/tendermint/libs/pubsub/query"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	rpctypes "github.com/tendermint/tendermint/rpc/jsonrpc/types"
	"github.com/tendermint/tendermint/state/txindex"
	ttypes "github.com/tendermint/tendermint/types"
)

func NewTx(indexer txindex.TxIndexer) interface{} {
	return func(ctx *rpctypes.Context, hash []byte) (*ctypes.ResultTx, error) {
		r, err := indexer.Get(hash)
		if err != nil {
			return nil, err
		}

		if r == nil {
			return nil, fmt.Errorf("tx (%X) not found", hash)
		}

		height := r.Height
		index := r.Index

		return &ctypes.ResultTx{
			Hash:     hash,
			Height:   height,
			Index:    index,
			TxResult: r.Result,
			Tx:       r.Tx,
		}, nil
	}
}

func NewTxSearch(indexer txindex.TxIndexer) interface{} {
	return func(
		ctx *rpctypes.Context,
		query string,
		pagePtr, perPagePtr *int,
		orderBy string,
	) (*ctypes.ResultTxSearch, error) {
		log.Printf("Query %s\n", query)

		q, err := tmquery.New(query)
		if err != nil {
			return nil, err
		}
		results, err := indexer.Search(ctx.Context(), q)
		if err != nil {
			return nil, err
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
			return nil, errors.New("expected order_by to be either `asc` or `desc` or empty")
		}

		// paginate results
		totalCount := len(results)
		perPage := validatePerPage(perPagePtr)

		page, err := validatePage(pagePtr, perPage, totalCount)
		if err != nil {
			return nil, err
		}

		skipCount := validateSkipCount(page, perPage)
		pageSize := tmmath.MinInt(perPage, totalCount-skipCount)

		apiResults := make([]*ctypes.ResultTx, 0, pageSize)
		for i := skipCount; i < skipCount+pageSize; i++ {
			r := results[i]

			apiResults = append(apiResults, &ctypes.ResultTx{
				Hash:     ttypes.Tx(r.Tx).Hash(),
				Height:   r.Height,
				Index:    r.Index,
				TxResult: r.Result,
				Tx:       r.Tx,
			})
		}

		return &ctypes.ResultTxSearch{Txs: apiResults, TotalCount: totalCount}, nil
	}
}
