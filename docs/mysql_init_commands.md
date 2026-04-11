# MySQL 初始化与启动命令（computility-ops）

## 推荐方式：外部 env 注入（不在仓库落地 `.env`）

### 1) 外部环境文件

创建：`~/.secrets/computility-ops.env`

```env
STORAGE_DRIVER=mysql
MYSQL_DSN=user:pass@tcp(127.0.0.1:3306)/computility_ops?parseTime=true&loc=Local&charset=utf8mb4
```

建议权限：

```bash
chmod 600 ~/.secrets/computility-ops.env
```

### 2) 初始化数据库

```bash
scripts/init_mysql.sh --host 127.0.0.1 --port 3306 --user user --db computility_ops
```

### 3) 启动服务

```bash
scripts/run_backend_with_env.sh ~/.secrets/computility-ops.env
```

### 4) 接口自检

```bash
scripts/check_api.sh
```

---

## 手动命令（备用）

```bash
export STORAGE_DRIVER=mysql
export MYSQL_DSN='user:pass@tcp(127.0.0.1:3306)/computility_ops?parseTime=true&loc=Local&charset=utf8mb4'

mysql -u user -p -h 127.0.0.1 -P 3306 -e "CREATE DATABASE IF NOT EXISTS computility_ops CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;"

mysql -u user -p computility_ops < backend/migrations/mysql_v1.sql
mysql -u user -p computility_ops < backend/migrations/mysql_v2_failure_dashboard.sql
mysql -u user -p computility_ops < backend/migrations/mysql_v3_ops_repo_tables.sql

cd backend
STORAGE_DRIVER=mysql MYSQL_DSN="$MYSQL_DSN" go run ./cmd/server
```
