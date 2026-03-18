package app

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	multiaddr "github.com/multiformats/go-multiaddr"
)

var (
	gDiscServerPeers []peer.ID
)

type DiscoveryRequestMsg struct {
	PeerIDs []string `json:"peer_ids"`
}

type DiscoveryPeerMsg struct {
	IPAddr     string   `json:"ip_addr"`
	Multiaddrs []string `json:"multi_addrs"`
}

type DiscoveryResponseMsg struct {
	PeerInfo map[string]DiscoveryPeerMsg `json:"peer_info"`
}

func discoveryService(ctx context.Context, h host.Host, allowlist *Allowlist) {
	peers := allowlist.GetAllPeers()

	// discover peers immediately
	doDiscoverPeers(ctx, h, allowlist, peers)

	// periodically udpate peer information
	ticker := time.NewTicker(DISCOVERY_TICKS)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			doDiscoverPeers(ctx, h, allowlist, peers)
		}
	}
}

func doDiscoverPeers(ctx context.Context, h host.Host, allowlist *Allowlist, peers []peer.ID) {
	for _, discPID := range gDiscServerPeers {
		if err := discoverPeers(ctx, h, allowlist, discPID, peers); err != nil {
			continue
		}
	}
}

func discoverPeers(ctx context.Context, h host.Host, allowlist *Allowlist, discPID peer.ID, peers []peer.ID) error {
	var reqMsg DiscoveryRequestMsg
	for _, peer := range peers {
		reqMsg.PeerIDs = append(reqMsg.PeerIDs, peer.String())
	}
	b, err := json.Marshal(&reqMsg)
	if err != nil {
		log.Println("discovery: json marshal err:", err)
		return err
	}

	stream, err := h.NewStream(ctx, discPID, PROTOCOL_DISCOVERY)
	if err != nil {
		return err
	}
	defer stream.Close()

	stream.SetWriteDeadline(time.Now().Add(P2P_WRITE_TIMEOUT))
	if _, err := stream.Write(b); err != nil {
		log.Printf("discovery: write to peer %s err: %v", discPID, err)
		stream.Reset()
		return err
	}

	stream.SetReadDeadline(time.Now().Add(P2P_READ_TIMEOUT))
	buf := make([]byte, 4096)
	n, err := stream.Read(buf)
	if err != nil {
		log.Printf("discovery: read from peer %s err: %v", discPID, err)
		stream.Reset()
		return err
	}

	var respMsg DiscoveryResponseMsg
	if err := json.Unmarshal(buf[:n], &respMsg); err != nil {
		log.Printf("discovery: json unmarshal err: %v", err)
		return err
	}

	// Add multi-addresses into the peerstore of the host
	for pidStr, peerMsg := range respMsg.PeerInfo {
		pid, err := peer.Decode(pidStr)
		if err != nil {
			log.Printf("discovery: invalid peer ID %s: %v", pidStr, err)
			continue
		}

		// check peer ID map
		_, ok := allowlist.GetIPByPeerID(pid)
		if !ok {
			continue
		}

		// update IP address mapping
		ip := net.ParseIP(peerMsg.IPAddr)
		if ip != nil {
			allowlist.Update(pid, ip)
		}

		// update multiaddrs in host peerstore
		var addrs []multiaddr.Multiaddr
		for _, addrStr := range peerMsg.Multiaddrs {
			addr, err := multiaddr.NewMultiaddr(addrStr)
			if err != nil {
				log.Printf("discovery: invalid multiaddr %s: %v", addrStr, err)
				continue
			}
			addrs = append(addrs, addr)
		}
		if len(addrs) > 0 {
			h.Peerstore().AddAddrs(pid, addrs, P2P_PEERSTORE_TTL)
		}
	}

	return nil
}
