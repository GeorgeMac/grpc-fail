#!/usr/bin/env bash

set -euxo pipefail

# Some lightly-edited explanation from the gogo docs:
#
# Generate gogo, gRPC-Gateway and swagger output.
#
# -I declares import folders, in order of importance
# This is how proto resolves the protofile imports.
# It will check for the protofile relative to each of these
# folders and use the first one it finds.
#
# --gogo_out generates GoGo Protobuf output with gRPC plugin enabled.
# --grpc-gateway_out generates gRPC-Gateway output.
# --swagger_out generates an OpenAPI 2.0 specification for our gRPC-Gateway endpoints.
protoc \
  -I. \
  --gogo_out=plugins=grpc:. \
  server.proto
