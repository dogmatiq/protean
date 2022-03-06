# Protean

[![Build Status](https://github.com/dogmatiq/protean/workflows/CI/badge.svg)](https://github.com/dogmatiq/protean/actions?workflow=CI)
[![Code Coverage](https://img.shields.io/codecov/c/github/dogmatiq/protean/main.svg)](https://codecov.io/github/dogmatiq/protean)
[![Latest Version](https://img.shields.io/github/tag/dogmatiq/protean.svg?label=semver)](https://semver.org)
[![Documentation](https://img.shields.io/badge/go.dev-reference-007d9c)](https://pkg.go.dev/github.com/dogmatiq/protean)
[![Go Report Card](https://goreportcard.com/badge/github.com/dogmatiq/protean)](https://goreportcard.com/report/github.com/dogmatiq/protean)

Protean is a framework for building browser-facing RPC services based on
[Protocol Buffers service definitions], with full streaming support.

Protean is inspired by [Twirp](https://github.com/twitchtv/twirp) but has
different goals. Specifically, Protean is intended to produce RPC services that
are easy to use directly in the browser with standard browser APIs. It can also
be used for server-to-server communication.

Both Protean and Twirp are alternatives to [gRPC].

## Getting Started

To use Protean, you will need a working knowledge of [Protocol Buffers],
[Protocol Buffers service definitions] and [generating Go code from .proto
files][protocol buffers go]. An understanding of [gRPC] is an advantage.

Protean provides a `protoc` plugin called `protoc-gen-go-protean`, which is used
in addition to the standard `protoc-gen-go` plugin. The latest version can be
installed by running:

```
go install github.com/dogmatiq/protean/cmd/protoc-gen-go-protean@latest
```

To generate Protean server and client code, pass the `--go-protean_out` flag to
`protoc`. An example demonstrating how to implement a server and use an RPC
client is available in [example_test.go](example_test.go).

## Transports

Protean exposes APIs via HTTP/1.1 and HTTP/2 using Go's standard HTTP server.

It has complete support for all method types that can be defined in a Protocol
Buffers service:

- **unary** methods, which accept a single request from the client and
  return a single response from the server
- **client streaming** methods, which accept a stream of requests from the
  client and return a single response from the server
- **server streaming** methods, which accept a single request from the client
  and return a stream of responses from the server
- **bidirectional streaming** methods, which accept a stream of requests from
  the client and return a stream of responses from the server

Depending on what kind of method is being called and what best suits the
application, the caller can choose from any of the transports described below on
a per-call basis.

| Tranport  | Supported Methods | Suitable Browser API | Implementation |
| --------- | ----------------- | -------------------- | -------------- |
| HTTP GET  | unary             | [fetch]              | ❌             |
| HTTP POST | unary             | [fetch]              | ✅             |
| JSON-RPC  | unary             | [fetch]              | ❌             |
| SSE       | server streaming  | [server-sent events] | ❌             |
| WebSocket | all               | [websocket]          | 🚧             |

All of the above transports are made available via the same HTTP handler. The
client uses content negotiation and other similar mechanisms to choose the
desired transport.

## Encoding

Protean supports all of the standard Protocol Buffers serialization formats:

- [native binary format][protocol buffers native]
- [canonical JSON format][protocol buffers json]
- human-readable text format (undocumented)

As with transports, the encoding is chosen via content negotiation. JSON is the
default encoding, allowing simpler use from the browser.

[fetch]: https://developer.mozilla.org/en-US/docs/Web/API/Fetch_API
[grpc]: https://grpc.io/
[protocol buffers go]: https://developers.google.com/protocol-buffers/docs/reference/go-generated
[protocol buffers json]: https://developers.google.com/protocol-buffers/docs/proto3#json
[protocol buffers native]: https://developers.google.com/protocol-buffers/docs/encoding
[protocol buffers service definitions]: https://developers.google.com/protocol-buffers/docs/proto3#services
[protocol buffers]: https://developers.google.com/protocol-buffers
[server-sent events]: https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events
[twirp]: https://github.com/twitchtv/twirp
[websocket]: https://developer.mozilla.org/en-US/docs/Web/API/WebSocket
