package app

import (
	"log"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/multiformats/go-multiaddr"
)

type connNotifyer struct{}

func (c *connNotifyer) Connected(_ network.Network, conn network.Conn) {
	log.Printf("ConnNotifier: connected to peer %v", conn.RemotePeer())
}

func (c *connNotifyer) Disconnected(_ network.Network, conn network.Conn) {
	pid := conn.RemotePeer()
	log.Printf("ConnNotifier: disconnected from peer %v", pid)

	// // notify static relay to retry connection
	// for _, ai := range gStaticRelayAddrs {
	// 	if ai.ID == pid {
	// 		em, err := gHost.EventBus().Emitter(new(RelayEvent))
	// 		if err != nil {
	// 			return
	// 		}
	// 		defer em.Close()

	// 		em.Emit(RelayEvent{eventType: RELAY_EVENT_TYPE_DISCONNCTED, ai: ai})

	// 		return
	// 	}
	// }
}

func (c *connNotifyer) Listen(_ network.Network, addrs multiaddr.Multiaddr) {
	log.Println("ConnNotifier: listen addr:", addrs)
}

func (c *connNotifyer) ListenClose(_ network.Network, addrs multiaddr.Multiaddr) {
	log.Println("ConnNotifier: close listen addr:", addrs)
}
