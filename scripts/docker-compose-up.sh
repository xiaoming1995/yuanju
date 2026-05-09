#!/usr/bin/env sh

set -eu

ROOT_DIR=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
ENV_FILE="${YUANJU_DOCKER_ENV_FILE:-$ROOT_DIR/backend/.env.docker}"
EXAMPLE_FILE="${YUANJU_DOCKER_ENV_EXAMPLE_FILE:-$ROOT_DIR/backend/.env.docker.example}"
PREPARE_ONLY=0

if [ "${1:-}" = "--prepare-only" ]; then
  PREPARE_ONLY=1
  shift
fi

rand_hex() {
  bytes="$1"
  if command -v openssl >/dev/null 2>&1; then
    openssl rand -hex "$bytes"
    return
  fi
  od -An -N"$bytes" -tx1 /dev/urandom | tr -d ' \n'
}

ensure_env_file() {
  if [ -f "$ENV_FILE" ]; then
    return
  fi

  if [ ! -f "$EXAMPLE_FILE" ]; then
    echo "缺少示例文件: $EXAMPLE_FILE" >&2
    exit 1
  fi

  mkdir -p "$(dirname "$ENV_FILE")"

  postgres_user="yuanju"
  postgres_db="yuanju"
  postgres_password="$(rand_hex 16)"
  jwt_secret="$(rand_hex 32)"
  admin_jwt_secret="$(rand_hex 32)"
  admin_encryption_key="$(rand_hex 16)"

  cat > "$ENV_FILE" <<EOF
# Docker Compose 专用环境变量
# 由 scripts/docker-compose-up.sh 首次自动生成；已有文件不会被覆盖
PORT=9002
POSTGRES_USER=$postgres_user
POSTGRES_DB=$postgres_db
POSTGRES_PASSWORD=$postgres_password
DATABASE_URL=postgres://$postgres_user:$postgres_password@postgres:5432/$postgres_db?sslmode=disable
REDIS_URL=redis://redis:6379
JWT_SECRET=$jwt_secret
ADMIN_JWT_SECRET=$admin_jwt_secret
ADMIN_ENCRYPTION_KEY=$admin_encryption_key
DEEPSEEK_API_KEY=
DEEPSEEK_BASE_URL=https://api.deepseek.com
OPENAI_API_KEY=
FRONTEND_URL=http://localhost:3000
EOF

  chmod 600 "$ENV_FILE"
  echo "已生成 $ENV_FILE"
}

validate_env_file() {
  set -a
  # shellcheck disable=SC1090
  . "$ENV_FILE"
  set +a

  postgres_user="${POSTGRES_USER:-yuanju}"
  postgres_db="${POSTGRES_DB:-yuanju}"
  postgres_password="${POSTGRES_PASSWORD:-}"
  database_url="${DATABASE_URL:-}"
  expected_database_url="postgres://$postgres_user:$postgres_password@postgres:5432/$postgres_db?sslmode=disable"

  if [ -z "$postgres_password" ]; then
    echo "$ENV_FILE 缺少 POSTGRES_PASSWORD" >&2
    exit 1
  fi

  if [ "$database_url" != "$expected_database_url" ]; then
    echo "$ENV_FILE 中 DATABASE_URL 与 POSTGRES_USER/POSTGRES_DB/POSTGRES_PASSWORD 不一致" >&2
    echo "期望值: $expected_database_url" >&2
    echo "当前值: $database_url" >&2
    exit 1
  fi

  if [ "${REDIS_URL:-}" != "redis://redis:6379" ]; then
    echo "$ENV_FILE 中 REDIS_URL 必须是 redis://redis:6379" >&2
    exit 1
  fi
}

ensure_env_file
validate_env_file

if [ "$PREPARE_ONLY" -eq 1 ]; then
  exit 0
fi

cd "$ROOT_DIR"
docker compose up -d --build "$@"
