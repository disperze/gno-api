package handler

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gnolang/gno/pkgs/amino"
	"github.com/gnolang/gno/pkgs/bft/rpc/client"
	ctypes "github.com/gnolang/gno/pkgs/bft/rpc/core/types"
	keysclient "github.com/gnolang/gno/pkgs/crypto/keys/client"
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

type SimulateResponse struct {
	GasInfo TxGasResult `json:"gas_info"`
}

type TxGasResult struct {
	GasUsed int64 `json:"gas_used"`
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

		var response *BroadcastTxResponse
		switch params.Mode {
		case "BROADCAST_MODE_ASYNC":
			response, err = broadcastAsync(cli, txBz)
		case "BROADCAST_MODE_SYNC":
			response, err = broadcastSync(cli, txBz)
		case "BROADCAST_MODE_BLOCK":
			response, err = broadcastCommit(cli, txBz)
		default:
			err = fmt.Errorf("%s, %s", "unknown mode", err.Error())
		}

		if err != nil {
			writeError(w, fmt.Errorf("%s, %s", "cannot broadcast tx", err.Error()))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

func SimulateTxHandler(cli client.ABCIClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var params struct {
			TxBytes string `json:"tx_bytes"`
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
		response, err := keysclient.SimulateTx(cli, txBz)
		if err != nil {
			writeError(w, fmt.Errorf("%s, %s", "cannot simulate tx", err.Error()))
			return
		}

		result := &SimulateResponse{
			GasInfo: TxGasResult{
				GasUsed: response.DeliverTx.GasUsed,
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

func broadcastAsync(cli client.ABCIClient, tx []byte) (*BroadcastTxResponse, error) {
	res, err := cli.BroadcastTxAsync(tx)
	if err != nil {
		return nil, err
	}

	code := uint32(0)
	if res.Error != nil {
		code = 1
	}

	result := &BroadcastTxResponse{
		TxResponse: &TxResponse{
			TxHash: fmt.Sprintf("%X", res.Hash),
			Code:   code,
			Data:   res.Data,
			RawLog: res.Log,
		},
	}

	return result, nil
}

func broadcastSync(cli client.ABCIClient, tx []byte) (*BroadcastTxResponse, error) {
	res, err := cli.BroadcastTxSync(tx)
	if err != nil {
		return nil, err
	}

	code := uint32(0)
	if res.Error != nil {
		code = 1
	}
	result := &BroadcastTxResponse{
		TxResponse: &TxResponse{
			TxHash: fmt.Sprintf("%X", res.Hash),
			Code:   code,
			Data:   res.Data,
			RawLog: res.Log,
		},
	}

	return result, nil
}

func broadcastCommit(cli client.ABCIClient, tx []byte) (*BroadcastTxResponse, error) {
	res, err := cli.BroadcastTxCommit(tx)
	if err != nil {
		return nil, err
	}

	code, log := getCodeLog(res)
	result := &BroadcastTxResponse{
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

	return result, nil
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
