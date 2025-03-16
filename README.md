# banserver

banserver is a distributed Teeworlds & DDNet banserver which connects to your servers via econ and parses log messages, reacts to them and executes econ commands in response to the parsed messages.

This application does not have any kind of persistence layer like a database or some kind of key value store. It solely depends on your configuration files.

These log messages may contain:

- player ips joining your server which might be banned from doing so, in which case they might be banned based on an ip blacklist.
- chat messages containing links to malicious websites, in which case the player may be banned automatically banned.
- propagate bans from one server to all other servers.

## Installation

Download the executable from the releases page or install it using the Go toolchain:

```shell
go install github.com/jxsl13/banserver@latest
```

## Setup tips

In case that the banserver and the game servers are running on different machines, it is recommended to have a secure connection between the banserver and the game servers, as the econ connection is unencrypted and sends the password and every command in plain text.
One way to achieve this is via `autossh`, which allows you to tunnel local server ports like the econ port `127.0.0.1:<port>`.
Another way is to have an overlay network like `tailscale` which allows you to connect to the server via a secure wireguard connection using  `<tailscale IP>:<port>`.

## Usage

```shell
$ banserver --help
Environment variables:
  ECON_ADDRESSES            comma separated list of econ addresses (<ip/hostname>:port)
  ECON_PASSWORDS            comma separated list of econ passwords
  ECON_RECONNECT_DELAY       (default: "10s")
  ECON_RECONNECT_TIMEOUT     (default: "24h0m0s")
  IP_BLACKLISTS             comma separated list of files containing ip ranges to blacklist
  CHAT_BLACKLISTS           comma separated list that contains regular expressions to check message blacklists
  PROPAGATE                 propagate bans and unbans from one game server to all other game servers (default: "false")
  PERMA_BAN_REASON          default reason for permabans (default: "permanently banned")
  PERMA_BAN_DURATION        default duration for permabans (default: "24h0m0s")
  CHAT_BAN_REASON           default reason for chat bans (default: "prohibited chat message")
  CHAT_BAN_DURATION         default duration for chat bans (default: "24h0m0s")

Usage:
  banserver [flags]
  banserver [command]

Available Commands:
  completion  Generate completion script
  help        Help about any command

Flags:
      --chat-ban-duration duration        default duration for chat bans (default 24h0m0s)
      --chat-ban-reason string            default reason for chat bans (default "prohibited chat message")
      --chat-blacklists string            comma separated list that contains regular expressions to check message blacklists
  -c, --config string                     .env config file path (or via env variable CONFIG)
      --econ-addresses string             comma separated list of econ addresses (<ip/hostname>:port)
      --econ-passwords string             comma separated list of econ passwords
      --econ-reconnect-delay duration      (default 10s)
      --econ-reconnect-timeout duration    (default 24h0m0s)
  -h, --help                              help for banserver
      --ip-blacklists string              comma separated list of files containing ip ranges to blacklist
      --perma-ban-duration duration       default duration for permabans (default 24h0m0s)
      --perma-ban-reason string           default reason for permabans (default "permanently banned")
      --propagate                         propagate bans and unbans from one game server to all other game servers

Use "banserver [command] --help" for more information about a command.
```
