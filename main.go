package main

import (
	"log"
)

// crashes in case of crucial errors
func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// This is an example of the aggregation system. This will be replaced with the actual
// CLI on a later milestone
// Most of the logic can be found in the aggregator.go and fetcher.go files
func main() {
	err := FetcherInit()
	must(err)

	// This assumes an instance of Supersim is running
	// More info: https://github.com/ethereum-optimism/Supersim
	ch1, err := NewChain("http://localhost:9545")
	must(err)

	ch2, err := NewChain("http://localhost:9546")
	must(err)

	cp := ContractPair{Sender: ch2, Receiver: ch1}

	// This function creates the aggregator, which is the first return value
	// As debug logging is currently enabled in the aggregator and we don't do anything with
	// the information, we ignore the return value for now
	_, errChan, err := cp.FetchAggregateCycle()
	must(err)

	for {
		select {
		case e := <-errChan:
			log.Printf("error: %v", e)
		}
	}
}
