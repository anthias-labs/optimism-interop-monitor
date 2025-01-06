package main

import (
	"log"

	"github.com/ethereum/go-ethereum/core/types"
)

type BlockStat struct{}

// Aggregates stats for a single sender -> receiver pair
type Aggregator struct {
	ContractPair
	messenger         map[Identifier]*types.Log
	inbox             map[Identifier]*types.Log // the key in the map refers to the sender message that is being received
	messengerContract Contract
	inboxContract     Contract
	BlockStats        map[uint64]BlockStat
}

func MakeAggregator(sender, receiver *Chain) (agg Aggregator) {
	agg.messenger = make(map[Identifier]*types.Log)
	agg.inbox = make(map[Identifier]*types.Log)
	agg.BlockStats = make(map[uint64]BlockStat)

	agg.Sender = sender
	agg.Receiver = receiver

	agg.inboxContract, agg.messengerContract = agg.GetContracts()

	return
}

// Add a message from the receiver
func (a Aggregator) AddInboxMessage(msg *types.Log) (err error) {
	name, data, err := a.inboxContract.ParseEventToDic(*msg)

	log.Printf("inbox: %s %v", name, data)
	return
}

// Add a message from the sender
func (a Aggregator) AddMessengerMessage(msg *types.Log) (err error) {
	name, data, err := a.messengerContract.ParseEventToDic(*msg)

	log.Printf("messenger: %s %v", name, data)
	return
}
