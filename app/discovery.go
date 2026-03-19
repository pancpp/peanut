package app

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	multiaddr "github.com/multiformats/go-multiaddr"
)

type DiscoveryRequestMsg struct {
	PeerIDs []string `json:"peer_ids"`
}

type DiscoveryPeerMsg struct {
	Multiaddrs []string `json:"multi_addrs"`
}

type DiscoveryResponseMsg struct {
	PeerInfo map[string]DiscoveryPeerMsg `json:"peer_info"`
}

type DiscoveryService struct {
	host         host.Host
	allowlist    *Allowlist
	discAddrInfo []peer.AddrInfo
}

func newDiscoveryService(host host.Host, discAddrInfo []peer.AddrInfo, allowlist *Allowlist) *DiscoveryService {
	return &DiscoveryService{
		host:         host,
		allowlist:    allowlist,
		discAddrInfo: discAddrInfo,
	}
}

func (ds *DiscoveryService) Start(ctx context.Context) {
	go ds.Run(ctx)
}

func (ds *DiscoveryService) Run(ctx context.Context) {
	// discover peers immediately
	for _, addrInfo := range ds.discAddrInfo {
		ds.discoverPeers(ctx, addrInfo.ID)
	}

	// periodically udpate peer information
	ticker := time.NewTicker(DISCOVERY_TICKS)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			for _, addrInfo := range ds.discAddrInfo {
				ds.discoverPeers(ctx, addrInfo.ID)
			}
		}
	}
}

func (ds *DiscoveryService) discoverPeers(ctx context.Context, discPID peer.ID) error {
	peers := ds.allowlist.GetAllPeers()

	var reqMsg DiscoveryRequestMsg
	for _, peer := range peers {
		reqMsg.PeerIDs = append(reqMsg.PeerIDs, peer.String())
	}
	b, err := json.Marshal(&reqMsg)
	if err != nil {
		log.Println("[discovery] json marshal err:", err)
		return err
	}

	stream, err := ds.host.NewStream(ctx, discPID, PROTOCOL_DISCOVERY)
	if err != nil {
		return err
	}
	defer stream.Close()

	stream.SetWriteDeadline(time.Now().Add(P2P_WRITE_TIMEOUT))
	if _, err := stream.Write(b); err != nil {
		log.Printf("[discovery] write to peer %s err: %v", discPID, err)
		stream.Reset()
		return err
	}
	stream.CloseWrite()

	stream.SetReadDeadline(time.Now().Add(P2P_READ_TIMEOUT))
	data, err := io.ReadAll(stream)
	if err != nil {
		log.Printf("[discovery] read from peer %s err: %v", discPID, err)
		stream.Reset()
		return err
	}

	var respMsg DiscoveryResponseMsg
	if err := json.Unmarshal(data, &respMsg); err != nil {
		log.Printf("[discovery] json unmarshal err: %v", err)
		return err
	}

	// Add multi-addresses into the peerstore of the host
	for pidStr, peerMsg := range respMsg.PeerInfo {
		pid, err := peer.Decode(pidStr)
		if err != nil {
			log.Printf("[discovery] invalid peer ID %s: %v", pidStr, err)
			continue
		}

		// check peer ID map
		if !ds.allowlist.PeerIDExists(pid) {
			continue
		}

		// update multiaddrs in host peerstore
		var addrs []multiaddr.Multiaddr
		for _, addrStr := range peerMsg.Multiaddrs {
			addr, err := multiaddr.NewMultiaddr(addrStr)
			if err != nil {
				log.Printf("[discovery] invalid multiaddr %s: %v", addrStr, err)
				continue
			}
			addrs = append(addrs, addr)
		}
		if len(addrs) > 0 {
			ds.host.Peerstore().AddAddrs(pid, addrs, P2P_PEERSTORE_TTL)
		}
	}

	log.Println("[discovery] discovered:", string(data))

	return nil
}
