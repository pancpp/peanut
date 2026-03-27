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
	"github.com/pancpp/peanut/p2p"
)

type AnnounceMsg struct {
	MultiAddrs []string `json:"multi_addrs"`
}

type AnnounceService struct {
	host         host.Host
	discAddrInfo []peer.AddrInfo
	denyList     []net.IPNet
}

func newAnnounceService(host host.Host, discAddrInfo []peer.AddrInfo, denyList []net.IPNet) *AnnounceService {
	return &AnnounceService{
		host:         host,
		discAddrInfo: discAddrInfo,
		denyList:     denyList,
	}
}

func (as *AnnounceService) Start(ctx context.Context) {
	go as.Run(ctx)
}

func (as *AnnounceService) Run(ctx context.Context) {
	// run the announce service immediately
	as.announce(ctx)

	// periodically run the announce service
	ticker := time.NewTicker(ANNOUNCE_TICKS)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			as.announce(ctx)
		}
	}
}

func (as *AnnounceService) denied(addr multiaddr.Multiaddr) bool {
	var ip net.IP
	multiaddr.ForEach(addr, func(c multiaddr.Component) bool {
		switch c.Protocol().Code {
		case multiaddr.P_IP4, multiaddr.P_IP6:
			ip = net.ParseIP(c.Value())
			return false // stop iteration once found
		}
		return true
	})

	for _, ipnet := range as.denyList {
		if ipnet.Contains(ip) {
			return true
		}
	}

	return false
}

func (as *AnnounceService) announce(ctx context.Context) {
	multiAddrs := as.host.Addrs()

	var m AnnounceMsg
	for _, addr := range multiAddrs {
		if !as.denied(addr) {
			m.MultiAddrs = append(m.MultiAddrs, addr.String())
		}
	}

	b, err := json.Marshal(&m)
	if err != nil {
		log.Println("[announce] json marshal err")
		return
	}
	for _, addrInfo := range as.discAddrInfo {
		as.postMsg(ctx, addrInfo.ID, b)
	}
}

func (as *AnnounceService) postMsg(ctx context.Context, discPID peer.ID, b []byte) {
	stream, err := as.host.NewStream(ctx, discPID, p2p.PROTOCOL_ANNOUNCE)
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

	log.Println("[announce] announced:", string(b))
}
