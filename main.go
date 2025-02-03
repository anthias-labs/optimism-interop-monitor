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

func main() {
	configFile := flag.String("config", "config.json", "path to config file")
	flag.Parse()

	data, err := os.ReadFile(*configFile)
	must(err)

	config, err := parseConfig(data)
	must(err)

	senderChain, err := NewChain(config.SenderChain)
	must(err)

	receiverChain, err := NewChain(config.ReceiverChain)
	must(err)

	cp := ContractPair{Sender: senderChain, Receiver: receiverChain}

	err = FetcherInit(config)
	must(err)

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
