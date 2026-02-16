#!/bin/bash
# Скрипт для снятия CPU и trace профилей с нагрузкой и JWT

HOST=${1:-localhost}
PORT=${2:-8080}
PPROF_PATH=${3:-/api/pprof}
JWT_TOKEN=${4:-}
OUT_DIR="pkg/pprof_profiles"
MAX_RETRIES=5

mkdir -p "$OUT_DIR"
BASE_HTTP_PORT=9000
COUNTER=0

# Проверка токена
if [[ -z "$JWT_TOKEN" ]]; then
    echo "Не указан JWT токен. Использование: ./fetch_pprof.sh host port /api/pprof <JWT_TOKEN>"
    exit 1
fi

CURL_AUTH=(-H "Authorization: Bearer $JWT_TOKEN")

# Функция для скачивания профиля с проверкой бинарности
download_profile() {
    local name=$1
    local url=$2
    local file="$OUT_DIR/$name.pprof"
    local attempt=1

    while [[ $attempt -le $MAX_RETRIES ]]; do
        echo "Попытка $attempt: Скачиваю $name -> $file"
        curl -s "${CURL_AUTH[@]}" "$url" -o "$file"

        if file "$file" | grep -qv 'ASCII text'; then
            echo "$name скачан успешно"

            # Веб-интерфейс только для CPU
            if [[ "$name" == "cpu" ]]; then
                local http_port=$((BASE_HTTP_PORT + COUNTER))
                echo "Запускаю go tool pprof -http=:$http_port для $file"
                nohup go tool pprof -http-handle=":$http_port" "$file" >/dev/null 2>&1 &
                echo "  ➜ http://localhost:$http_port"
                COUNTER=$((COUNTER + 1))
            fi
            return
        else
            echo "Файл $file пустой или содержит JSON с ошибкой"
            ((attempt++))
            sleep 2
        fi
    done

    echo "Не удалось скачать корректный профиль $name после $MAX_RETRIES попыток"
}

# --- 1. Создаём нагрузку на сервис ---
echo "Запускаем нагрузку на сервис..."

if command -v wrk >/dev/null 2>&1; then
    wrk -t4 -c100 -d15s "http://$HOST:$PORT/" >/dev/null 2>&1 &
    LOAD_PID=$!
    echo "wrk нагрузка запущена (PID $LOAD_PID)"
elif command -v ab >/dev/null 2>&1; then
    ab -n 1000 -c 20 "http://$HOST:$PORT/" >/dev/null 2>&1 &
    LOAD_PID=$!
    echo "ab нагрузка запущена (PID $LOAD_PID)"
else
    echo "wrk или ab не найдены, нагрузка пропущена"
    LOAD_PID=""
fi

# Ждём прогрева нагрузки
sleep 5

# --- 2. Снятие профилей ---
# CPU profile (30 секунд)
download_profile "cpu" "http://$HOST:$PORT$PPROF_PATH/profile?seconds=30"

# Trace profile (10 секунд)
download_profile "trace" "http://$HOST:$PORT$PPROF_PATH/trace?seconds=10"

# --- 3. Завершаем нагрузку ---
if [[ -n "$LOAD_PID" ]]; then
    kill $LOAD_PID >/dev/null 2>&1 || true
fi

echo
echo "Все профили сохранены в $OUT_DIR"
echo "CPU профиль можно анализировать через веб-интерфейс:"
echo "http://localhost:9001/ui/"
echo "Trace профиль анализируется через:"
echo "go tool trace $OUT_DIR/trace.pprof"
