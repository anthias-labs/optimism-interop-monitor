# Optimism Interop Monitoring

Interop monitoring for the Optimism Superchain


## About

This tool allows Superchain developers to track and monitor the state of interoperability between a sender-receiver pair. Given RPCs for both, it tracks message passing between them, aggregates different metrics, and provides an alert system for pathological states (e.g. increased message relaying latency). It also provides an API for querying statistics about interop. For more information on usage, check [usage](#usage).

## Building

Prebuilt binaries for all platforms will soon be available in releases.

Building requires `go` to be installed on the system.

To build, 
1. Clone the project.
2. Open a terminal on the root directory of the project and run `go get` to install dependencies
3. Run the project with `go run .`, or build a binary with `go build`

## Usage

To run, the CLI binary must be available in the path, or present in the current folder (in which case, run with `./optimism-interop-monitoring`). The CLI requires a config file, by default `config.json`. See [configuration](#configuration) for more information on the format and available fields.

To start the system, run:
```bash
optimism-interop-monitoring [-config <path>]
```

The optional `-config` flag specifies a different path for the config file (the default is `"config.json"`)

### Configuration

The `config.json` file has the following structure:
```jsonc
{
    "senderChain": "https://<RPC URL>", // (Required) URL of the sender chain RPC
    "receiverChain": "https://<RPC URL>", // (Required) URL of the receiver chain RPC
    "fetchTime": 1, // Frequency to poll to RPCs, in seconds (default: 1)
    "apiPort": 8800, // Port for the local API (default: 8800)
    "aggregateBlockAmount": 10, // How many blocks to aggregate for alerts (default: 10)
    "alertAvgLatencyMin": 0, // Minimum latency for emitting a high latency alert, disabled if set to 0 (default: 0)
    "alertMissingRelayMin": 0, // Minimum amount of messages received missing sender to emit alert, disabled if set to 0 (default: 0),
    "alertMissingReceptionMin": 0, // Minimum amount of messages sent missing reception to emit alert, disabled if set to 0 (default: 0),
    "telegramToken": "<Bot Token>", // Token for the Telegram bot for alerts (default: "")
    "telegramChatId": "<Chat ID>", // Chat ID for Telegram alerts (default: "")
    "discordWebhookURL": "<Webhook URL>", // URL of Discord webhook for alerts (default: "")
    "customWebhookURL": "<Webhook URL>", // URL of custom webhook for alerts (default: "")
    "purgeOldMessages": true, // Deletes messages without relay/reception after 2*aggregateBlockAmount to save memory (default: true)
    "purgeOldBlocks": false // Deletes block stats after 2*aggregateBlockAmount to save memory (default: true)
}
```

### API

All endpoints require a `GET` request and return JSON. All information is indexed on the block number of the **sender** chain. So for example, a `missingRelay` message on block `10`, means the `receiver` chain got a message from the sender chain for a transaction on block `10`, but no `sender` message was found yet.

#### `/all`

Optional params.:
- `from`: stats will be returned for block numbers above that value (default: `0`)
- `bin`: if set, stats will be aggregated in bins of that size. For example, a bin size of `5` means stats from blocks `10` to `14` will be aggregated on a single bin, labeled `10` (default: not set)

Returns:
```jsonc
{
  "<block number/bin number>": {
    "messageCount": 0, // Total number of successful messages
    "avgLatency": 0, // Average latency between the send and receive transactions
    "missingMessages": 0, // Messages with one of the parts missing (missingReception + missingRelay)
    "missingReception": 0, // Messages sent on the sender chain, but not yet received
    "missingRelay": 0 // Messages received on the receiver chain, but with no sender message found yet
  },
  ...
```

#### `/latest`

Optional param.:
- `count`: how many blocks back to aggregate from (default: `aggregateBlockAmount`)

Returns:
```jsonc
{
  "messageCount": 0, // Total number of successful messages
  "totalLatency": 0, // Total latency between the send and receive transactions
  "avgLatency": 0, // Average latency between the send and receive transactions
  "sentMessages": 0, // Number of `sent` messages found
  "receivedMessages": 0, // Number of `received` messages found
  "missingReception": 0, // Messages sent on the sender chain, but not yet received
  "missingRelay": 0 // Messages received on the receiver chain, but with no sender message found yet
}
```

### Alerts

Alerts measure for signs of failure among the latest `aggregateBlockAmount` blocks (default: `10`). That number also determines how often the system will check for alerts. Currently, the following alert types are supported:

- **High average latency**: triggers when the average latency between the `sent` and `received` transactions is above a custom threshold.
- **Message reception failure**: triggers when the amount of `sent` messages without reception is above a custom threshold.
- **Message relayed without sender transaction**: triggers when the amount of `received` messages without a corresponding `sent` message is above a custom threshold.

However, it is simple to add custom alerts for other possible tracking, requiring recompilation. For that, see [fetcher.go](./fetcher.go#L297). Note that the same information as in the `/latest` API endpoint can be used, with the `stats` struct.

Alerts are automatically relayed to specified alert channels. These are:

- **Discord**: Discord webhooks can be specified for alerts. Set the `discordWebhookURL` flag on `config.json` to enable.
- **Telegram**: Telegram bots are supported for relaying alerts. Set the `telegramToken` and `telegramChatId` on `config.json` to enable.
- **Custom Webhooks**: Other custom webhooks can be specified . Set the `customWebhookURL` flag on `config.json` to enable. By default, a `POST` request with a JSON body of `{ "text": <message>}` is sent, but this can be easily modified on the [alert_sender.go](./alert_sender.go#L92) file.
- **Slack**: The custom webhook format is intentionally compatible with Slack webhooks, so set the `customWebhookURL` flag on `config.json` to enable.

The alert format is as follows:
```
Alert: <Alert type> at <Value>

<Latest block statistics in JSON, same as `/latest` endpoint>
```