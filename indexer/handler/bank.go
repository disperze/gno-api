package handler

import (
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/gnolang/gno/pkgs/sdk/bank"
)

// BankEvents handle bank msgs
func BankEvents(msg bank.MsgSend) []abci.Event {
	return []abci.Event{
		{Type: "message", Attributes: []abci.EventAttribute{
			{Key: []byte("module"), Value: []byte("bank")},
			{Key: []byte("action"), Value: []byte("transfer")},
			{Key: []byte("sender"), Value: []byte(msg.FromAddress.String())},
		}},
		{Type: "transfer", Attributes: []abci.EventAttribute{
			{Key: []byte("from"), Value: []byte(msg.FromAddress.String())},
			{Key: []byte("to"), Value: []byte(msg.ToAddress.String())},
			{Key: []byte("amount"), Value: []byte(msg.Amount.String())},
		}},
	}
}

// MultiSendEvents handle bank multi send
func MultiSendEvents(msg bank.MsgMultiSend) []abci.Event {
	return []abci.Event{
		{Type: "message", Attributes: []abci.EventAttribute{
			{Key: []byte("module"), Value: []byte("bank")},
			{Key: []byte("action"), Value: []byte("multi_send")},
			{Key: []byte("sender"), Value: []byte(msg.GetSigners()[0].String())},
		}},
	}
}
