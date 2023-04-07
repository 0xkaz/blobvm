// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package vm

import (
	"net/http"

	"github.com/ava-labs/blobvm/chain"
	"github.com/ava-labs/blobvm/tdata"
)

type SuggestedFeeArgs struct {
	Input *chain.Input `serialize:"true" json:"input"`
}

type SuggestedFeeReply struct {
	TypedData *tdata.TypedData `serialize:"true" json:"typedData"`
	TotalCost uint64           `serialize:"true" json:"totalCost"`
}

func (svc *PublicService) SuggestedFee(
	_ *http.Request,
	args *SuggestedFeeArgs,
	reply *SuggestedFeeReply,
) error {
	if args.Input == nil {
		return ErrInputIsNil
	}
	utx, err := args.Input.Decode()
	if err != nil {
		return err
	}

	// Determine suggested fee
	price, cost, err := svc.vm.SuggestedFee()
	if err != nil {
		return err
	}
	g := svc.vm.genesis
	fu := utx.FeeUnits(g)
	price += cost / fu

	// Update meta
	utx.SetBlockID(svc.vm.lastAccepted.ID())
	utx.SetMagic(g.Magic)
	utx.SetPrice(price)

	reply.TypedData = utx.TypedData()
	reply.TotalCost = fu * price
	return nil
}
