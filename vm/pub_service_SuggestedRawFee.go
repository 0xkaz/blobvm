// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package vm

import (
	"net/http"
)

type SuggestedRawFeeReply struct {
	Price uint64 `serialize:"true" json:"price"`
	Cost  uint64 `serialize:"true" json:"cost"`
}

func (svc *PublicService) SuggestedRawFee(
	_ *http.Request,
	_ *struct{},
	reply *SuggestedRawFeeReply,
) error {
	price, cost, err := svc.vm.SuggestedFee()
	if err != nil {
		return err
	}
	reply.Price = price
	reply.Cost = cost
	return nil
}
