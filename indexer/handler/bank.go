package handler

import (
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/gnolang/gno/pkgs/sdk/bank"
)

// BankEvents handle bank msgs
func BankEvents(msg bank.MsgSend) []abci.Event {
	return []abci.Event{
		{Type: "message", Attributes: []abci.EventAttribute{
			{Key: []byte("module"), Value: []byte("bank"), Index: true},
			{Key: []byte("action"), Value: []byte("transfer"), Index: true},
			{Key: []byte("sender"), Value: []byte(msg.FromAddress.String()), Index: true},
		}},
		{Type: "transfer", Attributes: []abci.EventAttribute{
			{Key: []byte("from"), Value: []byte(msg.FromAddress.String()), Index: true},
			{Key: []byte("to"), Value: []byte(msg.ToAddress.String()), Index: true},
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
