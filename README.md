# optimism-interop-monitoring
Interop monitoring for the Optimism Superchain


## Milestone 2

In this milestone, we set up the aggregation pipeline to collect statistics from the interop performance for each pair of sender-receiver chains. 

The main focus of this milestone was:
1. We set up the aggregation logic for abstracting a series of send and receive interop transactions into messages, and messages into actual statistics. For this, the `Aggregator` object takes a pair of contracts (namely, `L2ToL2CrossDomainMessenger` from the sender chain and `CrossL2Inbox` from the receiver), and provides an API for adding incoming events from them.
2. The `Agregator` then finds the corresponding send-receive pairs and computes the latency between them, storing the information on a per-block basis. It also keeps track of each type of message to measure potential issues with message loss or adversarial false message received events.
   
An example of how this works can be found in `main.go`. To run this:
1. Make sure [Supersim](https://github.com/ethereum-optimism/Supersim) is running
2. Run `go get` to install dependencies, and start the system with `go run .`
3. Send some interop transactions from chain B to chain A. More info on this step can be found in the [Supersim repo](https://github.com/ethereum-optimism/Supersim)