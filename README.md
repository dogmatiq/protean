# Protean

[![Build Status](https://github.com/dogmatiq/protean/workflows/CI/badge.svg)](https://github.com/dogmatiq/protean/actions?workflow=CI)
[![Code Coverage](https://img.shields.io/codecov/c/github/dogmatiq/protean/main.svg)](https://codecov.io/github/dogmatiq/protean)
[![Latest Version](https://img.shields.io/github/tag/dogmatiq/protean.svg?label=semver)](https://semver.org)
[![Documentation](https://img.shields.io/badge/go.dev-reference-007d9c)](https://pkg.go.dev/github.com/dogmatiq/protean)
[![Go Report Card](https://goreportcard.com/badge/github.com/dogmatiq/protean)](https://goreportcard.com/report/github.com/dogmatiq/protean)

Protean is a framework for building RPC web services based on [Protocol Buffers
service definitions](https://developers.google.com/protocol-buffers/docs/proto3#services).

Protean is similar to [gRPC](https://grpc.io/) and draws inspiration from
[Twirp](https://github.com/twitchtv/twirp).

## Getting Started

To use Protean, you will need a working knowledge of [Protocol
Buffers](https://grpc.io/docs/protoc-installation/), and [generating Go code from .proto files](https://developers.google.com/protocol-buffers/docs/reference/go-generated).

Protean provides a `protoc` plugin called `protoc-gen-go-protean`. The latest
version can be installed by running:

```
go install github.com/dogmatiq/protean/cmd/protoc-gen-go-protean@latest
```

Add the `--go-protean_out` to `protoc` to generate Protean server interfaces and
RPC clients. An example demonstrating how to implement a server and use an RPC
client [example_test.go](example_test.go).

## Project Goals

### Goals

- Services to be consumable by web browsers using standard browser APIs.
- Services to be equally easy to consume from other servers in any language.
- Full support for client, server and bidirectional streaming.
- Provide an adequate level of cache control for use with service workers.
- Allow the client to choose the best encoding (protobuf, json or text) on a
  per-call basis.
- Allow the client to choose the best transport on a per-call basis. Options
  to include "conventional" HTTP GET and POST requests, websockets, server-sent
  events (SSE) and JSON-RPC.
- Produce [`http.Handler`](https://pkg.go.dev/net/http#Handler) implementations
  that work with Go's standard web server.
- Co-exist with gRPC services built from the same Protocol Buffers definitions.
- HTTP/1.1 and HTTP/2 support.

### Non-goals

- Hiding the fact that the server is RPC based.
- Providing RESTful APIs, changing behavior based on HTTP methods.
- Allowing fine-grained control over HTTP-specific behavior, such as setting headers.
- Non-HTTP transports.