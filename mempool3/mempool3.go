package mempool3

import (
	"container/heap"
	"sync"

	"github.com/ava-labs/avalanchego/ids"

	"github.com/ava-labs/blobvm/chain"
)

var _ chain.Mempool3 = &Mempool3{}

type Mempool3 struct {
	mu sync.RWMutex

	g       *chain.Genesis
	maxSize int
	maxHeap *txHeap
	minHeap *txHeap

	// Pending is a channel of length one, which the mempool ensures has an item on
	// it as long as there is an unissued transaction remaining in [txs]
	Pending chan struct{}
	// newTxs is an array of [Tx] that are ready to be gossiped.
	newTxs []*chain.Transaction
}

// New creates a new [Mempool3]. [maxSize] must be > 0 or else the
// implementation may panic.
func New(
	g *chain.Genesis,
	maxSize int,

) *Mempool3 {

	return &Mempool3{
		// collectionPath: collectionPath,
		// contractTxId:   contractTxId,
		// value:          value,
		g:       g,
		maxSize: maxSize,
		maxHeap: newTxHeap(maxSize, false),
		minHeap: newTxHeap(maxSize, true),
		Pending: make(chan struct{}, 1),
	}
}

func (th *Mempool3) Add(
	tx *chain.Transaction,
	collectionPath string,
	contractTxId string,
	value string) bool {

	txID := tx.ID()
	price := tx.GetPrice()

	th.mu.Lock()
	defer th.mu.Unlock()

	// Don't add duplicates
	if th.maxHeap.Has(txID) {
		return false
	}

	oldLen := th.maxHeap.Len()

	// Optimistically add tx to mempool
	heap.Push(th.maxHeap, &txEntry{
		id:             txID,
		price:          price,
		contractTxId:   contractTxId,
		collectionPath: collectionPath,
		value:          value,
		tx:             tx,
		index:          oldLen,
	})
	heap.Push(th.minHeap, &txEntry{
		id:             txID,
		price:          price,
		contractTxId:   contractTxId,
		collectionPath: collectionPath,
		value:          value,
		tx:             tx,
		index:          oldLen,
	})

	// Remove the lowest paying tx
	//
	// Note: we do this after adding the new transaction in case it is the new
	// lowest paying transaction
	if th.maxHeap.Len() > th.maxSize {
		t, _ := th.popMin()
		if t.ID() == txID {
			return false
		}
	}

	// When adding [tx] to the mempool make sure that there is an item in Pending
	// to signal the VM to produce a block. Note: if the VM's buildStatus has already
	// been set to something other than [dontBuild], this will be ignored and won't be
	// reset until the engine calls BuildBlock. This case is handled in IssueCurrentTx
	// and CancelCurrentTx.
	th.newTxs = append(th.newTxs, tx)
	th.addPending()
	return true
}

// Assumes there is non-zero items in [Mempool]
func (th *Mempool3) PeekMax() (*chain.Transaction, uint64) {
	th.mu.RLock()
	defer th.mu.RUnlock()

	txEntry := th.maxHeap.items[0]
	return txEntry.tx, txEntry.price
}

// // Assumes there is non-zero items in [Mempool]
// func (th *Mempool3) PeekMin() (*chain.Transaction, uint64) {
// 	th.mu.RLock()
// 	defer th.mu.RUnlock()

// 	txEntry := th.minHeap.items[0]
// 	return txEntry.tx, txEntry.price
// }

// Assumes there is non-zero items in [Mempool]
func (th *Mempool3) PopMax() (*chain.Transaction, uint64) { // O(log N)
	th.mu.Lock()
	defer th.mu.Unlock()

	item := th.maxHeap.items[0]
	return th.remove(item.id), item.price
}

// Assumes there is non-zero items in [Mempool]
func (th *Mempool3) PopMin() (*chain.Transaction, uint64) { // O(log N)
	th.mu.Lock()
	defer th.mu.Unlock()
	return th.popMin()
}

func (th *Mempool3) Remove(id ids.ID) *chain.Transaction { // O(log N)
	th.mu.Lock()
	defer th.mu.Unlock()

	return th.remove(id)
}

// Prune removes all transactions that are not found in "validHashes".
func (th *Mempool3) Prune(validHashes ids.Set) {
	th.mu.RLock()
	toRemove := []ids.ID{}
	for _, txE := range th.maxHeap.items { // O(N)
		if !validHashes.Contains(txE.tx.GetBlockID()) {
			toRemove = append(toRemove, txE.id)
		}
	}
	th.mu.RUnlock()

	for _, txID := range toRemove { // O(K * log N)
		th.Remove(txID)
	}
}

func (th *Mempool3) Len() int {
	th.mu.RLock()
	defer th.mu.RUnlock()

	return th.maxHeap.Len()
}

func (th *Mempool3) Get(id ids.ID) (*chain.Transaction, bool) {
	th.mu.RLock()
	defer th.mu.RUnlock()

	txEntry, ok := th.maxHeap.Get(id)
	if !ok {
		return nil, false
	}
	return txEntry.tx, true
}

func (th *Mempool3) Has(id ids.ID) bool {
	th.mu.RLock()
	defer th.mu.RUnlock()

	return th.maxHeap.Has(id)
}

// GetNewTxs returns the array of [newTxs] and replaces it with a new array.
func (th *Mempool3) NewTxs(maxUnits uint64) []*chain.Transaction {
	th.mu.Lock()
	defer th.mu.Unlock()

	// Note: this algorithm preserves the ordering of new transactions
	var (
		units    uint64
		selected = th.newTxs[:0]
	)
	for i, tx := range th.newTxs {
		// It is possible that a block may have been accepted that contains some
		// new transactions before [NewTxs] is called.
		if !th.maxHeap.Has(tx.ID()) {
			continue
		}
		txUnits := tx.LoadUnits(th.g)
		if txUnits > maxUnits-units {
			th.newTxs = th.newTxs[i:]
			return selected
		}
		units += txUnits
		selected = append(selected, tx)
	}
	th.newTxs = th.newTxs[len(th.newTxs):]
	return selected
}

// popMin assumes the write lock is held and takes O(log N) time to run.
func (th *Mempool3) popMin() (*chain.Transaction, uint64) { // O(log N)
	item := th.minHeap.items[0]
	return th.remove(item.id), item.price
}

// remove assumes the write lock is held and takes O(log N) time to run.
func (th *Mempool3) remove(id ids.ID) *chain.Transaction {
	maxEntry, ok := th.maxHeap.Get(id) // O(1)
	if !ok {
		return nil
	}
	heap.Remove(th.maxHeap, maxEntry.index) // O(log N)

	minEntry, ok := th.minHeap.Get(id) // O(1)
	if !ok {
		// This should never happen, as that would mean the heaps are out of
		// sync.
		return nil
	}
	txe, ok := heap.Remove(th.minHeap, minEntry.index).(*txEntry) // O(log N)
	if !ok {
		// This should never happen
		return nil
	}
	return txe.tx
}

// addPending makes sure that an item is in the Pending channel.
func (th *Mempool3) addPending() {
	select {
	case th.Pending <- struct{}{}:
	default:
	}
}
