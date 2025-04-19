#!/usr/bin/env bash
set -euo pipefail

# Project root
ROOT_DIR=$(cd "$(dirname "$0")/.." && pwd)

# Paths
PROTO_DIR="$ROOT_DIR/pkg/api/protos"
BIN_DIR="$ROOT_DIR/bin"
CMD_PATH="$ROOT_DIR/cmd/spdktgt-svr"

# Ensure output and bin directories exist
mkdir -p "$PROTO_DIR" "$BIN_DIR"

# 1. Generate gRPC code from .proto
protoc \
  --proto_path="$PROTO_DIR" \
  --go_out=paths=source_relative:"$PROTO_DIR" \
  --go-grpc_out=paths=source_relative:"$PROTO_DIR" \
  "$PROTO_DIR"/spdk.proto

# 2. Build the main executable
go build -o "$BIN_DIR/spdktgt-svr" "$CMD_PATH"

echo "✅ gRPC code generated and spdktgt-svr built at $BIN_DIR/spdktgt-svr"