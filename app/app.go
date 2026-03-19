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
	// get discovery server address info
	discoveryAddrInfo, err := getDiscoveryAddrs()
	if err != nil {
		return err
	}
	log.Println("app: discovery addr info:", discoveryAddrInfo)

	// get static relay addresse info
	staticRelayAddrInfo, err := getStaticRelayAddrs()
	if err != nil {
		return err
	}
	log.Println("app: static relay addr info:", staticRelayAddrInfo)

	// init allowlist
	allowlist, err := newAllowlist()
	if err != nil {
		return err
	}

	// init connection gater
	connGater, err := newConnGater(allowlist, discoveryAddrInfo, staticRelayAddrInfo)
	if err != nil {
		return err
	}

	// init p2p host
	p2pHost, err := newHost(connGater, discoveryAddrInfo, staticRelayAddrInfo)
	if err != nil {
		return err
	}

	// generate local IP address
	ipAddr := GenIPv4FromPeerID(p2pHost.ID())
	mask := net.CIDRMask(32, 32)
	ipNet := net.IPNet{
		IP:   ipAddr,
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
	go heartbeatService(ctx, p2pHost, discoveryAddrInfo, ipAddr)
	go discoveryService(ctx, p2pHost, discoveryAddrInfo, allowlist)

	return nil
}

func getDiscoveryAddrs() ([]peer.AddrInfo, error) {
	var discoveryAddrInfo []peer.AddrInfo
	for _, addrStr := range conf.GetStringSlice("p2p.discovery_multiaddrs") {
		addrInfo, err := peer.AddrInfoFromString(addrStr)
		if err != nil {
			return nil, err
		}
		discoveryAddrInfo = append(discoveryAddrInfo, *addrInfo)
	}

	return discoveryAddrInfo, nil
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
