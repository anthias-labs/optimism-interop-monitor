package main

import (
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// crashes in case of crucial errors
func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// This is an example of the fetching system. This will be replaced with the actual
// CLI on a later milestone
// Most of the logic can be found in the fetcher.go file
func main() {
	err := FetcherInit()
	must(err)

	// This assumes an instance of Supersim is running
	// More info: https://github.com/ethereum-optimism/Supersim
	ch1, err := NewChain("http://localhost:9545")
	must(err)

	ch2, err := NewChain("http://localhost:9546")
	must(err)

	// We set chain A as the receiver, and track its CrossL2Inbox contract
	inbox := Contract{
		ABI:     CrossL2InboxABI,
		Address: common.HexToAddress("0x4200000000000000000000000000000000000022"),
		Chain:   ch1,
	}

	// We set chain B as the sender, and track its L2ToL2CrossDomainMessenger contract
	messenger := Contract{
		ABI:     L2ToL2CrossDomainMessengerABI,
		Address: common.HexToAddress("0x4200000000000000000000000000000000000023"),
		Chain:   ch2,
	}

	errChan := make(chan error)

	// There are two available fetching methods that will be selectable through
	// config, either using the RPC native subscription or periodically polling for updates.
	// We don't have notifications available, so we use the latter here, but the code for the former
	// is in the Contract.SubscribeLogsNotification function
	inboxChan := inbox.CreateFetchChannel(big.NewInt(0), errChan)
	must(err)

	messengerChan := messenger.CreateFetchChannel(big.NewInt(0), errChan)
	must(err)

	// We read from both channels and log the events
	for {
		select {
		case inboxLog := <-inboxChan:
			name, data, err := inbox.ParseEventToDic(inboxLog)
			must(err)

			log.Printf("inbox: %s %v", name, data)
		case messengerLog := <-messengerChan:
			name, data, err := messenger.ParseEventToDic(messengerLog)
			must(err)

			log.Printf("messenger: %s %v", name, data)
		}
	}

}
