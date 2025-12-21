# Realtime API

Realtime API service using LiveKit as the transport layer. This service provides session management for real-time audio/video communication.

## Features

- **Session Management** - Create and manage realtime sessions (`POST /v1/realtime/sessions`)
- **LiveKit Integration** - Uses LiveKit for WebRTC-based real-time communication
- **LiveKit Polling** - Automatic session state sync via LiveKit Server API polling
- **JWT Authentication** - Integrates with Keycloak for secure access
- **In-Memory Session Store** - Goroutine-based actor pattern for thread-safe session management

## Session Lifecycle

Sessions follow a state-driven lifecycle synced with LiveKit via polling (default: every 15s):

1. **Created** - Session token generated, waiting for client connection
2. **Connected** - Client joined the LiveKit room (detected via `ListRooms` API)
3. **Deleted** - Session removed when room is empty or doesn't exist

Stale sessions (created but never connected) are cleaned up after `SESSION_STALE_TTL` (default: 10m).

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/v1/realtime/sessions` | Create a new realtime session |
| `GET` | `/v1/realtime/sessions` | List all sessions for the current user |
| `GET` | `/v1/realtime/sessions/:id` | Get a specific session |
| `DELETE` | `/v1/realtime/sessions/:id` | Delete a session |

### Health Endpoints (Public)

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/` | Service info |
| `GET` | `/healthz` | Health check |
| `GET` | `/readyz` | Readiness check |

## Quick Start

### Prerequisites

- Go 1.21+
- LiveKit server (or LiveKit Cloud account)
- Keycloak (optional, for authentication)

### Configuration

Set the following environment variables (or use the global `.env` file):

```bash
# Required - LiveKit Configuration
LIVEKIT_WS_URL=wss://your-livekit-server.com
LIVEKIT_API_KEY=your-api-key
LIVEKIT_API_SECRET=your-api-secret

# Optional - Service Configuration
REALTIME_API_PORT=8186
LIVEKIT_TOKEN_TTL=24h           # LiveKit token validity (default: 24 hours)
SESSION_STALE_TTL=10m           # How long before "created" sessions are cleaned up
SESSION_CLEANUP_INTERVAL=15s    # How often to poll LiveKit and cleanup

# Optional - Authentication (uses global Keycloak config)
AUTH_ENABLED=true
ISSUER=http://localhost:8085/realms/jan
AUDIENCE=account
JWKS_URL=http://keycloak:8085/realms/jan/protocol/openid-connect/certs
```

### Running Locally

```bash
# From the service directory
cd services/realtime-api

# Run directly
make run

# Or build and run
make build
./bin/realtime-api
```

### Running with Docker Compose

```bash
# From the repository root
docker compose --profile realtime up realtime-api
```

## Usage Example

### Create a Session

```bash
curl -X POST http://localhost:8186/v1/realtime/sessions \
  -H "Authorization: Bearer <your-jwt-token>"
```

> **Note**: No request body required. Session creation uses server defaults.

### Response

```json
{
  "id": "sess_abc123def456...",
  "object": "realtime.session",
  "client_secret": {
    "value": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": 1734567890
  },
  "ws_url": "wss://your-livekit-server.com",
  "room_id": "room_xyz789...",
  "user_id": "user-uuid-from-jwt"
}
```

### Get Session (shows status)

```bash
curl http://localhost:8186/v1/realtime/sessions/sess_abc123def456 \
  -H "Authorization: Bearer <your-jwt-token>"
```

```json
{
  "id": "sess_abc123def456...",
  "object": "realtime.session",
  "ws_url": "wss://your-livekit-server.com",
  "room_id": "room_xyz789...",
  "user_id": "user-uuid-from-jwt",
  "status": "connected"
}
```

### Connect with LiveKit Client

Use the `client_secret.value` (LiveKit token) and `ws_url` to connect:

```javascript
import { Room } from 'livekit-client';

const room = new Room();
await room.connect(response.ws_url, response.client_secret.value);
```

## Architecture

```
realtime-api/
├── cmd/server/           # Application entry point
├── internal/
│   ├── config/           # Environment configuration
│   ├── domain/session/   # Business logic & models
│   ├── infrastructure/
│   │   ├── auth/         # JWT/JWKS validation
│   │   ├── livekit/      # LiveKit token generation
│   │   ├── store/        # In-memory session store
│   │   ├── logger/       # Zerolog configuration
│   │   └── observability/# OpenTelemetry setup
│   ├── interfaces/httpserver/
│   │   ├── handlers/     # HTTP handlers
│   │   ├── middlewares/  # CORS, logging
│   │   └── routes/       # Route registration
│   └── utils/            # Shared utilities
└── docs/swagger/         # OpenAPI documentation
```

## Development

```bash
# Run tests
make test

# Format code
go fmt ./...

# Generate swagger docs
make swagger

# Build
make build
```

## License

Apache 2.0
