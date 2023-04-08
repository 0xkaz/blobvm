// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"bytes"
	"strconv"

	"github.com/ava-labs/blobvm/tdata"
	"github.com/ethereum/go-ethereum/common"
)

var _ UnsignedTransaction = &WriteTx{}

type WriteTx struct {
	*BaseTx `serialize:"true" json:"baseTx"`

	// To is the recipient of the [Units].
	To common.Address `serialize:"true" json:"to"`

	// Units are transferred to [To].
	Units uint64 `serialize:"true" json:"units"`

	//
	ContractTxId string `serialize:"true" json:"contract_tx_id"`
	//
	CollectionPath string `serialize:"true" json:"collection_path"`
	//
	Query string `serialize:"true" json:"query"`
}

func (t *WriteTx) Execute(c *TransactionContext) error {
	// Must transfer to someone
	if bytes.Equal(t.To[:], zeroAddress[:]) {
		return ErrNonActionable
	}

	// This prevents someone from transferring to themselves.
	if bytes.Equal(t.To[:], c.Sender[:]) {
		return ErrNonActionable
	}
	if t.Units == 0 {
		return ErrNonActionable
	}
	if _, err := ModifyBalance(c.Database, c.Sender, false, t.Units); err != nil {
		return err
	}
	if _, err := ModifyBalance(c.Database, t.To, true, t.Units); err != nil {
		return err
	}
	return nil
}

func (t *WriteTx) Copy() UnsignedTransaction {
	to := make([]byte, common.AddressLength)
	copy(to, t.To[:])
	return &WriteTx{
		BaseTx: t.BaseTx.Copy(),
		To:     common.BytesToAddress(to),
		Units:  t.Units,
	}
}

func (t *WriteTx) TypedData() *tdata.TypedData {
	return tdata.CreateTypedData(
		t.Magic, Transfer,
		[]tdata.Type{
			{Name: tdTo, Type: tdAddress},
			{Name: tdUnits, Type: tdUint64},
			{Name: tdPrice, Type: tdUint64},
			{Name: tdContractTxId, Type: tdString},
			{Name: tdCollectionPath, Type: tdString},
			{Name: tdBlockID, Type: tdString},
		},
		tdata.TypedDataMessage{
			tdTo:             t.To.Hex(),
			tdUnits:          strconv.FormatUint(t.Units, 10),
			tdPrice:          strconv.FormatUint(t.Price, 10),
			tdContractTxId:   t.ContractTxId,
			tdCollectionPath: t.CollectionPath,
			tdBlockID:        t.BlockID.String(),
		},
	)
}

func (t *WriteTx) Activity() *Activity {
	return &Activity{
		Typ:   Transfer,
		To:    t.To.Hex(),
		Units: t.Units,
		//
		Query:          t.Query,
		CollectionPath: t.CollectionPath,
		ContractTxId:   t.ContractTxId,
	}
}

func (b *WriteTx) GetContractTxId() string {
	return b.ContractTxId
}
func (b *WriteTx) GetCollectionPath() string {
	return b.ContractTxId
}
func (b *WriteTx) SetContractTxId(txid string) {
	b.ContractTxId = txid
}

func (b *WriteTx) SetCollectionPath(colPath string) {
	b.CollectionPath = colPath
}

func (b *WriteTx) SetQuery(q string) {
	b.Query = q
}
func (b *WriteTx) GetQuery() string {
	return b.Query
}
