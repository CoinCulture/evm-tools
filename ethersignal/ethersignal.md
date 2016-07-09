# EtherSignal Analysis

Following up on our introductory guide to the EVM, we now analyze a real contract.

The contract, `ethersignal.sol`, is as follows:


```
contract EtherSignal {
        event LogSignal(bytes32 indexed positionHash, bool pro, address addr);
        function setSignal(bytes32 positionHash, bool pro) {
                for (uint i = 0; i < 4096; i++) { } // burn some gas to increase DOS cost
                LogSignal(positionHash, pro, msg.sender);
                if(!msg.sender.send(this.balance)){ throw; }
    }

    function () {
        throw;
    }
}
```

After compiling with `solc --optimize --bin-runtime -o . ethersignal.sol` we get the following code:

```
$ echo $(cat EtherSignal.bin-runtime) | disasm 
60606040523615601d5760e060020a60003504637a6668bf81146023575b603f6002565b603f60043560243560005b6110008110156046575b600101602e565b005b505050565b606082815273ffffffffffffffffffffffffffffffffffffffff331660805283907fdfe72b85658ece2ea9a0485273e99806605f396deff71c0650ea0e4bb691ca8b90604090a273ffffffffffffffffffffffffffffffffffffffff33811690600090301631606082818181858883f193505050501515604157600256
0      PUSH1  => 60
2      PUSH1  => 40
4      MSTORE
5      CALLDATASIZE
6      ISZERO
7      PUSH1  => 1d
9      JUMPI
10     PUSH1  => e0
12     PUSH1  => 02
14     EXP
15     PUSH1  => 00
17     CALLDATALOAD
18     DIV
19     PUSH4  => 7a6668bf
24     DUP2
25     EQ
26     PUSH1  => 23
28     JUMPI
29     JUMPDEST
30     PUSH1  => 3f
32     PUSH1  => 02
34     JUMP
35     JUMPDEST
36     PUSH1  => 3f
38     PUSH1  => 04
40     CALLDATALOAD
41     PUSH1  => 24
43     CALLDATALOAD
44     PUSH1  => 00
46     JUMPDEST
47     PUSH2  => 1000
50     DUP2
51     LT
52     ISZERO
53     PUSH1  => 46
55     JUMPI
56     JUMPDEST
57     PUSH1  => 01
59     ADD
60     PUSH1  => 2e
62     JUMP
63     JUMPDEST
64     STOP
65     JUMPDEST
66     POP
67     POP
68     POP
69     JUMP
70     JUMPDEST
71     PUSH1  => 60
73     DUP3
74     DUP2
75     MSTORE
76     PUSH20  => ffffffffffffffffffffffffffffffffffffffff
97     CALLER
98     AND
99     PUSH1  => 80
101    MSTORE
102    DUP4
103    SWAP1
104    PUSH32  => dfe72b85658ece2ea9a0485273e99806605f396deff71c0650ea0e4bb691ca8b
137    SWAP1
138    PUSH1  => 40
140    SWAP1
141    LOG2
142    PUSH20  => ffffffffffffffffffffffffffffffffffffffff
163    CALLER
164    DUP2
165    AND
166    SWAP1
167    PUSH1  => 00
169    SWAP1
170    ADDRESS
171    AND
172    BALANCE
173    PUSH1  => 60
175    DUP3
176    DUP2
177    DUP2
178    DUP2
179    DUP6
180    DUP9
181    DUP4
182    CALL
183    SWAP4
184    POP
185    POP
186    POP
187    POP
188    ISZERO
189    ISZERO
190    PUSH1  => 41
192    JUMPI
193    PUSH1  => 02
195    JUMP
```

Let us go through it in chunks. First we have

```
0      PUSH1  => 60
2      PUSH1  => 40
4      MSTORE
```

which seems to be pointless, as the memory is later overwritten,
though the program counter location 0x02 is used later as an invalid jumpdest which serves as a generic yet
identifiable means for solidity to throw exceptions and revert all state transitions.
Next we load the size of the call-data and check if it is empty. 
If so, the caller has not correctly specified the function to call, and so the execution should throw.

```
5      CALLDATASIZE
6      ISZERO
7      PUSH1  => 1d
9      JUMPI
```

We see here that, if ISZERO returns a 0x01, we jump to 0x1d (29):

```
29     JUMPDEST
30     PUSH1  => 3f
32     PUSH1  => 02
34     JUMP
```

Program count 29 is a valid jump destination, but we then jump to pc 2, which is an invalid jump destination, 
and causes execution to halt.

We can confirm this behaviour with an execution:

```
$ evm --code $(cat EtherSignal.bin-runtime) --debug
VM STAT 12 OPs
PC 00000000: PUSH1 GAS: 9999999997 COST: 3
STACK = 0
MEM = 0
STORAGE = 0

PC 00000002: PUSH1 GAS: 9999999994 COST: 3
STACK = 1
0000: 0000000000000000000000000000000000000000000000000000000000000060
MEM = 0
STORAGE = 0

PC 00000004: MSTORE GAS: 9999999982 COST: 12
STACK = 2
0000: 0000000000000000000000000000000000000000000000000000000000000040
0001: 0000000000000000000000000000000000000000000000000000000000000060
MEM = 96
0000: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
0016: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
0032: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
0048: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
0064: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
0080: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
STORAGE = 0

PC 00000005: CALLDATASIZE GAS: 9999999980 COST: 2
STACK = 0
MEM = 96
0000: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
0016: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
0032: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
0048: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
0064: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
0080: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 60  ...............`
STORAGE = 0

PC 00000006: ISZERO GAS: 9999999977 COST: 3
STACK = 1
0000: 0000000000000000000000000000000000000000000000000000000000000000
MEM = 96
0000: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
0016: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
0032: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
0048: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
0064: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
0080: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 60  ...............`
STORAGE = 0

PC 00000007: PUSH1 GAS: 9999999974 COST: 3
STACK = 1
0000: 0000000000000000000000000000000000000000000000000000000000000001
MEM = 96
0000: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
0016: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
0032: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
0048: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
0064: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
0080: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 60  ...............`
STORAGE = 0

PC 00000009: JUMPI GAS: 9999999964 COST: 10
STACK = 2
0000: 000000000000000000000000000000000000000000000000000000000000001d
0001: 0000000000000000000000000000000000000000000000000000000000000001
MEM = 96
0000: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
0016: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
0032: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
0048: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
0064: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
0080: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 60  ...............`
STORAGE = 0

PC 00000029: JUMPDEST GAS: 9999999963 COST: 1
STACK = 0
MEM = 96
0000: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
0016: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
0032: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
0048: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
0064: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
0080: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 60  ...............`
STORAGE = 0

PC 00000030: PUSH1 GAS: 9999999960 COST: 3
STACK = 0
MEM = 96
0000: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
0016: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
0032: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
0048: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
0064: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
0080: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 60  ...............`
STORAGE = 0

PC 00000032: PUSH1 GAS: 9999999957 COST: 3
STACK = 1
0000: 000000000000000000000000000000000000000000000000000000000000003f
MEM = 96
0000: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
0016: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
0032: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
0048: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
0064: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
0080: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 60  ...............`
STORAGE = 0

PC 00000034: JUMP GAS: 9999999949 COST: 8
STACK = 2
0000: 0000000000000000000000000000000000000000000000000000000000000002
0001: 000000000000000000000000000000000000000000000000000000000000003f
MEM = 96
0000: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
0016: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
0032: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
0048: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
0064: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
0080: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 60  ...............`
STORAGE = 0

PC 00000034: JUMP GAS: 9999999949 COST: 8 ERROR: invalid jump destination (PUSH1) 2
STACK = 1
0000: 000000000000000000000000000000000000000000000000000000000000003f
MEM = 96
0000: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
0016: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
0032: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
0048: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
0064: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
0080: 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 60  ...............`
STORAGE = 0

OUT: 0x error: invalid jump destination (PUSH1) 2
```



If the call-data size is not zero, we load the first four bytes and check if they match the correct function identifier 

```
10     PUSH1  => e0
12     PUSH1  => 02
14     EXP
15     PUSH1  => 00
17     CALLDATALOAD
18     DIV
19     PUSH4  => 7a6668bf
24     DUP2
25     EQ
26     PUSH1  => 23
28     JUMPI
```

Here, we divide the 32-bytes of the call-data from position 0x0 by `2^224` to get the four-byte function identifier
and compare it with the function identifier for the `setSignal` function, `7a6668bf`.
We can confirm what the value should be in python:

```
>>> import sha3
>>> sha3.sha3_256("setSignal(bytes32,bool)").hexdigest()[:8]
'7a6668bf'
````

If the function identifier matches, we jump to pc 0x23, otherwise we continue to pc 0x1d, 
which we have already seen results in an invalid jump destination exception.

So, if the call-data is empty, or if the first four bytes are not equal to the function identifier,
the execution throws an exception and reverts all state. 
We can similarly confirm this with an execution by passing any arbitrary call-data that does not begin with `7a6668bf`.

If the first four bytes of the call-data are `7a6668bf`, then the function is being correctly called
and the `EQ` returns `0x01`, so we jump to pc 0x23 (35):

```
35     JUMPDEST
36     PUSH1  => 3f
38     PUSH1  => 04
40     CALLDATALOAD
41     PUSH1  => 24
43     CALLDATALOAD
44     PUSH1  => 00
```

It is a valid jump destination. We then load the arguments from the call-data (positions 0x04 and 0x24)
so they are available on the stack for later, and push a 0x00, which is the initial value of the counter for the loop.
Note we have also pushed a 0x3f to the stack - this will be used in a jump at the end, 
but pretty much just sits at the bottom of the stack until then.
Next up is the JUMPDEST which marks the top of the loop:

```
46     JUMPDEST
47     PUSH2  => 1000
50     DUP2
51     LT
52     ISZERO
53     PUSH1  => 46
55     JUMPI
56     JUMPDEST
57     PUSH1  => 01
59     ADD
60     PUSH1  => 2e
62     JUMP
63     JUMPDEST
64     STOP
65     JUMPDEST
66     POP
67     POP
68     POP
69     JUMP
70     JUMPDEST
```

The loop is expected to run 4096 (0x1000) times, so we push that to the stack,
and duplicate the second item on the stack (the counter), so we can check if it is less than 0x1000.
If not, the result of LT will be 0x0, so ISZERO will push 0x01, and we will jump to pc 0x46 (70), 
thus exiting the loop. Otherwise, we add 0x01 to the counter on the stack and jump to pc 0x2e (46), 
ie. back to the top of the loop.

Note that every time through, when we are at the JUMPDEST at the top of the loop, 
the stack should look like:

```
PC 00000046: JUMPDEST GAS: 9999999885 COST: 1
STACK = 5
0000: 0000000000000000000000000000000000000000000000000000000000000000
0001: 0000000000000000000000000000000000000000000000000000000000000000
0002: 0000000000000000000000000000000000000000000000000000000000000000
0003: 000000000000000000000000000000000000000000000000000000000000003f
0004: 000000000000000000000000000000000000000000000000000000007a6668bf
```

where the top element is the counter (increments each pass through the loop), 
and the next two elements are the two arguments passed in the call-data (`bool pro` and `bytes32 positionHash`, respectively)

The loop progresses in this way, doing nothing of interest but burning gas, until the counter value reaches 0x1000, 
at which point we jump to pc 0x46 (70),
which is a valid jump destination, and continue by logging an event as follows:

```
71     PUSH1  => 60
73     DUP3
74     DUP2
75     MSTORE
76     PUSH20  => ffffffffffffffffffffffffffffffffffffffff
97     CALLER
98     AND
99     PUSH1  => 80
101    MSTORE
102    DUP4
103    SWAP1
104    PUSH32  => dfe72b85658ece2ea9a0485273e99806605f396deff71c0650ea0e4bb691ca8b
137    SWAP1
138    PUSH1  => 40
140    SWAP1
141    LOG2
```

Note when the loop ends, we should still have the counter (now at 0x1000) and the two call-data arguments on the stack.
So here, we push 0x60, copy the second call-data argument to the top, copy the 0x60 to the top of the stack,
and store the call-data argument in memory at position 0x60.
Now we push 20-bytes of `0xff` and the caller address, and perform boolean AND, 
ensuring the resulting item on the stack has at most 20 non-zero bytes - 
though the result of CALLER should always be 20-bytes anyways, so not sure this is necessary.
We now store the resulting address at position 0x80 in memory.
A few more stack operations get everything prepared for the log operation. 
For instance, right before the LOG the stack looks something like:

```
0000: 0000000000000000000000000000000000000000000000000000000000000060 - position in memory
0001: 0000000000000000000000000000000000000000000000000000000000000040 - size in memory
0002: dfe72b85658ece2ea9a0485273e99806605f396deff71c0650ea0e4bb691ca8b - topic 1 (hard coded by solidity)
0003: 0000000000000000000000000000000000000000000000000000000000000000 - topic 2 (`bytes32 positionHash` argument)
0004: 0000000000000000000000000000000000000000000000000000000000001000
0005: 0000000000000000000000000000000000000000000000000000000000000000
0006: 0000000000000000000000000000000000000000000000000000000000000000
0007: 000000000000000000000000000000000000000000000000000000000000003f
0008: 000000000000000000000000000000000000000000000000000000007a6668bf
```
Here we are using `LOG2`, which takes a starting location and size for the byte-array to log from memory, 
as well as two additional items from the stack which serve as topics.
In this case, starting from position 0x60 in memory we grab 0x40 (64) bytes, 
which includes the `bool pro` argument and the caller address we just stored in memory.
In addition we pass two topics, one being the `bytes32 positionHash` argument, 
the other being a hardcoded hash provided by solidity.

Next up, we send any balance the contract has to the caller with a CALL:

```
142    PUSH20  => ffffffffffffffffffffffffffffffffffffffff
163    CALLER
164    DUP2
165    AND
166    SWAP1
167    PUSH1  => 00
169    SWAP1
170    ADDRESS
171    AND
172    BALANCE
173    PUSH1  => 60
175    DUP3
176    DUP2
177    DUP2
178    DUP2
179    DUP6
180    DUP9
181    DUP4
182    CALL
```

This is mostly just a mess of stack manipulation to arrange the arguments for the call.
The stack just before the call looks something like:

```
0000: 0000000000000000000000000000000000000000000000000000000000000000
0001: 000000000000000000000000000000000000000000000000000073656e646572
0002: 0000000000000000000000000000000000000000000000000000000000000000
0003: 0000000000000000000000000000000000000000000000000000000000000060
0004: 0000000000000000000000000000000000000000000000000000000000000000
0005: 0000000000000000000000000000000000000000000000000000000000000060
0006: 0000000000000000000000000000000000000000000000000000000000000000
0007: 0000000000000000000000000000000000000000000000000000000000000060
0008: 0000000000000000000000000000000000000000000000000000000000000000
0009: 0000000000000000000000000000000000000000000000000000000000000000
0010: 000000000000000000000000000000000000000000000000000073656e646572
0011: 0000000000000000000000000000000000000000000000000000000000001000
0012: 0000000000000000000000000000000000000000000000000000000000000000
0013: 0000000000000000000000000000000000000000000000000000000000000000
0014: 000000000000000000000000000000000000000000000000000000000000003f
0015: 000000000000000000000000000000000000000000000000000000007a6668bf
```

A call takes arguments of the stack in the following order:
gas, address, value, inputOffset, inputSize, returnOffset, returnSize;
where offset and size parameters refer to the memory array.
In this case, we pay 0x0 gas to send to the caller (address 000000000000000000000000000073656e646572) 0x0 wei.
The input is copied from position 0x60, length 0x0, and the return value is stored at position 0x60, length 0x0.
In essence, this is a dummy call.

If the call succeeds, it pushes 0x01 to the stack and writes the return to memory.
Otherwise, it pushes 0x0 to the stack.
Successful completion of a call would give us a stack like:

```
0000: 0000000000000000000000000000000000000000000000000000000000000001
0001: 0000000000000000000000000000000000000000000000000000000000000060
0002: 0000000000000000000000000000000000000000000000000000000000000000
0003: 0000000000000000000000000000000000000000000000000000000000000000
0004: 000000000000000000000000000000000000000000000000000073656e646572
0005: 0000000000000000000000000000000000000000000000000000000000001000
0006: 0000000000000000000000000000000000000000000000000000000000000000
0007: 0000000000000000000000000000000000000000000000000000000000000000
0008: 000000000000000000000000000000000000000000000000000000000000003f
0009: 000000000000000000000000000000000000000000000000000000007a6668bf
```

Before checking the return value, we remove three unneeded elements from the stack
by swaping out the return value below them and popping them.
Then we run ISZERO twice, which returns a 0x01 if the call returned a 0x1, and a 0x0 otherwise
- hence it is effectively redundant.

```
183    SWAP4
184    POP
185    POP
186    POP
187    POP
188    ISZERO
189    ISZERO
190    PUSH1  => 41
192    JUMPI
193    PUSH1  => 02
195    JUMP
```

If the call returned a 0x1, we jump to 0x41 (65). Otherwise, we throw an exception by jumping to the invalid destination at pc 0x2.

At pc 65, a valid jump destination, we pop a few more un-needed items from the stack:

```
65     JUMPDEST
66     POP
67     POP
68     POP
```

leaving us with:

```
0001: 000000000000000000000000000000000000000000000000000000000000003f
0002: 000000000000000000000000000000000000000000000000000000007a6668bf
```

This 0x3f (63) was pushed long ago, on pc 36, when we mentioned it would be used by a jump.
Here is that jump:

```
69     JUMP
```

It takes us to pc 63:

```
63     JUMPDEST
64     STOP
```

And that is that! Contract execution has completed successfully.
