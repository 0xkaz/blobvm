// Copyright (C) 2022-2023, WEAVEDB LTD, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package vm

import (
	log2 "log"
	"net/http"

	log "github.com/inconshreveable/log15"
)

// WeaveDBServiceSet are arguments for Set
type WeaveDBServiceSetArgs struct {
	ContractTxId   string `json:"contract_tx_id"`
	CollectionPath string `json:"collection_path"`
	Value          string `json:"value"`
}

// WeaveDBServiceReply is the reply from Set
type WeaveDBServiceReply struct {
	Success bool `serialize:"true" json:"success"`
}

func (svc *WeaveDBService) Set(_ *http.Request, args *WeaveDBServiceSetArgs, reply *GeneralPingReply) (err error) {
	log2.Printf("WeaveDBService.Set: txid=%s, path=%s, val=%s", args.ContractTxId, args.CollectionPath, args.Value)

	log.Info("WeaveDBService.Set")
	reply.Success = true
	return nil
}
