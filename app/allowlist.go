package app

import (
	"log"
	"net"
	"os"
	"sync"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/pancpp/peanut/conf"
	"go.yaml.in/yaml/v2"
)

type Allowlist struct {
	mtx           sync.RWMutex
	peerIdToIpMap map[peer.ID]string
	peerIpToIdMap map[string]peer.ID
}

func newAllowlist() (*Allowlist, error) {
	al := Allowlist{
		peerIdToIpMap: make(map[peer.ID]string),
		peerIpToIdMap: make(map[string]peer.ID),
	}

	// load peer IDs from allowlist file
	type AllowList struct {
		PeerIDs []string `yaml:"peer_ids"`
	}

	allowlistPath := conf.GetString("p2p.allowlist_path")
	data, err := os.ReadFile(allowlistPath)
	if err != nil {
		log.Printf("[allowlist] reading allowlist file err: %v, path: %s", err, allowlistPath)
		return nil, err
	}
	var alist AllowList
	if err := yaml.Unmarshal(data, &alist); err != nil {
		log.Printf("[allowlist] parsing allowlist file err: %v", err)
		return nil, err
	}
	for _, peerID := range alist.PeerIDs {
		id, err := peer.Decode(peerID)
		if err != nil {
			return nil, err
		}
		al.peerIdToIpMap[id] = ""
	}

	return &al, nil
}

func (al *Allowlist) Update(pid peer.ID, ip net.IP) {
	al.mtx.Lock()
	defer al.mtx.Unlock()

	if _, ok := al.peerIdToIpMap[pid]; !ok {
		return
	}

	ipstr := ip.String()
	al.peerIdToIpMap[pid] = ipstr
	al.peerIpToIdMap[ipstr] = pid
}

func (al *Allowlist) PeerIDExists(pid peer.ID) bool {
	al.mtx.RLock()
	defer al.mtx.RUnlock()

	_, ok := al.peerIdToIpMap[pid]
	return ok
}

func (al *Allowlist) IPExists(ip net.IP) bool {
	al.mtx.RLock()
	defer al.mtx.RUnlock()

	_, ok := al.peerIpToIdMap[ip.String()]
	return ok
}

func (al *Allowlist) GetIPByPeerID(pid peer.ID) (net.IP, bool) {
	al.mtx.RLock()
	defer al.mtx.RUnlock()

	ipstr, ok := al.peerIdToIpMap[pid]
	if !ok {
		return nil, false
	}

	return net.ParseIP(ipstr), true
}

func (al *Allowlist) GetPeerIDByIP(ip net.IP) (peer.ID, bool) {
	al.mtx.RLock()
	defer al.mtx.RUnlock()

	ipstr := ip.String()
	pid, ok := al.peerIpToIdMap[ipstr]

	return pid, ok
}

func (al *Allowlist) GetAllPeers() []peer.ID {
	var peers []peer.ID
	for pid := range al.peerIdToIpMap {
		peers = append(peers, pid)
	}

	return peers
}
