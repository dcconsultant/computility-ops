# computility-ops

Phase 1: renewal planning (no DB runtime, mysql abstraction reserved).

## 已完成（上线前四件套）

1. **异常提示优化**
   - 后端错误返回统一携带 `request_id`
   - 前端按错误码展示更友好的中文提示，并自动拼接请求 ID，便于排障
2. **导入字段映射兼容**
   - 支持标准列名 + 常见别名（含中文列名）自动映射到系统字段
   - 例如：`服务器编号/主机名/核数/价值分/到期日期/备注`
3. **日志审计**
   - 增加请求级审计日志（JSON 行格式）
   - 记录字段：请求ID、操作人、路径、状态、耗时、动作、结果、详情
4. **Docker 一键部署**
   - `docker-compose.yml` + 前后端 Dockerfile
   - `./deploy.sh` 一键构建并启动

## Backend quick start

```bash
cd backend
go mod tidy
go run ./cmd/server
```

Health check:

```bash
curl http://localhost:8080/api/v1/healthz
```

## Frontend quick start

```bash
cd frontend
npm install
npm run dev
```

- Frontend: http://localhost:5173
- Vite proxy: `/api` -> `http://localhost:8080`

## Docker 一键部署

```bash
./deploy.sh
```

- Frontend: `http://localhost:18080`
- 审计日志：`./logs/audit.log`

> 默认使用 memory 存储。若切换 mysql，可在 `docker-compose.yml` 中调整 `STORAGE_DRIVER` 和 `MYSQL_DSN`。

## 发布与版本号（P0）

版本格式：`V<大版本>.<中版本>.<YYMMDDHHMM>`，例如 `V1.0.2604100835`。

### 一键发版（推荐）

```bash
./release.sh "feat: xxx"
```

脚本会自动执行：
1. 更新 `frontend/src/version.ts`
2. 提交版本变更（commit message 自动附带版本号）
3. 推送 `origin/main`（失败时自动无代理重试）

### 自定义大/中版本

```bash
APP_VERSION_MAJOR=2 APP_VERSION_MINOR=3 ./release.sh "feat: xxx"
```

### 仅更新版本号（不提交）

```bash
APP_VERSION_MAJOR=1 APP_VERSION_MINOR=0 node ./scripts/bump-version.mjs
```
