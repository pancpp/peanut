package app

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/pancpp/peanut/p2p"
)

type AnnounceMsg struct {
	MultiAddrs []string `json:"multi_addrs"`
}

type AnnounceService struct {
	host         host.Host
	discAddrInfo []peer.AddrInfo
}

func newAnnounceService(host host.Host, discAddrInfo []peer.AddrInfo) *AnnounceService {
	return &AnnounceService{
		host:         host,
		discAddrInfo: discAddrInfo,
	}
}

func (hs *AnnounceService) Start(ctx context.Context) {
	go hs.Run(ctx)
}

func (hs *AnnounceService) Run(ctx context.Context) {
	// run the announce service immediately
	hs.announce(ctx)

	// periodically run the announce service
	ticker := time.NewTicker(ANNOUNCE_TICKS)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			hs.announce(ctx)
		}
	}
}

func (hs *AnnounceService) announce(ctx context.Context) {
	multiAddrs := hs.host.Addrs()

	var m AnnounceMsg
	for _, addr := range multiAddrs {
		m.MultiAddrs = append(m.MultiAddrs, addr.String())
	}
	b, err := json.Marshal(&m)
	if err != nil {
		log.Println("[announce] json marshal err")
		return
	}
	for _, addrInfo := range hs.discAddrInfo {
		hs.postMsg(ctx, addrInfo.ID, b)
	}
}

func (hs *AnnounceService) postMsg(ctx context.Context, discPID peer.ID, b []byte) {
	stream, err := hs.host.NewStream(ctx, discPID, p2p.PROTOCOL_ANNOUNCE)
	if err != nil {
		log.Println("[announce] new stream to discovery server err:", err)
		return
	}

	stream.SetWriteDeadline(time.Now().Add(p2p.P2P_WRITE_TIMEOUT))
	if _, err := stream.Write(b); err != nil {
		log.Printf("[announce] write to peer %s err: %v", discPID, err)
		stream.Reset()
		return
	}

	stream.Close()

	log.Println("[announce] reported:", string(b))
}
