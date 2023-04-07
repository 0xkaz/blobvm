// Copyright (C) 2022-2023, WEAVEDB LTD. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"github.com/ava-labs/avalanchego/ids"
)

type Mempool3 interface {
	Len() int
	Prune(ids.Set)
	PopMax() (*Transaction, uint64)
	// Add(*Transaction) bool
	Add(
		tx *Transaction,
		collectionPath string,
		contractTxId string,
		value string) bool
	NewTxs(uint64) []*Transaction
}
