// (c) 2019-2022, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package vm

import (
	log2 "log"
)

func init() {
	log2.SetFlags(log2.Ldate | log2.Lshortfile)
}

// BytesToData converts a byte slice to an array. If the byte slice input is
// larger than [QueueDataLen], it will be truncated.
func BytesToData(input []byte) [QueueDataLen]byte {
	log2.Printf("BytesToData")
	data := [QueueDataLen]byte{}
	lim := len(input)
	if lim > QueueDataLen {
		lim = QueueDataLen
	}
	copy(data[:], input[:lim])
	return data
}
