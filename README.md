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

## Goals

- Full support for client, server and bidirectional streaming.
- Services should be consumable by web browsers using standard browser APIs.
- Services should be equally easy to consume from other servers in any language.
- Provide adequate level of cache control for use with service workers.
- Allow the client to choose the best encoding and transport on a per-call
  basis. Options include "conventional" HTTP GET and POST requests, websockets,
  server-sent events and JSON-RPC.
- Produce [`http.Handler`](https://pkg.go.dev/net/http#Handler) implementations
  that work with Go's standard web server.
- Co-exist with gRPC services built from the same Protocol Buffers definitions.

## Non-goals

- Hiding the fact that the server is RPC based.
- Providing RESTful APIs, changing behavior based on HTTP methods.
- Allowing fine-grained control over HTTP-specific behavior, such as setting headers.

