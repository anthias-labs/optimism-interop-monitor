# optimism-interop-monitoring
Interop monitoring for the Optimism Superchain


## Milestone 1

We set up the tracking system to collect cross-chain messages (both initializing and finalizing) from a series of L2 chains.

This milestone mostly focuses on two important parts of the system:

1. We set up the actual fetching logic. This is the heart of the program, in which all other parts rely. The only direct dependency we use for this is the official geth library for communicating with the RPCs and parsing the events according to the contract ABI. We use a different Goroutine (lightweight managed green thread) for every tracked contract, which allows better scaling with little overhead, and prevents one RPC from lagging the performance of other networks
2. We set up the necessary types and API for the aggregation milestone. By using a good architecture for storing the events in memory and setting up relevant functions, we create the necessary abstractions for building the aggregation pipeline rapidly and reliably.

An example of how this works can be found in `main.go`. To run this:
1. Make sure [Supersim](https://github.com/ethereum-optimism/Supersim) is running
2. Run `go get` to install dependencies, and start the system with `go run .`
3. Send an interop transaction from chain B to chain A. More info on this step can be found in the [Supersim repo](https://github.com/ethereum-optimism/Supersim)