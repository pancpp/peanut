package app

import (
	"log"

	"github.com/libp2p/go-libp2p/core/control"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	multiaddr "github.com/multiformats/go-multiaddr"
)

type ConnGater struct {
	allowList map[peer.ID]struct{}
}

func newConnGater(allowlist *Allowlist,
	discoveryAddrInfo []peer.AddrInfo,
	staticRelayAddrInfo []peer.AddrInfo) (*ConnGater, error) {
	var peerIdList []peer.ID

	// add discovery servers to Peer ID list
	for _, addrInfo := range discoveryAddrInfo {
		peerIdList = append(peerIdList, addrInfo.ID)
	}

	// add relay servers to Peer ID list
	for _, addrInfo := range staticRelayAddrInfo {
		peerIdList = append(peerIdList, addrInfo.ID)
	}

	// add allowlist to Peer ID list
	for _, pid := range allowlist.GetAllPeers() {
		peerIdList = append(peerIdList, pid)
	}

	// generate connection gater allowlist
	allowList := make(map[peer.ID]struct{}, len(peerIdList))
	for _, pid := range peerIdList {
		allowList[pid] = struct{}{}
	}

	return &ConnGater{allowList: allowList}, nil
}

func (a *ConnGater) InterceptPeerDial(peer.ID) (allow bool) {
	return true
}

func (a *ConnGater) InterceptAddrDial(peer.ID, multiaddr.Multiaddr) (allow bool) {
	return true
}

func (a *ConnGater) InterceptAccept(network.ConnMultiaddrs) (allow bool) {
	return true
}

func (a *ConnGater) InterceptSecured(dir network.Direction, p peer.ID, connAddr network.ConnMultiaddrs) (allow bool) {
	_, ok := a.allowList[p]
	if !ok {
		log.Printf("denied peer %s from %s", p, connAddr.RemoteMultiaddr())
	}
	return ok
}

func (a *ConnGater) InterceptUpgraded(network.Conn) (allow bool, reason control.DisconnectReason) {
	return true, 0
}
