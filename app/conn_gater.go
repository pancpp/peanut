package app

import (
	"log"
	"os"

	"github.com/libp2p/go-libp2p/core/control"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	multiaddr "github.com/multiformats/go-multiaddr"
	"github.com/pancpp/peanut/conf"
	"go.yaml.in/yaml/v2"
)

type ConnGater struct {
	allowList map[peer.ID]struct{}
}

func newConnGater() (*ConnGater, error) {
	var peerIdList []peer.ID

	// load peer IDs from discovery servers
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

	// load peer IDs from relay servers
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

	// load peer IDs from allowlist file
	type AllowList struct {
		PeerIDs []string `yaml:"peer_ids"`
	}

	path := conf.GetString("p2p.allowlist_path")
	if path == "" {
		return nil, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		log.Printf("reading allowlist file err: %v, path: %s", err, path)
		return nil, err
	}

	var alist AllowList
	if err := yaml.Unmarshal(data, &alist); err != nil {
		log.Printf("parsing allowlist file err: %v", err)
		return nil, err
	}

	for _, peerID := range alist.PeerIDs {
		id, err := peer.Decode(peerID)
		if err != nil {
			return nil, err
		}
		peerIdList = append(peerIdList, id)
	}

	allowList := make(map[peer.ID]struct{}, len(peerIdList))
	for _, id := range peerIdList {
		allowList[id] = struct{}{}
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
