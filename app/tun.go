package app

import (
	"context"
	"log"
	"time"

	"github.com/songgao/water"
	"github.com/vishvananda/netlink"
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

	return &TunIface{
		mtu:       mtu,
		timeout:   timeout,
		iface:     iface,
		readChan:  make(chan []byte),
		writeChan: make(chan []byte),
	}, nil
}

func (t *TunIface) Close() error {
	if link, err := netlink.LinkByName(t.iface.Name()); err == nil {
		netlink.LinkSetDown(link)
	}
	return t.iface.Close()
}

func (t *TunIface) Start(ctx context.Context) {
	go t.read(ctx)
	go t.write(ctx)
}

func (t *TunIface) ReplaceIPAddr(ipCIDR string) error {
	// Parse IP addr
	ipAddr, err := netlink.ParseAddr(ipCIDR)
	if err != nil {
		log.Printf("(peanut) parse addr %s err: %v", ipCIDR, err)
		return err
	}

	// Get link device
	link, err := netlink.LinkByName(t.iface.Name())
	if err != nil {
		log.Printf("(peanut) get link by name %s err: %v", t.iface.Name(), err)
		return err
	}

	// Replace IP address
	if err := netlink.AddrReplace(link, ipAddr); err != nil {
		return err
	}

	return nil
}

func (t *TunIface) DeleteIPAddr(ipCIDR string) error {
	// Parse IP addr
	ipAddr, err := netlink.ParseAddr(ipCIDR)
	if err != nil {
		log.Printf("(peanut) parse addr %s err: %v", ipCIDR, err)
		return err
	}

	// Get link device
	link, err := netlink.LinkByName(t.iface.Name())
	if err != nil {
		log.Printf("(peanut) get link by name %s err: %v", t.iface.Name(), err)
		return err
	}

	// Delete IP address
	if err := netlink.AddrDel(link, ipAddr); err != nil {
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
