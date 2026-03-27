package netif

import (
	"context"
	"log"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/pancpp/peanut/p2p"
	"github.com/songgao/water/waterutil"
)

type Forwarder struct {
	host        host.Host
	routeTab    *RouteTable
	toTunChan   chan []byte
	fromTunChan chan []byte
	enableTrace bool
}

func NewForwarder(h host.Host, rt *RouteTable, toTunChan, fromTunChan chan []byte) (*Forwarder, error) {
	f := &Forwarder{
		host:        h,
		routeTab:    rt,
		toTunChan:   toTunChan,
		fromTunChan: fromTunChan,
		enableTrace: true,
	}
	h.SetStreamHandler(p2p.PROTOCOL_FORWARD, f.handleP2PData)

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
		f.toTunChan <- pkt
	}
}

func (f *Forwarder) handleTunData(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return

		case b := <-f.fromTunChan:
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

	if f.enableTrace {
		log.Println("[forwarder] dstIPAddr:", dstIPAddr)
	}

	pid, ok := f.routeTab.FindPeerID(dstIPAddr)
	if !ok {
		log.Println("[forwarder] pid not found")
		return
	}

	stream, err := f.host.NewStream(ctx, pid, p2p.PROTOCOL_FORWARD)
	if err != nil {
		log.Printf("[forwarder] new stream to peer %s err: %v", pid, err)
		return
	}
	defer stream.Close()

	stream.SetWriteDeadline(time.Now().Add(p2p.P2P_WRITE_TIMEOUT))
	if _, err := stream.Write(b); err != nil {
		stream.Reset()
		log.Printf("[forwarder] write to peer %s err: %v", pid, err)
		return
	}
}
