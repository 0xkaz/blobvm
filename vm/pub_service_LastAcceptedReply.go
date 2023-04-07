// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package vm

import (
	"net/http"

	"github.com/ava-labs/avalanchego/ids"
)

type LastAcceptedReply struct {
	Height  uint64 `serialize:"true" json:"height"`
	BlockID ids.ID `serialize:"true" json:"blockId"`
}

func (svc *PublicService) LastAccepted(_ *http.Request, _ *struct{}, reply *LastAcceptedReply) error {
	la := svc.vm.lastAccepted
	reply.Height = la.Hght
	reply.BlockID = la.ID()
	return nil
}
