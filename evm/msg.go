package main

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
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
