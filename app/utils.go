package app

import (
	"crypto/sha256"
	"net"

	"github.com/libp2p/go-libp2p/core/peer"
)

func GenIPv4FromPeerID(pid peer.ID) net.IP {
	hash := sha256.Sum256([]byte(pid))
	if (hash[0] == 0 && hash[1] == 0 && hash[2] == 0) ||
		(hash[0] == 255 && hash[1] == 255 && hash[2] == 255) {
		hash[0] = hash[3]
		hash[1] = hash[4]
		hash[2] = hash[5]
	}
	return net.IPv4(10, hash[0], hash[1], hash[2])
}
