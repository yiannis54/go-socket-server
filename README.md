# go-socket-server

A real-time notification server built in Go using WebSockets and gRPC. Clients connect over WebSocket to receive notifications, while backend services push notifications through gRPC. Supports broadcasting to all clients, targeting specific rooms, and sending private messages to individual users.

## Architecture

```
┌──────────────┐         gRPC          ┌─────────────────────┐        WebSocket        ┌─────────────┐
│              │ ────────────────────> │                     │ ───────────────────────>│             │
│   Backend    │  Broadcast / Notify   │  go-socket-server   │   Real-time messages    │   Browser   │
│   Services   │                       │     Server          │<────────────────────────│   Clients   │
│              │                       │                     │   Subscribe to rooms    │             │
└──────────────┘                       └─────────────────────┘                         └─────────────┘
```

Backend services send notifications via **gRPC**. The server routes them through a central **Hub** and delivers them to connected **WebSocket** clients in real time.

## Features

- **Broadcast** — Send a notification to every connected client.
- **Room notifications** — Send to all clients subscribed to a specific room.
- **Private notifications** — Send to a single user by ID.
- **Room subscriptions** — Clients can join and leave rooms dynamically over their WebSocket connection.
- **Graceful shutdown** — Coordinated shutdown of HTTP, gRPC, and the hub via `errgroup`.

## Project Structure

```
.
├── cmd/
│   └── main.go                  # Application entry point
├── internal/
│   ├── app/
│   │   ├── grpc.go              # gRPC service implementation
│   │   └── run.go               # HTTP server, gRPC server, hub orchestration
│   ├── config/
│   │   └── config.go            # Environment-based configuration
│   ├── middleware/
│   │   └── authsocket.go        # WebSocket authentication middleware
│   ├── notifications/
│   │   └── client.go            # In-process notification client
│   └── sockets/
│       ├── client.go            # WebSocket client (read/write pumps)
│       ├── hub.go               # Central hub for routing messages
│       ├── message.go           # Message type definitions
│       ├── messagetype.go       # Proto enum to string mapping
│       └── sockets.go           # WebSocket upgrade handler
├── notificationspb/
│   ├── message.proto            # Protobuf/gRPC service definitions
│   ├── message.pb.go            # Generated protobuf code
│   └── message_grpc.pb.go       # Generated gRPC code
├── .env.dist                    # Environment variable template
├── go.mod
└── go.sum
```

## Getting Started

### Prerequisites

- Go 1.23+
- (Optional) `protoc` with Go plugins if you need to regenerate protobuf files

### Configuration

Copy the environment template and set the variables:

```bash
cp .env.dist .env
```

| Variable    | Description                              | Default |
|-------------|------------------------------------------|---------|
| `TOKEN_KEY` | Query parameter name used for auth token | `t`     |
| `GRPC_PORT` | Port for the gRPC server                 | `9003`  |
| `HTTP_PORT` | Port for the HTTP/WebSocket server       | `3003`  |

### Run

```bash
export TOKEN_KEY=t GRPC_PORT=9003 HTTP_PORT=3003
go run ./cmd/main.go
```

### Run Tests

```bash
go test ./...
```

## Usage

### WebSocket Client

Connect to the WebSocket endpoint:

```
ws://localhost:3003/ws?t=<token>
```

Once connected, subscribe to a room by sending:

```json
{ "action": "enter", "room": "order-updates" }
```

Leave a room:

```json
{ "action": "leave", "room": "order-updates" }
```

Notifications are pushed to the client as JSON messages.

### gRPC — Sending Notifications

The server exposes a `NotificationService` with three RPCs:

| RPC              | Description                                 |
|------------------|---------------------------------------------|
| `Broadcast`      | Send a message to all connected clients     |
| `NotifyRoom`     | Send a message to all clients in a room     |
| `PrivateNotify`  | Send a message to a specific user by ID     |

See `notificationspb/message.proto` for the full service and message definitions.

### Example: Broadcast via gRPC (using grpcurl)

```bash
grpcurl -plaintext -d '{
  "type": "TYPE_INFO",
  "entityId": "order-123",
  "message": {"@type": "type.googleapis.com/google.protobuf.StringValue", "value": "Your order has shipped!"}
}' localhost:9003 notificationspb.NotificationService/Broadcast
```

## Tech Stack

- **Go** — Application language
- **gorilla/websocket** — WebSocket connections
- **gRPC + Protocol Buffers** — Backend-to-server communication
- **errgroup** — Concurrent goroutine lifecycle management

## Notes

- The authentication middleware (`validateToken`) is a stub that always accepts connections. Replace it with real token validation before using in production.
- User ID extraction from the auth token is not yet implemented — private notifications require this to work correctly.

## License

This project is provided as-is for educational and demonstration purposes.
