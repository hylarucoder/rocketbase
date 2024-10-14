#!/bin/bash

# 数据库连接信息
DB_HOST="localhost"
DB_PORT="5432"
DB_USER="lucasay"
DB_NAME="postgres"

# 颜色代码
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 函数：执行 PostgreSQL 命令
execute_psql() {
  PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "$1"
}

# 检查 psql 是否可用
if ! command -v psql &>/dev/null; then
  echo -e "${RED}Error: psql is not installed or not in PATH${NC}"
  exit 1
fi

# 请求数据库密码
read -sp "Enter PostgreSQL password for user $DB_USER: " DB_PASSWORD
echo

# 列出要删除的数据库
echo -e "${YELLOW}Databases to be dropped:${NC}"
DB_LIST=$(execute_psql "SELECT datname FROM pg_database WHERE datname LIKE 'rb_test_%';")
echo "$DB_LIST"

# 请求用户确认
read -p "Are you sure you want to drop these databases? (y/N) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
  echo -e "${YELLOW}Operation cancelled.${NC}"
  exit 1
fi

# 终止到这些数据库的连接
echo "Terminating existing connections..."
execute_psql "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname LIKE 'rb_test_%';"

# 删除数据库
echo "Dropping databases..."
DROP_COMMANDS=$(execute_psql "SELECT format('DROP DATABASE IF EXISTS %I;', datname) FROM pg_database WHERE datname LIKE 'rb_test_%';")

IFS=$'\n'
for cmd in $DROP_COMMANDS; do
  echo "Executing: $cmd"
  execute_psql "$cmd"
done

echo -e "${GREEN}All matching databases have been dropped.${NC}"
