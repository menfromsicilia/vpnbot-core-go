# Deployment Guide

## 📦 Prerequisites

- Go 1.22+ (for building from source)
- Docker & Docker Compose (for containerized deployment)
- SQLite support (CGO_ENABLED=1)

---

## 🚀 Deployment Options

### Option 1: Docker (Recommended)

#### 1. Clone and configure
```bash
cd /path/to/vpnbot-core-go
cp .env.example .env
# Edit .env with your API keys
```

#### 2. Build and run
```bash
docker-compose up -d
```

#### 3. Check logs
```bash
docker-compose logs -f
```

#### 4. Stop
```bash
docker-compose down
```

---

### Option 2: Build from Source

#### 1. Install Go 1.22+
```bash
# macOS
brew install go

# Ubuntu/Debian
wget https://go.dev/dl/go1.22.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
```

#### 2. Build binary
```bash
cd vpnbot-core-go
go mod download
make build-prod
```

#### 3. Setup environment
```bash
export API_KEY_REQUESTS="your-api-key"
export XRAY_NODE_TOKEN="your-xray-token"
export DB_PATH="/var/lib/vpnbot/vpnbot.db"
```

#### 4. Run
```bash
./vpnbot-core
```

---

### Option 3: Systemd Service

#### 1. Create service file
```bash
sudo nano /etc/systemd/system/vpnbot-core.service
```

```ini
[Unit]
Description=VPN Bot Core API
After=network.target

[Service]
Type=simple
User=vpnbot
WorkingDirectory=/opt/vpnbot-core
ExecStart=/opt/vpnbot-core/vpnbot-core
Restart=always
RestartSec=5

Environment="API_KEY_REQUESTS=your-api-key"
Environment="XRAY_NODE_TOKEN=your-xray-token"
Environment="DB_PATH=/var/lib/vpnbot/vpnbot.db"
Environment="PORT=8080"

[Install]
WantedBy=multi-user.target
```

#### 2. Create user and directories
```bash
sudo useradd -r -s /bin/false vpnbot
sudo mkdir -p /opt/vpnbot-core
sudo mkdir -p /var/lib/vpnbot
sudo chown vpnbot:vpnbot /var/lib/vpnbot
```

#### 3. Copy binary and migrations
```bash
sudo cp vpnbot-core /opt/vpnbot-core/
sudo cp -r migrations /opt/vpnbot-core/
sudo chown -R vpnbot:vpnbot /opt/vpnbot-core
```

#### 4. Enable and start
```bash
sudo systemctl daemon-reload
sudo systemctl enable vpnbot-core
sudo systemctl start vpnbot-core
sudo systemctl status vpnbot-core
```

#### 5. View logs
```bash
sudo journalctl -u vpnbot-core -f
```

---

## 🔄 Migration from Swift Version

### Prerequisites
1. Both services can run in parallel (on different ports if needed)
2. Backup current server list from Swift version

### Step-by-Step Migration

#### 1. Export servers from Swift version
```bash
# Get current servers
curl http://localhost:8080/api/servers \
  -H "X-Api-Key: your-api-key" > servers_backup.json
```

#### 2. Transform data format

The Swift version stores servers in UserDefaults. If you need to migrate:

**Swift format:**
```json
[
  {
    "id": "some-uuid",
    "countryCode": "DE",
    "cityName": "Frankfurt",
    "extName": "FRA-1",
    "endpoint": "84.201.14.167",
    "active": true
  }
]
```

**Go format** (remove `id`, add `inboundType`):
```json
{
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
}
```

Python script to transform:
```python
#!/usr/bin/env python3
import json
import sys

with open('servers_backup.json', 'r') as f:
    swift_servers = json.load(f)

go_servers = {
    "servers": [
        {
            "countryCode": s["countryCode"],
            "cityName": s["cityName"],
            "extName": s.get("extName", ""),
            "endpoint": s["endpoint"],
            "inboundType": "trojan",  # Set default
            "active": s["active"]
        }
        for s in swift_servers
    ]
}

with open('servers_go.json', 'w') as f:
    json.dump(go_servers, f, indent=2)

print("Transformed servers saved to servers_go.json")
```

#### 3. Start Go version
```bash
docker-compose up -d
```

#### 4. Import servers
```bash
curl -X POST http://localhost:8080/api/servers \
  -H "X-Api-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d @servers_go.json
```

#### 5. Test with new user creation
```bash
curl -X POST http://localhost:8080/api/create \
  -H "X-Api-Key: your-api-key"
```

#### 6. Update client applications
Change API calls from:
```
POST /api/create {"endpoint": "...", "id": "..."}
```
To:
```
POST /api/create {}
```

Use `uuid` and `configs` from response.

#### 7. Monitor for issues
```bash
docker-compose logs -f
```

#### 8. Stop Swift version
Once confirmed Go version works:
```bash
# Stop Swift container
docker stop vpnbot
```

---

## 🔧 Configuration

### Environment Variables

Create `.env` file:
```bash
PORT=8080
API_KEY_REQUESTS=vpnbot_rPHKly0DoS0pJbEWXCibiPXBkZ0sgVsQsvMav35zwkKSQOuUYBV7TiGghQVkgyao
XRAY_NODE_TOKEN=BSgFahXxzrL5yZcNo5wa4696QXAEfR2Ljf1MQKyq8V2KW6Fr4W6Xws98NvrQYeDp
DB_PATH=./vpnbot.db
REQUEST_TIMEOUT=10s
NODE_TIMEOUT=3s
LOG_LEVEL=info
```

### Timeout Tuning

If you have slow/distant nodes:
```bash
# Increase timeouts
REQUEST_TIMEOUT=30s  # Overall request timeout
NODE_TIMEOUT=10s     # Per-node timeout
```

For fast local nodes:
```bash
# Decrease for faster failures
REQUEST_TIMEOUT=5s
NODE_TIMEOUT=2s
```

---

## 📊 Monitoring

### Health Check
```bash
curl http://localhost:8080/api/health
```

### Logs
```bash
# Docker
docker-compose logs -f

# Systemd
sudo journalctl -u vpnbot-core -f

# JSON parsing
docker-compose logs -f | jq '.'
```

### Database Stats
```bash
sqlite3 vpnbot.db << EOF
SELECT COUNT(*) as total_servers FROM servers;
SELECT COUNT(*) as active_servers FROM servers WHERE active = 1;
SELECT COUNT(DISTINCT user_id) as total_users FROM user_nodes;
EOF
```

---

## 🔐 Security

### 1. API Key Management
```bash
# Generate secure API key
openssl rand -hex 32

# Update .env
API_KEY_REQUESTS=generated-key-here
```

### 2. Firewall
```bash
# Only allow access from trusted IPs
sudo ufw allow from 10.0.0.0/8 to any port 8080
sudo ufw enable
```

### 3. Reverse Proxy (Nginx)
```nginx
server {
    listen 443 ssl http2;
    server_name vpnbot.example.com;

    ssl_certificate /etc/letsencrypt/live/vpnbot.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/vpnbot.example.com/privkey.pem;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
```

---

## 🐛 Troubleshooting

### "command not found: go"
Install Go or use Docker deployment.

### "all nodes failed"
- Check node endpoints are reachable
- Verify XRAY_NODE_TOKEN is correct
- Check node logs: `docker logs xray-service-container`

### "failed to open database"
- Check DB_PATH directory exists and is writable
- For Docker: ensure volume is mounted correctly

### "Forbidden" (403)
- Verify X-Api-Key header is correct
- Check API_KEY_REQUESTS environment variable

### Database locked
```bash
# Check for stale locks
rm vpnbot.db-shm vpnbot.db-wal

# Restart service
docker-compose restart
```

---

## 📈 Performance Tuning

### For High Load (100+ nodes)
```go
// Increase connection pool in nodeclient/client.go
MaxConnsPerHost: 50

// Increase timeout
REQUEST_TIMEOUT=30s
```

### For Low Latency
```go
// Decrease timeouts
NODE_TIMEOUT=1s
REQUEST_TIMEOUT=5s
```

### Database Optimization
```bash
sqlite3 vpnbot.db << EOF
PRAGMA journal_mode = WAL;
PRAGMA synchronous = NORMAL;
PRAGMA cache_size = 20000;
PRAGMA temp_store = MEMORY;
VACUUM;
ANALYZE;
EOF
```

---

## 🔄 Backup & Restore

### Backup Database
```bash
# Stop service
docker-compose down

# Backup
cp vpnbot.db vpnbot.db.backup.$(date +%Y%m%d)

# Or use SQLite backup
sqlite3 vpnbot.db ".backup vpnbot.db.backup"

# Restart
docker-compose up -d
```

### Restore Database
```bash
docker-compose down
cp vpnbot.db.backup vpnbot.db
docker-compose up -d
```

---

## 📞 Support

For issues or questions:
1. Check logs first
2. Verify configuration
3. Test with simple curl commands
4. Review xray-service node status

Happy deploying! 🚀

