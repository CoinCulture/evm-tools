// Copyright 2014 The go-ethereum Authors
// This file is part of go-ethereum.
//
// go-ethereum is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-ethereum is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-ethereum. If not, see <http://www.gnu.org/licenses/>.

// evm executes EVM code snippets.
package main

import (
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/logger/glog"
	"gopkg.in/urfave/cli.v1"
)

const (
	LatestStateRootKey = "evm:LatestStateRootKey"
)

type message struct {
	from     *common.Address
	to       *common.Address
	gasPrice *big.Int
	gas      *big.Int
	value    *big.Int
	nonce    uint64
	data     []byte
}

func (m *message) From() (common.Address, error)         { return *m.from, nil }
func (m *message) FromFrontier() (common.Address, error) { return *m.from, nil }
func (m *message) To() *common.Address                   { return m.to }
func (m *message) GasPrice() *big.Int                    { return m.gasPrice }
func (m *message) Gas() *big.Int                         { return m.gas }
func (m *message) Value() *big.Int                       { return m.value }
func (m *message) Nonce() uint64                         { return m.nonce }
func (m *message) Data() []byte                          { return m.data }

var (
	app *cli.App
	// output and logging
	DebugFlag = cli.BoolFlag{
		Name:  "debug",
		Usage: "output full trace logs",
	}
	VerbosityFlag = cli.IntFlag{
		Name:  "verbosity",
		Usage: "sets the verbosity level",
	}
	SysStatFlag = cli.BoolFlag{
		Name:  "sysstat",
		Usage: "display system stats",
	}
	DumpFlag = cli.BoolFlag{
		Name:  "dump",
		Usage: "dumps the state after the run",
	}

	// jit options
	ForceJitFlag = cli.BoolFlag{
		Name:  "forcejit",
		Usage: "forces jit compilation",
	}
	DisableJitFlag = cli.BoolFlag{
		Name:  "nojit",
		Usage: "disabled jit compilation",
	}

	// call arguments
	CodeFlag = cli.StringFlag{
		Name:  "code",
		Usage: "EVM code",
	}
	GasFlag = cli.StringFlag{
		Name:  "gas",
		Usage: "gas limit for the evm",
		Value: "10000000000",
	}
	PriceFlag = cli.StringFlag{
		Name:  "price",
		Usage: "price set for the evm",
		Value: "0",
	}
	ValueFlag = cli.StringFlag{
		Name:  "value",
		Usage: "value set for the evm",
		Value: "0",
	}
	InputFlag = cli.StringFlag{
		Name:  "input",
		Usage: "input for the EVM",
	}
	FromFlag = cli.StringFlag{
		Name:  "from",
		Usage: "address sending the call",
		Value: common.StringToAddress("sender").Hex(),
	}
	ToFlag = cli.StringFlag{
		Name:  "to",
		Usage: "destination address receiving the call",
		Value: common.StringToAddress("evmuser").Hex(),
	}

	// state options
	DataDirFlag = cli.StringFlag{
		Name:  "datadir",
		Usage: "directory to load/store persistent state",
	}
	RootFlag = cli.StringFlag{
		Name:  "root",
		Usage: "state root to load",
	}
	CreateFlag = cli.BoolFlag{
		Name:  "create",
		Usage: "set to create contract",
	}

	// block arguments
	HashFlag = cli.StringFlag{
		Name:  "block_hash",
		Usage: "specify block hash",
		Value: common.ToHex([]byte("nothing")),
	}
	CoinbaseFlag = cli.StringFlag{
		Name:  "coinbase",
		Usage: "set coinbase address",
		Value: common.StringToAddress("coinbase").Hex(),
	}
	DifficultyFlag = cli.StringFlag{
		Name:  "difficulty",
		Usage: "mining difficulty",
		Value: "0",
	}
	NumberFlag = cli.StringFlag{
		Name:  "number",
		Usage: "block number",
		Value: "0",
	}
	GasLimitFlag = cli.StringFlag{
		Name:  "gas-limit",
		Usage: "set the per-block gas-limit",
		Value: "10000000",
	}
	TimeFlag = cli.StringFlag{
		Name:  "time",
		Usage: "last block time",
		Value: "0",
	}
)

func init() {
	app = utils.NewApp("0.2", "the evm command line interface")
	app.Flags = []cli.Flag{
		DebugFlag,
		VerbosityFlag,
		SysStatFlag,
		DumpFlag,

		ForceJitFlag,
		DisableJitFlag,

		CodeFlag,
		GasFlag,
		PriceFlag,
		ValueFlag,
		InputFlag,
		FromFlag,
		ToFlag,

		DataDirFlag,
		RootFlag,
		CreateFlag,

		DifficultyFlag,
		NumberFlag,
		GasLimitFlag,
		TimeFlag,
	}
	app.Action = run
}

func run(ctx *cli.Context) error {
	glog.SetToStderr(true)
	glog.SetV(ctx.GlobalInt(VerbosityFlag.Name))

	dataDir := ctx.GlobalString(DataDirFlag.Name)
	rootHash := common.Hash{}
	var db ethdb.Database
	if dataDir == "" {
		db, _ = ethdb.NewMemDatabase()
	} else {
		var exists bool
		// check if dir already exists
		p := filepath.Join(dataDir, "evm")
		if _, err := os.Stat(p); err == nil {
			exists = true
			fmt.Println("Datadir already exists")

		}

		fmt.Println("Loading database")
		// load db
		var err error
		db, err = ethdb.NewLDBDatabase(p, 128, 64) // cache, handles
		if err != nil {
			panic(err)
		}

		// if state already exists, load latest specified or latest hash
		if exists {
			if rootFlag := ctx.GlobalString(RootFlag.Name); rootFlag != "" {
				rootHash = common.HexToHash(rootFlag)
			} else {
				// load the latest block, grab hash
				root, err := db.Get([]byte(LatestStateRootKey))
				if err == nil {
					rootHash = common.BytesToHash(root)
				}

			}
		}
	}

	fmt.Printf("Loading root hash %X\n", rootHash)
	statedb, _ := state.New(rootHash, db)

	senderAddress := common.HexToAddress(ctx.GlobalString(FromFlag.Name))
	receiverAddress := common.HexToAddress(ctx.GlobalString(ToFlag.Name))
	gasBig := common.Big(ctx.GlobalString(GasFlag.Name))
	gasPriceBig := common.Big(ctx.GlobalString(PriceFlag.Name))
	valueBig := common.Big(ctx.GlobalString(ValueFlag.Name))
	dataBytes := common.FromHex(ctx.GlobalString(InputFlag.Name))

	var sender vm.Account
	// if sender doesn't exist, create
	if !statedb.HasAccount(senderAddress) {
		sender = statedb.CreateAccount(senderAddress)
	} else {
		sender = statedb.GetAccount(senderAddress)
	}

	// make chain config with log and jit options
	chainConfig := core.MakeChainConfig()
	chainConfig.VmConfig = vm.Config{
		Debug:     ctx.GlobalBool(DebugFlag.Name),
		ForceJit:  ctx.GlobalBool(ForceJitFlag.Name),
		EnableJit: !ctx.GlobalBool(DisableJitFlag.Name),
	}

	// make a phony blockchain
	// TODO: track blocks
	chainDb, _ := ethdb.NewMemDatabase()
	evmux := &event.TypeMux{}
	bc, err := core.NewBlockChain(chainDb, chainConfig, core.FakePow{}, evmux)
	if err != nil {
		panic(err)
	}

	// make the message from the flags
	msg := &message{
		from:     &senderAddress,
		to:       &receiverAddress,
		gasPrice: gasPriceBig,
		gas:      gasBig,
		value:    valueBig,
		nonce:    0,
		data:     dataBytes,
	}

	hash := common.HexToHash(ctx.GlobalString(HashFlag.Name))
	coinbase := common.HexToAddress(ctx.GlobalString(CoinbaseFlag.Name))
	difficulty := common.Big(ctx.GlobalString(DifficultyFlag.Name))
	number := common.Big(ctx.GlobalString(NumberFlag.Name))
	gasLimit := common.Big(ctx.GlobalString(GasLimitFlag.Name))
	timestamp := common.Big(ctx.GlobalString(TimeFlag.Name))

	header := &types.Header{
		ParentHash: hash,       // common.Hash
		Coinbase:   coinbase,   // common.Address
		Difficulty: difficulty, // rest are *big.Int
		Number:     number,
		GasLimit:   gasLimit,
		Time:       timestamp,
	}

	vmenv := core.NewEnv(statedb, chainConfig, bc, msg, header, chainConfig.VmConfig)

	tstart := time.Now()

	var (
		ret []byte
	)

	if ctx.GlobalBool(CreateFlag.Name) {
		input := append(common.Hex2Bytes(ctx.GlobalString(CodeFlag.Name)), common.Hex2Bytes(ctx.GlobalString(InputFlag.Name))...)
		ret, _, err = vmenv.Create(
			sender,
			input,
			common.Big(ctx.GlobalString(GasFlag.Name)),
			common.Big(ctx.GlobalString(PriceFlag.Name)),
			common.Big(ctx.GlobalString(ValueFlag.Name)),
		)
	} else {
		var receiver vm.Account
		// if receiver doesn't exist, create
		if !statedb.HasAccount(receiverAddress) {
			receiver = statedb.CreateAccount(receiverAddress)
		} else {
			receiver = statedb.GetAccount(receiverAddress)
		}
		receiver.SetCode(common.Hex2Bytes(ctx.GlobalString(CodeFlag.Name)))
		ret, err = vmenv.Call(
			sender,
			receiver.Address(),
			common.Hex2Bytes(ctx.GlobalString(InputFlag.Name)),
			common.Big(ctx.GlobalString(GasFlag.Name)),
			common.Big(ctx.GlobalString(PriceFlag.Name)),
			common.Big(ctx.GlobalString(ValueFlag.Name)),
		)
	}
	vmdone := time.Since(tstart)

	rootHash, err = statedb.Commit()
	if err != nil {
		panic(err)
	}
	if err := db.Put([]byte(LatestStateRootKey), rootHash.Bytes()); err != nil {
		panic(err)
	}

	if ctx.GlobalBool(DumpFlag.Name) {
		statedb.Commit()
		fmt.Println(string(statedb.Dump()))
	}
	vm.StdErrFormat(vmenv.StructLogs())

	if ctx.GlobalBool(SysStatFlag.Name) {
		var mem runtime.MemStats
		runtime.ReadMemStats(&mem)
		fmt.Printf("vm took %v\n", vmdone)
		fmt.Printf(`alloc:      %d
tot alloc:  %d
no. malloc: %d
heap alloc: %d
heap objs:  %d
num gc:     %d
`, mem.Alloc, mem.TotalAlloc, mem.Mallocs, mem.HeapAlloc, mem.HeapObjects, mem.NumGC)
	}

	fmt.Printf("OUT: 0x%x", ret)
	if err != nil {
		fmt.Printf(" error: %v", err)
	}
	fmt.Println()
	return nil
}

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
