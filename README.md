# VPN Bot Core - Go

REST API service for managing VPN nodes and user configurations.

## Features

- **Multi-node management**: Automatically creates users on all active nodes in parallel
- **Best-effort strategy**: Returns configs from available nodes, skips failed ones
- **Protocol support**: Trojan and VLESS with Reality
- **SQLite database**: Lightweight storage for servers and user tracking
- **Fast & efficient**: Built with Fiber and fasthttp

## Architecture

```
┌─────────────────────────────────┐
│  Fiber HTTP Server (8080)       │
│  - API Key authentication       │
└────────────┬────────────────────┘
             │
┌────────────▼────────────────────┐
│  Service Layer                  │
│  - Parallel node orchestration  │
│  - Config generation            │
└────┬────────────────────┬───────┘
     │                    │
┌────▼──────┐    ┌────────▼───────┐
│ SQLite DB │    │ Node Client    │
│           │    │ (fasthttp)     │
└───────────┘    └────────────────┘
```

## API Endpoints

### Public
- `GET /api/health` - Health check

### Protected (require X-Api-Key header)

#### User Operations
- `POST /api/create` - Create user on all nodes
- `DELETE /api/deleteUser` - Delete user from all nodes
- `POST /api/getUsers` - Get users from specific node
- `POST /api/getInbounds` - Get inbound configs from node

#### Server Management
- `GET /api/servers` - List active servers
- `POST /api/servers` - Create/replace servers
- `PUT /api/servers` - Update servers
- `DELETE /api/servers` - Delete servers

#### Statistics & Monitoring
- `GET /api/stats` - Comprehensive statistics (all nodes + users)
- `GET /api/users` - List all users with details
- `GET /api/users/count` - Total users count
- `GET /api/nodes/users` - All nodes with their users

#### Cleanup & Reliability
- `GET /api/cleanup/pending` - View pending deletion attempts
- `POST /api/cleanup` - Execute manual cleanup (recommended weekly)
- `DELETE /api/cleanup/pending` - Manually remove specific pending deletion

## Quick Start

### 1. Setup Environment

```bash
cp .env.example .env
# Edit .env with your values
```

### 2. Run Locally

```bash
go mod download
go run cmd/server/main.go
```

### 3. Build Binary

```bash
go build -o vpnbot-core cmd/server/main.go
./vpnbot-core
```

### 4. Docker

```bash
docker build -t vpnbot-core-go .
docker run -p 8080:8080 --env-file .env vpnbot-core-go
```

## Configuration

All configuration via environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `8080` |
| `API_KEY_REQUESTS` | API key for authentication | **required** |
| `XRAY_NODE_TOKEN` | Token for xray-service nodes | **required** |
| `DB_PATH` | SQLite database path | `./vpnbot.db` |
| `REQUEST_TIMEOUT` | Overall request timeout | `10s` |
| `NODE_TIMEOUT` | Timeout per node | `3s` |
| `LOG_LEVEL` | Log level (debug/info/warn/error) | `info` |

## API Examples

### Create User

**Request:**
```bash
curl -X POST http://localhost:8080/api/create \
  -H "X-Api-Key: your-api-key"
```

**Response:**
```json
{
  "uuid": "a1b2c3d4-...",
  "configs": [
    {
      "countryCode": "DE",
      "config": "trojan://pass@1.2.3.4:443?..."
    },
    {
      "countryCode": "US",
      "config": "vless://id@5.6.7.8:8443?..."
    }
  ]
}
```

### Delete User

**Request:**
```bash
curl -X DELETE http://localhost:8080/api/deleteUser \
  -H "X-Api-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{"id": "user-uuid"}'
```

### Add Servers

**Request:**
```bash
curl -X POST http://localhost:8080/api/servers \
  -H "X-Api-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "servers": [
      {
        "countryCode": "DE",
        "cityName": "Frankfurt",
        "extName": "FRA-1",
        "endpoint": "84.201.14.167",
        "inboundType": "trojan",
        "active": true
      }
    ]
  }'
```

## Database Schema

### servers
```sql
country_code TEXT
city_name TEXT
ext_name TEXT
endpoint TEXT PRIMARY KEY
inbound_type TEXT (trojan/vless/vmess/shadowsocks)
active BOOLEAN
created_at TIMESTAMP
```

### user_nodes
```sql
user_id TEXT
endpoint TEXT
inbound TEXT
created_at TIMESTAMP
PRIMARY KEY (user_id, endpoint, inbound)
```

## Development

### Project Structure

```
.
├── cmd/server/          # Entry point
├── internal/
│   ├── api/            # HTTP handlers & routing
│   ├── config/         # Configuration
│   ├── middleware/     # API key auth
│   ├── models/         # Data structures
│   ├── nodeclient/     # HTTP client for nodes
│   ├── repository/     # Database layer
│   └── service/        # Business logic
└── migrations/         # SQL migrations
```

### Testing

```bash
# Unit tests
go test ./...

# With coverage
go test -cover ./...
```

## Logging

Structured JSON logs with slog and automatic rotation:

```json
{"level":"info","msg":"user created","user_id":"uuid","successful":8,"failed":2}
{"level":"warn","msg":"node failed","endpoint":"1.2.3.4","error":"timeout"}
```

**Two modes:**
- **stdout** (Docker) - Docker handles rotation
- **file** (systemd) - Automatic rotation with lumberjack

**Configuration:**
```bash
LOG_OUTPUT=stdout         # or "file"
LOG_MAX_SIZE=100         # MB per file
LOG_MAX_BACKUPS=3        # Old files to keep
LOG_MAX_AGE=7            # Days retention
```

See [LOGGING.md](LOGGING.md) for details.

## Migration from Swift Version

This Go version is **backward compatible** with the Swift/Vapor version. The main differences:

1. **UUID generation**: Now done server-side (no need to provide in request)
2. **Multi-node creation**: Users created on all active nodes automatically
3. **Simplified deletion**: No need to specify endpoint, deletes from all nodes
4. **Better performance**: Parallel requests with goroutines

## License

MIT

