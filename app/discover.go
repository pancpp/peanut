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
	"github.com/pancpp/peanut/p2p"
)

type DiscoverRequestMsg struct {
	PeerIDs []string `json:"peer_ids"`
}

type DiscoverPeerMsg struct {
	Multiaddrs []string `json:"multi_addrs"`
}

type DiscoverResponseMsg struct {
	PeerInfo map[string]DiscoverPeerMsg `json:"peer_info"`
}

type DiscoverService struct {
	host         host.Host
	allowlist    []peer.ID
	discAddrInfo []peer.AddrInfo
}

func newDiscoverService(host host.Host, discAddrInfo []peer.AddrInfo, allowlist []peer.ID) *DiscoverService {
	return &DiscoverService{
		host:         host,
		allowlist:    allowlist,
		discAddrInfo: discAddrInfo,
	}
}

func (ds *DiscoverService) Start(ctx context.Context) {
	go ds.Run(ctx)
}

func (ds *DiscoverService) Run(ctx context.Context) {
	// discover peers immediately
	for _, addrInfo := range ds.discAddrInfo {
		ds.discoverPeers(ctx, addrInfo.ID)
	}

	// periodically udpate peer information
	ticker := time.NewTicker(DISCOVER_TICKS)
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

func (ds *DiscoverService) discoverPeers(ctx context.Context, discPID peer.ID) error {
	var reqMsg DiscoverRequestMsg
	for _, peer := range ds.allowlist {
		reqMsg.PeerIDs = append(reqMsg.PeerIDs, peer.String())
	}
	b, err := json.Marshal(&reqMsg)
	if err != nil {
		log.Println("[discover] json marshal err:", err)
		return err
	}

	stream, err := ds.host.NewStream(ctx, discPID, p2p.PROTOCOL_DISCOVER)
	if err != nil {
		return err
	}
	defer stream.Close()

	stream.SetWriteDeadline(time.Now().Add(p2p.P2P_WRITE_TIMEOUT))
	if _, err := stream.Write(b); err != nil {
		log.Printf("[discover] write to peer %s err: %v", discPID, err)
		stream.Reset()
		return err
	}
	stream.CloseWrite()

	stream.SetReadDeadline(time.Now().Add(p2p.P2P_READ_TIMEOUT))
	data, err := io.ReadAll(stream)
	if err != nil {
		log.Printf("[discover] read from peer %s err: %v", discPID, err)
		stream.Reset()
		return err
	}

	var respMsg DiscoverResponseMsg
	if err := json.Unmarshal(data, &respMsg); err != nil {
		log.Printf("[discover] json unmarshal err: %v", err)
		return err
	}

	// Add multi-addresses into the peerstore of the host
	for pidStr, peerMsg := range respMsg.PeerInfo {
		pid, err := peer.Decode(pidStr)
		if err != nil {
			log.Printf("[discover] invalid peer ID %s: %v", pidStr, err)
			continue
		}

		// update multiaddrs in host peerstore
		var addrs []multiaddr.Multiaddr
		for _, addrStr := range peerMsg.Multiaddrs {
			addr, err := multiaddr.NewMultiaddr(addrStr)
			if err != nil {
				log.Printf("[discover] invalid multiaddr %s: %v", addrStr, err)
				continue
			}
			addrs = append(addrs, addr)
		}
		if len(addrs) > 0 {
			ds.host.Peerstore().AddAddrs(pid, addrs, p2p.P2P_PEERSTORE_TTL)
		}
	}

	log.Println("[discover] discovered:", string(data))

	return nil
}
