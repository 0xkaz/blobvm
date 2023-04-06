// Copyright (C) 2022-2023, WEAVEDB LTD, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package vm

import (
	"fmt"
	log2 "log"
	"net/http"
)

// GeneralHelloArgs are arguments for Hello
type GeneralHelloArgs struct {
	Name string `json:"name"`
}

// GeneralHelloReply is the reply from Hello
type GeneralHelloReply struct {
	Message string `json:"message"`
}

func (svc *GeneralService) Hello(_ *http.Request, args *GeneralHelloArgs, reply *GeneralHelloReply) (err error) {
	log2.Printf("GeneralService.Hello")
	reply.Message = fmt.Sprintf("General.Hello Hello %s", args.Name)
	return nil
}
