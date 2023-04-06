// Copyright (C) 2022-2023, WEAVEDB LTD, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package vm

import (
	log2 "log"
	"net/http"

	log "github.com/inconshreveable/log15"
)

// WeaveDBServiceGetAargs are arguments for Get
type WeaveDBServiceGetAargs struct {
	ContractTxId   string `json:"contract_tx_id"`
	CollectionPath string `json:"collection_path"`
	Value          string `json:"value"`
}

// WeaveDBServiceGetReply is the reply from Get
type WeaveDBServiceGetReply struct {
	Success bool `serialize:"true" json:"success"`
}

func (svc *WeaveDBService) Get(_ *http.Request, args *WeaveDBServiceGetAargs, reply *WeaveDBServiceGetReply) (err error) {
	log2.Printf("WeaveDBService.Get: txid=%s, path=%s, val=%s", args.ContractTxId, args.CollectionPath, args.Value)

	log.Info("WeaveDBService.Get")
	reply.Success = true
	return nil
}
