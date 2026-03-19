package app

import (
	"log"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/multiformats/go-multiaddr"
)

type ConnNotifier struct {
	host host.Host
}

func newConnNotifier(host host.Host) *ConnNotifier {
	return &ConnNotifier{host: host}
}

func (c *ConnNotifier) Connected(_ network.Network, conn network.Conn) {
	log.Printf("[ConnNotifier] connected to peer %v", conn.RemotePeer())
}

func (c *ConnNotifier) Disconnected(_ network.Network, conn network.Conn) {
	pid := conn.RemotePeer()
	log.Printf("[ConnNotifier] disconnected from peer %v", pid)

}

func (c *ConnNotifier) Listen(_ network.Network, addrs multiaddr.Multiaddr) {
	log.Println("[ConnNotifier] listen addr:", addrs)
}

func (c *ConnNotifier) ListenClose(_ network.Network, addrs multiaddr.Multiaddr) {
	log.Println("[ConnNotifier] close listen addr:", addrs)
}
