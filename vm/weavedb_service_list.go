// Copyright (C) 2022-2023, WEAVEDB LTD, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package vm

import (
	"fmt"
	log2 "log"
	"net/http"
)

// // WeaveDBServiceListArgs are arguments for Set
// type WeaveDBServiceListArgs struct {
// 	ContractTxId   string `json:"contract_tx_id"`
// 	CollectionPath string `json:"collection_path"`
// 	Value          string `json:"value"`
// }

// WeaveDBServiceMempooleply is the reply from Set
type WeaveDBServiceMempoolReply struct {
	Success bool   `serialize:"true" json:"success"`
	Result  string `json:"result"`
}

func (svc *WeaveDBService) Mempool(_ *http.Request, _ *struct{}, reply *WeaveDBServiceMempoolReply) (err error) {
	log2.Printf("WeaveDBService.Mempool")
	reply.Result = fmt.Sprintf("mempool2=%v", svc.vm.mempool2)
	reply.Success = true
	return nil
}
