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
	connGater := newConnGater(allowlist, discoveryAddrInfo, staticRelayAddrInfo)

	// init p2p host
	p2pHost, err := newHost(connGater, discoveryAddrInfo, staticRelayAddrInfo)
	if err != nil {
		log.Println("[app] create p2p host err:", err)
		return err
	}

	// create IP getter
	ipGetter := newBasicIPGetter()

	// get local IPNet
	localIPNet, err := ipGetter.GetIPv4ByPeerID(p2pHost.ID())
	if err != nil {
		log.Println("[app] get local IP err:", err)
		return err
	}
	log.Println("app: local IP address:", localIPNet.String())

	// init tun interface
	tunName := conf.GetString("vpn.tun_name")
	tunIface, err := newTunIface(tunName, TUN_DEFAULT_MTU, TUN_DEFAULT_TIMEOUT)
	if err != nil {
		log.Printf("[app] create tun %s err: %v", tunName, err)
		return err
	}

	// set IP address
	if err := tunIface.ReplaceIPAddr(localIPNet); err != nil {
		log.Printf("[app] set local IPNet %v err: %v", localIPNet, err)
		return err
	}

	// set ip route and add IP to allowlist
	for _, pid := range allowlist.GetAllPeers() {
		ipNet, err := ipGetter.GetIPv4ByPeerID(pid)
		if err != nil {
			log.Printf("[app] get IPNet of peer %s err: %v", pid, err)
			return err
		}

		allowlist.Update(pid, ipNet.IP)

		if err := tunIface.ReplaceRoute(ipNet); err != nil {
			log.Printf("[app] set IPNet %v to route err: %v", ipNet, err)
			return err
		}
	}

	// init forwarder
	forwarder, err := newForwarder(p2pHost, tunIface, allowlist)
	if err != nil {
		return err
	}

	// create discovery service
	discoveryService := newDiscoveryService(p2pHost, discoveryAddrInfo, allowlist)

	// create heartbeat service
	heartbeatService := newHeartbeatService(p2pHost, discoveryAddrInfo)

	// start services
	tunIface.Start(ctx)
	forwarder.Start(ctx)
	discoveryService.Start(ctx)
	heartbeatService.Start(ctx)

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
