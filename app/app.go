package app

import (
	"context"
	"log"
	"net"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/pancpp/peanut/conf"
)

var (
	gLocalIP net.IP
)

func Init(ctx context.Context) error {
	// init allowlist
	allowlist, err := newAllowlist()
	if err != nil {
		return err
	}

	// init static relay addresses
	staticRelay, err := getStaticRelayAddrs()
	if err != nil {
		return err
	}

	// init connection gater
	connGater, err := newConnGater(allowlist)
	if err != nil {
		return err
	}

	// init p2p host
	p2pHost, err := newHost(connGater, staticRelay)
	if err != nil {
		return err
	}

	// generate local IP address
	ip := GenIPv4FromPeerID(p2pHost.ID())
	mask := net.CIDRMask(32, 32)
	ipNet := net.IPNet{
		IP:   ip,
		Mask: mask,
	}
	ipCIDR := ipNet.String()
	log.Println("app: local IP address:", ipCIDR)

	// init tun interface
	tunName := conf.GetString("vpn.tun_name")
	tunIface, err := newTunIface(tunName, TUN_DEFAULT_MTU, TUN_DEFAULT_TIMEOUT)
	if err != nil {
		return err
	}
	if err := tunIface.ReplaceIPAddr(ipCIDR); err != nil {
		return err
	}

	// init forwarder
	forwarder, err := newForwarder(p2pHost, tunIface, allowlist)
	if err != nil {
		return err
	}

	// start services
	tunIface.Start(ctx)
	forwarder.Start(ctx)
	go heartbeatService(ctx, p2pHost)
	go discoveryService(ctx, p2pHost, allowlist)

	return nil
}

func getStaticRelayAddrs() ([]peer.AddrInfo, error) {
	var staticRelayAddrInfo []peer.AddrInfo
	for _, addrStr := range conf.GetStringSlice("p2p.relay_multiaddrs") {
		addrInfo, err := peer.AddrInfoFromString(addrStr)
		if err != nil {
			return nil, err
		}
		staticRelayAddrInfo = append(staticRelayAddrInfo, *addrInfo)
	}

	return staticRelayAddrInfo, nil
}
