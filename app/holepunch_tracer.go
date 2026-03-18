package app

import (
	"log"

	"github.com/libp2p/go-libp2p/p2p/protocol/holepunch"
)

type holepunchTracer struct{}

func (t *holepunchTracer) Trace(evt *holepunch.Event) {
	switch evt.Type {
	case holepunch.DirectDialEvtT:
		ev := evt.Evt.(*holepunch.DirectDialEvt)
		log.Printf("Hole-punching: event-type: %s, local: %s, remote: %s, success: %v, err: %v",
			evt.Type, evt.Peer, evt.Remote, ev.Success, ev.Error)

	case holepunch.ProtocolErrorEvtT:
		ev := evt.Evt.(*holepunch.ProtocolErrorEvt)
		log.Printf("Hole-punching: event-type: %s, local: %s, remote: %s,  err: %v",
			evt.Type, evt.Peer, evt.Remote, ev.Error)

	case holepunch.StartHolePunchEvtT:
		ev := evt.Evt.(*holepunch.StartHolePunchEvt)
		log.Printf("Hole-punching: event-type: %s, local: %s, remote: %s, remoteAddrs: %v, RTT: %v",
			evt.Type, evt.Peer, evt.Remote, ev.RemoteAddrs, ev.RTT)

	case holepunch.EndHolePunchEvtT:
		ev := evt.Evt.(*holepunch.EndHolePunchEvt)
		log.Printf("Hole-punching: event-type: %s, local: %s, remote: %s, success: %v, ellapsedTime: %v, err: %v",
			evt.Type, evt.Peer, evt.Remote, ev.Success, ev.EllapsedTime, ev.Error)

	case holepunch.HolePunchAttemptEvtT:
		ev := evt.Evt.(*holepunch.HolePunchAttemptEvt)
		log.Printf("Hole-punching: event-type: %s, local: %s, remote: %s, attempt: %v",
			evt.Type, evt.Peer, evt.Remote, ev.Attempt)

	default:
		log.Println("Hole-punching: unknown event")
	}
}
