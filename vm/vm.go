// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

// Package vm implements custom VM.
package vm

import (
	"context"
	ejson "encoding/json"
	"fmt"
	log2 "log"
	"net/http"
	"time"

	"github.com/ava-labs/avalanchego/cache"
	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/database/manager"
	"github.com/ava-labs/avalanchego/database/versiondb"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow"
	"github.com/ava-labs/avalanchego/snow/choices"
	"github.com/ava-labs/avalanchego/snow/consensus/snowman"
	"github.com/ava-labs/avalanchego/snow/engine/common"
	snowmanblock "github.com/ava-labs/avalanchego/snow/engine/snowman/block"
	"github.com/ava-labs/avalanchego/utils"
	"github.com/ava-labs/avalanchego/utils/json"
	"github.com/gorilla/rpc/v2"
	log "github.com/inconshreveable/log15"

	avagoversion "github.com/ava-labs/avalanchego/version"
	"github.com/ava-labs/blobvm/chain"
	"github.com/ava-labs/blobvm/mempool"
	"github.com/ava-labs/blobvm/mempool3"
	"github.com/ava-labs/blobvm/version"
)

const (
	Name            = "blobvm"
	PublicEndpoint  = "/public"
	WeaveDBEndpoint = "/weavedb"
	QueueDataLen    = 1024
	MaxMempoolSize  = 4096
)

var (
	_ snowmanblock.ChainVM = &VM{}
	_ chain.VM             = &VM{}
)

type VM struct {
	snowCtx     *snow.Context
	db          database.Database
	config      Config
	genesis     *chain.Genesis
	AirdropData []byte

	// // State of WeaveDB VM
	// state WeaveDBState

	bootstrapped utils.AtomicBool

	mempool *mempool.Mempool
	// mempool2 [][QueueDataLen]byte
	mempool3 *mempool3.Mempool3

	appSender common.AppSender
	network   *PushNetwork

	// cache block objects to optimize "GetBlockStateless"
	// only put when a block is accepted
	// key: block ID, value: *chain.StatelessBlock
	blocks *cache.LRU

	// Block ID --> Block
	// Each element is a block that passed verification but
	// hasn't yet been accepted/rejected
	verifiedBlocks map[ids.ID]*chain.StatelessBlock

	toEngine chan<- common.Message
	builder  BlockBuilder

	preferred    ids.ID
	lastAccepted *chain.StatelessBlock

	// Recent activity
	activityCacheCursor uint64
	activityCache       []*chain.Activity

	// Execution checks
	targetRangeUnits uint64

	stop chan struct{}

	builderStop chan struct{}
	doneBuild   chan struct{}
	doneGossip  chan struct{}
}

const (
	blocksLRUSize = 128
)

// implements "snowmanblock.ChainVM.common.VM"
func (vm *VM) Initialize(
	ctx context.Context,
	snowCtx *snow.Context,
	dbManager manager.Manager,
	genesisBytes []byte,
	upgradeBytes []byte,
	configBytes []byte,
	toEngine chan<- common.Message,
	_ []*common.Fx,
	appSender common.AppSender,
) error {
	log.Info("initializing blobvm", "version", version.Version)

	// Load config
	vm.config.SetDefaults()
	if len(configBytes) > 0 {
		if err := ejson.Unmarshal(configBytes, &vm.config); err != nil {
			return fmt.Errorf("failed to unmarshal config %s: %w", string(configBytes), err)
		}
	}

	vm.snowCtx = snowCtx
	vm.db = dbManager.Current().Database
	vm.activityCache = make([]*chain.Activity, vm.config.ActivityCacheSize)

	// Init channels before initializing other structs
	vm.stop = make(chan struct{})
	vm.builderStop = make(chan struct{})
	vm.doneBuild = make(chan struct{})
	vm.doneGossip = make(chan struct{})
	vm.appSender = appSender
	vm.network = vm.NewPushNetwork()

	vm.blocks = &cache.LRU{Size: blocksLRUSize}
	vm.verifiedBlocks = make(map[ids.ID]*chain.StatelessBlock)

	vm.toEngine = toEngine
	vm.builder = vm.NewTimeBuilder()

	// Try to load last accepted
	has, err := chain.HasLastAccepted(vm.db)
	if err != nil {
		log.Error("could not determine if have last accepted")
		return err
	}

	// Parse genesis data
	vm.genesis = new(chain.Genesis)
	if err := ejson.Unmarshal(genesisBytes, vm.genesis); err != nil {
		log.Error("could not unmarshal genesis bytes")
		return err
	}
	if err := vm.genesis.Verify(); err != nil {
		log.Error("genesis is invalid")
		return err
	}
	targetUnitsPerSecond := vm.genesis.TargetBlockSize / uint64(vm.genesis.TargetBlockRate)
	vm.targetRangeUnits = targetUnitsPerSecond * uint64(vm.genesis.LookbackWindow)
	log.Debug("loaded genesis", "genesis", string(genesisBytes), "target range units", vm.targetRangeUnits)

	vm.mempool = mempool.New(vm.genesis, vm.config.MempoolSize)
	vm.mempool3 = mempool3.New(vm.genesis, vm.config.Mempool3Size)

	if has { //nolint:nestif
		blkID, err := chain.GetLastAccepted(vm.db)
		if err != nil {
			log.Error("could not get last accepted", "err", err)
			return err
		}

		blk, err := vm.GetStatelessBlock(blkID)
		if err != nil {
			log.Error("could not load last accepted", "err", err)
			return err
		}

		vm.preferred, vm.lastAccepted = blkID, blk
		log.Info("initialized blobvm from last accepted", "block", blkID)
	} else {
		genesisBlk, err := chain.ParseStatefulBlock(
			vm.genesis.StatefulBlock(),
			nil,
			choices.Accepted,
			vm,
		)
		if err != nil {
			log.Error("unable to init genesis block", "err", err)
			return err
		}

		// Set Balances
		if err := vm.genesis.Load(vm.db, vm.AirdropData); err != nil {
			log.Error("could not set genesis allocation", "err", err)
			return err
		}

		if err := chain.SetLastAccepted(vm.db, genesisBlk); err != nil {
			log.Error("could not set genesis as last accepted", "err", err)
			return err
		}
		gBlkID := genesisBlk.ID()
		vm.preferred, vm.lastAccepted = gBlkID, genesisBlk
		log.Info("initialized blobvm from genesis", "block", gBlkID)
	}
	vm.AirdropData = nil

	go vm.builder.Build()
	go vm.builder.Gossip()
	return nil
}

func (vm *VM) SetState(ctx context.Context, state snow.State) error {
	switch state {
	case snow.Bootstrapping:
		return vm.onBootstrapStarted()
	case snow.NormalOp:
		return vm.onNormalOperationsStarted()
	default:
		return snow.ErrUnknownState
	}
}

// onBootstrapStarted marks this VM as bootstrapping
func (vm *VM) onBootstrapStarted() error {
	vm.bootstrapped.SetValue(false)
	return nil
}

// onNormalOperationsStarted marks this VM as bootstrapped
func (vm *VM) onNormalOperationsStarted() error {
	if vm.bootstrapped.GetValue() {
		return nil
	}
	vm.bootstrapped.SetValue(true)
	return nil
}

// implements "snowmanblock.ChainVM.common.VM"
func (vm *VM) Shutdown(ctx context.Context) error {
	close(vm.stop)
	<-vm.doneBuild
	<-vm.doneGossip
	if vm.snowCtx == nil {
		return nil
	}
	return vm.db.Close()
}

// implements "snowmanblock.ChainVM.common.VM"
func (vm *VM) Version(ctx context.Context) (string, error) { return version.Version.String(), nil }

// NewHandler returns a new Handler for a service where:
//   - The handler's functionality is defined by [service]
//     [service] should be a gorilla RPC service (see https://www.gorillatoolkit.org/pkg/rpc/v2)
//   - The name of the service is [name]
//   - The LockOption is the first element of [lockOption]
//     By default the LockOption is WriteLock
//     [lockOption] should have either 0 or 1 elements. Elements beside the first are ignored.
func newHandler(name string, service interface{}, lockOption ...common.LockOption) (*common.HTTPHandler, error) {
	server := rpc.NewServer()
	server.RegisterCodec(json.NewCodec(), "application/json")
	server.RegisterCodec(json.NewCodec(), "application/json;charset=UTF-8")
	if err := server.RegisterService(service, name); err != nil {
		return nil, err
	}

	var lock common.LockOption = common.WriteLock
	if len(lockOption) != 0 {
		lock = lockOption[0]
	}
	return &common.HTTPHandler{LockOptions: lock, Handler: server}, nil
}

// implements "snowmanblock.ChainVM.common.VM"
// for "ext/vm/[chainID]"
func (vm *VM) CreateHandlers(ctx context.Context) (map[string]*common.HTTPHandler, error) {
	// publicServer := rpc.NewServer()
	// publicServer.RegisterCodec(json.NewCodec(), "application/json")
	// publicServer.RegisterCodec(json.NewCodec(), "application/json;charset=UTF-8")
	// if err := publicServer.RegisterService(&PublicService{}, Name); err != nil {
	// 	return nil, err
	// }
	// generalServer := rpc.NewServer()
	// generalServer.RegisterCodec(json.NewCodec(), "application/json")
	// generalServer.RegisterCodec(json.NewCodec(), "application/json;charset=UTF-8")
	// if err := generalServer.RegisterService(&GeneralService{}, Name); err != nil {
	// 	return nil, err
	// }
	// weavedbServer := rpc.NewServer()
	// weavedbServer.RegisterCodec(json.NewCodec(), "application/json")
	// weavedbServer.RegisterCodec(json.NewCodec(), "application/json;charset=UTF-8")
	// if err := generalServer.RegisterService(&WeaveDBService{}, Name); err != nil {
	// 	return nil, err
	// }

	publicServcie, err := newHandler(Name, &PublicService{vm: vm})
	if err != nil {
		return nil, err
	}
	generalServcie, err := newHandler(Name, &GeneralService{vm: vm})
	if err != nil {
		return nil, err
	}
	weavedbService, err := newHandler(Name, &WeaveDBService{vm: vm})
	if err != nil {
		return nil, err
	}

	apis := map[string]*common.HTTPHandler{}
	apis[PublicEndpoint] = publicServcie
	apis[WeaveDBEndpoint] = weavedbService
	apis[""] = generalServcie
	return apis, nil

	// return map[string]*common.HTTPHandler{
	// 	PublicEndpoint: {
	// 		LockOptions: common.NoLock,
	// 		Handler:     publicServer,
	// 	},
	// 	"": {
	// 		LockOptions: common.NoLock,
	// 		Handler:     generalServer,
	// 	},
	// 	WeaveDBEndpoint: {
	// 		LockOptions: common.NoLock,
	// 		Handler:     weavedbServer,
	// 	},
	// }, nil
}

// implements "snowmanblock.ChainVM.common.VM"
// for "ext/vm/[vmID]"
func (vm *VM) CreateStaticHandlers(ctx context.Context) (map[string]*common.HTTPHandler, error) {
	log2.Printf("CreateStaticHandlers")
	server := rpc.NewServer()
	server.RegisterCodec(json.NewCodec(), "application/json")
	server.RegisterCodec(json.NewCodec(), "application/json;charset=UTF-8")
	if err := server.RegisterService(&StaticService{}, Name); err != nil {
		return nil, err
	}

	return map[string]*common.HTTPHandler{
		"": {
			LockOptions: common.NoLock,
			Handler:     server,
		},
		"public": {
			LockOptions: common.NoLock,
			Handler:     server,
		},
	}, nil
	// return nil, nil
}

// implements "snowmanblock.ChainVM.commom.VM.AppHandler"
func (vm *VM) AppRequest(ctx context.Context, nodeID ids.NodeID, requestID uint32, deadline time.Time, request []byte) error {
	// (currently) no app-specific messages
	return nil
}

// implements "snowmanblock.ChainVM.commom.VM.AppHandler"
func (vm *VM) AppRequestFailed(ctx context.Context, nodeID ids.NodeID, requestID uint32) error {
	// (currently) no app-specific messages
	return nil
}

// implements "snowmanblock.ChainVM.commom.VM.AppHandler"
func (vm *VM) AppResponse(ctx context.Context, nodeID ids.NodeID, requestID uint32, response []byte) error {
	// (currently) no app-specific messages
	return nil
}

// implements "snowmanblock.ChainVM.commom.VM.AppHandler"
func (vm *VM) CrossChainAppRequest(ctx context.Context, chainID ids.ID, requestID uint32, deadline time.Time, request []byte) error {
	// (currently) no app-specific messages
	return nil
}

// implements "snowmanblock.ChainVM.commom.VM.AppHandler"
func (vm *VM) CrossChainAppRequestFailed(ctx context.Context, chainID ids.ID, requestID uint32) error {
	// (currently) no app-specific messages
	return nil
}

// implements "snowmanblock.ChainVM.commom.VM.AppHandler"
func (vm *VM) CrossChainAppResponse(ctx context.Context, chainID ids.ID, requestID uint32, request []byte) error {
	// (currently) no app-specific messages
	return nil
}

// implements "snowmanblock.ChainVM.commom.VM.health.Checkable"
func (vm *VM) HealthCheck(ctx context.Context) (interface{}, error) {
	// log2.Printf("HealthCheck")
	return http.StatusOK, nil
}

// implements "snowmanblock.ChainVM.commom.VM.validators.Connector"
func (vm *VM) Connected(ctx context.Context, id ids.NodeID, nodeVersion *avagoversion.Application) error {
	log2.Printf("Connected")
	// no-op
	return nil
}

// implements "snowmanblock.ChainVM.commom.VM.validators.Connector"
func (vm *VM) Disconnected(ctx context.Context, id ids.NodeID) error {
	log2.Printf("Disconnected")
	// no-op
	return nil
}

// implements "snowmanblock.ChainVM.commom.VM.Getter"
// replaces "core.SnowmanVM.GetBlock"

func (vm *VM) GetBlock(ctx context.Context, id ids.ID) (snowman.Block, error) {
	log2.Printf("VM.GetBlock")
	b, err := vm.GetStatelessBlock(id)
	if err != nil {
		log.Warn("failed to get block", "err", err)
	}
	return b, err
}

// // GetBlock implements the snowman.ChainVM interface
// func (vm *VM) GetBlock(ctx context.Context, blkID ids.ID) (snowman.Block, error) {
// 	log2.Printf("GetBlock")
// 	return vm.getBlock(blkID)
// }

// func (vm *VM) getBlock(blkID ids.ID) (*Block, error) {
// 	log2.Printf("getBlock")
// 	// If block is in memory, return it.
// 	if blk, exists := vm.verifiedBlocks[blkID]; exists {
// 		return blk, nil
// 	}
// 	// return vm.state.GetBlock(blkID)
// 	return vm.state.GetBlock(blkID)
// }

func (vm *VM) GetStatelessBlock(blkID ids.ID) (*chain.StatelessBlock, error) {
	log2.Printf("VM.GetStatelessBlock")
	// has the block been cached from previous "Accepted" call
	bi, exist := vm.blocks.Get(blkID)
	if exist {
		blk, ok := bi.(*chain.StatelessBlock)
		if !ok {
			return nil, fmt.Errorf("unexpected entry %T found in LRU cache, expected *chain.StatelessBlock", bi)
		}
		return blk, nil
	}

	// has the block been verified, not yet accepted
	if blk, exists := vm.verifiedBlocks[blkID]; exists {
		return blk, nil
	}

	// not found in memory, fetch from disk if accepted
	stBlk, err := chain.GetBlock(vm.db, blkID)
	if err != nil {
		return nil, err
	}
	// If block on disk, it must've been accepted
	return chain.ParseStatefulBlock(stBlk, nil, choices.Accepted, vm)
}

// implements "snowmanblock.ChainVM.commom.VM.Parser"
// replaces "core.SnowmanVM.ParseBlock"
func (vm *VM) ParseBlock(ctx context.Context, source []byte) (snowman.Block, error) {
	log2.Printf("VM.ParseBlock")
	newBlk, err := chain.ParseBlock(
		source,
		choices.Processing,
		vm,
	)
	if err != nil {
		log.Error("could not parse block", "err", err)
		return nil, err
	}
	log.Debug("parsed block", "id", newBlk.ID())

	// If we have seen this block before, return it with the most
	// up-to-date info
	if oldBlk, err := vm.GetBlock(ctx, newBlk.ID()); err == nil {
		log.Debug("returning previously parsed block", "id", oldBlk.ID())
		return oldBlk, nil
	}

	return newBlk, nil
}

// implements "snowmanblock.ChainVM"
// called via "avalanchego" node over RPC
func (vm *VM) BuildBlock(ctx context.Context) (snowman.Block, error) {
	log.Debug("BuildBlock triggered")
	blk, err := chain.BuildBlock(vm, vm.preferred)
	vm.builder.HandleGenerateBlock()
	if err != nil {
		log.Debug("BuildBlock failed", "error", err)
		return nil, err
	}
	sblk, ok := blk.(*chain.StatelessBlock)
	if !ok {
		return nil, fmt.Errorf("unexpected snowman.Block %T, expected *StatelessBlock", blk)
	}

	log.Debug("BuildBlock success", "blkID", blk.ID(), "txs", len(sblk.Txs))
	return blk, nil
}

func (vm *VM) Submit(txs ...*chain.Transaction) (errs []error) {
	blk, err := vm.GetStatelessBlock(vm.preferred)
	if err != nil {
		return []error{err}
	}
	now := time.Now().Unix()
	ctx, err := vm.ExecutionContext(now, blk)
	if err != nil {
		return []error{err}
	}
	vdb := versiondb.New(vm.db)

	for _, tx := range txs {
		if err := vm.submit(tx, vdb, now, ctx); err != nil {
			log.Debug("failed to submit transaction",
				"tx", tx.ID(),
				"error", err,
			)
			errs = append(errs, err)
			continue
		}
		vdb.Abort()
	}
	return errs
}

func (vm *VM) submit(tx *chain.Transaction, db database.Database, blkTime int64, ctx *chain.Context) error {
	if err := tx.Init(vm.genesis); err != nil {
		return err
	}
	if err := tx.ExecuteBase(vm.genesis); err != nil {
		return err
	}
	dummy := chain.DummyBlock(blkTime, tx)
	if err := tx.Execute(vm.genesis, db, dummy, ctx); err != nil {
		return err
	}
	vm.mempool.Add(tx)
	return nil
}

// "SetPreference" implements "snowmanblock.ChainVM"
// replaces "core.SnowmanVM.SetPreference"
func (vm *VM) SetPreference(ctx context.Context, id ids.ID) error {
	log.Debug("set preference", "id", id)
	vm.preferred = id
	return nil
}

// "LastAccepted" implements "snowmanblock.ChainVM"
// replaces "core.SnowmanVM.LastAccepted"
func (vm *VM) LastAccepted(ctx context.Context) (ids.ID, error) {
	return vm.lastAccepted.ID(), nil
}

// AddWeaveDBQueue
func (vm *VM) AddWeaveDBQueue(
// txs ...*chain.Transaction

collectionPath string,
contractTxId string,
value string,

) (errs []error) {
	log2.Printf("AddWeaveDBQueue..")
	blk, err := vm.GetStatelessBlock(vm.preferred)
	log2.Printf("blk: %v", blk)
	if err != nil {
		return []error{err}
	}
	now := time.Now().Unix()
	log2.Printf("now: %v", now)
	ctx, err := vm.ExecutionContext(now, blk)
	log2.Printf("ctx: %v", ctx)
	if err != nil {
		return []error{err}
	}
	vdb := versiondb.New(vm.db)
	log2.Printf("vdb: %v",vdb)
	if err := vm.addWeaveDBQueue(tx, vdb, now, ctx, collectionName, contractTxId, value); err != nil {
	// for _, tx := range txs {
	// 	if err := vm.addWeaveDBQueue(tx, vdb, now, ctx); err != nil {
	// 		log.Debug("failed to submit transaction",
	// 			"tx", tx.ID(),
	// 			"error", err,
	// 		)
	// 		errs = append(errs, err)
	// 		continue
	// 	}
	// 	vdb.Abort()
	// }
	return errs
}

// addWeaveDBQueue ...
func (vm *VM) addWeaveDBQueue(
	tx *chain.Transaction,
db database.Database, 
	blkTime int64,
	ctx *chain.Context,
	collectionPath string,
	contractTxId string,
	value string,
) error {
	if err := tx.Init(vm.genesis); err != nil {
		return err
	}
	if err := tx.ExecuteBase(vm.genesis); err != nil {
		return err
	}
	dummy := chain.DummyBlock(blkTime, tx)
	if err := tx.Execute(vm.genesis, db, dummy, ctx); err != nil {
		return err
	}

	vm.mempool3.Add(tx, collectionPath, contractTxId, value)

	// log2.Printf("addWeaveDBQueue: data=%v", data)
	// if len(vm.mempool2) > MaxMempoolSize {
	// 	return false
	// }
	// log2.Printf("addWeaveDBQueue")
	// vm.mempool2 = append(vm.mempool2, data)
	// log2.Printf("addWeaveDBQueue")

	vm.NotifyBlockReady()
	return nil
}

// NotifyBlockReady tells the consensus engine that a new block
// is ready to be created
func (vm *VM) NotifyBlockReady() {
	log2.Printf("NotifyBlockReady")

	select {
	case vm.toEngine <- common.PendingTxs:
	default:
		vm.snowCtx.Log.Debug("dropping message to consensus engine")
	}
}
