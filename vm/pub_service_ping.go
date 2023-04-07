// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package vm

import (
	"net/http"

	log "github.com/inconshreveable/log15"
)

type PingReply struct {
	Success bool `serialize:"true" json:"success"`
}

func (svc *PublicService) Ping(_ *http.Request, _ *struct{}, reply *PingReply) (err error) {
	log.Info("ping")
	reply.Success = true
	return nil
}
