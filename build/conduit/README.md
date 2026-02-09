# Conduit Binary Management

This directory contains scripts for downloading and managing the Conduit Matrix homeserver binary.

## Overview

Conduit is a lightweight Matrix homeserver written in Rust. Ayo uses Conduit for inter-agent communication via the Matrix protocol.

## Usage

### Download Conduit

```bash
# Download for current platform
./download.sh

# Download for specific platform
./download.sh --platform linux-amd64
./download.sh --platform darwin-arm64

# Specify output directory
./download.sh --output /path/to/bin
```

### Supported Platforms

| Platform | Description |
|----------|-------------|
| `linux-amd64` | Linux x86_64 |
| `linux-arm64` | Linux ARM64 (aarch64) |
| `darwin-amd64` | macOS Intel |
| `darwin-arm64` | macOS Apple Silicon |

## Manual Installation

If the download script fails, you can install Conduit manually:

### Using Cargo (Rust)

```bash
cargo install conduit
```

### From Source

```bash
git clone https://gitlab.com/famedly/conduit.git
cd conduit
cargo build --release
cp target/release/conduit ~/.local/share/ayo/bin/
```

### From Releases

Download the appropriate binary from:
https://gitlab.com/famedly/conduit/-/releases

## Version

Current pinned version: **0.8.0**

Update the `CONDUIT_VERSION` in `download.sh` and the checksums in `checksums.txt` when upgrading.

## Checksums

The `checksums.txt` file contains SHA256 checksums for verifying downloaded binaries. Always verify the checksum before trusting a downloaded binary.

## Integration with Ayo

Ayo looks for the Conduit binary in the following locations (in order):

1. Embedded in the ayo binary (release builds with `embed_conduit` tag)
2. `~/.local/share/ayo/bin/conduit`
3. `conduit` in PATH

The daemon will attempt to download Conduit automatically if not found (development mode only).
