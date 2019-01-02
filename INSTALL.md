# Install

## Go

- [Install Go](https://golang.org/doc/install)
- [Set the `$GOPATH` variable](https://github.com/golang/go/wiki/SettingGOPATH)
- Add `$GOPATH/bin` to your `$PATH`
- [Install dep](https://github.com/golang/dep#installation)


## evm-tools

Download this repo and install the tools:

```
mkdir -p $GOPATH/src/github.com/CoinCulture
git clone https://github.com/CoinCulture/evm-tools $GOPATH/src/github.com/CoinCulture/evm-tools
cd $GOPATH/src/github.com/CoinCulture/evm-tools
dep ensure
make
```

This will install the `evm`, `disasm`, and `evm-deploy` tools to your
`$GOPATH/bin`.

## Python

We also use some python, and in particular the `sha3` and `rlp` packages.
It's recommended to use `virtualenv` for installing them, with specific versions:

```
virtualenv evm-tools
source evm-tools/bin/activate
pip install pysha3==0.3 rlp==0.6.0
```

Note there are two incompatible versions of `sha3`, 
and Ethereum uses the `Keccak` version, so make sure you have the right one. 
You can verify with:

```
$ python -c "import sha3; print sha3.sha3_256('').hexdigest()"
c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470
```

More recent versions of the `sha3` package have a different API.
