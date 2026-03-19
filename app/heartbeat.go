package app

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
)

type HeartbeatMsg struct {
	IPAddr     string   `json:"ip_addr"`
	MultiAddrs []string `json:"multi_addrs"`
}

func heartbeatService(ctx context.Context, h host.Host, discAddrInfo []peer.AddrInfo, ipAddr net.IP) {
	// run the heartbeat service immediately
	doHeartbeat(ctx, h, discAddrInfo, ipAddr)

	// periodically run the heartbeat service
	ticker := time.NewTicker(HEARTBEAT_TICKS)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			doHeartbeat(ctx, h, discAddrInfo, ipAddr)
		}
	}
}

func doHeartbeat(ctx context.Context, h host.Host, discAddrInfo []peer.AddrInfo, ipAddr net.IP) {
	multiAddrs := h.Addrs()

	m := HeartbeatMsg{
		IPAddr: ipAddr.String(),
	}
	for _, addr := range multiAddrs {
		m.MultiAddrs = append(m.MultiAddrs, addr.String())
	}
	b, err := json.Marshal(&m)
	if err != nil {
		log.Println("json marshal err")
		return
	}
	for _, addrInfo := range discAddrInfo {
		postHeartbeatMsg(ctx, h, addrInfo.ID, b)
	}
}

func postHeartbeatMsg(ctx context.Context, h host.Host, discPID peer.ID, b []byte) {
	stream, err := h.NewStream(ctx, discPID, PROTOCOL_HEARTBEAT)
	if err != nil {
		log.Println("[heartbeat] new stream to discovery server err:", err)
		return
	}

	stream.SetWriteDeadline(time.Now().Add(P2P_WRITE_TIMEOUT))
	if _, err := stream.Write(b); err != nil {
		log.Printf("[heartbeat] write to peer %s err: %v", discPID, err)
		stream.Reset()
		return
	}

	stream.Close()

	log.Println("[heartbeat] reported:", string(b))
}
