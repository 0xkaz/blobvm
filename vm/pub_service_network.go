// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package vm

import (
	"net/http"

	"github.com/ava-labs/avalanchego/ids"
)

type NetworkReply struct {
	NetworkID uint32 `serialize:"true" json:"networkId"`
	SubnetID  ids.ID `serialize:"true" json:"subnetId"`
	ChainID   ids.ID `serialize:"true" json:"chainId"`
}

func (svc *PublicService) Network(_ *http.Request, _ *struct{}, reply *NetworkReply) (err error) {
	reply.NetworkID = svc.vm.snowCtx.NetworkID
	reply.SubnetID = svc.vm.snowCtx.SubnetID
	reply.ChainID = svc.vm.snowCtx.ChainID
	return nil
}
