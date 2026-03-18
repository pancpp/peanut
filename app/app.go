package app

import (
	"context"
	"encoding/base64"
	"log"
	"os"
	"strings"
	"time"

	"github.com/libp2p/go-libp2p"
	coreconnmgr "github.com/libp2p/go-libp2p/core/connmgr"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/pnet"
	"github.com/libp2p/go-libp2p/p2p/protocol/holepunch"
	quic "github.com/libp2p/go-libp2p/p2p/transport/quic"
	"github.com/pancpp/peanut/conf"
)

const (
	P2P_DIAL_TIMEOUT = 10 * time.Second
)

var (
	gTunIface         *TunIface
	gConnGater        coreconnmgr.ConnectionGater
	gStaticRelayAddrs []peer.AddrInfo
	gHost             host.Host
)

func Init(ctx context.Context) error {
	// init static relay addresses
	if err := initStaticRelayAddrs(); err != nil {
		return err
	}

	// init connection gater
	if err := initConnGater(); err != nil {
		return err
	}

	// init p2p host
	if err := initHost(ctx); err != nil {
		return err
	}

	// // init tun interface
	// if err := initTun(ctx); err != nil {
	// 	return err
	// }

	// // set streamer handlers
	// gHost.SetStreamHandler(PROTOCOL_TUN, gTunIface.handleTunService)

	return nil
}

func initStaticRelayAddrs() error {
	var staticRelayAddrInfo []peer.AddrInfo
	for _, addrStr := range conf.GetStringSlice("p2p.relay_multiaddrs") {
		addrInfo, err := peer.AddrInfoFromString(addrStr)
		if err != nil {
			return err
		}
		staticRelayAddrInfo = append(staticRelayAddrInfo, *addrInfo)
	}

	gStaticRelayAddrs = staticRelayAddrInfo

	return nil
}

func initConnGater() error {
	connGater, err := newConnGater()
	if err != nil {
		return err
	}

	gConnGater = connGater

	return nil
}

func initHost(_ context.Context) error {
	// libp2p host options
	var opts []libp2p.Option

	// option: private key
	privateKeyPath := conf.GetString("p2p.private_key_path")
	privateKeyB64, err := os.ReadFile(privateKeyPath)
	if err != nil {
		log.Printf("reading private key err: %v, path: %s", err, privateKeyPath)
		return err
	}
	privateKeyBytes, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(privateKeyB64)))
	if err != nil {
		log.Printf("base64 unmarshal err: %v, string: %s", err, string(privateKeyB64))
		return err
	}
	privateKey, err := crypto.UnmarshalPrivateKey(privateKeyBytes)
	if err != nil {
		log.Printf("invalid private key, err: %v, string: %s", err, string(privateKeyBytes))
		return err
	}
	opts = append(opts, libp2p.Identity(privateKey))

	// option: private network PSK
	pskPath := conf.GetString("p2p.pnet_psk_path")
	if pskPath != "" {
		pskFile, err := os.Open(pskPath)
		if err != nil {
			return err
		}
		defer pskFile.Close()

		psk, err := pnet.DecodeV1PSK(pskFile)
		if err != nil {
			return err
		}

		opts = append(opts, libp2p.PrivateNetwork(psk))
		log.Println("private network is enabled")
	}

	// option: listening addresses
	listenAddrs := conf.GetStringSlice("p2p.listen_multiaddrs")
	opts = append(opts,
		libp2p.Transport((quic.NewTransport)),
		libp2p.ListenAddrStrings(listenAddrs...))

	// option: NAT service
	opts = append(opts, libp2p.EnableNATService())

	// option: Attempt to open ports using uPNP for NATed hosts
	opts = append(opts, libp2p.NATPortMap())

	// option: force reachability private
	opts = append(opts, libp2p.ForceReachabilityPrivate())

	// option: static relays
	opts = append(opts, libp2p.EnableAutoRelayWithStaticRelays(gStaticRelayAddrs))

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
		return err
	}

	// register connection tracker
	if conf.GetBool("p2p.enable_conn_notifier") {
		h.Network().Notify(&connNotifyer{})
	}

	gHost = h

	return nil
}

func initTun(ctx context.Context) error {
	name := conf.GetString("tun_name")
	tunIface, err := NewTunIface(name, DEFAULT_TUN_MTU, DEFAULT_TUN_TIMEOUT)
	if err != nil {
		return err
	}
	go tunIface.Run(ctx)

	gTunIface = tunIface

	return nil
}
