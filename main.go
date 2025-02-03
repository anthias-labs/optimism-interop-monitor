package main

import (
	"flag"
	"log"
	"os"
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
	configFile := flag.String("config", "config.json", "path to config file")
	flag.Parse()

	data, err := os.ReadFile(*configFile)
	must(err)

	config, err := parseConfig(data)
	must(err)

	// This assumes an instance of Supersim is running
	// More info: https://github.com/ethereum-optimism/Supersim
	senderChain, err := NewChain(config.SenderChain)
	must(err)

	receiverChain, err := NewChain(config.ReceiverChain)
	must(err)

	cp := ContractPair{Sender: senderChain, Receiver: receiverChain}

	err = FetcherInit(config)
	must(err)

	// This function creates the aggregator, which is the first return value
	// As debug logging is currently enabled in the aggregator and we don't do anything with
	// the information, we ignore the return value for now
	agg, errChan, err := cp.FetchAggregateCycle(config)
	must(err)

	go StartApi(config, &agg)

	for {
		select {
		case e := <-errChan:
			log.Printf("error: %v", e)
		}
	}
}
