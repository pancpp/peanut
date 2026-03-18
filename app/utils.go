package app

import (
	"encoding/base64"
	mrand "math/rand"
	"net"
	"net/netip"
	"time"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/songgao/water/waterutil"
)

func IP4ToAddr(ip net.IP) netip.Addr {
	ip = ip.To4()
	return netip.AddrFrom4([4]byte{ip[0], ip[1], ip[2], ip[3]})
}

func IPAddrTo4(addr netip.Addr) net.IP {
	b := addr.As4()
	return net.IPv4(b[0], b[1], b[2], b[3])
}

func IPv4SourceAddr(b []byte) netip.Addr {
	srcIP := waterutil.IPv4Source(b)
	return IP4ToAddr(srcIP)
}

func IPv4DestAddr(b []byte) netip.Addr {
	destIP := waterutil.IPv4Destination(b)
	return IP4ToAddr(destIP)
}

func GenEd25519Key(seed int64) (crypto.PrivKey, error) {
	cryptoRand := mrand.New(mrand.NewSource(seed))
	privKey, _, err := crypto.GenerateEd25519Key(cryptoRand)
	if err != nil {
		return nil, err
	}

	return privKey, nil
}

func GenRandRSAKey(bits int) (crypto.PrivKey, error) {
	cryptoRand := mrand.New(mrand.NewSource(time.Now().UnixMicro()))
	privKey, _, err := crypto.GenerateRSAKeyPair(bits, cryptoRand)
	if err != nil {
		return nil, err
	}

	return privKey, nil
}

func EncodePrivKey(privKey crypto.PrivKey) (string, error) {
	privKeyBytes, err := crypto.MarshalPrivateKey(privKey)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(privKeyBytes), nil
}

func DecodePrivKey(privKeyB64 string) (crypto.PrivKey, error) {
	privKeyBytes, err := base64.StdEncoding.DecodeString(privKeyB64)
	if err != nil {
		return nil, err
	}

	return crypto.UnmarshalPrivateKey(privKeyBytes)
}

func PeerIDFromPrivKey(privKey crypto.PrivKey) (peer.ID, error) {
	return peer.IDFromPrivateKey(privKey)
}
