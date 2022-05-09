package handler

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gnolang/gno/pkgs/amino"
	"github.com/gnolang/gno/pkgs/bft/rpc/client"
	ctypes "github.com/gnolang/gno/pkgs/bft/rpc/core/types"
	"github.com/gnolang/gno/pkgs/std"
)

type BroadcastTxResponse struct {
	TxResponse *TxResponse `json:"tx_response,omitempty"`
}

type TxResponse struct {
	Height    int64  `json:"height,omitempty"`
	TxHash    string `json:"txhash,omitempty"`
	Code      uint32 `json:"code,omitempty"`
	Data      []byte `json:"data,omitempty"`
	RawLog    string `json:"raw_log,omitempty"`
	GasWanted int64  `json:"gas_wanted,omitempty"`
	GasUsed   int64  `json:"gas_used,omitempty"`
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

		code, log := getCodeLog(res)
		result := BroadcastTxResponse{
			TxResponse: &TxResponse{
				TxHash:    fmt.Sprintf("%X", res.Hash),
				Height:    res.Height,
				Code:      code,
				Data:      res.DeliverTx.Data,
				RawLog:    log,
				GasWanted: res.DeliverTx.GasWanted,
				GasUsed:   res.DeliverTx.GasUsed,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(result)
	}
}

func ProtoTxsHandler(cli client.ABCIClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var params struct {
			TxBytes string `json:"tx_bytes"`
			Mode    string `json:"mode"`
		}
		err := json.NewDecoder(r.Body).Decode(&params)
		if err != nil {
			writeError(w, fmt.Errorf("%s, %s", "unmarshaling json params", err.Error()))
			return
		}

		txBz, err := base64.StdEncoding.DecodeString(params.TxBytes)
		if err != nil {
			writeError(w, fmt.Errorf("%s, %s", "cannot decode tx", err.Error()))
			return
		}

		res, err := cli.BroadcastTxCommit(txBz)
		if err != nil {
			writeError(w, fmt.Errorf("%s, %s", err.Error(), "broadcasting bytes"))
			return
		}

		code, log := getCodeLog(res)
		result := BroadcastTxResponse{
			TxResponse: &TxResponse{
				TxHash:    fmt.Sprintf("%X", res.Hash),
				Height:    res.Height,
				Code:      code,
				Data:      res.DeliverTx.Data,
				RawLog:    log,
				GasWanted: res.DeliverTx.GasWanted,
				GasUsed:   res.DeliverTx.GasUsed,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(result)
	}
}

func getCodeLog(res *ctypes.ResultBroadcastTxCommit) (uint32, string) {
	if res.CheckTx.IsErr() {
		return 1, res.CheckTx.Log
	}

	if res.DeliverTx.IsErr() {
		return 2, res.DeliverTx.Log
	}

	return 0, ""
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
