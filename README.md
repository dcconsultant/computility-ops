# computility-ops

运维计算资源管理与故障率分析系统（Go + React）。

## 快速启动

### Backend

```bash
cd backend
go mod tidy
go run ./cmd/server
```

健康检查：

```bash
curl http://127.0.0.1:8080/api/v1/healthz
```

### Frontend

```bash
cd frontend
npm install
npm run dev
```

- Frontend: http://localhost:5173
- Vite 代理: `/api` -> `http://localhost:8080`

---

## MySQL 生产化运行（推荐：外部 env 注入）

### 1) 准备外部 secrets 文件（不放仓库）

示例：`~/.secrets/computility-ops.env`

```env
STORAGE_DRIVER=mysql
MYSQL_DSN=user:pass@tcp(127.0.0.1:3306)/computility_ops?parseTime=true&loc=Local&charset=utf8mb4
```

### 2) 初始化数据库表

```bash
scripts/init_mysql.sh --host 127.0.0.1 --port 3306 --user user --db computility_ops
```

> 脚本会依次执行：
> - `backend/migrations/mysql_v1.sql`
> - `backend/migrations/mysql_v2_failure_dashboard.sql`
> - `backend/migrations/mysql_v3_ops_repo_tables.sql`

### 3) 注入 env 启动后端（不拷贝 `.env` 到代码目录）

```bash
scripts/run_backend_with_env.sh ~/.secrets/computility-ops.env
```

### 4) 接口自检

```bash
scripts/check_api.sh
```

---

## 系统配置入口（前端）

- 顶部右上角有一个低显眼度的 ⚙️ 按钮（系统配置）
- 可直接在抽屉中填写 MySQL 参数并点击“测试 MySQL 连接”

## 关键 API（故障率 & 系统）

- `POST /api/v1/failure-rates/analyze/import`
  - 上传故障清单 xlsx，重算并落库
- `GET /api/v1/failure-rates/overall`
  - 历史/当年故障率汇总
- `GET /api/v1/failure-rates/overview-cards`
  - 概览卡片（当年 vs 历史平均，按 storage/non_storage）
- `GET /api/v1/failure-rates/age-trend`
  - 机龄 1~10 年趋势（storage/non_storage）
- `POST /api/v1/system/mysql/test`
  - 测试 MySQL 连接可达性（用于配置管理抽屉）

---

## 脚本说明

- `scripts/init_mysql.sh`：创建库并执行迁移
- `scripts/run_backend_with_env.sh`：从外部 env 文件注入配置并启动 backend
- `scripts/check_api.sh`：快速验证健康检查与故障率相关接口
- `scripts/bump-version.mjs`：版本号生成

---

## 安全建议

- 不要把账密写进仓库
- 保持 `.env` / `.env.*` 在 `.gitignore`
- secrets 文件权限建议 `chmod 600 ~/.secrets/computility-ops.env`
