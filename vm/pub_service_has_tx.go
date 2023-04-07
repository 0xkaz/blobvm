// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package vm

import (
	"net/http"

	"github.com/ava-labs/avalanchego/ids"

	"github.com/ava-labs/blobvm/chain"
)

type HasTxArgs struct {
	TxID ids.ID `serialize:"true" json:"txId"`
}

type HasTxReply struct {
	Accepted bool `serialize:"true" json:"accepted"`
}

func (svc *PublicService) HasTx(_ *http.Request, args *HasTxArgs, reply *HasTxReply) error {
	has, err := chain.HasTransaction(svc.vm.db, args.TxID)
	if err != nil {
		return err
	}
	reply.Accepted = has
	return nil
}
