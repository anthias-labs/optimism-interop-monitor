package main

import (
	"github.com/ethereum/go-ethereum/core/types"
)

type BlockStat struct{}

// Aggregates stats for a single sender -> receiver pair
type Aggregator struct {
	Sender     Chain
	Receiver   Chain
	messenger  map[Identifier]*types.Log
	inbox      map[Identifier]*types.Log // the key in the map refers to the sender message that is being received
	BlockStats map[uint64]BlockStat
}

// Add a message from the receiver
func (a Aggregator) AddInboxMessage(msg *types.Log) (err error) {
	return
}

// Add a message from the sender
func (a Aggregator) AddMessengerMessage(msg *types.Log) (err error) {
	return
}
