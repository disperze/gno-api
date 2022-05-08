package handler

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gnolang/gno/pkgs/amino"
	abci "github.com/gnolang/gno/pkgs/bft/abci/types"
	"github.com/gnolang/gno/pkgs/bft/rpc/client"
	ctypes "github.com/gnolang/gno/pkgs/bft/rpc/core/types"
	"github.com/gnolang/gno/pkgs/std"
)

type BroadcastResponse struct {
	Result abci.ResponseDeliverTx `json:"result"`
	Hash   string                 `json:"hash"`
	Height int64                  `json:"height"`
}

func TxsHandler(cli client.ABCIClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var params map[string]json.RawMessage
		err := json.NewDecoder(r.Body).Decode(&params)
		if err != nil {
			writeError(w, fmt.Errorf("%s, %s", "unmarshaling json params", err.Error()))
			return
		}

		txData, ok := params["tx"]
		if !ok {
			writeError(w, fmt.Errorf("%s, %s", "Missing tx param", err.Error()))
			return
		}

		txBz, _ := txData.MarshalJSON()
		var tx std.Tx
		err = amino.UnmarshalJSON(txBz, &tx)
		if err != nil {
			writeError(w, fmt.Errorf("%s, %s", "unmarshaling tx json bytes", err.Error()))
			return
		}

		res, err := BroadcastHandler(cli, tx)
		if err != nil {
			writeError(w, err)
			return
		}

		if res.CheckTx.IsErr() {
			writeError(w, fmt.Errorf("transaction failed %#v\nlog %s", res, res.CheckTx.Log))
			return
		}

		if res.DeliverTx.IsErr() {
			writeError(w, fmt.Errorf("transaction failed %#v\nlog %s", res, res.DeliverTx.Log))
			return
		}

		result := BroadcastResponse{
			Hash:   fmt.Sprintf("%X", res.Hash),
			Height: res.Height,
			Result: res.DeliverTx,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(result)
	}
}

func ProtoTxsHandler(cli client.ABCIClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var params map[string]string
		err := json.NewDecoder(r.Body).Decode(&params)
		if err != nil {
			writeError(w, fmt.Errorf("%s, %s", "unmarshaling json params", err.Error()))
			return
		}

		txData, ok := params["tx"]
		if !ok {
			writeError(w, fmt.Errorf("%s, %s", "Missing tx param", err.Error()))
			return
		}

		txBz, err := base64.StdEncoding.DecodeString(txData)
		if err != nil {
			writeError(w, fmt.Errorf("%s, %s", "cannot decode tx", err.Error()))
			return
		}

		res, err := cli.BroadcastTxCommit(txBz)
		if err != nil {
			writeError(w, fmt.Errorf("%s, %s", err.Error(), "broadcasting bytes"))
			return
		}

		if res.CheckTx.IsErr() {
			writeError(w, fmt.Errorf("transaction failed %#v\nlog %s", res, res.CheckTx.Log))
			return
		}

		if res.DeliverTx.IsErr() {
			writeError(w, fmt.Errorf("transaction failed %#v\nlog %s", res, res.DeliverTx.Log))
			return
		}

		result := BroadcastResponse{
			Hash:   fmt.Sprintf("%X", res.Hash),
			Height: res.Height,
			Result: res.DeliverTx,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(result)
	}
}

func BroadcastHandler(cli client.ABCIClient, tx std.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
	bz, err := amino.Marshal(tx)
	if err != nil {
		return nil, fmt.Errorf("%s, %s", err.Error(), "remarshaling tx binary bytes")
	}

	bres, err := cli.BroadcastTxCommit(bz)
	if err != nil {
		return nil, fmt.Errorf("%s, %s", err.Error(), "broadcasting bytes")
	}

	return bres, nil
}
