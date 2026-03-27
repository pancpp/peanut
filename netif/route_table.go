package netif

import (
	"net"
	"sync"

	"github.com/libp2p/go-libp2p/core/peer"
)

type RouteTable struct {
	peerIdToIpMtx sync.RWMutex
	peerIdToIpMap map[peer.ID]net.IP

	peerIpToIdMtx sync.RWMutex
	peerIpToIdMap map[string]peer.ID
}

func NewRouteTable() *RouteTable {
	return &RouteTable{
		peerIdToIpMap: make(map[peer.ID]net.IP),
		peerIpToIdMap: make(map[string]peer.ID),
	}
}

func (rt *RouteTable) DelPeerIP(ip net.IP) {
	rt.peerIdToIpMtx.Lock()
	defer rt.peerIdToIpMtx.Unlock()

	delete(rt.peerIpToIdMap, ip.String())
}

func (rt *RouteTable) DelPeerID(pid peer.ID) {
	rt.peerIpToIdMtx.Lock()
	defer rt.peerIpToIdMtx.Unlock()

	delete(rt.peerIdToIpMap, pid)
}

func (rt *RouteTable) Set(pid peer.ID, ip net.IP) {
	rt.peerIdToIpMtx.Lock()
	rt.peerIdToIpMap[pid] = ip
	rt.peerIdToIpMtx.Unlock()

	rt.peerIpToIdMtx.Lock()
	rt.peerIpToIdMap[ip.String()] = pid
	rt.peerIpToIdMtx.Unlock()
}

func (rt *RouteTable) FindPeerIP(pid peer.ID) (net.IP, bool) {
	rt.peerIdToIpMtx.Lock()
	defer rt.peerIdToIpMtx.Unlock()

	ip, ok := rt.peerIdToIpMap[pid]
	if !ok {
		return nil, false
	}

	return ip, true
}

func (rt *RouteTable) FindPeerID(ip net.IP) (peer.ID, bool) {
	rt.peerIpToIdMtx.Lock()
	defer rt.peerIpToIdMtx.Unlock()

	ipstr := ip.String()
	pid, ok := rt.peerIpToIdMap[ipstr]

	return pid, ok
}
