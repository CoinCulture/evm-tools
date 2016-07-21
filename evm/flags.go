package main

import (
	"github.com/ethereum/go-ethereum/common"
	"gopkg.in/urfave/cli.v1"
)

var appFlags = []cli.Flag{
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

	DifficultyFlag,
	NumberFlag,
	GasLimitFlag,
	TimeFlag,
}

var (
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
		Value: "1",
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
		Value: common.StringToAddress("receiver").Hex(),
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
