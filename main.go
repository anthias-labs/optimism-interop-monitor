package main

import (
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// crashes in case of error
func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	err := FetcherInit()
	must(err)

	// Test:

	ch1, err := NewChain("http://localhost:9545")
	must(err)

	ch2, err := NewChain("http://localhost:9546")
	must(err)

	inbox := Contract{
		ABI:     CrossL2InboxABI,
		Address: common.HexToAddress("0x4200000000000000000000000000000000000022"),
		Chain:   ch1,
	}

	messenger := Contract{
		ABI:     L2ToL2CrossDomainMessengerABI,
		Address: common.HexToAddress("0x4200000000000000000000000000000000000023"),
		Chain:   ch2,
	}

	errChan := make(chan error)

	inboxChan := inbox.CreateFetchChannel(big.NewInt(0), errChan)
	must(err)

	messengerChan := messenger.CreateFetchChannel(big.NewInt(0), errChan)
	must(err)

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
