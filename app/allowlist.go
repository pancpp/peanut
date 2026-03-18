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
	store := Allowlist{
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
		log.Printf("reading allowlist file err: %v, path: %s", err, allowlistPath)
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
		store.peerIdToIpMap[id] = ""
	}

	return &store, nil
}

func (store *Allowlist) Update(pid peer.ID, ip net.IP) {
	store.mtx.Lock()
	defer store.mtx.Unlock()

	if _, ok := store.peerIdToIpMap[pid]; !ok {
		return
	}

	ipstr := ip.String()
	store.peerIdToIpMap[pid] = ipstr
	store.peerIpToIdMap[ipstr] = pid
}

func (store *Allowlist) GetIPByPeerID(pid peer.ID) (net.IP, bool) {
	store.mtx.RLock()
	defer store.mtx.RUnlock()

	ipstr, ok := store.peerIdToIpMap[pid]
	if !ok {
		return nil, false
	}

	return net.ParseIP(ipstr), true
}

func (store *Allowlist) GetPeerIDByIP(ip net.IP) (peer.ID, bool) {
	store.mtx.RLock()
	defer store.mtx.RUnlock()

	ipstr := ip.String()
	pid, ok := store.peerIpToIdMap[ipstr]

	return pid, ok
}

func (store *Allowlist) GetAllPeers() []peer.ID {
	var peers []peer.ID
	for pid := range store.peerIdToIpMap {
		peers = append(peers, pid)
	}

	return peers
}
