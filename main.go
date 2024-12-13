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

	inboxLogs, _ := inbox.FetchLogs(big.NewInt(0))
	messengerLogs, _ := messenger.FetchLogs(big.NewInt(0))

	log.Println("Inbox: ")
	for _, l := range inboxLogs {
		name, data, err := inbox.ParseEventToDic(l)

		if err != nil {
			log.Println(err)
			continue
		}

		log.Printf("Inbox#%s: %+v", name, data)
	}

	log.Println("Messenger: ")
	for _, l := range messengerLogs {
		name, data, err := messenger.ParseEventToDic(l)

		if err != nil {
			log.Println(err)
			continue
		}

		log.Printf("Messenger#%s: %+v", name, data)
	}
}
