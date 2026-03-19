package app

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
)

type HeartbeatMsg struct {
	MultiAddrs []string `json:"multi_addrs"`
}

type HeartbeatService struct {
	host         host.Host
	discAddrInfo []peer.AddrInfo
}

func newHeartbeatService(host host.Host, discAddrInfo []peer.AddrInfo) *HeartbeatService {
	return &HeartbeatService{
		host:         host,
		discAddrInfo: discAddrInfo,
	}
}

func (hs *HeartbeatService) Start(ctx context.Context) {
	go hs.Run(ctx)
}

func (hs *HeartbeatService) Run(ctx context.Context) {
	// run the heartbeat service immediately
	hs.doHeartbeat(ctx)

	// periodically run the heartbeat service
	ticker := time.NewTicker(HEARTBEAT_TICKS)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			hs.doHeartbeat(ctx)
		}
	}
}

func (hs *HeartbeatService) doHeartbeat(ctx context.Context) {
	multiAddrs := hs.host.Addrs()

	var m HeartbeatMsg
	for _, addr := range multiAddrs {
		m.MultiAddrs = append(m.MultiAddrs, addr.String())
	}
	b, err := json.Marshal(&m)
	if err != nil {
		log.Println("[heartbeat] json marshal err")
		return
	}
	for _, addrInfo := range hs.discAddrInfo {
		hs.postHeartbeatMsg(ctx, addrInfo.ID, b)
	}
}

func (hs *HeartbeatService) postHeartbeatMsg(ctx context.Context, discPID peer.ID, b []byte) {
	stream, err := hs.host.NewStream(ctx, discPID, PROTOCOL_HEARTBEAT)
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
