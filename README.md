# computility-ops

算力资源运营管理系统(L0)，覆盖服务器资产、故障率分析、续保、资源规划、重配规划、价值评分、合同供应商与元模型管理等运营管理场景。

## 功能模块

- 资产与基础数据：服务器清单、宿主机套餐、特殊规则、机柜配置。
- 故障率分析：按机型、套餐、套餐+机型导入故障率，支持故障清单分析、概览卡片、机龄趋势、特征事实、温存储高故障服务器导出。
- 价值评分与成本：成本参数、原始价值、性能参数、月度 TCO、性能评分计算。
- 合同与供应商：供应商导入/导出，合同、附件与基础 CRUD。
- 交付方式决策：自建与公有云成本对标、动态配比、国家模板与价格敏感度分析。
- 续保管理：续保计划、非续保清单导出、续保设置、单价维护。
- 资源与重配规划：资源规划配置/计算/最新结果，重配计划异步计算、进度、结果与动作导出。
- 运维决策建议：替换、重配、自愈等建议入口。
- 元模型管理：模型、字段、引用、版本、记录、导入任务与错误导出。
- 系统能力：健康检查、MySQL 连接测试、导入错误查询、审计日志。

## 技术栈与目录

- Backend：Go 1.24、Gin、Excelize，入口为 `backend/cmd/server`。
- Frontend：React 18、TypeScript、Vite、Ant Design。
- Storage：本地直跑默认 `memory`，Docker/生产口径默认 `mysql`。
- API 前缀：`/api/v1`。

主要目录：

```text
backend/              Go 后端服务
  cmd/server/         服务入口
  internal/           应用、领域、服务、HTTP、存储与模块化单体代码
  migrations/         MySQL 迁移 SQL
frontend/             Vite + React 前端
docs/                 PRD、OpenAPI、指标、架构边界与专题文档
scripts/              初始化、启动、自检与版本脚本
```

## 本地快速启动

### Backend：默认内存存储

适合本地体验和开发联调，不需要 MySQL。

```bash
cd backend
go mod tidy
go run ./cmd/server
```

健康检查：

```bash
curl http://127.0.0.1:8080/api/v1/healthz
```

返回中的 `storage_driver` 默认为 `memory`。

### Backend：本地连接 MySQL

如果需要持久化数据，显式注入环境变量后启动：

```bash
export STORAGE_DRIVER=mysql
export MYSQL_DSN='user:pass@tcp(127.0.0.1:3306)/computility_ops?parseTime=true&loc=Local&charset=utf8mb4'
cd backend
go run ./cmd/server
```

Windows PowerShell 示例：

```powershell
$env:STORAGE_DRIVER = "mysql"
$env:MYSQL_DSN = "user:pass@tcp(127.0.0.1:3306)/computility_ops?parseTime=true&loc=Local&charset=utf8mb4"
cd backend
go run ./cmd/server
```

### Frontend

```bash
cd frontend
npm install
npm run dev
```

- Frontend: `http://localhost:5173`
- Vite 代理：`/api` -> `http://localhost:8080`

## MySQL 初始化

推荐将数据库账号密码放在仓库外部，例如 `~/.secrets/computility-ops.env`：

```env
STORAGE_DRIVER=mysql
MYSQL_DSN=user:pass@tcp(127.0.0.1:3306)/computility_ops?parseTime=true&loc=Local&charset=utf8mb4
```

初始化数据库：

```bash
scripts/init_mysql.sh --host 127.0.0.1 --port 3306 --user user --db computility_ops
```

当前脚本会创建数据库，并按脚本中写死的顺序执行部分迁移。`backend/migrations/` 下存在更多历史迁移文件，部署新环境前应确认目标版本需要的完整 schema，并保持迁移脚本与迁移目录同步。

注入外部 env 启动后端：

```bash
scripts/run_backend_with_env.sh ~/.secrets/computility-ops.env
```

接口自检：

```bash
scripts/check_api.sh
```

> Windows 环境下，`.sh` 脚本建议在 WSL、Git Bash 或 CI/CD 的 Linux runner 中执行。

## Windows 部署

Windows 上有两种推荐方式。

### 方式一：Windows + Docker Desktop

适合本机演示、小规模部署或接近生产形态的验证。

1. 安装 Docker Desktop，并确保 Docker Compose 可用。
2. 准备 MySQL。可以使用 Windows 本机 MySQL，也可以使用外部 MySQL。
3. 复制 `.env.example` 为 `.env`，调整 `MYSQL_DSN`。从容器访问 Windows 本机 MySQL 时，可使用 `host.docker.internal`。
4. 在 WSL 或 Git Bash 中执行：

```bash
./deploy.sh
```

启动后访问：

- Frontend: `http://localhost:18080`
- Backend: 容器内 `:8080`，通过前端 Nginx 代理 `/api`
- Audit log: `./logs/audit.log`

### 方式二：Windows 原生进程

适合开发机或不使用容器的内部环境。

1. 安装 Go、Node.js、MySQL。
2. 使用 MySQL 客户端初始化数据库。
3. PowerShell 中设置后端环境变量并启动：

建议先确认 MySQL 大包参数，避免生成续保方案或导入大数据时触发 `max_allowed_packet` 限制：

```powershell
mysql -u root -p -e "SHOW VARIABLES LIKE 'max_allowed_packet';"
mysql -u root -p -e "SET GLOBAL max_allowed_packet=67108864;"
```

如需永久生效，在 MySQL 配置文件的 `[mysqld]` 下增加 `max_allowed_packet=64M` 后重启 MySQL。

```powershell
$env:STORAGE_DRIVER = "mysql"
$env:MYSQL_DSN = "user:pass@tcp(127.0.0.1:3306)/computility_ops?parseTime=true&loc=Local&charset=utf8mb4"
$env:APP_ADDR = ":8080"
cd backend
go run ./cmd/server
```

4. 另开终端启动前端：

```powershell
cd frontend
npm install
npm run dev
```

如需长期运行，可将后端编译为 Windows 可执行文件，并交给 Windows Service、NSSM、计划任务或企业内部进程管理工具托管；前端可用 `npm run build` 后由 IIS/Nginx/Caddy 托管静态文件，并将 `/api` 反向代理到后端。

## 云原生部署

项目已经提供后端、前端 Dockerfile，可作为 Kubernetes、Helm、Kustomize 或云厂商容器平台的基础镜像构建来源。

推荐部署形态：

- Backend：无状态 Deployment，暴露 `APP_ADDR=:8080`，通过 ConfigMap/Secret 注入 `STORAGE_DRIVER=mysql` 与 `MYSQL_DSN`。
- Frontend：Nginx 静态站点镜像，`/api/` 反向代理到后端 Service。
- MySQL：优先使用云数据库或托管 MySQL；如自建，请独立 StatefulSet/云盘/备份策略。
- Logs：容器标准输出与 `AUDIT_LOG_PATH` 挂载或采集到平台日志系统。
- Ingress：统一暴露前端域名，由前端 Nginx 或 Ingress 路由 `/api` 到后端。
- Secrets：数据库账号、密码、DSN 使用 Kubernetes Secret 或云平台密钥服务，不写入镜像和 Git。
- Health check：使用 `GET /api/v1/healthz` 做 readiness/liveness 探测。

部署前检查：

- 数据库 schema 已按目标版本完成迁移。
- `MYSQL_DSN` 的主机名在集群内可解析，例如云数据库地址或 MySQL Service DNS。
- 前端 `/api/` 代理目标与后端 Service 名称一致。
- 审计日志路径具备写入权限，或改为平台日志采集方式。

## 环境变量

| 变量 | 默认值 | 说明 |
| --- | --- | --- |
| `APP_ADDR` | `:8080` | 后端监听地址 |
| `STORAGE_DRIVER` | 本地代码默认 `memory`，Compose 默认 `mysql` | 存储驱动，支持 `memory` / `mysql` |
| `MYSQL_DSN` | 空 | Go MySQL driver DSN |
| `AUDIT_LOG_PATH` | Compose 中为 `/var/log/computility/audit.log` | 审计日志路径 |
| `META_IMPORT_CLEAN_DAYS` | `7` | 元模型导入任务自动清理天数 |
| `META_IMPORT_KEEP_LATEST` | `200` | 清理时保留最近任务数量 |
| `META_IMPORT_UNIQUE_MODE` | `strict` | 导入唯一校验模式：`strict` / `off` |

## 常用脚本

- `scripts/init_mysql.sh`：创建数据库并执行脚本中配置的 MySQL 迁移。
- `scripts/run_backend_with_env.sh`：从仓库外部 env 文件注入配置并启动 backend。
- `scripts/check_api.sh`：快速验证健康检查与故障率概览接口。
- `scripts/check_e2e_ops.sh`：运维链路端到端检查脚本。
- `scripts/bump-version.mjs`：版本号生成。
- `deploy.sh`：Docker Compose 构建并启动后端、前端。
- `release.sh`：发布辅助脚本。

## 关键 API

完整接口以 `backend/internal/http/router.go` 和 `docs/openapi.yaml` 为准。常用入口：

- `GET /api/v1/healthz`：健康检查。
- `POST /api/v1/servers/import`：导入服务器清单。
- `POST /api/v1/host-packages/import`：导入宿主机套餐。
- `POST /api/v1/failure-rates/analyze/import`：上传故障清单 xlsx，重算并落库。
- `GET /api/v1/failure-rates/overview-cards`：故障率概览卡片。
- `GET /api/v1/failure-rates/age-trend`：机龄 1~10 年趋势。
- `POST /api/v1/value-score/tco/calculate`：计算月度 TCO。
- `POST /api/v1/resource-planning/calculate`：资源规划计算。
- `POST /api/v1/reconfig/plan/start`：启动重配计划计算。
- `POST /api/v1/system/mysql/test`：测试 MySQL 连接。
- `GET /api/v1/system/import-errors`：查询导入错误。

## 安全建议

- 不要把数据库账号、密码、DSN 写进仓库。
- 保持 `.env` / `.env.*` 在 `.gitignore`。
- secrets 文件权限建议设置为仅当前用户可读，例如 `chmod 600 ~/.secrets/computility-ops.env`。
- 生产环境使用最小权限数据库账号，并开启数据库备份。
- 云原生环境中使用 Secret/密钥服务管理敏感配置。
