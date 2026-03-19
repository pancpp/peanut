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

## Issues
### UDP buffer size
If you have such error message: `sys_conn.go:62: failed to sufficiently increase receive buffer size (was: 208 kiB, wanted: 7168 kiB, got: 416 kiB). See https://github.com/quic-go/quic-go/wiki/UDP-Buffer-Sizes for details.`. Follow the instruction to increase the UDP receive buffer size.
```bash
sudo sysctl -w net.core.rmem_max=7500000
sudo sysctl -w net.core.wmem_max=7500000
sudo sysctl -p
```
