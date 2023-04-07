// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/blobvm/tdata"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type Transaction3 struct {
	UnsignedTransaction `serialize:"true" json:"unsignedTransaction"`
	Signature           []byte `serialize:"true" json:"signature"`

	digestHash []byte
	bytes      []byte
	id         ids.ID
	size       uint64
	sender     common.Address
}

func NewTx3(utx UnsignedTransaction, sig []byte) *Transaction {
	return &Transaction{
		UnsignedTransaction: utx,
		Signature:           sig,
	}
}

func (t *Transaction3) Copy() *Transaction {
	sig := make([]byte, len(t.Signature))
	copy(sig, t.Signature)
	return &Transaction{
		UnsignedTransaction: t.UnsignedTransaction.Copy(),
		Signature:           sig,
	}
}

func DigestHash3(utx UnsignedTransaction) ([]byte, error) {
	return tdata.DigestHash(utx.TypedData())
}

func (t *Transaction3) Init(g *Genesis) error {
	stx, err := Marshal(t)
	if err != nil {
		return err
	}
	t.bytes = stx

	id, err := ids.ToID(crypto.Keccak256(t.bytes))
	if err != nil {
		return err
	}
	t.id = id

	// Compute digest hash
	dh, err := DigestHash(t.UnsignedTransaction)
	if err != nil {
		return err
	}
	t.digestHash = dh

	// Derive sender
	pk, err := DeriveSender(t.digestHash, t.Signature)
	if err != nil {
		return err
	}
	t.sender = crypto.PubkeyToAddress(*pk)

	t.size = uint64(len(t.Bytes()))
	return nil
}

func (t *Transaction3) Bytes() []byte { return t.bytes }

func (t *Transaction3) Size() uint64 { return t.size }

func (t *Transaction3) ID() ids.ID { return t.id }

func (t *Transaction3) DigestHash() []byte { return t.digestHash }

func (t *Transaction3) Sender() common.Address { return t.sender }

func (t *Transaction3) Execute(g *Genesis, db database.Database, blk *StatelessBlock, context *Context) error {
	if err := t.UnsignedTransaction.ExecuteBase(g); err != nil {
		return err
	}
	if !context.RecentBlockIDs.Contains(t.GetBlockID()) {
		// Hash must be recent to be any good
		// Should not happen beause of mempool cleanup
		return ErrInvalidBlockID
	}
	if context.RecentTxIDs.Contains(t.ID()) {
		// Tx hash must not be recently executed (otherwise could be replayed)
		//
		// NOTE: We only need to keep cached tx hashes around as long as the
		// block hash referenced in the tx is valid
		return ErrDuplicateTx
	}

	// Ensure sender has balance
	if _, err := ModifyBalance(db, t.sender, false, t.FeeUnits(g)*t.GetPrice()); err != nil {
		return err
	}
	if t.GetPrice() < context.NextPrice {
		return ErrInsufficientPrice
	}
	if err := t.UnsignedTransaction.Execute(&TransactionContext{
		Genesis:   g,
		Database:  db,
		BlockTime: uint64(blk.Tmstmp),
		TxID:      t.id,
		Sender:    t.sender,
	}); err != nil {
		return err
	}
	if err := SetTransaction3(db, t); err != nil {
		return err
	}
	return nil
}

func (t *Transaction3) Activity() *Activity {
	activity := t.UnsignedTransaction.Activity()
	activity.Sender = t.sender.Hex()
	activity.TxID = t.id
	return activity
}
