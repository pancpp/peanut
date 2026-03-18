# Peanut

A Go library for building peer-to-peer virtual networks using libp2p and TUN interfaces.

## peanut-key

Utility tool for generating P2P keys and private network PSKs.

```bash
# Generate a P2P Ed25519 private key (prints to stdout, peer ID to stderr)
go run scripts/peanut-key/main.go pkey

# Write private key to file
go run scripts/peanut-key/main.go pkey -o private.key

# Show peer ID from a private key file
go run scripts/peanut-key/main.go pid -i private.key

# Generate a private network PSK (prints to stdout)
go run scripts/peanut-key/main.go pnet

# Write PSK to file
go run scripts/peanut-key/main.go pnet -o swarm.key
```
