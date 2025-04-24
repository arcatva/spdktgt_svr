#!/usr/bin/env bash
set -euo pipefail

# Project root
ROOT_DIR=$(cd "$(dirname "$0")/.." && pwd)

# Paths
PROTO_DIR="$ROOT_DIR/pkg/api/protos"
BIN_DIR="$ROOT_DIR/deb-build/usr/bin"
CMD_PATH="$ROOT_DIR/cmd/spdktgt-svr"

# 1. Generate gRPC code from .proto
protoc \
  --proto_path="$PROTO_DIR" \
  --go_out=paths=source_relative:"$PROTO_DIR" \
  --go-grpc_out=paths=source_relative:"$PROTO_DIR" \
  "$PROTO_DIR"/spdk.proto

echo "✅ gRPC code generated successfully."

# 2. Build the main executable
go build -o "$BIN_DIR/spdktgt-svr" "$CMD_PATH"

echo "✅ spdktgt-svr built successfully."

# 3. Build the Debian package
chmod 775 "$ROOT_DIR/deb-build/DEBIAN/postinst"
dpkg-deb --build "$ROOT_DIR/deb-build" "$ROOT_DIR/spdktgt-svr.deb"
echo "✅ Debian package built successfully."