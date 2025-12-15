# API Examples

Base URL: `http://localhost:8080`

All protected endpoints require `X-Api-Key` header.

---

## 🔓 Public Endpoints

### Health Check
```bash
curl http://localhost:8080/api/health
```

**Response:** `200 OK`

---

## 🔒 Protected Endpoints

### 1. Create User on All Nodes

Creates a user with auto-generated UUID on all active nodes in parallel.

**Request:**
```bash
curl -X POST http://localhost:8080/api/create \
  -H "X-Api-Key: vpnbot_rPHKly0DoS0pJbEWXCibiPXBkZ0sgVsQsvMav35zwkKSQOuUYBV7TiGghQVkgyao"
```

**Response:**
```json
{
  "uuid": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "configs": [
    {
      "countryCode": "DE",
      "config": "trojan://password123@84.201.14.167:443?type=raw&security=reality&fp=chrome&sni=www.google.com&pbk=xxx&sid=yyy&spx=zzz#Config:a1b2c3d4-e5f6-7890-abcd-ef1234567890"
    },
    {
      "countryCode": "US",
      "config": "vless://a1b2c3d4-e5f6-7890-abcd-ef1234567890@85.202.15.168:8443?flow=xtls-rprx-vision&type=tcp&security=reality&fp=chrome&sni=cloudflare.com&pbk=aaa&sid=bbb&spx=ccc#XTLS-Reality:a1b2c3d4-e5f6-7890-abcd-ef1234567890"
    }
  ]
}
```

---

### 2. Delete User from All Nodes

Deletes user from all nodes where it was created (tracked in database).

**Request:**
```bash
curl -X DELETE http://localhost:8080/api/deleteUser \
  -H "X-Api-Key: vpnbot_rPHKly0DoS0pJbEWXCibiPXBkZ0sgVsQsvMav35zwkKSQOuUYBV7TiGghQVkgyao" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
  }'
```

**Response:** `200 OK`

---

### 3. Get Servers

Returns list of active servers.

**Request:**
```bash
curl http://localhost:8080/api/servers \
  -H "X-Api-Key: vpnbot_rPHKly0DoS0pJbEWXCibiPXBkZ0sgVsQsvMav35zwkKSQOuUYBV7TiGghQVkgyao"
```

**Response:**
```json
[
  {
    "countryCode": "DE",
    "cityName": "Frankfurt",
    "extName": "FRA-1",
    "endpoint": "84.201.14.167",
    "inboundType": "trojan",
    "active": true,
    "createdAt": "2024-01-01T12:00:00Z"
  },
  {
    "countryCode": "US",
    "cityName": "New York",
    "extName": "NYC-1",
    "endpoint": "85.202.15.168",
    "inboundType": "vless",
    "active": true,
    "createdAt": "2024-01-01T12:05:00Z"
  }
]
```

---

### 4. Add/Replace Servers

Creates new servers or replaces existing ones (by endpoint).

**Request:**
```bash
curl -X POST http://localhost:8080/api/servers \
  -H "X-Api-Key: vpnbot_rPHKly0DoS0pJbEWXCibiPXBkZ0sgVsQsvMav35zwkKSQOuUYBV7TiGghQVkgyao" \
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
      },
      {
        "countryCode": "GB",
        "cityName": "London",
        "extName": "LON-1",
        "endpoint": "86.203.16.169",
        "inboundType": "vless",
        "active": true
      }
    ]
  }'
```

**Response:** `201 Created`

**Note:** `inboundType` defaults to `"trojan"` if not specified.

---

### 5. Update Servers

Updates existing servers.

**Request:**
```bash
curl -X PUT http://localhost:8080/api/servers \
  -H "X-Api-Key: vpnbot_rPHKly0DoS0pJbEWXCibiPXBkZ0sgVsQsvMav35zwkKSQOuUYBV7TiGghQVkgyao" \
  -H "Content-Type: application/json" \
  -d '{
    "servers": [
      {
        "countryCode": "DE",
        "cityName": "Frankfurt",
        "extName": "FRA-2",
        "endpoint": "84.201.14.167",
        "inboundType": "vless",
        "active": false
      }
    ]
  }'
```

**Response:** `200 OK`

---

### 6. Delete Servers

Deletes servers by endpoint.

**Request:**
```bash
curl -X DELETE http://localhost:8080/api/servers \
  -H "X-Api-Key: vpnbot_rPHKly0DoS0pJbEWXCibiPXBkZ0sgVsQsvMav35zwkKSQOuUYBV7TiGghQVkgyao" \
  -H "Content-Type: application/json" \
  -d '{
    "servers": [
      {
        "endpoint": "84.201.14.167"
      },
      {
        "endpoint": "85.202.15.168"
      }
    ]
  }'
```

**Response:** `200 OK`

---

### 7. Get Users from Node

Returns list of users from a specific node (for monitoring/debug).

**Request:**
```bash
curl -X POST http://localhost:8080/api/getUsers \
  -H "X-Api-Key: vpnbot_rPHKly0DoS0pJbEWXCibiPXBkZ0sgVsQsvMav35zwkKSQOuUYBV7TiGghQVkgyao" \
  -H "Content-Type: application/json" \
  -d '{
    "endpoint": "84.201.14.167"
  }'
```

**Response:**
```json
[
  {
    "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "inbound": "trojan"
  },
  {
    "id": "b2c3d4e5-f6a7-8901-bcde-f12345678901",
    "inbound": "vless"
  }
]
```

---

### 8. Get Inbound Configs from Node

Returns inbound configuration details from a specific node.

**Request:**
```bash
curl -X POST http://localhost:8080/api/getInbounds \
  -H "X-Api-Key: vpnbot_rPHKly0DoS0pJbEWXCibiPXBkZ0sgVsQsvMav35zwkKSQOuUYBV7TiGghQVkgyao" \
  -H "Content-Type: application/json" \
  -d '{
    "endpoint": "84.201.14.167"
  }'
```

**Response:**
```json
{
  "inbounds": [
    {
      "inbound": "trojan",
      "connection_config": {
        "level": 0,
        "password": "generated-password",
        "tcp": {
          "port": 443,
          "reality": {
            "fingerprint": "chrome",
            "private_key": "xxx",
            "public_key": "yyy",
            "server_name": "www.google.com",
            "short_id": "zzz",
            "spider_x": "/"
          }
        }
      }
    }
  ]
}
```

---

### 9. Get Statistics

Returns comprehensive statistics about all nodes and users.

**Request:**
```bash
curl http://localhost:8080/api/stats \
  -H "X-Api-Key: vpnbot_rPHKly0DoS0pJbEWXCibiPXBkZ0sgVsQsvMav35zwkKSQOuUYBV7TiGghQVkgyao"
```

**Response:**
```json
{
  "totalUsers": 150,
  "nodes": [
    {
      "endpoint": "84.201.14.167",
      "countryCode": "DE",
      "cityName": "Frankfurt",
      "extName": "FRA-1",
      "inboundType": "trojan",
      "active": true,
      "usersCount": 45
    },
    {
      "endpoint": "85.202.15.168",
      "countryCode": "US",
      "cityName": "New York",
      "extName": "NYC-1",
      "inboundType": "vless",
      "active": true,
      "usersCount": 52
    },
    {
      "endpoint": "86.203.16.169",
      "countryCode": "GB",
      "cityName": "London",
      "extName": "LON-1",
      "inboundType": "trojan",
      "active": false,
      "usersCount": 0
    }
  ],
  "byProtocol": {
    "trojan": 89,
    "vless": 61
  }
}
```

---

### 10. Get All Users

Returns all users with detailed information about nodes where they were created.

**Request:**
```bash
curl http://localhost:8080/api/users \
  -H "X-Api-Key: vpnbot_rPHKly0DoS0pJbEWXCibiPXBkZ0sgVsQsvMav35zwkKSQOuUYBV7TiGghQVkgyao"
```

**Response:**
```json
{
  "users": [
    {
      "userId": "724d0470-f308-40c6-b7fc-941fa348f56c",
      "nodesCount": 6,
      "createdAt": "2024-12-15T00:50:23Z",
      "nodes": [
        {
          "endpoint": "84.201.14.167",
          "countryCode": "DE",
          "cityName": "Frankfurt",
          "inbound": "trojan",
          "createdAt": "2024-12-15T00:50:23Z"
        },
        {
          "endpoint": "87.120.244.189",
          "countryCode": "NL",
          "cityName": "Amsterdam",
          "inbound": "trojan",
          "createdAt": "2024-12-15T00:50:23Z"
        }
      ]
    },
    {
      "userId": "846cd8fb-2719-4316-a633-4b2f1d569ba8",
      "nodesCount": 6,
      "createdAt": "2024-12-14T23:45:12Z",
      "nodes": [...]
    }
  ]
}
```

---

### 11. Get Users Count

Returns total count of unique users.

**Request:**
```bash
curl http://localhost:8080/api/users/count \
  -H "X-Api-Key: vpnbot_rPHKly0DoS0pJbEWXCibiPXBkZ0sgVsQsvMav35zwkKSQOuUYBV7TiGghQVkgyao"
```

**Response:**
```json
{
  "count": 150
}
```

---

## 🔒 Authentication

All protected endpoints require `X-Api-Key` header:

```bash
-H "X-Api-Key: your-api-key-here"
```

**401 Forbidden** response if key is missing or invalid:
```json
{
  "error": "Forbidden"
}
```

---

## ⚠️ Error Responses

### 400 Bad Request
Invalid request body or missing parameters.
```json
{
  "error": "Invalid request body"
}
```

### 503 Service Unavailable
All nodes failed to respond.
```json
{
  "error": "all nodes failed: 10/10 nodes unavailable"
}
```

---

## 📊 Response Codes

| Code | Meaning |
|------|---------|
| 200 | Success |
| 201 | Created |
| 400 | Bad Request |
| 403 | Forbidden (invalid API key) |
| 500 | Internal Server Error |
| 503 | Service Unavailable (all nodes down) |

---

## 🧪 Testing Script

Save as `test.sh`:

```bash
#!/bin/bash

API_KEY="vpnbot_rPHKly0DoS0pJbEWXCibiPXBkZ0sgVsQsvMav35zwkKSQOuUYBV7TiGghQVkgyao"
BASE_URL="http://localhost:8080"

echo "1. Health check..."
curl -s "$BASE_URL/api/health"
echo -e "\n"

echo "2. Adding servers..."
curl -s -X POST "$BASE_URL/api/servers" \
  -H "X-Api-Key: $API_KEY" \
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
echo -e "\n"

echo "3. Creating user..."
USER_RESPONSE=$(curl -s -X POST "$BASE_URL/api/create" \
  -H "X-Api-Key: $API_KEY")
echo $USER_RESPONSE | jq '.'

USER_ID=$(echo $USER_RESPONSE | jq -r '.uuid')
echo "User ID: $USER_ID"
echo ""

echo "4. Getting servers..."
curl -s "$BASE_URL/api/servers" \
  -H "X-Api-Key: $API_KEY" | jq '.'
echo ""

echo "5. Deleting user..."
curl -s -X DELETE "$BASE_URL/api/deleteUser" \
  -H "X-Api-Key: $API_KEY" \
  -H "Content-Type: application/json" \
  -d "{\"id\": \"$USER_ID\"}"
echo -e "\n"
```

Make it executable and run:
```bash
chmod +x test.sh
./test.sh
```

---

## 10. Get Pending Deletions

View users that failed to be deleted from specific nodes (for manual cleanup).

**GET** `/api/cleanup/pending`

### Request

```bash
curl http://localhost:8080/api/cleanup/pending \
  -H "X-Api-Key: your-api-key-here"
```

### Response

```json
{
  "count": 2,
  "pendingDeletions": [
    {
      "userId": "724d0470-f308-40c6-b7fc-941fa348f56c",
      "endpoint": "158.160.87.175",
      "inbound": "trojan",
      "attempts": 3,
      "lastAttempt": "2024-01-15T14:30:00Z",
      "createdAt": "2024-01-15T10:00:00Z",
      "errorMessage": "connection timeout"
    },
    {
      "userId": "846cd8fb-2719-4316-a633-4b2f1d569ba8",
      "endpoint": "84.201.14.167",
      "inbound": "trojan",
      "attempts": 1,
      "lastAttempt": "2024-01-15T14:00:00Z",
      "createdAt": "2024-01-15T14:00:00Z",
      "errorMessage": "node unavailable"
    }
  ]
}
```

---

## 11. Execute Manual Cleanup

Manually retry deletion of users from nodes that previously failed. Recommended to run weekly.

**POST** `/api/cleanup`

### Request

```bash
curl -X POST http://localhost:8080/api/cleanup \
  -H "X-Api-Key: your-api-key-here"
```

### Response (Success)

```json
{
  "totalAttempted": 5,
  "successful": 3,
  "failed": 2,
  "stillPending": 2,
  "errors": [
    "user=724d0470-f308-40c6-b7fc-941fa348f56c, endpoint=158.160.87.175, inbound=trojan: connection timeout",
    "user=846cd8fb-2719-4316-a633-4b2f1d569ba8, endpoint=84.201.14.167, inbound=trojan: node unavailable"
  ]
}
```

### Response (No Pending Deletions)

```json
{
  "totalAttempted": 0,
  "successful": 0,
  "failed": 0,
  "stillPending": 0,
  "errors": []
}
```

---

## 12. Delete Specific Pending Deletion

Manually remove a specific pending deletion record (useful when a node is permanently dead).

**DELETE** `/api/cleanup/pending`

### Request (with specific inbound)

```bash
curl -X DELETE http://localhost:8080/api/cleanup/pending \
  -H "X-Api-Key: your-api-key-here" \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "724d0470-f308-40c6-b7fc-941fa348f56c",
    "endpoint": "158.160.87.175",
    "inbound": "trojan"
  }'
```

### Request (delete all inbounds for user+endpoint)

```bash
curl -X DELETE http://localhost:8080/api/cleanup/pending \
  -H "X-Api-Key: your-api-key-here" \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "724d0470-f308-40c6-b7fc-941fa348f56c",
    "endpoint": "158.160.87.175"
  }'
```

### Response

```
OK (HTTP 200)
```

### Use Cases

1. **Node permanently dead**: Remove all pending deletions for that node
2. **User recreated**: Clear old pending deletions before recreating
3. **False positives**: Remove records that shouldn't be retried

---

## Recommended Cleanup Schedule

It's recommended to run the manual cleanup **once per week** via cron or manually:

```bash
#!/bin/bash
# weekly-cleanup.sh

API_KEY="your-api-key-here"
BASE_URL="http://localhost:8080"

echo "Running weekly cleanup..."
curl -X POST "$BASE_URL/api/cleanup" \
  -H "X-Api-Key: $API_KEY" | jq '.'
```

Add to crontab (every Sunday at 3 AM):
```bash
0 3 * * 0 /path/to/weekly-cleanup.sh
```

