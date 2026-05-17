#!/bin/bash
set -e

# ===================================================
# Hysteria2 Server Setup Script
# Для работы с Hysteria2 Admin Panel
# ===================================================

usage() {
    echo "Использование:"
    echo "  $0 --panel <URL> [--port 443] [--domain domain.com] [--cert file] [--key file]"
    echo ""
    echo "Обязательно:"
    echo "  --panel URL          Адрес панели (например: http://panel.example.com:8080)"
    echo ""
    echo "Опционально:"
    echo "  --port PORT          Порт Hysteria2 (по умолчанию: 443)"
    echo "  --domain DOMAIN      Домен для Let's Encrypt (авто-сертификат)"
    echo "  --cert FILE          Путь к существующему сертификату (.pem)"
    echo "  --key FILE           Путь к ключу сертификата (.pem)"
    echo "  --bandwidth-up       Лимит аплоада (по умолчанию: 100 mbps)"
    echo "  --bandwidth-down     Лимит даунлоада (по умолчанию: 500 mbps)"
    echo ""
    echo "Пример:"
    echo "  $0 --panel http://194.26.xxx.xxx:8080 --port 443 --domain ru-server.example.com"
    exit 1
}

# Парсинг аргументов
PANEL=""
PORT=443
DOMAIN=""
CERT_FILE=""
KEY_FILE=""
BW_UP="100 mbps"
BW_DOWN="500 mbps"

while [[ $# -gt 0 ]]; do
    case "$1" in
        --panel) PANEL="$2"; shift 2 ;;
        --port) PORT="$2"; shift 2 ;;
        --domain) DOMAIN="$2"; shift 2 ;;
        --cert) CERT_FILE="$2"; shift 2 ;;
        --key) KEY_FILE="$2"; shift 2 ;;
        --bandwidth-up) BW_UP="$2"; shift 2 ;;
        --bandwidth-down) BW_DOWN="$2"; shift 2 ;;
        *) echo "Неизвестный аргумент: $1"; usage ;;
    esac
done

if [[ -z "$PANEL" ]]; then
    echo "Ошибка: укажите --panel"
    usage
fi

echo "========================================"
echo " Установка Hysteria2 для работы с панелью"
echo " Панель: $PANEL"
echo " Порт:   $PORT"
echo "========================================"

# 1. Установка Hysteria2
echo "[1/5] Установка Hysteria2..."
bash <(curl -fsSL https://get.hy2.sh/)

# 2. Сертификат
echo "[2/5] Настройка сертификата..."
if [[ -n "$DOMAIN" ]]; then
    # Let's Encrypt через acme.sh или просто openssl
    echo "  Генерирую самоподписанный сертификат для домена $DOMAIN..."
    echo "  (рекомендуется заменить на Let's Encrypt для production)"
    CERT_FILE="/etc/hysteria/cert.pem"
    KEY_FILE="/etc/hysteria/key.pem"
    openssl req -x509 -nodes -newkey ec -pkeyopt ec_paramgen_curve:prime256v1 \
        -days 3650 -keyout "$KEY_FILE" -out "$CERT_FILE" \
        -subj "/CN=$DOMAIN" -addext "subjectAltName=DNS:$DOMAIN"
elif [[ -z "$CERT_FILE" || -z "$KEY_FILE" ]]; then
    echo "  Генерирую самоподписанный сертификат (ip-address)..."
    CERT_FILE="/etc/hysteria/cert.pem"
    KEY_FILE="/etc/hysteria/key.pem"
    openssl req -x509 -nodes -newkey ec -pkeyopt ec_paramgen_curve:prime256v1 \
        -days 3650 -keyout "$KEY_FILE" -out "$CERT_FILE" \
        -subj "/CN=hysteria-server"
else
    echo "  Использую существующие сертификаты"
fi

# 3. Конфигурация
echo "[3/5] Создание конфига /etc/hysteria/config.yaml..."
mkdir -p /etc/hysteria

cat > /etc/hysteria/config.yaml <<EOF
listen: :$PORT
protocol: hysteria2
cert: $CERT_FILE
key: $KEY_FILE

auth:
  type: http
  url: "$PANEL/api/sub/{uuid}"
  method: GET

bandwidth:
  up: "$BW_UP"
  down: "$BW_DOWN"

quic:
  initStreamReceiveWindow: 8388608
  maxStreamReceiveWindow: 8388608
  initConnReceiveWindow: 20971520
  maxConnReceiveWindow: 20971520
  maxIdleTimeout: 30s
EOF

echo "  Конфиг создан:"
cat /etc/hysteria/config.yaml

# 4. Systemd service
echo "[4/5] Настройка systemd..."
systemctl enable hysteria-server
systemctl restart hysteria-server

sleep 2

# 5. Проверка и вывод результата
echo "[5/5] Проверка статуса..."
if systemctl is-active --quiet hysteria-server; then
    echo ""
    echo "========================================"
    echo " ✅ Hysteria2 ЗАПУЩЕН!"
    echo "========================================"
    echo ""
    echo "Добавьте этот сервер в панель:"
    echo "------------------------------"
    echo "  Название:  $(hostname)"
    echo "  Адрес:     $(curl -4 -s ifconfig.me 2>/dev/null || dig +short myip.opendns.com @resolver1.opendns.com 2>/dev/null || echo '<PUBLIC_IP>')"
    echo "  Порт:      $PORT"
    echo "  API Порт:  9443"
    echo "  Локация:   <укажите сами>"
    echo ""
    echo "После добавления в панель:"
    echo "  1. Создайте пользователя"
    echo "  2. Назначьте ему ключ на этот сервер"
    echo "  3. Сгенерируйте подписку"
    echo "========================================"
else
    echo "❌ Ошибка: Hysteria2 не запустился"
    echo "Проверьте логи: journalctl -u hysteria-server -n 50"
    systemctl status hysteria-server --no-pager
    exit 1
fi
