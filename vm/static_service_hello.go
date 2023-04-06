// (c) 2019-2022, WEAVEDB LTD. All rights reserved.
// See the file LICENSE for licensing terms.

package vm

import (
	"fmt"
	"net/http"

	log2 "log"
)

func init() {
	log2.SetFlags(log2.Ldate | log2.Lshortfile)
}

// HelloArgs are arguments for Hello
type HelloArgs struct {
	Name string `json:"name"`
}

// HelloReply is the reply from Hello
type HelloReply struct {
	Message string `json:"message"`
}

// Hello returns reply
func (ss *StaticService) Hello(_ *http.Request, args *HelloArgs, reply *HelloReply) error {
	log2.Printf("StaticService.Hello")

	reply.Message = fmt.Sprintf("Hello Hello %s", args.Name)
	return nil
}
