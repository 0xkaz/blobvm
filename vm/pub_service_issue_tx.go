// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package vm

import (
	"fmt"
	"net/http"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/ava-labs/blobvm/chain"
	"github.com/ava-labs/blobvm/tdata"
)

type IssueTxArgs struct {
	TypedData *tdata.TypedData `serialize:"true" json:"typedData"`
	Signature hexutil.Bytes    `serialize:"true" json:"signature"`
}

type IssueTxReply struct {
	TxID ids.ID `serialize:"true" json:"txId"`
}

func (svc *PublicService) IssueTx(_ *http.Request, args *IssueTxArgs, reply *IssueTxReply) error {
	if args.TypedData == nil {
		return ErrTypedDataIsNil
	}
	utx, err := chain.ParseTypedData(args.TypedData)
	if err != nil {
		return err
	}
	tx := chain.NewTx(utx, args.Signature[:])

	// otherwise, unexported tx.id field is empty
	if err := tx.Init(svc.vm.genesis); err != nil {
		return err
	}
	reply.TxID = tx.ID()

	errs := svc.vm.Submit(tx)
	if len(errs) == 0 {
		return nil
	}
	if len(errs) == 1 {
		return errs[0]
	}
	return fmt.Errorf("%v", errs)
}
