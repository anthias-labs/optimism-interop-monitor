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

type DetailedIntervalStat struct {
	MessageCount     uint64   `json:"messageCount"`
	TotalLatency     *big.Int `json:"totalLatency"`
	AvgLatency       float64  `json:"avgLatency"`
	SentMesssages    uint64   `json:"sentMessages"`
	ReceivedMessages uint64   `json:"receivedMessages"`
	MissingRelay     uint64   `json:"missingRelay"`
	MissingReception uint64   `json:"missingReception"`
}

// Aggregates stats for a single sender -> receiver pair
type Aggregator struct {
	ContractPair
	config            *Config
	messenger         map[Identifier]*types.Log
	inbox             map[Identifier]*types.Log // the key in the map refers to the sender message that is being received
	messengerContract Contract
	inboxContract     Contract
	BlockStats        map[uint64]BlockStat // with respect to sender blocknum
	LatestBlock       *uint64
}

func MakeAggregator(sender, receiver *Chain, config *Config) (agg Aggregator) {
	var LatestBlock uint64
	agg.messenger = make(map[Identifier]*types.Log)
	agg.inbox = make(map[Identifier]*types.Log)
	agg.BlockStats = make(map[uint64]BlockStat)

	agg.Sender = sender
	agg.Receiver = receiver

	agg.inboxContract, agg.messengerContract = agg.GetContracts()

	agg.config = config
	agg.LatestBlock = &LatestBlock

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
	bs := agg.GetBlockStats(senderId.BlockNumber)

	bs.ReceivedMessages += 1

	agg.BlockStats[senderId.BlockNumber] = *bs

	if senderId.BlockNumber > *agg.LatestBlock {
		*agg.LatestBlock = senderId.BlockNumber
	}

	if ok {
		agg.AddMessagePair(messageLog, msg)
		delete(agg.messenger, senderId)
	} else {
		agg.inbox[senderId] = msg
	}

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
	bs := agg.GetBlockStats(msg.BlockNumber)

	bs.SentMesssages += 1

	agg.BlockStats[msg.BlockNumber] = *bs

	if msg.BlockNumber > *agg.LatestBlock {
		*agg.LatestBlock = msg.BlockNumber
	}

	if ok {
		agg.AddMessagePair(msg, messageLog)
		delete(agg.inbox, id)
	} else {
		agg.messenger[id] = msg
	}

	log.Printf("messenger: %s %v", name, data)
	return
}

func (agg *Aggregator) GetBlockStats(blockNumber uint64) (bs *BlockStat) {
	bs_v, ok := agg.BlockStats[blockNumber]

	if !ok {
		bs_v = BlockStat{
			TotalLatency: big.NewInt(0),
		}
		agg.BlockStats[blockNumber] = bs_v
	}

	return &bs_v
}

func (agg *Aggregator) AddMessagePair(senderMsg, receiverMsg *types.Log) (err error) {
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
	bs.MessageCount += 1
	agg.BlockStats[senderMsg.BlockNumber] = *bs

	log.Printf("addMessagePair: found pair, timestamps %d %d", senderTimestamp.Uint64(), receiverTimestamp.Uint64())
	return
}

func (agg *Aggregator) AggregateLatestBlocks(blockAmount uint64) (ds DetailedIntervalStat) {
	ds = DetailedIntervalStat{
		TotalLatency:     big.NewInt(0),
		SentMesssages:    0,
		ReceivedMessages: 0,
		MessageCount:     0,
		MissingRelay:     0,
		MissingReception: 0,
	}

	if agg.config.PurgeOldMessages {
		for key := range agg.inbox {
			if key.BlockNumber <= *agg.LatestBlock-2*agg.config.AggregateBlockAmount {
				delete(agg.inbox, key)
			}
		}
		for key := range agg.messenger {
			if key.BlockNumber <= *agg.LatestBlock-2*agg.config.AggregateBlockAmount {
				delete(agg.messenger, key)
			}
		}

	}

	for key, val := range agg.BlockStats {
		if agg.config.PurgeOldBlocks && key <= *agg.LatestBlock-2*agg.config.AggregateBlockAmount {
			delete(agg.BlockStats, key)
		}

		if int64(key) >= int64(*agg.LatestBlock)-int64(blockAmount) {
			ds.MessageCount += val.MessageCount
			ds.TotalLatency.Add(ds.TotalLatency, val.TotalLatency)
			ds.MissingReception += val.SentMesssages - val.MessageCount
			ds.MissingRelay += val.ReceivedMessages - val.MessageCount
			ds.ReceivedMessages += val.ReceivedMessages
			ds.SentMesssages += val.SentMesssages
		}
	}

	if ds.MessageCount == 0 {
		ds.AvgLatency = 0
	} else {
		ds.AvgLatency = float64(ds.TotalLatency.Uint64()) / float64(ds.MessageCount)
	}

	return
}
