# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Peanut is a P2P virtual network application built on libp2p with TUN interface packet forwarding. Linux-only (requires TUN device and netlink). Configuration via YAML file (default `/etc/peanut/peanut.yaml`), using Viper with pflag for CLI flags.

## Build & Run Commands

```bash
go build -o peanut .    # Build the binary
go build ./...          # Build all packages
go vet ./...            # Static analysis
go test ./...           # Run all tests
go test -run TestName   # Run a single test
```

CLI flags: `-c <config_path>` (default `/etc/peanut/peanut.yaml`), `-V` (version info).

## Architecture

The application is structured into these packages:

### `main.go`
Entry point. Initializes config, logger, and app in sequence, then blocks on signal (SIGTERM/SIGINT/SIGHUP).

### `conf/` — Configuration
Wraps Viper. Parsed from YAML config file. Key config keys:
- `p2p.private_key_path` — base64-encoded Ed25519 private key file
- `p2p.pnet_psk_path` — private network pre-shared key file
- `p2p.listen_multiaddrs` — libp2p listen addresses (default QUIC on port 19882)
- `p2p.discovery_multiaddrs` — discovery server multiaddrs
- `p2p.relay_multiaddrs` — relay server multiaddrs
- `p2p.allowlist_path` — YAML file of allowed peer IDs
- `tun_name` — TUN interface name (default `peanut0`)

### `logger/` — Logging
Configures `log` standard library output to lumberjack for file rotation. Optional console output via `log.enable_console_log`.

### `app/` — Application Core
Initializes the libp2p host with QUIC transport, NAT traversal, hole punching, static relays, and a connection gater (allowlist-based).

Key files:
- `app.go` — `Init()` orchestrates startup: static relays → connection gater → libp2p host → TUN
- `conn_gater.go` — `ConnGater` implements `ConnectionGater`; allows only allowlisted peer IDs at the `InterceptSecured` stage
- `tun.go` — `TunIface` manages TUN device for IP packet forwarding between peers
- `host_relay.go` — relay reservation loop
- `conn_notifyer.go` — connection event notifications
- `holepunch_tracer.go` — hole-punching event tracer
- `protocols.go` — protocol IDs: `/peanut/heartbeat/1.0`, `/peanut/discovery/1.0`, `/peanut/tun/1.0`

### `peerstore/` — Peer Storage
Placeholder package (empty).

### `scripts/peanut-key/`
Key generation tool with subcommands:
- `peanut-key pkey [-o file]` — generate Ed25519 private key
- `peanut-key pnet [-o file]` — generate private network PSK
- `peanut-key pid -i <file>` — show peer ID from private key file

## Key Dependencies

- `github.com/libp2p/go-libp2p` — P2P networking framework (QUIC transport)
- `github.com/songgao/water` — TUN/TAP device creation
- `github.com/vishvananda/netlink` — Linux network interface configuration
- `github.com/spf13/viper` + `pflag` — configuration and CLI flags
- `gopkg.in/natefinch/lumberjack.v2` — log file rotation
