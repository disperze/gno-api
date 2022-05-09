package handler

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gnolang/gno/pkgs/amino"
	"github.com/gnolang/gno/pkgs/bft/rpc/client"
	"github.com/gnolang/gno/pkgs/crypto"
	"github.com/gnolang/gno/pkgs/std"

	"github.com/gorilla/mux"
)

type Coin struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

type PaginationInfo struct {
	NextKey []byte `json:"next_key"`
	Total   string `json:"total"`
}

type BankResult struct {
	Balances   []Coin         `json:"balances"`
	Pagination PaginationInfo `json:"pagination"`
}

type AuthAccount struct {
	Account CosmosAccount `json:"account"`
}

type CosmosAccount struct {
	Type          string         `json:"@type"`
	Address       crypto.Address `json:"address"`
	PubKey        crypto.PubKey  `json:"pub_key"`
	AccountNumber uint64         `json:"account_number"`
	Sequence      uint64         `json:"sequence"`
}

type GnoAccount struct {
	std.BaseAccount
}

func GnoRenderQueryHandler(cli client.ABCIClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()
		data := []byte(fmt.Sprintf("%s\n%s", params.Get("realm"), params.Get("query")))
		res, err := cli.ABCIQuery("vm/qrender", data)
		if err != nil {
			writeError(w, err)
			return
		}

		if res.Response.Error != nil {
			writeError(w, res.Response.Error)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, string(res.Response.Data))
	}
}

func AuthQueryHandler(cli client.ABCIClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		authPath := fmt.Sprintf("auth/accounts/%s", vars["address"])
		res, err := cli.ABCIQuery(authPath, nil)
		if err != nil {
			writeError(w, err)
			return
		}

		if res.Response.Error != nil {
			writeError(w, res.Response.Error)
			return
		}

		var account GnoAccount
		err = amino.UnmarshalJSON(res.Response.Data, &account)
		if err != nil {
			writeError(w, err)
			return
		}

		if account.Address.IsZero() {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		result := AuthAccount{
			Account: CosmosAccount{
				Type:          "/cosmos.auth.v1beta1.BaseAccount",
				Address:       account.GetAddress(),
				PubKey:        account.GetPubKey(),
				AccountNumber: account.GetAccountNumber(),
				Sequence:      account.GetSequence(),
			},
		}

		json, _ := amino.MarshalJSONIndent(result, "", "  ")
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, string(json))
	}
}

func TxDecodeHandler(cli client.ABCIClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()
		txParam := params.Get("tx")
		if txParam == "" {
			writeError(w, fmt.Errorf("tx param is required"))
			return
		}

		txData, err := base64.StdEncoding.DecodeString(txParam)
		if err != nil {
			writeError(w, err)
			return
		}

		var tx std.Tx
		err = amino.Unmarshal(txData, &tx)
		if err != nil {
			writeError(w, err)
			return
		}

		jsonData, err := amino.MarshalJSON(&tx)
		if err != nil {
			writeError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, string(jsonData))
	}
}

func BankQueryHandler(cli client.ABCIClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		denom := "gnot"

		authPath := fmt.Sprintf("bank/balances/%s", vars["address"])
		res, err := cli.ABCIQuery(authPath, nil)
		if err != nil {
			writeError(w, err)
			return
		}

		if res.Response.Error != nil {
			writeError(w, err)
			return
		}

		var balance string
		err = json.Unmarshal(res.Response.Data, &balance)
		if err != nil {
			writeError(w, err)
			return
		}

		coins := []Coin{}
		if balance != "" {
			coins = append(coins, Coin{
				Denom:  denom,
				Amount: strings.TrimRight(balance, denom),
			})
		}

		result := BankResult{
			Balances: coins,
			Pagination: PaginationInfo{
				NextKey: nil,
				Total:   strconv.Itoa(len(coins)),
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(result)
	}
}

func StakingQueryHandler(cli client.ABCIClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		template := `{
	"delegation_responses": [
	],
	"pagination": {
		"next_key": null,
		"total": "0"
	}
}`
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, template)
	}
}

func StakingUnbondingQueryHandler(cli client.ABCIClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		template := `{
	"unbonding_responses": [
	],
	"pagination": {
		"next_key": null,
		"total": "0"
	}
}`
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, template)
	}
}

func writeError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprintf(w, "Error: %s", err.Error())
}
