package app

import (
	"context"
	"io"
	"log"
	"net/netip"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/songgao/water"
	"github.com/songgao/water/waterutil"
	"github.com/vishvananda/netlink"
)

const (
	DEFAULT_TUN_MTU     = 1280
	DEFAULT_TUN_TIMEOUT = 10 * time.Second
)

type TunIface struct {
	mtu       int
	timeout   time.Duration
	iface     *water.Interface
	peerMtx   sync.RWMutex // lock peerAddrs and peerIDs
	peerAddrs map[peer.ID]netip.Addr
	peerIDs   map[netip.Addr]peer.ID
}

func NewTunIface(name string, mtu int, timeout time.Duration) (*TunIface, error) {
	// Create tun interface
	var tunconf water.Config
	tunconf.DeviceType = water.TUN
	tunconf.Name = name
	iface, err := water.New(tunconf)
	if err != nil {
		return nil, err
	}

	t := &TunIface{
		mtu:       mtu,
		timeout:   timeout,
		iface:     iface,
		peerAddrs: make(map[peer.ID]netip.Addr),
		peerIDs:   make(map[netip.Addr]peer.ID),
	}

	// Configure the tun interface
	link, err := netlink.LinkByName(iface.Name())
	if err != nil {
		return nil, err
	}

	if err := netlink.LinkSetMTU(link, mtu); err != nil {
		return nil, err
	}

	if err := netlink.LinkSetUp(link); err != nil {
		return nil, err
	}

	// Set PROTOCOL_TUN handler
	gHost.SetStreamHandler(PROTOCOL_TUN, t.handleTunService)

	return t, nil
}

func (t *TunIface) Close() error {
	// Remove PROTOCO_TUN handler
	gHost.RemoveStreamHandler(PROTOCOL_TUN)

	if link, err := netlink.LinkByName(t.iface.Name()); err == nil {
		netlink.LinkSetDown(link)
	}
	return t.iface.Close()
}

func (t *TunIface) SetIPAddr(ipCIDR string) error {
	// Parse IP addr
	ipAddr, err := netlink.ParseAddr(ipCIDR)
	if err != nil {
		log.Printf("(peanut) parse addr %s err: %v", ipCIDR, err)
		return err
	}

	// Get link device
	linkName := t.iface.Name()
	link, err := netlink.LinkByName(linkName)
	if err != nil {
		log.Printf("(peanut) get link by name %s err: %v", linkName, err)
		return err
	}

	// Check whether the IP addr exists
	addrList, err := netlink.AddrList(link, netlink.FAMILY_V4)
	if err != nil {
		log.Println("(peanut) get addr list err:", err)
		return err
	}
	for _, addr := range addrList {
		if addr.Equal(*ipAddr) {
			log.Println("(peanut) addr exists, do not change:", ipAddr)
			return nil
		}
	}

	// Remove old IP addrs
	log.Println("(peanut) del addr list:", addrList)
	for _, addr := range addrList {
		netlink.AddrDel(link, &addr)
	}

	// Add the IP addr
	if err := netlink.AddrAdd(link, ipAddr); err != nil {
		log.Println("(peanut) add addr:", *ipAddr)
		return err
	}

	return nil
}

func (t *TunIface) RemoveIPAddr() error {
	// Remove IP addr
	linkName := t.iface.Name()
	link, err := netlink.LinkByName(linkName)
	if err != nil {
		log.Printf("(peanut) get link by name %s err: %v", linkName, err)
		return err
	}

	addrs, err := netlink.AddrList(link, netlink.FAMILY_V4)
	if err != nil {
		log.Println("(peanut) get addr list err:", err)
		return err
	}
	log.Println("(peanut) del addr list:", addrs)
	for _, addr := range addrs {
		netlink.AddrDel(link, &addr)
	}

	return nil
}

// map key: peerID, map value: ipCIDR
func (t *TunIface) SetPeers(peers map[string]string) error {
	t.peerMtx.Lock()
	defer t.peerMtx.Unlock()

	if len(peers) == 0 {
		return nil
	}

	t.peerAddrs = make(map[peer.ID]netip.Addr)
	t.peerIDs = make(map[netip.Addr]peer.ID)
	for pid, ipCIDR := range peers {
		peerID, err := peer.Decode(pid)
		if err != nil {
			log.Printf("invalid peer ID %v, err: %v", pid, err)
			continue
		}

		prefix, err := netip.ParsePrefix(ipCIDR)
		if err != nil {
			log.Printf("invalid ipCIDR %v, err: %v", ipCIDR, err)
			continue
		}
		ipAddr := prefix.Addr()

		t.peerAddrs[peerID] = ipAddr
		t.peerIDs[ipAddr] = peerID
	}

	return nil
}

func (t *TunIface) AddPeer(pid string, ipCIDR string) error {
	t.peerMtx.Lock()
	defer t.peerMtx.Unlock()

	peerID, err := peer.Decode(pid)
	if err != nil {
		log.Printf("invalid peer ID %v, err: %v", pid, err)
		return err
	}

	prefix, err := netip.ParsePrefix(ipCIDR)
	if err != nil {
		log.Printf("invalid ipCIDR %v, err: %v", ipCIDR, err)
		return err
	}
	ipAddr := prefix.Addr()

	t.peerAddrs[peerID] = ipAddr
	t.peerIDs[ipAddr] = peerID

	return nil
}

func (t *TunIface) RemovePeer(pid string) error {
	t.peerMtx.Lock()
	defer t.peerMtx.Unlock()

	peerID, err := peer.Decode(pid)
	if err != nil {
		log.Printf("invalid peer ID %v, err: %v", pid, err)
		return err
	}

	if ipAddr, ok := t.peerAddrs[peerID]; ok {
		delete(t.peerAddrs, peerID)
		delete(t.peerIDs, ipAddr)
	}

	return nil
}

func (t *TunIface) ClearPeers() {
	t.peerMtx.Lock()
	defer t.peerMtx.Unlock()

	t.peerAddrs = make(map[peer.ID]netip.Addr)
	t.peerIDs = make(map[netip.Addr]peer.ID)
}

func (t *TunIface) Name() string {
	return t.iface.Name()
}

func (t *TunIface) MTU() int {
	return t.mtu
}

func (t *TunIface) Run(ctx context.Context) {
	t.readFromTun(ctx)
}

// process inbound packet from other peer
func (t *TunIface) handleTunService(stream network.Stream) {
	defer stream.Reset()

	remotePeer := stream.Conn().RemotePeer()

	t.peerMtx.RLock()
	_, ok := t.peerAddrs[remotePeer]
	t.peerMtx.RUnlock()

	if !ok {
		log.Printf("remote peer %v not allowed, discard the packet", remotePeer)
		return
	}

	b, err := io.ReadAll(stream)
	if err != nil {
		log.Printf("read data from peer %v err: %v", remotePeer, err)
		return
	}

	// drop all IPv6 packets
	if waterutil.IsIPv6(b) {
		return
	}

	// write packet to tun interface
	go t.iface.Write(b)
}

func (t *TunIface) readFromTun(ctx context.Context) {
	quit := false
	go func() {
		<-ctx.Done()
		quit = true
		t.iface.Close()
	}()

	for !quit {
		b := make([]byte, t.mtu)
		n, err := t.iface.Read(b)
		if err != nil {
			continue
		}

		t.processOutboundPacket(ctx, b[:n])
	}
}

// process outbound packet to other peer
func (t *TunIface) processOutboundPacket(ctx context.Context, b []byte) {
	// drop all IPv6 packets
	if waterutil.IsIPv6(b) {
		return
	}

	// get destination IP
	dstAddr := IPv4DestAddr(b)

	// TODO(Leyuan Pan)
	// 1. support broadcast address 10.254.6.255, 255.255.255.255
	// 2. support multicast address 224.0.0.0 - 239.255.255.255
	// Temporarily disable the multicast packet forwarding
	// if dstAddr.IsMulticast() {
	// 	t.peerMtx.RLock()
	// 	for pid := range t.peerAddrs {
	// 		go forward(t.ctx, t.host, pid, b)
	// 	}
	// 	t.peerMtx.RUnlock()
	// 	return
	// }

	// get peer with dstAddr
	t.peerMtx.RLock()
	if dstPeerID, ok := t.peerIDs[dstAddr]; ok {
		go forward(ctx, gHost, dstPeerID, b, t.timeout)
	}
	t.peerMtx.RUnlock()
}

func forward(ctx context.Context, host host.Host, pid peer.ID, b []byte, timeout time.Duration) {
	stream, err := host.NewStream(ctx, pid, PROTOCOL_TUN)
	if err != nil {
		return
	}

	stream.SetWriteDeadline(time.Now().Add(timeout))
	if _, err := stream.Write(b); err != nil {
		log.Println("write to stream err:", err)
		stream.Reset()
		return
	}

	stream.Close()
}
