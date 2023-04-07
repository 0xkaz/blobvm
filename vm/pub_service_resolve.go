// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package vm

import (
	"net/http"

	"github.com/ethereum/go-ethereum/common"

	"github.com/ava-labs/blobvm/chain"
)

type ResolveArgs struct {
	Key common.Hash `serialize:"true" json:"key"`
}

type ResolveReply struct {
	Exists    bool             `serialize:"true" json:"exists"`
	Value     []byte           `serialize:"true" json:"value"`
	ValueMeta *chain.ValueMeta `serialize:"true" json:"valueMeta"`
}

func (svc *PublicService) Resolve(_ *http.Request, args *ResolveArgs, reply *ResolveReply) error {
	vmeta, exists, err := chain.GetValueMeta(svc.vm.db, args.Key)
	if err != nil {
		return err
	}
	if !exists {
		// Avoid value lookup if doesn't exist
		return nil
	}
	v, exists, err := chain.GetValue(svc.vm.db, args.Key)
	if err != nil {
		return err
	}
	if !exists {
		return ErrCorruption
	}

	// Set values properly
	reply.Exists = true
	reply.Value = v
	reply.ValueMeta = vmeta
	return nil
}
