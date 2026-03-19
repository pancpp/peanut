package app

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/songgao/water"
	"github.com/vishvananda/netlink"
)

const (
	IP_RULE_PRIORITY = 6000
)

type TunIface struct {
	mtu       int
	timeout   time.Duration
	iface     *water.Interface
	readChan  chan []byte
	writeChan chan []byte
}

func newTunIface(name string, mtu int, timeout time.Duration) (*TunIface, error) {
	// Create tun interface
	tunconf := water.Config{
		DeviceType: water.TUN,
		PlatformSpecificParams: water.PlatformSpecificParams{
			Name:       name,
			Persist:    false,
			MultiQueue: false,
		},
	}
	iface, err := water.New(tunconf)
	if err != nil {
		return nil, err
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

	// add IP rule
	rule := netlink.NewRule()
	rule.Table = 52
	rule.Priority = IP_RULE_PRIORITY
	if exists, err := ruleExists(rule); err != nil {
		return nil, err
	} else if !exists {
		if err := netlink.RuleAdd(rule); err != nil {
			return nil, err
		}
	}

	return &TunIface{
		mtu:       mtu,
		timeout:   timeout,
		iface:     iface,
		readChan:  make(chan []byte),
		writeChan: make(chan []byte),
	}, nil
}

func ruleExists(rule *netlink.Rule) (bool, error) {
	rules, err := netlink.RuleList(netlink.FAMILY_ALL)
	if err != nil {
		return false, err
	}

	for _, r := range rules {
		if r.Table == rule.Table && r.Priority == rule.Priority {
			return true, nil
		}
	}

	return false, nil
}

func (t *TunIface) Close() error {
	// delete tun
	if link, err := netlink.LinkByName(t.iface.Name()); err == nil {
		netlink.LinkSetDown(link)
	}

	// delete ip rule
	rule := netlink.NewRule()
	rule.Table = 52
	rule.Priority = IP_RULE_PRIORITY
	netlink.RuleDel(rule)

	return t.iface.Close()
}

func (t *TunIface) Start(ctx context.Context) {
	go t.read(ctx)
	go t.write(ctx)
}

func (t *TunIface) ReplaceIPAddr(ipNet net.IPNet) error {
	// Get link device
	link, err := netlink.LinkByName(t.iface.Name())
	if err != nil {
		log.Printf("(peanut) get link by name %s err: %v", t.iface.Name(), err)
		return err
	}

	// Replace IP address
	if err := netlink.AddrReplace(link, &netlink.Addr{IPNet: &ipNet}); err != nil {
		return err
	}

	return nil
}

func (t *TunIface) DeleteIPAddr(ipNet net.IPNet) error {
	// Get link device
	link, err := netlink.LinkByName(t.iface.Name())
	if err != nil {
		log.Printf("(peanut) get link by name %s err: %v", t.iface.Name(), err)
		return err
	}

	// Delete IP address
	if err := netlink.AddrDel(link, &netlink.Addr{IPNet: &ipNet}); err != nil {
		return err
	}

	return nil
}

func (t *TunIface) ReplaceRoute(ipNet net.IPNet) error {
	// Get link device
	link, err := netlink.LinkByName(t.iface.Name())
	if err != nil {
		log.Printf("(peanut) get link by name %s err: %v", t.iface.Name(), err)
		return err
	}

	// Create route
	route := netlink.Route{
		LinkIndex: link.Attrs().Index,
		Dst:       &ipNet,
		Table:     52,
		Scope:     netlink.SCOPE_UNIVERSE,
	}

	// Replace route
	if err := netlink.RouteReplace(&route); err != nil {
		return err
	}

	return nil
}

func (t *TunIface) DeleteRoute(ipNet net.IPNet) error {
	// Get link device
	link, err := netlink.LinkByName(t.iface.Name())
	if err != nil {
		log.Printf("(peanut) get link by name %s err: %v", t.iface.Name(), err)
		return err
	}

	// Create route
	route := netlink.Route{
		LinkIndex: link.Attrs().Index,
		Dst:       &ipNet,
		Table:     52,
	}

	// Replace route
	if err := netlink.RouteDel(&route); err != nil {
		return err
	}

	return nil
}

func (t *TunIface) Name() string {
	return t.iface.Name()
}

func (t *TunIface) MTU() int {
	return t.mtu
}

func (t *TunIface) GetReadTunChan() <-chan []byte {
	return t.readChan
}

func (t *TunIface) GetWriteTunChan() chan []byte {
	return t.writeChan
}

func (t *TunIface) read(ctx context.Context) {
	for {
		b := make([]byte, t.mtu)
		n, err := t.iface.Read(b)
		if err != nil {
			if ctx.Err() != nil {
				return
			} else {
				continue
			}
		}

		select {
		case <-ctx.Done():
			return

		case t.readChan <- b[:n]:
		}
	}
}

func (t *TunIface) write(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return

		case b := <-t.writeChan:
			t.iface.Write(b)
		}
	}
}
