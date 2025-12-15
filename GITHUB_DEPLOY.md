# 🚀 Деплой через GitHub Actions

Автоматический деплой на DigitalOcean при каждом push в main ветку.

---

## 📋 **Инструкция по настройке**

### **1. Создай DigitalOcean Droplet**

```bash
# На DigitalOcean:
# - Ubuntu 22.04 LTS
# - Минимум: 1GB RAM ($6/month)
# - Рекомендуется: 2GB RAM ($12/month)
```

---

### **2. Настрой сервер**

```bash
# Подключись к серверу
ssh root@YOUR_DROPLET_IP

# Установи Docker
apt update && apt upgrade -y
apt install -y docker.io docker-compose git
systemctl enable docker
systemctl start docker

# Создай директорию для проекта
mkdir -p /opt/vpnbot-core-go
cd /opt/vpnbot-core-go

# Инициализируй git (клонируем репо позже через Actions)
git init
git remote add origin https://github.com/YOUR_USERNAME/vpnbot-core-go.git

# Создай директории
mkdir -p data logs
```

---

### **3. Сгенерируй SSH ключ для GitHub Actions**

```bash
# На сервере создай отдельного пользователя для деплоя (опционально, но безопаснее)
adduser deploy
usermod -aG docker deploy
su - deploy

# Сгенерируй SSH ключ
ssh-keygen -t ed25519 -C "github-actions-deploy" -f ~/.ssh/github_deploy -N ""

# Добавь публичный ключ в authorized_keys
cat ~/.ssh/github_deploy.pub >> ~/.ssh/authorized_keys
chmod 600 ~/.ssh/authorized_keys

# СКОПИРУЙ ПРИВАТНЫЙ КЛЮЧ (нужен для GitHub Secrets)
cat ~/.ssh/github_deploy
# Скопируй весь вывод включая BEGIN и END строки
```

**ИЛИ используй root (проще):**

```bash
# На сервере как root
ssh-keygen -t ed25519 -C "github-actions" -f ~/.ssh/github_deploy -N ""
cat ~/.ssh/github_deploy.pub >> ~/.ssh/authorized_keys

# СКОПИРУЙ ПРИВАТНЫЙ КЛЮЧ
cat ~/.ssh/github_deploy
```

---

### **4. Создай GitHub репозиторий**

```bash
# На твоем Mac
cd /Users/anatoly/Desktop/coding/vpn-bot/vpnbot-core-go

# Инициализируй git (если еще не сделал)
git init
git add .
git commit -m "Initial commit"

# Создай приватный репозиторий на GitHub:
# https://github.com/new
# Название: vpnbot-core-go
# Visibility: Private (ВАЖНО!)

# Добавь remote и запуш
git remote add origin https://github.com/YOUR_USERNAME/vpnbot-core-go.git
git branch -M main
git push -u origin main
```

---

### **5. Настрой GitHub Secrets**

Зайди в настройки репозитория:
```
https://github.com/YOUR_USERNAME/vpnbot-core-go/settings/secrets/actions
```

Добавь следующие секреты (Settings → Secrets and variables → Actions → New repository secret):

| Name | Value | Описание |
|------|-------|----------|
| `DO_HOST` | `YOUR_DROPLET_IP` | IP адрес DigitalOcean сервера |
| `DO_USERNAME` | `root` или `deploy` | Пользователь для SSH |
| `DO_SSH_KEY` | `содержимое ~/.ssh/github_deploy` | Приватный SSH ключ (весь текст) |
| `PORT` | `8080` | Порт приложения |
| `API_KEY_REQUESTS` | `3f75f092-4019-4d98-8def-9d0af8fec75e` | Твой API ключ |
| `XRAY_NODE_TOKEN` | `твой-реальный-токен` | Токен для Xray нод |
| `REQUEST_TIMEOUT` | `10s` | Таймаут запросов |
| `NODE_TIMEOUT` | `3s` | Таймаут для нод |
| `LOG_LEVEL` | `info` | Уровень логирования |
| `LOG_MAX_SIZE` | `100` | Макс размер лог файла (MB) |
| `LOG_MAX_BACKUPS` | `3` | Количество бэкапов логов |
| `LOG_MAX_AGE` | `28` | Срок хранения логов (дни) |

---

### **6. Первый деплой**

```bash
# На сервере вручную склонируй репо первый раз
cd /opt
rm -rf vpnbot-core-go  # если уже существует
git clone https://github.com/YOUR_USERNAME/vpnbot-core-go.git
cd vpnbot-core-go
mkdir -p data logs

# Создай .env вручную (GitHub Actions потом будет обновлять)
cat > .env << 'EOF'
PORT=8080
API_KEY_REQUESTS=3f75f092-4019-4d98-8def-9d0af8fec75e
XRAY_NODE_TOKEN=твой-реальный-токен
DB_PATH=/app/data/vpnbot.db
REQUEST_TIMEOUT=10s
NODE_TIMEOUT=3s
LOG_LEVEL=info
LOG_OUTPUT=file
LOG_FILE=/app/logs/vpnbot.log
LOG_MAX_SIZE=100
LOG_MAX_BACKUPS=3
LOG_MAX_AGE=28
LOG_COMPRESS=true
EOF

# Запусти первый раз
docker compose up -d --build

# Проверь
curl http://localhost:8080/api/health
```

---

### **7. Настрой firewall**

```bash
# На сервере
ufw allow 22/tcp    # SSH
ufw allow 8080/tcp  # VPN Bot Core
ufw enable
ufw status
```

---

### **8. Тест автодеплоя**

```bash
# На твоем Mac - внеси любое изменение
cd /Users/anatoly/Desktop/coding/vpn-bot/vpnbot-core-go
echo "# Test deploy" >> README.md
git add .
git commit -m "Test auto-deploy"
git push

# Проверь GitHub Actions:
# https://github.com/YOUR_USERNAME/vpnbot-core-go/actions

# Через ~2 минуты проверь что обновилось:
curl http://YOUR_DROPLET_IP:8080/api/health
```

---

## ✅ **Готово!**

Теперь при каждом `git push` на `main` ветку:
1. ✅ GitHub Actions запускается автоматически
2. ✅ Подключается к твоему серверу
3. ✅ Делает `git pull`
4. ✅ Пересобирает Docker образ
5. ✅ Перезапускает контейнер
6. ✅ Проверяет health check

---

## 🔄 **Рабочий процесс**

```bash
# На твоем Mac
cd /Users/anatoly/Desktop/coding/vpn-bot/vpnbot-core-go

# Внеси изменения
vim internal/service/service.go

# Закоммить и запушить
git add .
git commit -m "Added new feature"
git push

# Все! Автоматически задеплоится на сервер
```

---

## 🛠️ **Ручной деплой (если нужно)**

Можешь запустить деплой вручную:
1. Зайди на https://github.com/YOUR_USERNAME/vpnbot-core-go/actions
2. Выбери "Deploy to DigitalOcean"
3. Нажми "Run workflow"
4. Выбери ветку `main`
5. Нажми "Run workflow"

---

## 📊 **Мониторинг деплоя**

Смотри логи деплоя в реальном времени:
```
https://github.com/YOUR_USERNAME/vpnbot-core-go/actions
```

Или на сервере:
```bash
ssh root@YOUR_DROPLET_IP "cd /opt/vpnbot-core-go && docker compose logs -f"
```

---

## 🔧 **Troubleshooting**

### **Деплой не работает**

1. Проверь GitHub Actions логи
2. Проверь что SSH ключ правильно добавлен в Secrets
3. Проверь что на сервере есть `/opt/vpnbot-core-go`
4. Проверь права доступа к директории

### **Контейнер не запускается**

```bash
ssh root@YOUR_DROPLET_IP
cd /opt/vpnbot-core-go
docker compose logs
```

### **Health check fails**

```bash
ssh root@YOUR_DROPLET_IP
curl http://localhost:8080/api/health -v
docker compose ps
```

---

## 💾 **Backup стратегия**

Бэкапы **НЕ включены** в автодеплой (чтобы не потерять данные).

База данных и логи сохраняются в:
- `/opt/vpnbot-core-go/data/vpnbot.db`
- `/opt/vpnbot-core-go/logs/`

При деплое:
- ✅ Код обновляется
- ✅ Контейнер пересобирается
- ✅ БД и логи **сохраняются**

---

## 🎯 **Преимущества этого подхода**

1. ✅ **Автоматизация** - push и готово
2. ✅ **История** - все деплои в GitHub Actions
3. ✅ **Rollback** - легко откатиться на предыдущий коммит
4. ✅ **Безопасность** - секреты в GitHub Secrets
5. ✅ **CI/CD** - можно добавить тесты перед деплоем
6. ✅ **Команда** - все разработчики могут деплоить

---

## 🚀 **Дальнейшие улучшения**

Можно добавить:
- Pre-deploy тесты (unit tests, linting)
- Slack/Telegram уведомления о деплое
- Staging окружение (отдельная ветка)
- Blue-green deployment
- Automatic rollback при ошибках

