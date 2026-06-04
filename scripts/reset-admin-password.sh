#!/usr/bin/env bash
# 重置后台管理员密码。
# 交互输入邮箱与新密码（密码隐藏不回显），用项目同款 bcrypt 生成哈希后写入数据库。
# 不可逆：旧密码无法找回，本脚本是「设一个新的覆盖掉」。
#
# 用法:  bash scripts/reset-admin-password.sh
# 可选环境变量(默认值): PG_CONTAINER=yuanju_postgres  PG_USER=yuanju  PG_DB=yuanju
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PG_CONTAINER="${PG_CONTAINER:-yuanju_postgres}"
PG_USER="${PG_USER:-yuanju}"
PG_DB="${PG_DB:-yuanju}"

# 1) 选账号
read -rp "管理员邮箱 [admin@yuanju.com]: " EMAIL
EMAIL="${EMAIL:-admin@yuanju.com}"

# 2) 输入新密码（隐藏，二次确认）
read -rsp "新密码: " NEWPW; echo
read -rsp "再输一次确认: " NEWPW2; echo
[ -n "$NEWPW" ] || { echo "✗ 密码不能为空，已取消。"; exit 1; }
[ "$NEWPW" = "$NEWPW2" ] || { echo "✗ 两次输入不一致，已取消。"; exit 1; }

# 3) 用项目 bcrypt 生成 UPDATE SQL（密码只通过环境变量传给子进程，不进命令行历史）
SQL="$(cd "$ROOT/backend" && NEWPW="$NEWPW" EMAIL="$EMAIL" go run ./cmd/adminpw)"

# 4) 写库
RESULT="$(docker exec -i "$PG_CONTAINER" psql -U "$PG_USER" -d "$PG_DB" -t -A -c "$SQL" 2>&1)"

if [ "$RESULT" = "UPDATE 1" ]; then
  echo "✅ 已重置 [$EMAIL] 的密码。直接用新密码登录后台即可（无需重启）。"
else
  echo "⚠️  未更新（数据库返回：$RESULT）。"
  echo "   请确认邮箱是否存在。现有管理员："
  docker exec -i "$PG_CONTAINER" psql -U "$PG_USER" -d "$PG_DB" -t -A -c "SELECT email FROM admins;" 2>&1 | sed 's/^/   - /'
  exit 1
fi
