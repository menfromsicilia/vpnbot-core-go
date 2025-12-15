# Logging Guide

## 📋 Overview

VPN Bot Core supports flexible logging with automatic rotation to prevent disk space issues.

---

## 🎯 Two Logging Modes

### **1. STDOUT Mode (Default, for Docker)**

Logs go to standard output, Docker handles rotation.

**Best for:**
- Docker deployments
- Kubernetes
- Cloud platforms with log aggregation

**Configuration:**
```bash
LOG_OUTPUT=stdout
```

**Docker handles rotation automatically:**
```yaml
# docker-compose.yml
logging:
  driver: "json-file"
  options:
    max-size: "10m"   # Rotate after 10MB
    max-file: "3"     # Keep 3 files
```

**View logs:**
```bash
docker-compose logs -f
docker-compose logs --tail 100
```

---

### **2. File Mode (for systemd/local)**

Logs go to file with automatic rotation via lumberjack.

**Best for:**
- Systemd services
- Bare metal servers
- Local development

**Configuration:**
```bash
LOG_OUTPUT=file
LOG_FILE=./logs/vpnbot.log
LOG_MAX_SIZE=100        # MB per file
LOG_MAX_BACKUPS=3       # Number of old files to keep
LOG_MAX_AGE=7           # Days to keep old files
LOG_COMPRESS=true       # Compress old files
```

**How it works:**
```
logs/
├── vpnbot.log           # Current log (max 100MB)
├── vpnbot.log.1         # Previous log (compressed)
├── vpnbot.log.2         # Older log (compressed)
└── vpnbot.log.3         # Oldest log (compressed)
```

When `vpnbot.log` reaches 100MB:
1. Rename to `vpnbot.log.1`
2. Compress to `vpnbot.log.1.gz`
3. Create new `vpnbot.log`
4. Delete files older than 7 days or beyond 3 backups

---

## 🎚️ Log Levels

Control verbosity:

```bash
LOG_LEVEL=debug   # All messages (very verbose)
LOG_LEVEL=info    # Normal operations (default)
LOG_LEVEL=warn    # Warnings only
LOG_LEVEL=error   # Errors only
```

**What each level logs:**

### DEBUG
```json
{"level":"debug","msg":"starting goroutine","node":"1.2.3.4"}
{"level":"debug","msg":"request completed","duration":"523ms"}
{"level":"info","msg":"user created","user_id":"uuid","successful":8}
{"level":"warn","msg":"node failed","endpoint":"1.2.3.4","error":"timeout"}
{"level":"error","msg":"database error","error":"connection lost"}
```

### INFO (recommended)
```json
{"level":"info","msg":"user created","user_id":"uuid","successful":8}
{"level":"warn","msg":"node failed","endpoint":"1.2.3.4","error":"timeout"}
{"level":"error","msg":"database error","error":"connection lost"}
```

### WARN
```json
{"level":"warn","msg":"node failed","endpoint":"1.2.3.4","error":"timeout"}
{"level":"error","msg":"database error","error":"connection lost"}
```

### ERROR
```json
{"level":"error","msg":"database error","error":"connection lost"}
```

---

## 📊 Disk Space Estimates

### File Mode

**Example: 1000 requests/day**

- Average log entry: ~200 bytes
- Logs per day: ~200KB
- With rotation (100MB, 3 backups): max **400MB** disk space

**Heavy load: 10,000 requests/day**

- Logs per day: ~2MB
- With rotation: max **400MB** disk space (same, rotation kicks in)

### Docker Mode

**Depends on Docker settings:**

```yaml
max-size: "10m"   # 10MB per file
max-file: "3"     # Keep 3 files
# Total: 30MB max
```

---

## 🛠️ Configuration Examples

### Development (verbose, no rotation needed)
```bash
LOG_LEVEL=debug
LOG_OUTPUT=stdout
```

### Production Docker (efficient)
```bash
LOG_LEVEL=info
LOG_OUTPUT=stdout
# Docker handles rotation
```

### Production Systemd (file rotation)
```bash
LOG_LEVEL=info
LOG_OUTPUT=file
LOG_FILE=/var/log/vpnbot/vpnbot.log
LOG_MAX_SIZE=100
LOG_MAX_BACKUPS=5
LOG_MAX_AGE=30
LOG_COMPRESS=true
```

### Debugging Issues (temporary)
```bash
LOG_LEVEL=debug
LOG_OUTPUT=file
LOG_FILE=./debug.log
LOG_MAX_SIZE=50
LOG_MAX_BACKUPS=1
LOG_MAX_AGE=1
LOG_COMPRESS=false  # Faster, easier to read
```

---

## 📈 Monitoring Logs

### Docker
```bash
# Follow logs
docker-compose logs -f vpnbot-core

# Last 100 lines
docker-compose logs --tail 100 vpnbot-core

# Filter by level
docker-compose logs vpnbot-core | grep '"level":"error"'

# Parse JSON
docker-compose logs vpnbot-core | jq 'select(.level=="error")'
```

### File
```bash
# Follow logs
tail -f logs/vpnbot.log

# Filter errors
grep '"level":"error"' logs/vpnbot.log

# Parse JSON
cat logs/vpnbot.log | jq 'select(.msg=="user created")'

# Count by level
cat logs/vpnbot.log | jq -r '.level' | sort | uniq -c
```

---

## 🔍 Log Analysis Examples

### Find all user creations
```bash
cat logs/vpnbot.log | jq 'select(.msg=="user created")'
```

### Find failed nodes
```bash
cat logs/vpnbot.log | jq 'select(.msg=="node failed")'
```

### Count errors by endpoint
```bash
cat logs/vpnbot.log | \
  jq -r 'select(.level=="error") | .endpoint' | \
  sort | uniq -c
```

### Show only high-impact errors
```bash
cat logs/vpnbot.log | \
  jq 'select(.level=="error" and (.msg | contains("all nodes failed")))'
```

---

## 🚨 Troubleshooting

### Logs not appearing?
```bash
# Check log level
echo $LOG_LEVEL

# Check output mode
echo $LOG_OUTPUT

# For file mode, check directory exists
ls -la logs/
```

### Logs growing too fast?
```bash
# Reduce log level
LOG_LEVEL=warn

# Decrease file size
LOG_MAX_SIZE=50

# Decrease retention
LOG_MAX_BACKUPS=2
LOG_MAX_AGE=3
```

### Need to see old logs?
```bash
# File mode: decompress
gunzip -c logs/vpnbot.log.1.gz | less

# Docker mode: not available (logs rotated away)
```

### Disk space issues?
```bash
# Check log size
du -h logs/

# Force cleanup (file mode)
rm logs/vpnbot.log.*

# Docker: prune logs
docker system prune --volumes
```

---

## 🎯 Best Practices

1. **Production**: Use `LOG_LEVEL=info` for normal operations
2. **Docker**: Let Docker handle rotation (`LOG_OUTPUT=stdout`)
3. **Systemd**: Use file rotation with compression
4. **Debugging**: Temporarily set `LOG_LEVEL=debug`, then revert
5. **Monitoring**: Set up alerts for ERROR level logs
6. **Retention**: Keep 7-30 days for troubleshooting
7. **Size**: 100MB per file is good for most cases

---

## 📊 Log Format

All logs are in JSON for easy parsing:

```json
{
  "time": "2024-12-15T10:30:45Z",
  "level": "info",
  "msg": "user created",
  "user_id": "uuid-here",
  "successful": 8,
  "failed": 2
}
```

**Fields:**
- `time` - ISO 8601 timestamp
- `level` - debug/info/warn/error
- `msg` - Human-readable message
- Additional context fields (user_id, endpoint, error, etc.)

---

## 🔗 Integration Examples

### Logrotate (alternative to lumberjack)
```bash
# /etc/logrotate.d/vpnbot
/var/log/vpnbot/vpnbot.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
    postrotate
        systemctl reload vpnbot-core
    endscript
}
```

### Prometheus Alert
```yaml
- alert: HighErrorRate
  expr: rate(vpnbot_errors_total[5m]) > 10
  annotations:
    summary: "High error rate in VPN Bot"
```

### ELK Stack
Ship logs to Elasticsearch:
```bash
# Filebeat configuration
filebeat.inputs:
  - type: log
    paths:
      - /var/log/vpnbot/*.log
    json.keys_under_root: true
```

---

Happy logging! 📝

