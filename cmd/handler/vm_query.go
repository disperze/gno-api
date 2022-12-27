package handler

import (
	"fmt"
	"net/http"

	"github.com/gnolang/gno/pkgs/bft/rpc/client"
)

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

func GnoEvalQueryHandler(cli client.ABCIClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()
		data := []byte(fmt.Sprintf("%s\n%s", params.Get("realm"), params.Get("func")))
		res, err := cli.ABCIQuery("vm/qeval", data)
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

func GnoFuncsQueryHandler(cli client.ABCIClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()
		data := []byte(params.Get("realm"))
		res, err := cli.ABCIQuery("vm/qeval", data)
		if err != nil {
			writeError(w, err)
			return
		}

		if res.Response.Error != nil {
			writeError(w, res.Response.Error)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, string(res.Response.Data))
	}
}
