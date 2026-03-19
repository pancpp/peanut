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
	b2 := (hash[0] & 0b01111111) | 0b01000000
	b3 := hash[1]
	b4 := hash[2]
	if (b2 == 64 && b3 == 0 && b4 == 0) ||
		(b2 == 127 && b3 == 255 && b4 == 255) {
		b2 = (hash[3] & 0b01111111) | 0b01000000
		b3 = hash[4]
		b4 = hash[5]
	}
	ip := net.IPv4(100, b2, b3, b4)
	mask := net.CIDRMask(32, 32)

	return net.IPNet{IP: ip, Mask: mask}, nil
}
