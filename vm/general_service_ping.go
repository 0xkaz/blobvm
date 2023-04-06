// Copyright (C) 2022-2023, WEAVEDB LTD, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package vm

import (
	log2 "log"
	"net/http"

	log "github.com/inconshreveable/log15"
)

// GeneralPingReply are arguments for Ping
type GeneralPingReply struct {
	Success bool `serialize:"true" json:"success"`
}

func (svc *GeneralService) Ping(_ *http.Request, _ *struct{}, reply *GeneralPingReply) (err error) {
	log2.Printf("GeneralService.Ping")
	log.Info("ping")
	reply.Success = true
	return nil
}
