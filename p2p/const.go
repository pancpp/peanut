package p2p

import "time"

const (
	PROTOCOL_ANNOUNCE = "/peanut/heartbeat/1.0" // /peanut/announce/1.0
	PROTOCOL_DISCOVER = "/peanut/discovery/1.0"
	PROTOCOL_FORWARD  = "/peanut/forward/1.0"
)

const (
	P2P_DIAL_TIMEOUT  = 10 * time.Second
	P2P_READ_TIMEOUT  = 10 * time.Second
	P2P_WRITE_TIMEOUT = 10 * time.Second
	P2P_PEERSTORE_TTL = 30 * time.Minute
)
