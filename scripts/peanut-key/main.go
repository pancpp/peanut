package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: peanut-key <command> [options]\n\nCommands:\n  pkey  Generate a P2P Ed25519 private key\n  pnet  Generate a private network PSK\n  pid   Show peer ID from a private key file\n\nOptions:\n  -o string\n    \toutput file path (default: stdout)\n  -i string\n    \tinput file path\n")
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	cmd := os.Args[1]

	fs := flag.NewFlagSet(cmd, flag.ExitOnError)
	outputPath := fs.String("o", "", "output file path (default: stdout)")
	inputPath := fs.String("i", "", "input file path")
	fs.Parse(os.Args[2:])

	switch cmd {
	case "pkey":
		genPrivateKey(*outputPath)
	case "pnet":
		genPNetPSK(*outputPath)
	case "pid":
		showPeerID(*inputPath)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		usage()
		os.Exit(1)
	}
}

func genPrivateKey(outputPath string) {
	priv, pub, err := crypto.GenerateKeyPair(crypto.Ed25519, -1)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to generate key pair: %v\n", err)
		os.Exit(1)
	}

	id, err := peer.IDFromPublicKey(pub)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to derive peer ID: %v\n", err)
		os.Exit(1)
	}

	privBytes, err := crypto.MarshalPrivateKey(priv)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to marshal private key: %v\n", err)
		os.Exit(1)
	}

	encoded := base64.StdEncoding.EncodeToString(privBytes)

	if outputPath != "" {
		if err := os.WriteFile(outputPath, []byte(encoded+"\n"), 0600); err != nil {
			fmt.Fprintf(os.Stderr, "failed to write key file: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Key written to %s\n", outputPath)
	} else {
		fmt.Println(encoded)
	}

	fmt.Fprintf(os.Stderr, "Peer ID: %s\n", id.String())
}

func showPeerID(inputPath string) {
	if inputPath == "" {
		fmt.Fprintf(os.Stderr, "Usage: peanut-key pid -i <private_key_file>\n")
		os.Exit(1)
	}

	data, err := os.ReadFile(inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read key file: %v\n", err)
		os.Exit(1)
	}

	privBytes, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(data)))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to decode key: %v\n", err)
		os.Exit(1)
	}

	priv, err := crypto.UnmarshalPrivateKey(privBytes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to unmarshal private key: %v\n", err)
		os.Exit(1)
	}

	id, err := peer.IDFromPublicKey(priv.GetPublic())
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to derive peer ID: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(id.String())
}

func genPNetPSK(outputPath string) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		fmt.Fprintf(os.Stderr, "failed to generate PSK: %v\n", err)
		os.Exit(1)
	}

	content := fmt.Sprintf("/key/swarm/psk/1.0.0/\n/base16/\n%s\n", hex.EncodeToString(key))

	if outputPath != "" {
		if err := os.WriteFile(outputPath, []byte(content), 0600); err != nil {
			fmt.Fprintf(os.Stderr, "failed to write key file: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "PSK written to %s\n", outputPath)
	} else {
		fmt.Print(content)
	}
}
