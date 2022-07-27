package main

import (
	"flag"
	"fmt"

	"github.com/gnolang/gno/pkgs/amino"
	"github.com/gnolang/gno/pkgs/bft/rpc/client"
	"github.com/gnolang/gno/pkgs/std"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/state/txindex/kv"
	dbm "github.com/tendermint/tm-db"

	_ "github.com/gnolang/gno/pkgs/sdk/auth"
	"github.com/gnolang/gno/pkgs/sdk/bank"
	"github.com/gnolang/gno/pkgs/sdk/vm"

	"github.com/disperze/gno-api/indexer/handler"
)

func start(index *kv.TxIndex, remote string, startHeight int64) error {
	c := client.NewHTTP(remote, "/websocket")
	status, err := c.Status()
	if err != nil {
		panic(err)
	}
	last := status.SyncInfo.LatestBlockHeight

	for height := startHeight; height <= last; height++ {
		fmt.Printf("Height: %d\n", height)
		block, err := c.Block(&height)
		if err != nil {
			return err
		}
		txs := block.Block.Data.Txs
		if len(txs) == 0 {
			continue
		}
		bres, err := c.BlockResults(&height)
		if err != nil {
			return err
		}

		for i := 0; i < len(txs); i++ {
			txResult := bres.Results.DeliverTxs[i]
			var code uint32 = 0
			if txResult.Error != nil {
				code = 1
			}

			tx := txs[i]
			stdtx := std.Tx{}
			amino.MustUnmarshal(tx, &stdtx)

			events := []abci.Event{}
			for _, msg := range stdtx.Msgs {
				switch msg := msg.(type) {
				case bank.MsgSend:
					events = append(events, handler.BankEvents(msg)...)
				case bank.MsgMultiSend:
					events = append(events, handler.MultiSendEvents(msg)...)
				case vm.MsgAddPackage:
					events = append(events, handler.VmAddPackageEvents(msg)...)
				case vm.MsgCall:
					events = append(events, handler.VmCallEvents(msg)...)
				}
			}

			data := &abci.TxResult{
				Height: height,
				Index:  uint32(i),
				Tx:     tx,
				Result: abci.ResponseDeliverTx{
					Code:      code,
					Data:      txResult.Data,
					Log:       txResult.Log,
					Info:      txResult.Info,
					GasWanted: txResult.GasWanted,
					GasUsed:   txResult.GasUsed,
					Events:    events,
					Codespace: "",
				},
			}

			err = index.Index(data)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

var (
	remotePtr      = flag.String("remote", "http://localhost:26657", "Remote rpc")
	startHeightPtr = flag.Int64("start", 1, "Start height")
)

func main() {
	flag.Parse()

	store, err := dbm.NewDB("tx_index", "goleveldb", "data")
	if err != nil {
		panic(err)
	}
	defer store.Close()
	indexer := kv.NewTxIndex(store)

	if err = start(indexer, *remotePtr, *startHeightPtr); err != nil {
		panic(err)
	}
}
