package main

import (
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// will be exposed via config later on
var FETCH_SLEEP_TIME = 1

// Represents either the sender chain, or any chain in the dependency set
type Chain struct {
	RPC            string
	Client         *ethclient.Client
	ChainId        *big.Int
	timestampCache map[uint64]*big.Int
}

// Any contract, knows how to fetch events and decode them
type Contract struct {
	ABI     *abi.ABI
	Address common.Address
	Chain   *Chain
}

type Identifier struct {
	Origin      common.Address
	BlockNumber uint64
	LogIndex    uint64
	Timestamp   uint64
	ChainId     uint64
}

type ContractPair struct {
	Sender   *Chain
	Receiver *Chain
}

func FetcherInit() (err error) {
	CrossL2InboxABI, err = crossL2InboxMetaData.GetAbi()

	if err != nil {
		return
	}

	L2ToL2CrossDomainMessengerABI, err = l2ToL2CrossDomainMessengerMetaData.GetAbi()

	if err != nil {
		return
	}

	return nil
}

func NewChain(RPC string) (c *Chain, err error) {
	client, err := ethclient.Dial(RPC)
	if err != nil {
		return nil, err
	}

	chainId, err := client.ChainID(context.Background())
	if err != nil {
		return nil, err
	}

	c = &Chain{
		RPC:            RPC,
		Client:         client,
		ChainId:        chainId,
		timestampCache: make(map[uint64]*big.Int),
	}

	return
}

func (c *Chain) FetchLogs(address common.Address, from *big.Int) (logs []types.Log, err error) {
	logs, err = c.Client.FilterLogs(context.Background(), ethereum.FilterQuery{
		FromBlock: from,
		Addresses: []common.Address{address},
	})

	return
}

func (c *Chain) SubscribeLogsNotification(address common.Address, from *big.Int, logsChan chan<- types.Log, errChan chan<- error) (ethereum.Subscription, error) {
	// Create a filter query for the subscription
	query := ethereum.FilterQuery{
		FromBlock: from,
		Addresses: []common.Address{address},
	}

	// Subscribe to the logs
	subscription, err := c.Client.SubscribeFilterLogs(context.Background(), query, logsChan)
	if err != nil {
		return nil, err
	}

	// Start a goroutine to handle errors
	go func() {
		for {
			select {
			case err := <-subscription.Err():
				// Send subscription errors to the error channel
				errChan <- err
				return
			}
		}
	}()

	return subscription, nil
}

func (c *Chain) GetBlockTimestamp(blockNumber *big.Int) (timestamp *big.Int, err error) {
	time, ok := c.timestampCache[blockNumber.Uint64()]

	if ok {
		return time, nil
	}

	header, err := c.Client.HeaderByNumber(context.Background(), blockNumber)

	if err != nil {
		return nil, err
	}

	c.timestampCache[blockNumber.Uint64()] = big.NewInt(int64(header.Time))

	return big.NewInt(int64(header.Time)), nil
}

func (c *Chain) GetCurrentBlockNumber() (blockNum *big.Int, err error) {
	b, err := c.Client.BlockNumber(context.Background())

	return big.NewInt(int64(b)), err
}

func (c Contract) FetchLogs(from *big.Int) (logs []types.Log, err error) {
	logs, err = c.Chain.FetchLogs(c.Address, from)
	return
}

func (c Contract) SubscribeLogsNotification(from *big.Int, logsChan chan<- types.Log, errChan chan<- error) (ethereum.Subscription, error) {
	return c.Chain.SubscribeLogsNotification(c.Address, from, logsChan, errChan)
}

func (c Contract) CreateFetchChannel(from *big.Int, errChan chan error) (logsChan chan types.Log) {
	logsChan = make(chan types.Log)
	lastFetch := from.Uint64()

	go func() {
		for {
			logs, err := c.FetchLogs(big.NewInt(int64(lastFetch)))

			if err != nil {
				errChan <- err
			} else {
				for _, l := range logs {
					lastFetch = max(lastFetch, l.BlockNumber+1)
					logsChan <- l
				}
			}

			time.Sleep(time.Second * time.Duration(FETCH_SLEEP_TIME))
		}
	}()

	return
}

func (c Contract) ParseEventToDic(eventLog types.Log) (eventName string, logData map[string]interface{}, err error) {
	event, err := c.ABI.EventByID(eventLog.Topics[0])

	if err != nil {
		return "", nil, err
	}

	logData = make(map[string]interface{})

	err = event.Inputs.UnpackIntoMap(logData, eventLog.Data)
	if err != nil {
		return "", nil, err
	}

	return event.Name, logData, nil
}

func (chain *Chain) GetEventIdentifier(eventLog types.Log) (id Identifier, err error) {
	id.BlockNumber = eventLog.BlockNumber
	id.ChainId = chain.ChainId.Uint64()
	id.LogIndex = uint64(eventLog.Index)
	id.Origin = eventLog.Address
	ts, err := chain.GetBlockTimestamp(big.NewInt(int64(eventLog.BlockNumber)))

	if err != nil {
		return Identifier{}, err
	}

	id.Timestamp = ts.Uint64()

	return
}

func (cp ContractPair) GetContracts() (inbox Contract, messenger Contract) {
	inbox = Contract{
		ABI:     CrossL2InboxABI,
		Address: common.HexToAddress("0x4200000000000000000000000000000000000022"),
		Chain:   cp.Receiver,
	}

	messenger = Contract{
		ABI:     L2ToL2CrossDomainMessengerABI,
		Address: common.HexToAddress("0x4200000000000000000000000000000000000023"),
		Chain:   cp.Sender,
	}

	return
}

func (cp ContractPair) FetchAggregateCycle() (agg Aggregator, errChan chan error, err error) {
	agg = MakeAggregator(cp.Sender, cp.Receiver)

	errChan = make(chan error)

	inbox, messenger := agg.inboxContract, agg.messengerContract

	inboxCurrentBlock, err := cp.Receiver.GetCurrentBlockNumber()

	if err != nil {
		return agg, nil, err
	}

	messengerCurrentBlock, err := cp.Sender.GetCurrentBlockNumber()

	if err != nil {
		return agg, nil, err
	}

	inboxChan := inbox.CreateFetchChannel(inboxCurrentBlock, errChan)
	if err != nil {
		return agg, nil, err
	}

	messengerChan := messenger.CreateFetchChannel(messengerCurrentBlock, errChan)
	if err != nil {
		return agg, nil, err
	}

	// We read from both channels and log the events
	go func() {
		for {
			select {
			case inboxLog := <-inboxChan:
				err := agg.AddInboxMessage(&inboxLog)
				if err != nil {
					errChan <- err
				}
			case messengerLog := <-messengerChan:
				err := agg.AddMessengerMessage(&messengerLog)
				if err != nil {
					errChan <- err
				}

			}
		}

	}()

	return
}
