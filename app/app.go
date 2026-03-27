package app

import (
	"context"
	"log"
	"net"
	"os"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/pancpp/peanut/conf"
	"github.com/pancpp/peanut/netif"
	"github.com/pancpp/peanut/p2p"
	"go.yaml.in/yaml/v2"
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
	allowlist, err := getAllowlist()
	if err != nil {
		return err
	}

	// init connection gater
	connGater := p2p.NewConnGater(allowlist, discoveryAddrInfo, staticRelayAddrInfo)

	// init p2p host
	p2pHost, err := p2p.NewHost(connGater, discoveryAddrInfo, staticRelayAddrInfo)
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
	fromTunChan := make(chan []byte)
	toTunChan := make(chan []byte)
	tunName := conf.GetString("vpn.tun_name")
	tunIface, err := netif.NewTunIface(tunName, netif.TUN_DEFAULT_MTU, netif.TUN_DEFAULT_TIMEOUT, fromTunChan, toTunChan)
	if err != nil {
		log.Printf("[app] create tun %s err: %v", tunName, err)
		return err
	}

	// set IP address
	if err := tunIface.ReplaceIPAddr(localIPNet); err != nil {
		log.Printf("[app] set local IPNet %v err: %v", localIPNet, err)
		return err
	}

	// init IP rule
	if err := netif.InitIPRule(); err != nil {
		log.Println("[app] init IP rule err:", err)
		return err
	}

	// create route table
	routeTab := netif.NewRouteTable()

	// set ip route and add IP to allowlist
	for _, pid := range allowlist {
		ipNet, err := ipGetter.GetIPv4ByPeerID(pid)
		if err != nil {
			log.Printf("[app] get IPNet of peer %s err: %v", pid, err)
			return err
		}
		routeTab.Set(pid, ipNet.IP)
		if err := tunIface.ReplaceRoute(ipNet); err != nil {
			log.Printf("[app] set IPNet %v to route err: %v", ipNet, err)
			return err
		}
	}

	// init forwarder
	forwarder, err := netif.NewForwarder(p2pHost, routeTab, toTunChan, fromTunChan)
	if err != nil {
		return err
	}

	// create discover service
	discoverService := newDiscoverService(p2pHost, discoveryAddrInfo, allowlist)

	// create announce service
	_, localhostIPNet, _ := net.ParseCIDR("127.0.0.1/8")
	announceService := newAnnounceService(p2pHost, discoveryAddrInfo, []net.IPNet{localIPNet, *localhostIPNet})

	// start services
	tunIface.Start(ctx)
	forwarder.Start(ctx)
	discoverService.Start(ctx)
	announceService.Start(ctx)

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

func getAllowlist() ([]peer.ID, error) {
	// load peer IDs from allowlist file
	type AllowList struct {
		PeerIDs []string `yaml:"peer_ids"`
	}

	allowlistPath := conf.GetString("p2p.allowlist_path")
	data, err := os.ReadFile(allowlistPath)
	if err != nil {
		log.Printf("[allowlist] reading allowlist file err: %v, path: %s", err, allowlistPath)
		return nil, err
	}
	var alist AllowList
	if err := yaml.Unmarshal(data, &alist); err != nil {
		log.Printf("[allowlist] parsing allowlist file err: %v", err)
		return nil, err
	}

	peerIdList := make([]peer.ID, 0, len(alist.PeerIDs))
	for _, peerID := range alist.PeerIDs {
		id, err := peer.Decode(peerID)
		if err != nil {
			return nil, err
		}
		peerIdList = append(peerIdList, id)
	}

	return peerIdList, nil
}
