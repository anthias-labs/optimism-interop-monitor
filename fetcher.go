package main

import (
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// will be exposed via config later on
var FETCH_SLEEP_TIME = 5

var l2ToL2CrossDomainMessengerMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"function\",\"name\":\"crossDomainMessageContext\",\"inputs\":[],\"outputs\":[{\"name\":\"sender_\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"source_\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"crossDomainMessageSender\",\"inputs\":[],\"outputs\":[{\"name\":\"sender_\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"crossDomainMessageSource\",\"inputs\":[],\"outputs\":[{\"name\":\"source_\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"messageNonce\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"messageVersion\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint16\",\"internalType\":\"uint16\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"relayMessage\",\"inputs\":[{\"name\":\"_id\",\"type\":\"tuple\",\"internalType\":\"structIdentifier\",\"components\":[{\"name\":\"origin\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"blockNumber\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"logIndex\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"timestamp\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"chainId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"_sentMessage\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"payable\"},{\"type\":\"function\",\"name\":\"sendMessage\",\"inputs\":[{\"name\":\"_destination\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_target\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_message\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"successfulMessages\",\"inputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"version\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"string\",\"internalType\":\"string\"}],\"stateMutability\":\"view\"},{\"type\":\"event\",\"name\":\"RelayedMessage\",\"inputs\":[{\"name\":\"source\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"messageNonce\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"messageHash\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"SentMessage\",\"inputs\":[{\"name\":\"destination\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"target\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"messageNonce\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"sender\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"},{\"name\":\"message\",\"type\":\"bytes\",\"indexed\":false,\"internalType\":\"bytes\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"EventPayloadNotSentMessage\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"IdOriginNotL2ToL2CrossDomainMessenger\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidChainId\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"MessageAlreadyRelayed\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"MessageDestinationNotRelayChain\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"MessageDestinationSameChain\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"MessageTargetCrossL2Inbox\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"MessageTargetL2ToL2CrossDomainMessenger\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NotEntered\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ReentrantCall\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"TargetCallFailed\",\"inputs\":[]}]",
}

var L2ToL2CrossDomainMessengerABI *abi.ABI

// Represents either the sender chain, or any chain in the dependency set
type Chain struct {
	RPC     string
	Client  *ethclient.Client
	ChainId *big.Int
}

// Any contract, knows how to fetch events and decode them
type Contract struct {
	ABI     *abi.ABI
	Address common.Address
	Chain   *Chain
}

type Identifier struct {
	Origin      common.Address
	BlockNumber *big.Int
	LogIndex    *big.Int
	Timestamp   *big.Int
	ChainId     *big.Int
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

	c = &Chain{
		RPC:    RPC,
		Client: client,
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
	header, err := c.Client.HeaderByNumber(context.Background(), blockNumber)

	if err != nil {
		return nil, err
	}

	return big.NewInt(int64(header.Time)), nil
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

func GetEventIdentifier(eventLog types.Log, chain *Chain) (id Identifier, err error) {
	id.BlockNumber = big.NewInt(int64(eventLog.BlockNumber))
	id.ChainId = chain.ChainId
	id.LogIndex = big.NewInt(int64(eventLog.Index))
	id.Origin = eventLog.Address
	id.Timestamp, err = chain.GetBlockTimestamp(big.NewInt(int64(eventLog.BlockNumber)))

	if err != nil {
		return Identifier{}, err
	}

	return
}
