# Changelog

All notable changes to vpnbot-core-go will be documented in this file.

## [1.0.0] - 2024-12-14

### 🎉 Initial Release

Complete rewrite from Swift/Vapor to Go/Fiber with major improvements.

### ✨ Features

- **Multi-node user creation**: Automatically creates users on all active nodes in parallel
- **Best-effort strategy**: Returns configs from available nodes, skips failed ones
- **Auto UUID generation**: No need to provide UUID in request
- **Protocol support**: Trojan and VLESS with Reality
- **Server-side inbound type**: Configure protocol per server (trojan/vless/vmess/shadowsocks)
- **User tracking**: Database tracks which users are on which nodes
- **Smart deletion**: Deletes users from all nodes automatically
- **Simplified API**: Cleaner request/response format

### 🏗️ Architecture

- **Fast & efficient**: Built with Fiber framework and fasthttp client
- **Parallel processing**: Goroutines for concurrent node requests
- **SQLite database**: Lightweight storage with WAL mode
- **Structured logging**: JSON logs with slog
- **Graceful shutdown**: Proper cleanup on termination

### 📡 API Endpoints

#### User Operations
- `POST /api/create` - Create user on all nodes (auto-generates UUID)
- `DELETE /api/deleteUser` - Delete user from all tracked nodes
- `POST /api/getUsers` - Get users from specific node
- `POST /api/getInbounds` - Get inbound configs from node

#### Server Management
- `GET /api/servers` - List active servers
- `POST /api/servers` - Create/replace servers
- `PUT /api/servers` - Update servers
- `DELETE /api/servers` - Delete servers

#### Health
- `GET /api/health` - Health check

### 🔄 Breaking Changes from Swift Version

1. **UUID generation**: Now server-side (don't send in request)
2. **Create response format**: Changed to `{uuid, configs}` structure
3. **Delete endpoint**: No longer requires `endpoint` parameter
4. **Multi-node by default**: Creates on all active nodes automatically

### 🔧 Configuration

All via environment variables:
- `PORT` - Server port (default: 8080)
- `API_KEY_REQUESTS` - API authentication key
- `XRAY_NODE_TOKEN` - Token for xray-service nodes
- `DB_PATH` - SQLite database path
- `REQUEST_TIMEOUT` - Overall request timeout (default: 10s)
- `NODE_TIMEOUT` - Per-node timeout (default: 3s)
- `LOG_LEVEL` - Logging level (debug/info/warn/error)

### 📦 Deployment

- Docker & Docker Compose support
- Systemd service example
- Migration guide from Swift version
- Health check endpoint for monitoring

### 🔐 Security

- API key middleware for protected endpoints
- Bearer token authentication to xray nodes
- Proper error handling without leaking internals

### 📊 Database Schema

**servers:**
- `country_code`, `city_name`, `ext_name`
- `endpoint` (PRIMARY KEY)
- `inbound_type` (trojan/vless/vmess/shadowsocks)
- `active` (boolean)
- `created_at`

**user_nodes:**
- `user_id`, `endpoint`, `inbound` (composite PRIMARY KEY)
- `created_at`

### 🐛 Known Issues

None currently. This is a stable initial release.

### 🎯 Future Enhancements

Potential features for future versions:
- [ ] Background health checks for nodes
- [ ] Prometheus metrics export
- [ ] Rate limiting per API key
- [ ] WebSocket support for real-time updates
- [ ] Admin dashboard UI
- [ ] User statistics and analytics
- [ ] Automatic node failover
- [ ] Config caching for faster responses

---

## Migration from Swift Version

See [DEPLOYMENT.md](DEPLOYMENT.md) for detailed migration guide.

**Quick summary:**
1. Export servers from Swift version
2. Transform JSON format (remove `id`, add `inboundType`)
3. Start Go version
4. Import servers via POST /api/servers
5. Update client applications to use new API format
6. Stop Swift version after verification

---

## Backward Compatibility

While the Go version introduces breaking changes to the API, it can be deployed alongside the Swift version during migration. Both versions can coexist temporarily.

---

## Performance

Based on testing with 10 nodes:
- **Create user**: ~500-700ms (all nodes responding)
- **Delete user**: ~300-500ms
- **Get servers**: <10ms (SQLite query)
- **Memory usage**: ~15-20MB idle
- **Binary size**: ~15MB (static build)

---

## Credits

- Original Swift version: vpnbot-core
- Go rewrite: Complete reimplementation with architectural improvements
- Framework: [Fiber](https://gofiber.io/) - Express-inspired web framework
- HTTP Client: [fasthttp](https://github.com/valyala/fasthttp) - Fast HTTP implementation
- Database: SQLite with [go-sqlite3](https://github.com/mattn/go-sqlite3)

