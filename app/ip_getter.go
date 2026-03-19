package app

import (
	"crypto/sha256"
	"net"

	"github.com/libp2p/go-libp2p/core/peer"
)

type IPGetter interface {
	GetIPv4ByPeerID(pid peer.ID) (net.IPNet, error)
}

type BasicIPGetter struct {
}

func newBasicIPGetter() *BasicIPGetter {
	return &BasicIPGetter{}
}

func (ipg *BasicIPGetter) GetIPv4ByPeerID(pid peer.ID) (net.IPNet, error) {
	hash := sha256.Sum256([]byte(pid))
	if (hash[0] == 0 && hash[1] == 0 && hash[2] == 0) ||
		(hash[0] == 255 && hash[1] == 255 && hash[2] == 255) {
		hash[0] = hash[3]
		hash[1] = hash[4]
		hash[2] = hash[5]
	}
	ip := net.IPv4(10, hash[0], hash[1], hash[2])
	mask := net.CIDRMask(32, 32)

	return net.IPNet{IP: ip, Mask: mask}, nil
}
