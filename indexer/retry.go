package main

import (
	"fmt"

	"github.com/avast/retry-go/v4"
	"github.com/gnolang/gno/pkgs/bft/rpc/client"
	ctypes "github.com/gnolang/gno/pkgs/bft/rpc/core/types"
)

type RetryClient struct {
	cli *client.HTTP
}

func NewRetryClient(cli *client.HTTP) *RetryClient {
	return &RetryClient{cli: cli}
}

func (r *RetryClient) Status() (*ctypes.ResultStatus, error) {
	var result *ctypes.ResultStatus
	err := retry.Do(
		func() error {
			var err error
			result, err = r.cli.Status()
			return err
		},
		r.getRetryOptions()...,
	)

	return result, err
}

func (r *RetryClient) Block(height *int64) (*ctypes.ResultBlock, error) {
	var result *ctypes.ResultBlock
	err := retry.Do(
		func() error {
			var err error
			result, err = r.cli.Block(height)
			return err
		},
		r.getRetryOptions()...,
	)

	return result, err
}

func (r *RetryClient) BlockResults(height *int64) (*ctypes.ResultBlockResults, error) {
	var result *ctypes.ResultBlockResults
	err := retry.Do(
		func() error {
			var err error
			result, err = r.cli.BlockResults(height)
			return err
		},
		r.getRetryOptions()...,
	)

	return result, err
}

func (r *RetryClient) getRetryOptions() []retry.Option {
	return []retry.Option{
		retry.OnRetry(func(n uint, err error) {
			fmt.Printf("Retry #%d: %v\n", n, err)
		}),
		retry.DelayType(retry.BackOffDelay),
	}
}
