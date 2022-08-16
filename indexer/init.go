package main

import (
	"io"

	"github.com/tendermint/tendermint/libs/log"
	ttypes "github.com/tendermint/tendermint/types"
)

func NewLogger(w io.Writer) log.Logger {
	return log.NewTMLogger(log.NewSyncWriter(w))
}

func createAndStartEventBus(logger log.Logger) (*ttypes.EventBus, error) {
	eventBus := ttypes.NewEventBus()
	eventBus.SetLogger(logger.With("module", "events"))
	if err := eventBus.Start(); err != nil {
		return nil, err
	}
	return eventBus, nil
}
