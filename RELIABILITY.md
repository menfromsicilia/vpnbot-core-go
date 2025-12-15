# Reliability & Error Handling

## 🔒 **Удаление пользователей при недоступных нодах**

### **Текущее поведение (Best-effort + Tracking)**

Когда вы удаляете пользователя:

```
DELETE /api/deleteUser {"id": "uuid"}
  ↓
1. Получаем список нод где был создан (из user_nodes)
2. Параллельно шлем DELETE на каждую ноду
3. Если нода отвечает → OK ✅
   - Удаляем из user_nodes
4. Если нода НЕ отвечает → Warning в логах ⚠️
   - Записываем в pending_deletions
   - НЕ удаляем из user_nodes
5. Возвращаем OK (best-effort)
```

### **Возможная проблема и решение**

**Если нода была недоступна во время удаления:**
- ❌ Пользователь остался на недоступной ноде
- ✅ Но записан в `pending_deletions` для повторной попытки
- ⚠️ НЕ удален из `user_nodes` (сохраняется связь)

**Когда это происходит:**
- Нода перезагружается
- Сетевые проблемы
- Нода выключена

**Что происходит потом:**
- Раз в неделю вы запускаете `POST /api/cleanup`
- Система пытается удалить из всех pending нод
- Если успешно → удаляет из `pending_deletions` и `user_nodes`
- Если снова fail → обновляет счетчик попыток

---

## 🧹 **Ручная очистка (Weekly Cleanup)**

### **1. Проверить что нужно почистить**

```bash
curl http://localhost:8080/api/cleanup/pending \
  -H "X-Api-Key: your-api-key-here" | jq '.'
```

**Ответ:**
```json
{
  "count": 3,
  "pendingDeletions": [
    {
      "userId": "724d0470-f308-40c6-b7fc-941fa348f56c",
      "endpoint": "158.160.87.175",
      "inbound": "trojan",
      "attempts": 3,
      "lastAttempt": "2024-01-15T14:30:00Z",
      "createdAt": "2024-01-15T10:00:00Z",
      "errorMessage": "connection timeout"
    }
  ]
}
```

### **2. Запустить cleanup**

```bash
curl -X POST http://localhost:8080/api/cleanup \
  -H "X-Api-Key: your-api-key-here" | jq '.'
```

### **3. Удалить конкретную запись вручную (опционально)**

Если нода мертва навсегда или запись не должна быть повторена:

```bash
# Удалить конкретную запись
curl -X DELETE http://localhost:8080/api/cleanup/pending \
  -H "X-Api-Key: your-api-key-here" \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "724d0470-f308-40c6-b7fc-941fa348f56c",
    "endpoint": "158.160.87.175",
    "inbound": "trojan"
  }'

# Или удалить все для user+endpoint (если несколько inbound)
curl -X DELETE http://localhost:8080/api/cleanup/pending \
  -H "X-Api-Key: your-api-key-here" \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "724d0470-f308-40c6-b7fc-941fa348f56c",
    "endpoint": "158.160.87.175"
  }'
```

**Ответ:**
```json
{
  "totalAttempted": 3,
  "successful": 2,
  "failed": 1,
  "stillPending": 1,
  "errors": [
    "user=724d0470-f308-40c6-b7fc-941fa348f56c, endpoint=158.160.87.175, inbound=trojan: connection timeout"
  ]
}
```

---

## 🤖 **Автоматизация (Рекомендуется)**

### **Еженедельный cleanup (каждое воскресенье)**

Создайте скрипт:

```bash
#!/bin/bash
# weekly-cleanup.sh

API_KEY="your-api-key-here"
BASE_URL="http://localhost:8080"

echo "=== Weekly VPN Cleanup ==="
echo "Date: $(date)"
echo ""

echo "1. Checking pending deletions..."
PENDING=$(curl -s "$BASE_URL/api/cleanup/pending" -H "X-Api-Key: $API_KEY")
COUNT=$(echo $PENDING | jq -r '.count')
echo "Found $COUNT pending deletions"
echo ""

if [ "$COUNT" -gt 0 ]; then
    echo "2. Running cleanup..."
    RESULT=$(curl -s -X POST "$BASE_URL/api/cleanup" -H "X-Api-Key: $API_KEY")
    echo $RESULT | jq '.'
    echo ""
    
    echo "3. Summary:"
    echo "- Total attempted: $(echo $RESULT | jq -r '.totalAttempted')"
    echo "- Successful: $(echo $RESULT | jq -r '.successful')"
    echo "- Failed: $(echo $RESULT | jq -r '.failed')"
    echo "- Still pending: $(echo $RESULT | jq -r '.stillPending')"
else
    echo "Nothing to clean up! ✅"
fi

echo ""
echo "=== Cleanup Complete ==="
```

### **Добавить в crontab**

```bash
# Сделать исполняемым
chmod +x /path/to/weekly-cleanup.sh

# Добавить в crontab (каждое воскресенье в 3:00 AM)
crontab -e
```

Добавьте строку:
```
0 3 * * 0 /path/to/weekly-cleanup.sh >> /var/log/vpn-cleanup.log 2>&1
```

---

## 📊 **Мониторинг**

### **Логи удаления**

Смотреть failed deletions в реальном времени:
```bash
docker compose logs -f | grep "failed to delete user from node"
```

Подсчитать сколько fail за последний час:
```bash
docker compose logs --since 1h | grep "failed to delete" | wc -l
```

Посмотреть cleanup логи:
```bash
docker compose logs | grep "cleanup"
```

### **Метрики**

В логах есть счетчики:
```json
{"msg":"user deleted","successful":5,"failed":1}
{"msg":"cleanup completed","total":3,"successful":2,"failed":1,"still_pending":1}
```

Если `failed > 0` → нода была недоступна.

---

## 🗄️ **База данных**

### **Таблица pending_deletions**

```sql
CREATE TABLE pending_deletions (
    user_id TEXT NOT NULL,
    endpoint TEXT NOT NULL,
    inbound TEXT NOT NULL,
    attempts INT DEFAULT 1,
    last_attempt TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    error_message TEXT,
    PRIMARY KEY (user_id, endpoint, inbound)
);
```

### **Проверка БД вручную**

```bash
# Посмотреть pending deletions
docker exec vpnbot-core-go sqlite3 /app/data/vpnbot.db \
  "SELECT user_id, endpoint, attempts, error_message FROM pending_deletions;"

# Посмотреть кто на каких нодах (user_nodes)
docker exec vpnbot-core-go sqlite3 /app/data/vpnbot.db \
  "SELECT user_id, endpoint FROM user_nodes ORDER BY created_at DESC LIMIT 10;"

# Подсчитать pending deletions
docker exec vpnbot-core-go sqlite3 /app/data/vpnbot.db \
  "SELECT COUNT(*) FROM pending_deletions;"
```

---

## 🚨 **Алерты**

### **Когда беспокоиться:**

- ✅ **1-3 pending deletions** - норма (перезагрузки)
- ⚠️ **5-10 pending deletions** - проверить ноды
- 🚨 **>20 pending deletions** - серьезная проблема с нодами
- 🔴 **Одна нода постоянно в pending** - нода не работает

### **Проверка здоровья нод:**

```bash
# Проверить доступность всех нод
curl http://localhost:8080/api/stats -H "X-Api-Key: xxx" | \
  jq '.nodes[] | {endpoint, active, usersCount}'

# Проверить pending по нодам
curl http://localhost:8080/api/cleanup/pending -H "X-Api-Key: xxx" | \
  jq '.pendingDeletions | group_by(.endpoint) | map({endpoint: .[0].endpoint, count: length})'
```

---

## 🔍 **Поиск и диагностика**

### **Проблема: Нода постоянно недоступна**

1. Проверить pending для конкретной ноды:
```bash
curl http://localhost:8080/api/cleanup/pending \
  -H "X-Api-Key: xxx" | \
  jq '.pendingDeletions[] | select(.endpoint=="158.160.87.175")'
```

2. Проверить доступность ноды напрямую:
```bash
curl http://158.160.87.175:8000/api/health \
  -H "Authorization: Bearer $XRAY_NODE_TOKEN"
```

3. Временно деактивировать проблемную ноду:
```bash
curl -X PUT http://localhost:8080/api/servers \
  -H "X-Api-Key: xxx" \
  -H "Content-Type: application/json" \
  -d '{
    "servers": [{
      "endpoint": "158.160.87.175",
      "active": false
    }]
  }'
```

4. Если нода мертва навсегда - удалить все pending для нее:
```bash
# Получить все pending для ноды
PENDING=$(curl -s http://localhost:8080/api/cleanup/pending -H "X-Api-Key: xxx" | \
  jq -r '.pendingDeletions[] | select(.endpoint=="158.160.87.175")')

# Удалить каждую запись
echo "$PENDING" | jq -r '{userId, endpoint, inbound}' | while read record; do
  curl -X DELETE http://localhost:8080/api/cleanup/pending \
    -H "X-Api-Key: xxx" \
    -H "Content-Type: application/json" \
    -d "$record"
done

# Или вручную через SQLite (быстрее для массового удаления)
docker exec vpnbot-core-go sqlite3 /app/data/vpnbot.db \
  "DELETE FROM pending_deletions WHERE endpoint='158.160.87.175';"
```

---

## 💡 **Best Practices**

1. **Weekly Cleanup**: Запускайте `POST /api/cleanup` раз в неделю
2. **Мониторинг**: Проверяйте `GET /api/cleanup/pending` перед cleanup
3. **Бэкап**: Еженедельный бэкап БД перед cleanup
4. **Логирование**: Сохраняйте результаты cleanup в файл
5. **Алерты**: Настройте уведомления при >10 pending deletions

---

## 📈 **Статистика и отчеты**

### **Еженедельный отчет**

```bash
#!/bin/bash
# weekly-report.sh

API_KEY="your-api-key"
BASE_URL="http://localhost:8080"

echo "=== Weekly VPN Report ==="
echo "Date: $(date)"
echo ""

# Total stats
echo "1. Overall Statistics:"
curl -s "$BASE_URL/api/stats" -H "X-Api-Key: $API_KEY" | \
  jq '{totalUsers, nodes: [.nodes[] | {endpoint, countryCode, usersCount}]}'
echo ""

# Pending deletions
echo "2. Pending Deletions:"
curl -s "$BASE_URL/api/cleanup/pending" -H "X-Api-Key: $API_KEY" | \
  jq '{count, byNode: [.pendingDeletions | group_by(.endpoint) | .[] | {endpoint: .[0].endpoint, count: length}]}'
echo ""

# All users
echo "3. Users Distribution:"
curl -s "$BASE_URL/api/users" -H "X-Api-Key: $API_KEY" | \
  jq '.users | length'
```

---

## ✅ **Преимущества текущего решения**

### **Простота:**
- ✅ Best-effort - не блокирует удаление
- ✅ Автоматическое tracking failed deletions
- ✅ Ручной контроль через weekly cleanup

### **Надежность:**
- ✅ Не теряем информацию о неудачных удалениях
- ✅ Счетчик попыток и ошибки сохраняются
- ✅ Можно повторять cleanup много раз

### **Гибкость:**
- ✅ Ручной запуск - полный контроль
- ✅ Легко автоматизировать через cron
- ✅ Подробная информация в логах и API

---

## 🎯 **Workflow**

```
User deletion request
  ↓
Try delete from all nodes (parallel)
  ↓
Some nodes fail
  ↓
Add to pending_deletions ⚠️
  ↓
Weekly cleanup (Sunday)
  ↓
Retry pending deletions
  ↓
Success → Remove from pending ✅
Fail → Increment attempts counter ⚠️
```

---

## 🔧 **Troubleshooting**

### **Cleanup не работает**

1. Проверить логи:
```bash
docker compose logs | grep "cleanup"
```

2. Проверить pending deletions:
```bash
curl http://localhost:8080/api/cleanup/pending -H "X-Api-Key: xxx"
```

3. Запустить cleanup вручную:
```bash
curl -X POST http://localhost:8080/api/cleanup -H "X-Api-Key: xxx"
```

### **Много pending deletions**

1. Проверить какие ноды проблемные:
```bash
curl http://localhost:8080/api/cleanup/pending -H "X-Api-Key: xxx" | \
  jq '.pendingDeletions | group_by(.endpoint)'
```

2. Проверить доступность проблемных нод
3. Если нода мертва → деактивировать и удалить pending вручную:
```bash
docker exec vpnbot-core-go sqlite3 /app/data/vpnbot.db \
  "DELETE FROM pending_deletions WHERE endpoint='dead-node-ip';"
```

---

Этот подход дает вам **полный контроль** и **визуализацию** проблем, не перегружая систему автоматикой! 🎯
