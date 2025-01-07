package main

import (
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
)

type BlockStat struct {
	MessageCount     uint64
	TotalLatency     *big.Int
	SentMesssages    uint64
	ReceivedMessages uint64
}

// Aggregates stats for a single sender -> receiver pair
type Aggregator struct {
	ContractPair
	messenger         map[Identifier]*types.Log
	inbox             map[Identifier]*types.Log // the key in the map refers to the sender message that is being received
	messengerContract Contract
	inboxContract     Contract
	BlockStats        map[uint64]BlockStat // with respect to sender blocknum
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
func (agg *Aggregator) AddInboxMessage(msg *types.Log) (err error) {
	name, data, err := agg.inboxContract.ParseEventToDic(*msg)
	if err != nil {
		return err
	}

	senderId, err := coerceToIdentifier(data["id"])
	if err != nil {
		return err
	}

	// check if message is in messenger outbox
	messageLog, ok := agg.messenger[senderId]

	if ok {
		agg.AddMessagePair(messageLog, msg)
		delete(agg.messenger, senderId)
	} else {
		agg.inbox[senderId] = msg
	}

	bs := agg.GetBlockStats(senderId.BlockNumber)

	bs.MessageCount += 1
	bs.ReceivedMessages += 1

	agg.BlockStats[senderId.BlockNumber] = *bs

	log.Printf("inbox: %s %v", name, data)
	return
}

// Add a message from the sender
func (agg *Aggregator) AddMessengerMessage(msg *types.Log) (err error) {
	name, data, err := agg.messengerContract.ParseEventToDic(*msg)
	if err != nil {
		return err
	}

	id, err := agg.Sender.GetEventIdentifier(*msg)
	if err != nil {
		return err
	}

	// check if message is in receiver inbox
	messageLog, ok := agg.inbox[id]

	if ok {
		agg.AddMessagePair(msg, messageLog)
		delete(agg.inbox, id)
	} else {
		agg.messenger[id] = msg
	}

	bs := agg.GetBlockStats(msg.BlockNumber)

	bs.MessageCount += 1
	bs.SentMesssages += 1

	agg.BlockStats[msg.BlockNumber] = *bs

	log.Printf("messenger: %s %v %v", name, data, id)
	log.Printf("map currently: %v", agg.inbox[id])
	return
}

func (agg Aggregator) GetBlockStats(blockNumber uint64) (bs *BlockStat) {
	bs_v, ok := agg.BlockStats[blockNumber]

	if !ok {
		bs_v = BlockStat{
			TotalLatency: big.NewInt(0),
		}
		agg.BlockStats[blockNumber] = bs_v
	}

	return &bs_v
}

func (agg Aggregator) AddMessagePair(senderMsg, receiverMsg *types.Log) (err error) {
	bs := agg.GetBlockStats(senderMsg.BlockNumber)

	// We don't want to rely on the reported identifier time, so just in case we fetch the timestamp
	// We keep a cache so we only ever fetch once per block
	senderTimestamp, err := agg.Sender.GetBlockTimestamp(big.NewInt(int64(senderMsg.BlockNumber)))
	if err != nil {
		return
	}

	receiverTimestamp, err := agg.Receiver.GetBlockTimestamp(big.NewInt(int64(receiverMsg.BlockNumber)))
	if err != nil {
		return
	}

	latency := big.NewInt(0)
	latency.Sub(receiverTimestamp, senderTimestamp)
	bs.TotalLatency.Add(bs.TotalLatency, latency)

	log.Printf("addMessagePair: found pair, stats: %v", agg.BlockStats)
	return
}
