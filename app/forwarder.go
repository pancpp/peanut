package app

import (
	"context"
	"log"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/songgao/water/waterutil"
)

type Forwarder struct {
	host         host.Host
	allowlist    *Allowlist
	readTunChan  <-chan []byte
	writeTunChan chan []byte
}

func newForwarder(h host.Host, t *TunIface, allowlist *Allowlist) (*Forwarder, error) {
	f := &Forwarder{
		host:         h,
		allowlist:    allowlist,
		readTunChan:  t.GetReadTunChan(),
		writeTunChan: t.GetWriteTunChan(),
	}
	h.SetStreamHandler(PROTOCOL_FORWARD, f.handleP2PData)

	return f, nil
}

func (f *Forwarder) Start(ctx context.Context) {
	go f.handleTunData(ctx)
}

func (f *Forwarder) handleP2PData(stream network.Stream) {
	defer stream.Close()

	buf := make([]byte, TUN_DEFAULT_MTU)
	for {
		n, err := stream.Read(buf)
		if err != nil {
			return
		}
		pkt := make([]byte, n)
		copy(pkt, buf[:n])
		f.writeTunChan <- pkt
	}
}

func (f *Forwarder) handleTunData(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return

		case b := <-f.readTunChan:
			f.processTunData(ctx, b)
		}
	}
}

func (f *Forwarder) processTunData(ctx context.Context, b []byte) {
	// drop all IPv6 packets
	if waterutil.IsIPv6(b) {
		log.Println("[forwarder] drop IPv6 packets")
		return
	}

	dstIPAddr := waterutil.IPv4Destination(b).To4()
	log.Println("[forwarder] dstIPAddr:", dstIPAddr)
	pid, ok := f.allowlist.GetPeerIDByIP(dstIPAddr)
	if !ok {
		log.Println("[forwarder] pid not found")
		return
	}

	stream, err := f.host.NewStream(ctx, pid, PROTOCOL_FORWARD)
	if err != nil {
		log.Printf("[forwarder] new stream to peer %s err: %v", pid, err)
		return
	}
	defer stream.Close()

	stream.SetWriteDeadline(time.Now().Add(P2P_WRITE_TIMEOUT))
	if _, err := stream.Write(b); err != nil {
		stream.Reset()
		log.Printf("[forwarder] write to peer %s err: %v", pid, err)
		return
	}
}
