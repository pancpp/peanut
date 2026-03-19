package app

import (
	"encoding/base64"
	"log"
	"os"
	"strings"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/pnet"
	"github.com/libp2p/go-libp2p/p2p/protocol/holepunch"
	quic "github.com/libp2p/go-libp2p/p2p/transport/quic"
	"github.com/pancpp/peanut/conf"
)

func newHost(connGater *ConnGater,
	discoveryAddrInfo []peer.AddrInfo,
	staticRelayAddrInfo []peer.AddrInfo) (host.Host, error) {
	// libp2p host options
	var opts []libp2p.Option

	// option: private key
	privateKeyPath := conf.GetString("p2p.private_key_path")
	privateKeyB64, err := os.ReadFile(privateKeyPath)
	if err != nil {
		log.Printf("reading private key err: %v, path: %s", err, privateKeyPath)
		return nil, err
	}
	privateKeyBytes, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(privateKeyB64)))
	if err != nil {
		log.Printf("base64 unmarshal err: %v, string: %s", err, string(privateKeyB64))
		return nil, err
	}
	privateKey, err := crypto.UnmarshalPrivateKey(privateKeyBytes)
	if err != nil {
		log.Printf("invalid private key, err: %v, string: %s", err, string(privateKeyBytes))
		return nil, err
	}
	opts = append(opts, libp2p.Identity(privateKey))

	// option: private network PSK
	pskPath := conf.GetString("p2p.pnet_psk_path")
	if pskPath != "" {
		pskFile, err := os.Open(pskPath)
		if err != nil {
			return nil, err
		}
		defer pskFile.Close()

		psk, err := pnet.DecodeV1PSK(pskFile)
		if err != nil {
			return nil, err
		}

		opts = append(opts, libp2p.PrivateNetwork(psk))
		log.Println("private network is enabled")
	}

	// option: listening addresses
	listenAddrs := conf.GetStringSlice("p2p.listen_multiaddrs")
	opts = append(opts,
		libp2p.Transport((quic.NewTransport)),
		libp2p.ListenAddrStrings(listenAddrs...))

	// add connection gater
	if connGater != nil {
		opts = append(opts, libp2p.ConnectionGater(connGater))
	}

	// option: NAT service
	opts = append(opts, libp2p.EnableNATService())

	// option: Attempt to open ports using uPNP for NATed hosts
	opts = append(opts, libp2p.NATPortMap())

	// option: force reachability private
	opts = append(opts, libp2p.ForceReachabilityPrivate())

	// option: static relays
	opts = append(opts, libp2p.EnableAutoRelayWithStaticRelays(staticRelayAddrInfo))

	// option: enable holepunching service
	var holePunchOpts []holepunch.Option
	if conf.GetBool("p2p.enable_holepunch_tracer") {
		holePunchOpts = append(holePunchOpts, holepunch.WithTracer(&holepunchTracer{}))
	}
	opts = append(opts, libp2p.EnableHolePunching(holePunchOpts...))

	// option: set dial timeout
	opts = append(opts, libp2p.WithDialTimeout(P2P_DIAL_TIMEOUT))

	// create libp2p host
	h, err := libp2p.New(opts...)
	if err != nil {
		return nil, err
	}

	// add discovery server address info
	for _, addrInfo := range discoveryAddrInfo {
		h.Peerstore().AddAddrs(addrInfo.ID, addrInfo.Addrs, peerstore.PermanentAddrTTL)
	}

	// register connection tracker
	if conf.GetBool("p2p.enable_conn_notifier") {
		h.Network().Notify(newConnNotifier(h))
	}

	log.Println("host: PeerID:", h.ID())
	log.Println("host: listen addrs:", h.Addrs())

	return h, nil
}
