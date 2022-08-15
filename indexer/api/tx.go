package api

import (
	"crypto/sha256"
	"errors"
	"log"
	"net/http"
	"sort"

	"github.com/tendermint/tendermint/libs/bytes"
	tmmath "github.com/tendermint/tendermint/libs/math"
	tmquery "github.com/tendermint/tendermint/libs/pubsub/query"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	"github.com/tendermint/tendermint/state/txindex/kv"
)

type TxService struct {
	indexer *kv.TxIndex
}

type Args struct {
	query               string
	pagePtr, perPagePtr *int
	orderBy             string
}

type Result ctypes.ResultTxSearch

func NewTxService(indexer *kv.TxIndex) *TxService {
	return &TxService{indexer}
}

func (t *TxService) TxSearch(r *http.Request, args *Args, result *Result) error {
	log.Printf("Query %s\n", args.query)

	q, err := tmquery.New(args.query)
	if err != nil {
		return err
	}
	results, err := t.indexer.Search(r.Context(), q)
	if err != nil {
		return err
	}

	// sort results (must be done before pagination)
	switch args.orderBy {
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
		return errors.New("expected order_by to be either `asc` or `desc` or empty")
	}

	// paginate results
	totalCount := len(results)
	perPage := validatePerPage(args.perPagePtr)

	page, err := validatePage(args.pagePtr, perPage, totalCount)
	if err != nil {
		return err
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

	*result = Result(ctypes.ResultTxSearch{Txs: apiResults, TotalCount: totalCount})

	return nil
}

func NewSHA256(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}
