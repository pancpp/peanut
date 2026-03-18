package app

import "time"

const (
	P2P_DIAL_TIMEOUT  = 10 * time.Second
	P2P_READ_TIMEOUT  = 10 * time.Second
	P2P_WRITE_TIMEOUT = 10 * time.Second
	P2P_PEERSTORE_TTL = 10 * time.Minute
)

const (
	HEARTBEAT_TICKS = 30 * time.Second
	DISCOVERY_TICKS = 60 * time.Second
)

const (
	PROTOCOL_HEARTBEAT = "/peanut/heartbeat/1.0"
	PROTOCOL_DISCOVERY = "/peanut/discovery/1.0"
	PROTOCOL_FORWARD   = "/peanut/forward/1.0"
)

const (
	TUN_DEFAULT_MTU     = 1280
	TUN_DEFAULT_TIMEOUT = 10 * time.Second
)
