// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package vm

import (
	"net/http"

	"github.com/ethereum/go-ethereum/common"

	"github.com/ava-labs/blobvm/chain"
)

type BalanceArgs struct {
	Address common.Address `serialize:"true" json:"address"`
}

type BalanceReply struct {
	Balance uint64 `serialize:"true" json:"balance"`
}

func (svc *PublicService) Balance(_ *http.Request, args *BalanceArgs, reply *BalanceReply) error {
	bal, err := chain.GetBalance(svc.vm.db, args.Address)
	if err != nil {
		return err
	}
	reply.Balance = bal
	return err
}
