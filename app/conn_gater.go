package app

import (
	"log"

	"github.com/libp2p/go-libp2p/core/control"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	multiaddr "github.com/multiformats/go-multiaddr"
	"github.com/pancpp/peanut/conf"
)

type ConnGater struct {
	allowList map[peer.ID]struct{}
}

func newConnGater(allowlist *Allowlist) (*ConnGater, error) {
	var peerIdList []peer.ID

	// add discovery servers to Peer ID list
	discMultiAddrs := conf.GetStringSlice("p2p.discovery_multiaddrs")
	for _, addr := range discMultiAddrs {
		maddr, err := multiaddr.NewMultiaddr(addr)
		if err != nil {
			log.Printf("discovery server multi-addr parsing err: %v, %v", err, addr)
			return nil, err
		}

		info, err := peer.AddrInfoFromP2pAddr(maddr)
		if err != nil {
			return nil, err
		}

		peerIdList = append(peerIdList, info.ID)
	}

	// add relay servers to Peer ID list
	relayMultiAddrs := conf.GetStringSlice("p2p.relay_multiaddrs")
	for _, addr := range relayMultiAddrs {
		maddr, err := multiaddr.NewMultiaddr(addr)
		if err != nil {
			log.Printf("relay server multi-addr parsing err: %v, %v", err, addr)
			return nil, err
		}

		info, err := peer.AddrInfoFromP2pAddr(maddr)
		if err != nil {
			return nil, err
		}

		peerIdList = append(peerIdList, info.ID)
	}

	// ad allowlist to Peer ID list
	for _, pid := range allowlist.GetAllPeers() {
		peerIdList = append(peerIdList, pid)
	}

	// load peer IDs from allowlist file
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
