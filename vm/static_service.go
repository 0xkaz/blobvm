// (c) 2019-2022, WEAVEDB LTD. All rights reserved.
// See the file LICENSE for licensing terms.

package vm

import (
	log2 "log"
)

func init() {
	log2.SetFlags(log2.Ldate | log2.Lshortfile)
}

// StaticService ...
type StaticService struct{}

// CreateStaticService ...
func CreateStaticService() *StaticService {
	log2.Printf("CreateStaticService")
	return &StaticService{}
}
