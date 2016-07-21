# Notes on the EVM

This is a guide to understanding the EVM, its relationship with solidity, and how to use some debugging tools.

## Overview

The EVM is a stack-based virtual machine with a memory byte-array and key-value storage (persisted in a merkle tree).
Elements on the stack are 32-byte words, and all keys and values in storage are 32 bytes.
There are over 100 opcodes, divided into categories deliniated in multiples of 16.
Here is the list from the pyethereum client, annotated with rough category names.

```python
# schema: [opcode, ins, outs, gas]

opcodes = {

    # arithmetic
    0x00: ['STOP', 0, 0, 0],
    0x01: ['ADD', 2, 1, 3],
    0x02: ['MUL', 2, 1, 5],
    0x03: ['SUB', 2, 1, 3],
    0x04: ['DIV', 2, 1, 5],
    0x05: ['SDIV', 2, 1, 5],
    0x06: ['MOD', 2, 1, 5],
    0x07: ['SMOD', 2, 1, 5],
    0x08: ['ADDMOD', 3, 1, 8],
    0x09: ['MULMOD', 3, 1, 8],
    0x0a: ['EXP', 2, 1, 10],
    0x0b: ['SIGNEXTEND', 2, 1, 5],

    # boolean
    0x10: ['LT', 2, 1, 3],
    0x11: ['GT', 2, 1, 3],
    0x12: ['SLT', 2, 1, 3],
    0x13: ['SGT', 2, 1, 3],
    0x14: ['EQ', 2, 1, 3],
    0x15: ['ISZERO', 1, 1, 3],
    0x16: ['AND', 2, 1, 3],
    0x17: ['OR', 2, 1, 3],
    0x18: ['XOR', 2, 1, 3],
    0x19: ['NOT', 1, 1, 3],
    0x1a: ['BYTE', 2, 1, 3],

    # crypto
    0x20: ['SHA3', 2, 1, 30],
    
    # contract context
    0x30: ['ADDRESS', 0, 1, 2],
    0x31: ['BALANCE', 1, 1, 20],
    0x32: ['ORIGIN', 0, 1, 2],
    0x33: ['CALLER', 0, 1, 2],
    0x34: ['CALLVALUE', 0, 1, 2],
    0x35: ['CALLDATALOAD', 1, 1, 3],
    0x36: ['CALLDATASIZE', 0, 1, 2],
    0x37: ['CALLDATACOPY', 3, 0, 3],
    0x38: ['CODESIZE', 0, 1, 2],
    0x39: ['CODECOPY', 3, 0, 3],
    0x3a: ['GASPRICE', 0, 1, 2],
    0x3b: ['EXTCODESIZE', 1, 1, 20],
    0x3c: ['EXTCODECOPY', 4, 0, 20],

    # blockchain context
    0x40: ['BLOCKHASH', 1, 1, 20],
    0x41: ['COINBASE', 0, 1, 2],
    0x42: ['TIMESTAMP', 0, 1, 2],
    0x43: ['NUMBER', 0, 1, 2],
    0x44: ['DIFFICULTY', 0, 1, 2],
    0x45: ['GASLIMIT', 0, 1, 2],
  
    # storage and execution
    0x50: ['POP', 1, 0, 2],
    0x51: ['MLOAD', 1, 1, 3],
    0x52: ['MSTORE', 2, 0, 3],
    0x53: ['MSTORE8', 2, 0, 3],
    0x54: ['SLOAD', 1, 1, 50],
    0x55: ['SSTORE', 2, 0, 0],
    0x56: ['JUMP', 1, 0, 8],
    0x57: ['JUMPI', 2, 0, 10],
    0x58: ['PC', 0, 1, 2],
    0x59: ['MSIZE', 0, 1, 2],
    0x5a: ['GAS', 0, 1, 2],
    0x5b: ['JUMPDEST', 0, 0, 1],

    # logging
    0xa0: ['LOG0', 2, 0, 375],
    0xa1: ['LOG1', 3, 0, 750],
    0xa2: ['LOG2', 4, 0, 1125],
    0xa3: ['LOG3', 5, 0, 1500],
    0xa4: ['LOG4', 6, 0, 1875],
	
    # arbitrary length storage (proposal for metropolis hardfork)
    0xe1: ['SLOADBYTES', 3, 0, 50],
    0xe2: ['SSTOREBYTES', 3, 0, 0],
    0xe3: ['SSIZE', 1, 1, 50],

    # closures
    0xf0: ['CREATE', 3, 1, 32000],
    0xf1: ['CALL', 7, 1, 40],
    0xf2: ['CALLCODE', 7, 1, 40],
    0xf3: ['RETURN', 2, 0, 0],
    0xf4: ['DELEGATECALL', 6, 0, 40],
    0xff: ['SUICIDE', 1, 0, 0],
}

# push
for i in range(1, 33):
    opcodes[0x5f + i] = ['PUSH' + str(i), 0, 1, 3]

# duplicate and swap
for i in range(1, 17):
    opcodes[0x7f + i] = ['DUP' + str(i), i, i + 1, 3]
    opcodes[0x8f + i] = ['SWAP' + str(i), i + 1, i + 1, 3]
```

The table tells us how many arguments each opcode pops off the stack and pushes back onto the stack,
as well as how much gas is consumed. In addition
Most opcodes take some number of arguments off the stack, and push one or no results back onto the stack.
Some, like GAS and PC, take no arguments off the stack, and push the remaining gas and program counter,
respectively, onto the stack. 
A number of opcodes, like SHA3, CREATE, and RETURN, take arguments off the stack that refer to 
positions and sizes in memory, allowing them to operate on a contiguous array of memory.
Each

All arithmetic happens on big integers using elements on the stack (ie. 32-byte Big Endian integers).
Currently, the only crypto operation is the SHA3 hash function, 
which takes a position in memory and a length to read input from and outputs the hash on the stack.
Contract and blockchain level contexts give access to various useful environmental information - 
for instance CALLDATACOPY will copy the input data sent to the contract (known as call-data) into memory, and NUMBER can be used
to time-lock behaviour by block number.
The EVM operates on its ephemeral memory via MLOAD and MSTORE, and on its persistent storage via SLOAD and SSTORE.
JUMP can be used to jump to arbitrary points in the program, and those points must be a JUMPDEST.
The PUSH1-PUSH32 opcodes push anywhere from 1 to 32 bytes to the stack.
The DUP1-DUP16 opcodes push a duplicate of one of the top 16 elements of the stack to the top of the stack.
The SWAP1-SWAP16 opcodes swap the top element of the stack with any of the preceeding 16.
The LOG opcodes enable event logging which is recorded in blocks and can be verified efficiently by light clients.
Finally, CALL and CREATE allow contracts to call and create other contracts, respectively, while RETURN returns a chunk of memory from a call,
and SUICIDE causes the contract to be destroyed and return all funds to a specified address.

The specification for each opcode can be found in the [yellow paper](http://gavwood.com/paper.pdf)([source on github](https://github.com/ethereum/yellowpaper)), 
or in the implementation of the EVM in your favorite language.

Note the EVM is redundantly Turing complete - 
it has both the primitives of a Turing tape (ops for managing memory and jumping to arbitrary points in the program),
and those of an agent-based message passing system, 
where agents may have arbitrary code (ops for calling and creating other contracts, returning values).
To force all executions to terminate, each operation is tagged with an explicit cost, denominated in *gas*. 
Executions must specify a maximum amount of gas, such that using more than that amount throws an OutOfGas exception.
The list of opcodes specifies how much gas each opcode consumes. In addition, certain operations consume amounts of gas
that are parameterized by the following (see the yellow paper for more details):

```
# Non-opcode gas prices
GDEFAULT = 1
GMEMORY = 3
GQUADRATICMEMDENOM = 512  # 1 gas per 512 quadwords
GSTORAGEREFUND = 15000
GSTORAGEKILL = 5000
GSTORAGEMOD = 5000
GSTORAGEADD = 20000
GEXPONENTBYTE = 10    # cost of EXP exponent per byte
GCOPY = 3             # cost to copy one 32 byte word
GCONTRACTBYTE = 200   # one byte of code in contract creation
GCALLVALUETRANSFER = 9000   # non-zero-valued call
GLOGBYTE = 8          # cost of a byte of logdata

GTXCOST = 21000       # TX BASE GAS COST
GTXDATAZERO = 4       # TX DATA ZERO BYTE GAS COST
GTXDATANONZERO = 68   # TX DATA NON ZERO BYTE GAS COST
GSHA3WORD = 6         # Cost of SHA3 per word
GSHA256BASE = 60      # Base c of SHA256
GSHA256WORD = 12      # Cost of SHA256 per word
GRIPEMD160BASE = 600  # Base cost of RIPEMD160
GRIPEMD160WORD = 120  # Cost of RIPEMD160 per word
GIDENTITYBASE = 15    # Base cost of indentity
GIDENTITYWORD = 3     # Cost of identity per word
GECRECOVER = 3000     # Cost of ecrecover op

GSTIPEND = 2300

GCALLNEWACCOUNT = 25000
GSUICIDEREFUND = 24000
```

Other exceptions, besides out-of-gas, include invalid op codes, stack underflow, and invalid jump destinations.
There is also a stack size limit, such that the stack can only be so big, 
and a call-depth limit, such that chains of calls from contracts to other contracts can only be so long,
for instance causing recursive invocations of a contract to eventually halt, despite the amount of gas provided.
Ethereum transactions are atomic - if an exception is thrown, all state transitions are reverted.
The only exception to this rule (no pun intended) is gas payment - any gas used up until the OutOfGas exception is deducted and sent to the miners.
Note that such transactions are still included in blocks so that they can pay fees - if they were not, this would provide
a significant DoS attack vector against miners.

# Execution

Let us look at some simple executions. To do so, I have collected some useful tools in a single repo, 
including forks of some nice nice tools provided by go-ethereum. 
Make sure you have Go installed, set your `GOPATH` environment variable to whatever you want, 
and add `$GOPATH/bin` to your `PATH`. Then run:

```
go get github.com/ebuchman/evm-tools/...
```

This will install the following tools, now accessible from $GOPATH/bin: `evm`, `disasm`, `evm-deploy`.

TODO: We also use some python ...

Now, here is some very simple bytecode I wrote:

```
6005600401
```

To disassemble, run `echo 6005600401 | disasm`, which produces:

```
0      PUSH1  => 05
2      PUSH1  => 04
4      ADD
```

So, this simple program pushes the numbers `05` and `04` to the stack and adds them.

We can run it through the EVM with `evm --debug --code 6005600401`, we get something like:

```
VM STAT 4 OPs
PC 00000000: PUSH1 GAS: 9999999997 COST: 3
STACK = 0
MEM = 0
STORAGE = 0

PC 00000002: PUSH1 GAS: 9999999994 COST: 3
STACK = 1
0000: 0000000000000000000000000000000000000000000000000000000000000005
MEM = 0
STORAGE = 0

PC 00000004: ADD GAS: 9999999991 COST: 3
STACK = 2
0000: 0000000000000000000000000000000000000000000000000000000000000004
0001: 0000000000000000000000000000000000000000000000000000000000000005
MEM = 0
STORAGE = 0

PC 00000005: STOP GAS: 9999999991 COST: 0
STACK = 1
0000: 0000000000000000000000000000000000000000000000000000000000000009
MEM = 0
STORAGE = 0
```

The `--debug` flag prints the current state of the stack, memory, and storage for us at each step, 
and shows us each opcode and the gas cost. Note how the 0x04 and 0x05 are pushed to the stack (padded to 32-bytes)
and consumed by ADD, which leaves the result, 0x09, on the stack.
To have the value returned, instead of simply left on the stack,
we need to modify the bytecode so the value is copied into memory and then returned:

```
$ echo 60056004016000526001601ff3  | disasm
60056004016000526001601ff3
0      PUSH1  => 05
2      PUSH1  => 04
4      ADD
5      PUSH1  => 00
7      MSTORE
8      PUSH1  => 01
10     PUSH1  => 1f
12     RETURN
```

The value (0x09) is being stored in memory at position 0x0. 
However, since the element being stored comes from the stack, 
it is a 32-byte word, the Big Endian encoding (ie. left-padded with zeros) of 0x09.
So, to return just `0x09`, we return a byte-array of length 0x01, starting from position 0x1f.
Alternatively, we could return a byte-array of length 0x20 starting from position 0x00 - 
then the returned value would be left-padded with zeroes to 32-bytes.

Run the above code with `evm --debug --code 60056004016000526001601ff3`:

```
VM STAT 8 OPs
PC 00000000: PUSH1 GAS: 9999999997 COST: 3
STACK = 0
MEM = 0
STORAGE = 0

PC 00000002: PUSH1 GAS: 9999999994 COST: 3
STACK = 1
0000: 0000000000000000000000000000000000000000000000000000000000000005
MEM = 0
STORAGE = 0

PC 00000004: ADD GAS: 9999999991 COST: 3
STACK = 2
0000: 0000000000000000000000000000000000000000000000000000000000000004
0001: 0000000000000000000000000000000000000000000000000000000000000005
MEM = 0
STORAGE = 0

PC 00000005: PUSH1 GAS: 9999999988 COST: 3
STACK = 1
0000: 0000000000000000000000000000000000000000000000000000000000000009
MEM = 0
STORAGE = 0

PC 00000007: MSTORE GAS: 9999999982 COST: 6
STACK = 2
0000: 0000000000000000000000000000000000000000000000000000000000000000
0001: 0000000000000000000000000000000000000000000000000000000000000009
MEM = 32
0000: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
0016: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
STORAGE = 0

PC 00000008: PUSH1 GAS: 9999999979 COST: 3
STACK = 0
MEM = 32
0000: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
0016: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 09  ...............?
STORAGE = 0

PC 00000010: PUSH1 GAS: 9999999976 COST: 3
STACK = 1
0000: 0000000000000000000000000000000000000000000000000000000000000001
MEM = 32
0000: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
0016: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 09  ...............?
STORAGE = 0

PC 00000012: RETURN GAS: 9999999976 COST: 0
STACK = 2
0000: 000000000000000000000000000000000000000000000000000000000000001f
0001: 0000000000000000000000000000000000000000000000000000000000000001
MEM = 32
0000: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
0016: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 09  ...............?
STORAGE = 0

OUT: 0x09
```

Note how the 32-byte, big endian 0x09 is stored in memory, and how the program finally outputs a `0x09`.

If the arguments we are adding are larger than one byte, we can use a different PUSH operator.
For instance, to add two-byte numbers like 257 (0x0101) and 258 (0x0102), we use PUSH2 (0x61):

```
$ echo 61010161010201 | disasm
61010161010201
0      PUSH2  => 0101
3      PUSH2  => 0102
6      ADD
```

And the execution `evm --debug --code 61010161010201` gives

```
VM STAT 4 OPs
PC 00000000: PUSH2 GAS: 9999999997 COST: 3
STACK = 0
MEM = 0
STORAGE = 0

PC 00000003: PUSH2 GAS: 9999999994 COST: 3
STACK = 1
0000: 0000000000000000000000000000000000000000000000000000000000000101
MEM = 0
STORAGE = 0

PC 00000006: ADD GAS: 9999999991 COST: 3
STACK = 2
0000: 0000000000000000000000000000000000000000000000000000000000000102
0001: 0000000000000000000000000000000000000000000000000000000000000101
MEM = 0
STORAGE = 0

PC 00000007: STOP GAS: 9999999991 COST: 0
STACK = 1
0000: 0000000000000000000000000000000000000000000000000000000000000203
MEM = 0
STORAGE = 0
```

where `0x0203 = 515 = 257 + 258`

What if we want to pass the arguments as call-data, rather than hardcoding them?
We need to first agree on a formatting discipline - say, all input values 
are left-padded to 32-bytes, for convenience. Then we can do the following:

```
$ echo 60003560203501 | disasm
60003560203501
0      PUSH1  => 00
2      CALLDATALOAD
3      PUSH1  => 20
5      CALLDATALOAD
6      ADD
```

To execute, we must pass correctly padded input:

```
$ evm --debug --code 60003560203501 --input 00000000000000000000000000000000000000000000000000000000000000050000000000000000000000000000000000000000000000000000000000000004
VM STAT 6 OPs
PC 00000000: PUSH1 GAS: 9999999997 COST: 3
STACK = 0
MEM = 0
STORAGE = 0

PC 00000002: CALLDATALOAD GAS: 9999999994 COST: 3
STACK = 1
0000: 0000000000000000000000000000000000000000000000000000000000000000
MEM = 0
STORAGE = 0

PC 00000003: PUSH1 GAS: 9999999991 COST: 3
STACK = 1
0000: 0000000000000000000000000000000000000000000000000000000000000005
MEM = 0
STORAGE = 0

PC 00000005: CALLDATALOAD GAS: 9999999988 COST: 3
STACK = 2
0000: 0000000000000000000000000000000000000000000000000000000000000020
0001: 0000000000000000000000000000000000000000000000000000000000000005
MEM = 0
STORAGE = 0

PC 00000006: ADD GAS: 9999999985 COST: 3
STACK = 2
0000: 0000000000000000000000000000000000000000000000000000000000000004
0001: 0000000000000000000000000000000000000000000000000000000000000005
MEM = 0
STORAGE = 0

PC 00000007: STOP GAS: 9999999985 COST: 0
STACK = 1
0000: 0000000000000000000000000000000000000000000000000000000000000009
MEM = 0
STORAGE = 0
```

What if you want your program to have multiple possible functions?
The combination of these problems, of formatting call-data and calling
one of many functions, gave rise to an [Application Binary Interface (ABI)
standard](https://github.com/ethereum/wiki/wiki/Ethereum-Contract-ABI), 
respected by the high-level programming languages (solidity, serpent, etc.).
We will discuss this later.

First, how can we do control flow? Using boolean expressions and jumps, of course!
Here is a simple loop:

```
$ echo  6000356000525b600160005103600052600051600657 | disasm
6000356000525b600160005103600052600051600657
0      PUSH1  => 00
2      CALLDATALOAD
3      PUSH1  => 00
5      MSTORE
6      JUMPDEST
7      PUSH1  => 01
9      PUSH1  => 00
11     MLOAD
12     SUB
13     PUSH1  => 00
15     MSTORE
16     PUSH1  => 00
18     MLOAD
19     PUSH1  => 06
21     JUMPI
```

Here, we load some value (the counter) from the call-data and loop that many times 
by storing the counter in memory (at position 0x0) and decrementing on each pass through the loop.
The loop essentially starts at the `JUMPDEST`. The final opcode, `JUMPI`, takes a value and a location, 
and if the value is non-zero, jumps to the location in the program. If the location is not a `JUMPDEST`, 
the execution throws an exception. In this case, the `JUMPDEST` is at position `0x06`,
and the value it checks is the counter variable, loaded from memory.

Run the loop five times with 
`evm --debug --code 6000356000525b600160005103600052600051600657 --input 0000000000000000000000000000000000000000000000000000000000000005`
See if you can decipher the code - look for the counter variable decrementing in memory.

What happens if we run the code without any input, or with an input of zero? Will the loop run zero times?
Why or why not? (HINT: the EVM has no notion of negative numbers, so -1 is really `2^256 - 1` or `0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff`). To solve this, check if the input value is zero before entering the loop (hint: use ISZERO).

Note there is a significant inefficiency here in our use of memory, since we are constantly loading from, 
then storing to, the same spot in memory multiple times.
Instead of accessing the memory-byte array, we could just keep what we need on the stack, using the DUP and SWAP opcodes.
The solidity compiler makes these kinds of optimizations all the time. Here is the loop without using memory:

```
$ echo 6000355b6001900380600357 | disasm
6000355b6001900380600357
0      PUSH1  => 00
2      CALLDATALOAD
3      JUMPDEST
4      PUSH1  => 01
6      SWAP1
7      SUB
8      DUP1
9      PUSH1  => 03
11     JUMPI
```

Much simpler! The SWAP makes sure the counter (the value to be decremented) is at the top of the stack, 
which is what the SUB opcode expects.
The DUP1 is used to duplicate the counter on the stack, so it can be consumed by JUMPI and still available 
for subtraction the next time through the loop. Otherwise, the loop works exactly the same way.
Note, also, that since we do not store the counter to memory before the loop starts, the JUMPDEST is
at position 0x03 instead of 0x06.

Run this loop five times with
`evm --debug --code 6000355b6001900380600357 --input 0000000000000000000000000000000000000000000000000000000000000005`
and watch the counter persist and decrement on the stack, instead of in memory.

One more improvement before moving on. Passing a 32-byte padded input string is awful;
the call-data should be only as big as it needs to be. In this case, we would like a loop of 5 times to be called with
`--input 05` and one that run 257 times with `--input 0101`. 
Problem is, CALLDATALOAD loads 32-byte big-endian numbers, so `--input 05` 
becomes the massive `0500000000000000000000000000000000000000000000000000000000000000` on the stack.
Since there is no byte shifting operator in the EVM, we have to use division.
In this case, we want to divide by `256^(32-L)`, where `L` is the length of the call-data.
This has the effect of byte-shifting to the right by `(32-L)` bytes. 
The updated byte-code looks like:

```
$ echo  366020036101000a600035045b6001900380600c57 | disasm
366020036101000a600035045b6001900380600c57
0      CALLDATASIZE
1      PUSH1  => 20
3      SUB
4      PUSH2  => 0100
7      EXP
8      PUSH1  => 00
10     CALLDATALOAD
11     DIV
12     JUMPDEST
13     PUSH1  => 01
15     SWAP1
16     SUB
17     DUP1
18     PUSH1  => 0c
20     JUMPI
```

and we can run the loop five times with `evm --debug --code 366020036101000a600035045b6001900380600c57 --input 05`
or 257 times with `evm --debug --code 366020036101000a600035045b6001900380600c57 --input 0101`.
Make sure you understand how the EXP and DIV are being used to achieve byte shifting - this is a very common paradigm
used extensively by the higher level languages.

# Contracts

So far, we have only looked at the base execution environment of the EVM.
But the EVM is embedded in a blockchain state divided into accounts.
All accounts in ethereum are stored in a merkle radix tree.
Programs in the EVM live in *accounts* known as *contracts*. 
In addition to an address, a balance, and a sequence number (equal to the number of transactions sent by the account - also known as a nonce),
contracts keep the hash of their EVM bytecode, and the merkle root of their internal storage tree.
An account can have at most one program associated with it - 
any time a transaction is made to the contract, or it is the target of another contract executing the CALL opcode, 
the code of that contract will execute. 
Note that once deployed, the code of a contract may not be changed.
The merkle root of the account/contract storage is updated after any successful transaction where execution of the SSTORE opcode results
in a value being stored at a new key or a change to the value stored at an existing key.

Contract creation happens in a special way, by sending a transaction to the empty address with the contract code as data.
The ethereum state transition function will interpret this event as a contract creation event by creating a new account,
running the program specified in the call-data, and setting whatever is returned by the EVM as the code for the new contract.
That is, the code sent during creation is not the same as that which will be stored in the contract - it is instead the so called
"deploy-code", which contains the actual contract code wrapped in some operations that will copy it into memory and return it.

For instance, if we take one of the programs we have written (that does not return anything) and send it as data to the empty address, 
the program will execute, but the resulting account will have no code, so any transactions to that account will cause no code to run.

Looking at the simple addition program `6005600401` as an example, 
we can generate the deploy with the `evm-deploy` tool:

```
$ echo 6005600401 | evm-deploy | disasm 
600580600b6000396000f36005600401
0      PUSH1  => 05
2      DUP1
3      PUSH1  => 0b
5      PUSH1  => 00
7      CODECOPY
8      PUSH1  => 00
10     RETURN
11     PUSH1  => 05
13     PUSH1  => 04
15     ADD
```

Here, we know the program of interest is length `0x05`, and we know it is embedded in the larger deploy-code, 
starting at position 11 (0x0b). So we copy this chunk of code into memory (position 0x00) and return it. 
Note that using the `DUP1` keeps the length of the code (in this case, 0x05) on the stack for both the CODECOPY and the RETURN. 
When the deploy-code is run, the return value should be the code of interest, ie. `6005600401`:

```
$ evm --debug --code 600580600b6000396000f36005600401
VM STAT 7 OPs
PC 00000000: PUSH1 GAS: 9999999997 COST: 3
STACK = 0
MEM = 0
STORAGE = 0

PC 00000002: DUP1 GAS: 9999999994 COST: 3
STACK = 1
0000: 0000000000000000000000000000000000000000000000000000000000000005
MEM = 0
STORAGE = 0

PC 00000003: PUSH1 GAS: 9999999991 COST: 3
STACK = 2
0000: 0000000000000000000000000000000000000000000000000000000000000005
0001: 0000000000000000000000000000000000000000000000000000000000000005
MEM = 0
STORAGE = 0

PC 00000005: PUSH1 GAS: 9999999988 COST: 3
STACK = 3
0000: 000000000000000000000000000000000000000000000000000000000000000b
0001: 0000000000000000000000000000000000000000000000000000000000000005
0002: 0000000000000000000000000000000000000000000000000000000000000005
MEM = 0
STORAGE = 0

PC 00000007: CODECOPY GAS: 9999999979 COST: 9
STACK = 4
0000: 0000000000000000000000000000000000000000000000000000000000000000
0001: 000000000000000000000000000000000000000000000000000000000000000b
0002: 0000000000000000000000000000000000000000000000000000000000000005
0003: 0000000000000000000000000000000000000000000000000000000000000005
MEM = 32
0000: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
0016: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
STORAGE = 0

PC 00000008: PUSH1 GAS: 9999999976 COST: 3
STACK = 1
0000: 0000000000000000000000000000000000000000000000000000000000000005
MEM = 32
0000: 60 05 60 04 01 00 00 00 00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
STORAGE = 0

PC 00000010: RETURN GAS: 9999999976 COST: 0
STACK = 2
0000: 0000000000000000000000000000000000000000000000000000000000000000
0001: 0000000000000000000000000000000000000000000000000000000000000005
MEM = 32
0000: 60 05 60 04 01 00 00 00 00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
STORAGE = 0

OUT: 0x6005600401
```

Tada! 


# Stateful EVM

I made some modifications to the `evm` tool so it can persist state between invocations. 
Just use the `--datadir` flag.
For instance, let us use the contract that returns the sum of 0x5 and 0x4.
We can run the deploy code, so the contract is actually deployed 
(an account created with the correct code) and then we can interact with it:

```
$ evm --code $(echo "60056004016000526001601ff3" | evm-deploy)  --datadir evm-data
Loading database
Loading root hash 0000000000000000000000000000000000000000000000000000000000000000
Contract Address: 1F2A98889594024BFFDA3311CBE69728D392C06D
VM STAT 0 OPs
OUT: 0x60056004016000526001601ff3
```

Note it gave us the new contract address. Where did this address come from?
It is the sha3 hash of the [RLP](https://github.com/ethereum/wiki/wiki/RLP) encoding of the list `[address of sender, sequence number of sender]`.
The default sender is `0x000000000000000000000000000073656e646572` (that is hex for `sender`),
and the sequence number starts at `0x0`. In python:

```
>>> import rlp, sha3
>>> sha3.sha3_256(rlp.encode(["000000000000000000000000000073656e646572".decode('hex'), 0])).hexdigest()[24:]
'1f2a98889594024bffda3311cbe69728d392c06d'
```

Note this means the address is strictly deterministic starting from the same state. 
If we deploy the contract, again, the sequence number of the sender will be 0x1:

```
$ evm --code $(echo "60056004016000526001601ff3" | evm-deploy)  --datadir evm-data
Datadir already exists
Loading database
Loading root hash BFDDB19821CE2BFAB71C4BA9E8ADC6CF083DAE0EF9206AA506BC88B0F9064182
Contract Address: 14F6D12ECEBB7606C528880AD8B97C25AB7D4AD9
VM STAT 0 OPs
OUT: 0x60056004016000526001601ff3
```

Same output, new contract address:

```
>>> import rlp, sha3
>>> sha3.sha3_256(rlp.encode(["000000000000000000000000000073656e646572".decode('hex'), 1])).hexdigest()[24:]
'14f6d12ecebb7606c528880ad8b97c25ab7d4ad9'
```

Now we can send a transaction to the contract:

```
$ evm --to 14F6D12ECEBB7606C528880AD8B97C25AB7D4AD9 --datadir evm-data
Datadir already exists
Loading database
Loading root hash 60209E93FEFD3DD5CF1D6B3FBDC33DA1B020C5B880A51E8306A3F5FDF269122A
Loaded account for receiver 14F6D12ECEBB7606C528880AD8B97C25AB7D4AD9
CODE: 60056004016000526001601FF3
VM STAT 0 OPs
OUT: 0x09
```

Woopie!

# Exceptions

Here are some examples of exceptions.

Invalid opcode (`5f` is not a known opcode):

```
evm --debug --code 5f
```

Stack underflow (`JUMP` (0x56) expects at least one argument on the stack):

```
evm --debug --code 56
```

Invalid jump destination (the destination 0x0 is not a `JUMPDEST`, in this case its a `PUSH`):

```
evm --debug --code 600056
```

Here's an out-of-gas exception (`PUSH` requires 3 gas):

```
evm --debug --gas 1 --code 6000
```

See `evm --help` for options and defaults.

# Memory and Storage

In addition to the stack, the EVM comes with an emphemeral memory byte-array and persistent storage tree.
Access to the memory byte-array is relatively cheap, and out-of-bounds memory access is not an exception; 
memory grows as necessary when you access it, you simply pay the gas for the change in size. 

For instance, accessing memory location 0x1000 costs us quite a lot the first time, 
since the memory grows from size 0 to size 4096, and very little the second time, since
the size of the memory doesn't change:

```
evm --debug --code 611000805151
```

The storage size is practically infinite, or `2^256`, but is relatively expensive, at `20000` gas when writing a non-zero where there was previously a zero,
and `5000` otherwise. For instance, we store a 0x2 at position 0x0, then overwrite it with a 0x1:

```
evm --debug --code 60026000556001600055
```

# Solidity

Finally, we can talk about solidity. Solidity is a high-level, javascript-like, contract-oriented language
that compiles to EVM. It has many high-level features not found directly in the EVM, like types, arrays, and function calls.
It also conforms to the Ethereum ABI, a specification for how arguments and function calls should be encoded in the call-data.
In summary, the first four bytes of the call-data are the function identifier, 
corresponding to the first four bytes of the sha3 hash of the canonical version of the function signature.
The rest of the arguments are passed in padded to 32-bytes.

First, we need the solidity compiler, `solc`. Since it is written in C++, it is a pain to install.
Fortunately, you can use docker and the very nice image provided by Eris Industries.

First, let's create a working directory:

```
export SOLC_WORKSPACE=$HOME/solidity_workspace
mkdir $SOLC_WORKSPACE
```

Now we run the docker container (if you don't already have the eris/compilers image, this will download it):

```
docker run --name solc -v $SOLC_WORKSPACE:/home/eris/.eris -it quay.io/eris/compilers /bin/bash
```

The result of this command is to drop you in a container with the solidity compiler installed, 
where the directory `$SOLC_WORKSPACE` is mounted into the container at /home/eris/.eris
so that any new files or edits made in `$SOLC_WORKSPACE` will be reflected immediately in the container.
Run `solc --help` to ensure the compiler is properly installed.

Working with docker typically involves two terminal sessions, one in the container and one on your host.
The container session allows interactive access to whatever binaries you needed (in this case, `solc`),
while the host session allows files and changes to be made normally from the host, and immediately reflected in the container.

Open another window to be the host session, set the `SOLC_WORKSPACE` with `export SOLC_WORKSPACE=$HOME/solidity_workspace`,
and save the following simple solidity contract as `$SOLC_WORKSPACE/add.sol`:

```
contract Addition{
	int x;
        function add(int a, int b){
		x = a + b;
    	}
}
```

This contract allows users to call the `add` function, passing two arguments `a` and `b` whose sum is stored in the variable `x`.
Note that variables defined at the top of a contract are persisted in the contract storage tree (using `SSTORE`).

Back in the container terminal session, `cd /home/eris/.eris` and `ls`, you should see some folders and your solidity contract.

Compile the contract:

```
solc --bin-runtime --optimize -o . add.sol
```

In the host terminal session, you should see the contract under `$SOLC_WORKSPACE/Addition.bin-runtime`.

By using `--bin-runtime`, we get the code as it would be in the contract after having been deployed - 
we can test that with the `evm` tool. If we use `--bin` instead of `--bin-runtime`, and run that
through the `evm`, the output from the evm should be the same as the output from the compiler when using `--bin-runtime`,
ie. the return value of a contract compiled with `--bin` is the contract compiled with `--bin-runtime`.

Let's disassemble the solidity contract:

```
$ echo $(cat MyContract.bin-runtime)  | disasm 
606060405260e060020a6000350463a5f3c23b8114601a575b005b60243560043501600055601856
0      PUSH1  => 60
2      PUSH1  => 40
4      MSTORE
5      PUSH1  => e0
7      PUSH1  => 02
9      EXP
10     PUSH1  => 00
12     CALLDATALOAD
13     DIV
14     PUSH4  => a5f3c23b
19     DUP2
20     EQ
21     PUSH1  => 1a
23     JUMPI
24     JUMPDEST
25     STOP
26     JUMPDEST
27     PUSH1  => 24
29     CALLDATALOAD
30     PUSH1  => 04
32     CALLDATALOAD
33     ADD
34     PUSH1  => 00
36     SSTORE
37     PUSH1  => 18
39     JUMP
```

The addition itself happens towards the bottom. Note that just before the ADD, with the CALLDATALOADs, we are loading 32-byte arguments 
from positions 0x04 and 0x24 of the call-data, rather than 0x00 and 0x20, to make room in the first four bytes for the function identifier.
In this case, as you might guess, the function identifier for our sole function is `a5f3c23b`.
Everything before line 26 in the code is dealing with checking whether or not the first four bytes of the call-data equal `a5f3c23b` -
since CALLDATALOAD grabs a 32-byte word, and we only want four bytes, we have to byte shift by dividing by a large integer,
hence the `EXP` and `DIV`.
If the first four bytes of the call-data match `a5f3c23b`, we load the arguments, add them, and store at position 0x00.
Otherwise, we halt.

Note we can verify function identifiers by running `solc --hashes add.sol`, or in python by hashing the canonical signature:

```
>>> import sha3
>>> sha3.sha3_256("add(int256,int256)").hexdigest()[:8]
'a5f3c23b'
```

To call the function correctly, we can do `evm --debug --code $(cat Addition.bin-runtime) --input a5f3c23b00000000000000000000000000000000000000000000000000000000000000050000000000000000000000000000000000000000000000000000000000000004`

A more interesting version of the contract would have a `get` function, so we can find out the last value stored:

```
contract Addition{
	int x;
	function add(int a, int b){
		x = a + b;
	}

	function get() returns (int){
		return x;
	}

}
```

Using our persistent evm tool, lets deploy this contract, add two values, and then retrieve the stored value.
Save the contract as `add.sol`. Now, from within the container, compile the contract:

```
solc --bin --optimize -o . add.sol
```

This should create `Addition.bin`, which should be available outside the container in `$SOLC_WORKSPACE`.
From the `$SOLC_WORKSPACE` on the host, lets deploy the contract:

```
$ evm --code $(cat Addition.bin) --datadir evm-data
Loading database
Loading root hash 0000000000000000000000000000000000000000000000000000000000000000
Contract Address: 1F2A98889594024BFFDA3311CBE69728D392C06D
VM STAT 0 OPs
OUT: 0x606060405260e060020a60003504636d4ce63c81146024578063a5f3c23b146031575b005b6000546060908152602090f35b60243560043501600055602256
```

Now let's call the `add` function, and get it to store the result of `5+4`:

```
evm --datadir evm-data --to 0x1F2A98889594024BFFDA3311CBE69728D392C06D --input a5f3c23b00000000000000000000000000000000000000000000000000000000000000050000000000000000000000000000000000000000000000000000000000000004
```

Finally, let's call the `get` function. First, we get the functionid from solc:

```
$ solc --hashes add.sol
======= Addition =======
Function signatures: 
6d4ce63c: get()
a5f3c23b: add(int256,int256)
```

Now we can call the contract:

```
$ evm --datadir evm-data --to 0x1F2A98889594024BFFDA3311CBE69728D392C06D --input 6d4ce63c
Datadir already exists
Loading database
Loading root hash F3C30A7CD9769C45590C236816F2714E96198DBD7FEC33AE892E861816F548B2
Loaded account for receiver 1F2A98889594024BFFDA3311CBE69728D392C06D
CODE: 606060405260E060020A60003504636D4CE63C81146024578063A5F3C23B146031575B005B6000546060908152602090F35B60243560043501600055602256
VM STAT 0 OPs
OUT: 0x0000000000000000000000000000000000000000000000000000000000000009
```

Tada! Try running the code with `--debug` and make sure you can understand everything that is happening.


# Conclusion

That concludes this introducty guide to the Ethereum Virtual Machine.
Hopefully, you now have a much deeper understanding of how it works,
and will pass on this knowledge to others, both human and machine,
to increase the number of people who understand how Ethereum works at a low level
and to increase the number of tools for working with and analyzing Ethereum contracts.

Toodles!
