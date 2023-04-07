// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package vm

import (
	"net/http"

	"github.com/ava-labs/blobvm/chain"
)

type GenesisReply struct {
	Genesis *chain.Genesis `serialize:"true" json:"genesis"`
}

func (svc *PublicService) Genesis(_ *http.Request, _ *struct{}, reply *GenesisReply) (err error) {
	reply.Genesis = svc.vm.Genesis()
	return nil
}
