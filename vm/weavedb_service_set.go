// Copyright (C) 2022-2023, WEAVEDB LTD, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package vm

import (
	"fmt"
	log2 "log"
	"net/http"

	"github.com/ava-labs/avalanchego/utils/formatting"

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
	// id, err := svc.vm.state.GetLastAccepted()

	log.Info("WeaveDBService.Set")
	bytes, err := formatting.Decode(formatting.Hex, args.Value)
	log2.Printf("WeaveDBService.Set: len(bytes)=%v", len(bytes))

	if err != nil {
		// return errBadData
		return fmt.Errorf("WeaveDBService.Set: err=%v", err)
	}
	// if len(bytes) != QueueDataLen {
	// 	return fmt.Errorf("WeaveDBService.Set: len(bytes) != QueueDataLen")
	// }

	// svc.vm.AddWeaveDBQueue(
	// args.CollectionPath,
	// args.ContractTxId,
	// args.Value,
	// )
	// reply.Success = svc.vm.addWeaveDBQueue(BytesToData(bytes))

	reply.Success = true
	return nil
}
