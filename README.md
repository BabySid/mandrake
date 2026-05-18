# Mandrake

A lightweight HTTP reverse proxy written in Go with request header/body modification and SSE streaming support. Zero third-party dependencies.

## Features

- **Reverse Proxy** — Forward HTTP requests to HTTPS upstream backends
- **Header Modification** — Add, overwrite, or delete request/response headers
- **JSON Body Modification** — Set or delete fields using dot-path notation (e.g., `data.user.name`)
- **Array Operations** — Prepend or append items to JSON arrays
- **SSE Streaming** — Real-time streaming passthrough with `flush_interval: -1` (no buffering)
- **Multi-Route** — Path-based routing to different upstream backends
- **Graceful Shutdown** — Clean shutdown on SIGINT/SIGTERM
- **Standard Library Only** — Zero third-party dependencies

## Requirements

- Go 1.22+

## Build

```bash
go build -o mandrake .
```

## Usage

```bash
./mandrake -config config.json
```

## Configuration

Example `config.json`:

```json
{
  "server": {
    "listen": ":8888",
    "read_timeout": "3600s",
    "write_timeout": "3600s"
  },
  "routes": [
    {
      "path": "/",
      "upstream": "https://api.example.com",
      "flush_interval": -1,
      "transport": {
        "tls_skip_verify": false,
        "read_timeout": "3600s",
        "write_timeout": "3600s"
      },
      "request": {
        "headers": {
          "set": {
            "Host": "api.example.com",
            "User-Agent": "Go-http-client/1.1"
          },
          "delete": ["X-Unwanted-Header"]
        },
        "body": {
          "set": {
            "model": "claude-sonnet-4-20250514"
          },
          "delete": ["internal_field"],
          "prepend": {
            "system": {
              "type": "text",
              "text": "prepended item"
            }
          },
          "append": {
            "system": {
              "type": "text",
              "text": "appended item"
            }
          }
        }
      },
      "response": {
        "headers": {
          "set": {"X-Proxy": "mandrake"},
          "delete": ["Server"]
        }
      }
    }
  ],
  "log": {
    "level": "info"
  }
}
```

### Configuration Reference

| Field | Description |
|-------|-------------|
| `server.listen` | Address to listen on (default: `:8080`) |
| `server.read_timeout` | Server read timeout (Go duration string) |
| `server.write_timeout` | Server write timeout (Go duration string) |
| `routes[].path` | URL path pattern for routing |
| `routes[].upstream` | Backend URL (scheme + host) |
| `routes[].flush_interval` | `-1` = flush immediately (SSE), `0` = default |
| `routes[].transport.tls_skip_verify` | Skip TLS certificate verification |
| `routes[].request.headers.set` | Headers to add/overwrite on request |
| `routes[].request.headers.delete` | Headers to remove from request |
| `routes[].request.body.set` | JSON fields to set (dot-path supported) |
| `routes[].request.body.delete` | JSON fields to delete (dot-path supported) |
| `routes[].request.body.prepend` | Items to prepend to JSON arrays |
| `routes[].request.body.append` | Items to append to JSON arrays |
| `routes[].response.headers.set` | Headers to add/overwrite on response |
| `routes[].response.headers.delete` | Headers to remove from response |
| `log.level` | Log level: `debug`, `info`, `warn`, `error` |

### Body Modification

**Dot-path notation** allows nested field access:

- `"model": "new-value"` — sets top-level field
- `"data.user.name": "value"` — sets nested field, auto-creates intermediate objects

**Array operations:**

- `prepend` — inserts item at the beginning of an array
- `append` — adds item to the end of an array

### Logging

Set `log.level` to `debug` to see modified headers and body (pretty-printed JSON) for each request.

## Testing

```bash
go test -v ./...
go test -race ./...
```

## License

BSD 3-Clause License. See [LICENSE](LICENSE) for details.
