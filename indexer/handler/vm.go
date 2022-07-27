package handler

import (
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/gnolang/gno/pkgs/sdk/vm"
)

// VmAddPackageEvents handle vm add_package msgs
func VmAddPackageEvents(msg vm.MsgAddPackage) []abci.Event {
	return []abci.Event{
		{Type: "message", Attributes: []abci.EventAttribute{
			{Key: []byte("module"), Value: []byte("vm")},
			{Key: []byte("action"), Value: []byte("package")},
			{Key: []byte("sender"), Value: []byte(msg.Creator.String())},
		}},
		{Type: "package", Attributes: []abci.EventAttribute{
			{Key: []byte("creator"), Value: []byte(msg.Creator.String())},
			{Key: []byte("path"), Value: []byte(msg.Package.Path)},
			{Key: []byte("deposit"), Value: []byte(msg.Deposit.String())},
		}},
	}
}

// VmCallEvents handle vm add_package msgs
func VmCallEvents(msg vm.MsgCall) []abci.Event {
	return []abci.Event{
		{Type: "message", Attributes: []abci.EventAttribute{
			{Key: []byte("module"), Value: []byte("vm")},
			{Key: []byte("action"), Value: []byte("call")},
			{Key: []byte("sender"), Value: []byte(msg.Caller.String())},
		}},
		{Type: "package", Attributes: []abci.EventAttribute{
			{Key: []byte("caller"), Value: []byte(msg.Caller.String())},
			{Key: []byte("package"), Value: []byte(msg.PkgPath)},
			{Key: []byte("func"), Value: []byte(msg.Func)},
			{Key: []byte("send"), Value: []byte(msg.Send.String())},
		}},
	}
}
