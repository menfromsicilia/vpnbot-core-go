# Getting Started with VPN Bot Core (Go)

Welcome! This is your new VPN management API, rewritten from Swift to Go with significant improvements.

---

## 🎯 What's New?

### Major Improvements Over Swift Version

1. **Multi-Node Creation**: Creates users on ALL active nodes automatically (not just one)
2. **Best-Effort**: Returns configs from available nodes, skips failed ones
3. **Auto UUID**: Generates UUID server-side (simpler client code)
4. **Better Performance**: Parallel goroutines for faster node communication
5. **Protocol Flexibility**: Configure trojan/vless per server
6. **Smart Deletion**: Deletes from all tracked nodes automatically

---

## ⚡ Quick Start (5 minutes)

### Prerequisites
- Docker & Docker Compose installed
- Access to xray-service nodes

### Step 1: Configure Environment

```bash
cd vpnbot-core-go

# Copy environment template
cp .env.example .env

# Edit with your values
nano .env
```

**Required values:**
```bash
API_KEY_REQUESTS=your-secure-api-key-here
XRAY_NODE_TOKEN=your-xray-token-here
```

### Step 2: Start Service

```bash
docker-compose up -d
```

### Step 3: Verify It's Running

```bash
# Health check
curl http://localhost:8080/api/health

# Should return: 200 OK
```

### Step 4: Add Your Servers

```bash
curl -X POST http://localhost:8080/api/servers \
  -H "X-Api-Key: your-secure-api-key-here" \
  -H "Content-Type: application/json" \
  -d '{
    "servers": [
      {
        "countryCode": "DE",
        "cityName": "Frankfurt",
        "endpoint": "84.201.14.167",
        "inboundType": "trojan",
        "active": true
      }
    ]
  }'
```

### Step 5: Create Your First User

```bash
curl -X POST http://localhost:8080/api/create \
  -H "X-Api-Key: your-secure-api-key-here"
```

**Response:**
```json
{
  "uuid": "generated-uuid",
  "configs": [
    {
      "countryCode": "DE",
      "config": "trojan://password@84.201.14.167:443?..."
    }
  ]
}
```

🎉 **Done!** You now have a working VPN config.

---

## 📱 Update Your Client Application

### Old API (Swift version):
```javascript
// Old way - had to provide endpoint and UUID
POST /api/create
{
  "endpoint": "84.201.14.167",
  "id": "manually-generated-uuid"
}

// Response: {tcp, ws, reality} - all same string
```

### New API (Go version):
```javascript
// New way - simpler!
POST /api/create
{
  // Empty body - UUID generated server-side
}

// Response:
{
  "uuid": "auto-generated-uuid",
  "configs": [
    {
      "countryCode": "DE",
      "config": "trojan://..."  // Use this field (was "ws" in old version)
    },
    {
      "countryCode": "US",
      "config": "vless://..."
    }
  ]
}
```

### Client Code Update

**Before (Swift API):**
```javascript
const userID = generateUUID();
const response = await fetch('/api/create', {
  method: 'POST',
  headers: {
    'X-Api-Key': apiKey,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    endpoint: selectedServer.endpoint,
    id: userID
  })
});

const { ws } = await response.json();
// ws contains single config string
```

**After (Go API):**
```javascript
const response = await fetch('/api/create', {
  method: 'POST',
  headers: {
    'X-Api-Key': apiKey
  }
});

const { uuid, configs } = await response.json();
// uuid: auto-generated user ID
// configs: array of {countryCode, config} for all nodes
```

---

## 🔄 Migration from Swift Version

### Option A: Parallel Deployment (Recommended)

Run both services simultaneously during transition:

```bash
# Swift version on port 8080
# Go version on port 8081

# In docker-compose.yml, change:
PORT=8081
```

Update clients gradually, then stop Swift version.

### Option B: Direct Replacement

1. **Export servers** from Swift version
2. **Stop Swift** service
3. **Start Go** version
4. **Import servers**
5. **Update clients**

Detailed guide: [DEPLOYMENT.md](DEPLOYMENT.md)

---

## 🏗️ Project Structure

```
vpnbot-core-go/
├── cmd/server/          # Entry point (main.go)
├── internal/
│   ├── api/            # HTTP handlers & routing
│   ├── config/         # Configuration loading
│   ├── middleware/     # API key authentication
│   ├── models/         # Data structures
│   ├── nodeclient/     # HTTP client for xray nodes
│   ├── repository/     # Database layer (SQLite)
│   └── service/        # Business logic (orchestration)
├── migrations/         # SQL schema
├── .env               # Configuration (create from .env.example)
├── docker-compose.yml # Docker deployment
├── Dockerfile         # Container image
└── README.md          # Full documentation
```

---

## 📖 Documentation

- **[README.md](README.md)** - Full project documentation
- **[API_EXAMPLES.md](API_EXAMPLES.md)** - API usage examples
- **[DEPLOYMENT.md](DEPLOYMENT.md)** - Deployment & migration guide
- **[CHANGELOG.md](CHANGELOG.md)** - Version history

---

## 🐛 Troubleshooting

### "All nodes failed"
- Check nodes are running: `curl http://node-ip:8000/user`
- Verify XRAY_NODE_TOKEN is correct
- Check network connectivity

### "Forbidden" (403)
- Verify X-Api-Key header matches API_KEY_REQUESTS

### "No active servers"
- Add servers via POST /api/servers
- Check servers are marked `active: true`

### Database issues
```bash
# Restart service
docker-compose restart

# Check logs
docker-compose logs -f
```

---

## 🎓 Common Tasks

### View Logs
```bash
docker-compose logs -f
```

### Add More Servers
```bash
curl -X POST http://localhost:8080/api/servers \
  -H "X-Api-Key: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "servers": [
      {"countryCode": "US", "cityName": "New York", "endpoint": "1.2.3.4", "inboundType": "vless", "active": true},
      {"countryCode": "GB", "cityName": "London", "endpoint": "5.6.7.8", "inboundType": "trojan", "active": true}
    ]
  }'
```

### List All Servers
```bash
curl http://localhost:8080/api/servers \
  -H "X-Api-Key: $API_KEY"
```

### Delete User
```bash
curl -X DELETE http://localhost:8080/api/deleteUser \
  -H "X-Api-Key: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"id": "user-uuid"}'
```

### Disable Server (Without Deleting)
```bash
curl -X PUT http://localhost:8080/api/servers \
  -H "X-Api-Key: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "servers": [
      {"endpoint": "84.201.14.167", "active": false}
    ]
  }'
```

### Check Database Stats
```bash
sqlite3 vpnbot.db << EOF
SELECT COUNT(*) as total_servers FROM servers;
SELECT COUNT(DISTINCT user_id) as total_users FROM user_nodes;
EOF
```

---

## 🚀 Production Checklist

- [ ] Configure secure API_KEY_REQUESTS
- [ ] Set appropriate timeouts for your network
- [ ] Setup log monitoring
- [ ] Configure firewall rules
- [ ] Enable HTTPS with reverse proxy (Nginx/Caddy)
- [ ] Setup database backups
- [ ] Test with all target nodes
- [ ] Update client applications
- [ ] Monitor logs after deployment

---

## 💡 Tips

1. **Start with 1-2 nodes** to test, then add more
2. **Use LOG_LEVEL=debug** initially for troubleshooting
3. **Increase NODE_TIMEOUT** if nodes are geographically distant
4. **Monitor logs** for the first few hours after deployment
5. **Backup database** before major changes

---

## 📞 Need Help?

1. Check logs: `docker-compose logs -f`
2. Verify configuration in `.env`
3. Test nodes directly: `curl http://node-ip:8000/user -H "Authorization: Bearer $TOKEN"`
4. Review [API_EXAMPLES.md](API_EXAMPLES.md) for correct request format

---

## 🎉 Success Criteria

You'll know it's working when:
- ✅ Health check returns 200
- ✅ Can add/list servers
- ✅ Create user returns configs array
- ✅ Configs work in VPN client
- ✅ Logs show successful node communication

---

**Welcome to the new VPN Bot Core!** 🚀

This version is faster, more reliable, and easier to maintain. Enjoy! 🎊

