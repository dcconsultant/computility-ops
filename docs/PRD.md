# computility-ops PRD（Phase 1）

- **文档版本**：v1.0
- **更新日期**：2026-04-09
- **产品阶段**：Phase 1（可用版本，内存存储为默认）
- **适用读者**：产品、后端、前端、运维、交接同学

---

## 1. 背景与目标

computility-ops 用于支撑“服务器续保规划”场景：
- 把多来源资产表（服务器、套餐、特殊名单、故障率）统一导入系统
- 基于可解释规则生成续保方案
- 支持查询与导出，方便业务复核和落地执行

### 1.1 业务问题

当前资产续保决策依赖人工筛选，痛点是：
- 数据源分散，字段命名不统一
- 规则执行靠人，易错且不可追踪
- 结果复盘困难（谁、何时、为何选了这些机器）

### 1.2 Phase 1 目标

1. 打通从“导入数据 → 生成方案 → 查询/导出结果”的最小闭环
2. 提供可解释、可审计的规则引擎（基础版）
3. 支持容器化一键部署，便于快速上线验证

### 1.3 非目标（Phase 1 暂不做）

- 多租户与细粒度权限体系
- 复杂多目标优化算法（成本/风险/库存联动）
- 自动接入 CMDB/监控系统实时同步
- 故障率参与计算（当前仅录入与展示，不参与打分）

---

## 2. 用户与使用场景

### 2.1 目标用户

- **资产运营 / 交付团队**：上传台账、发起续保方案
- **技术管理 / 审批人**：查看方案结构、导出用于评审
- **运维 / 开发同学**：部署系统、排障、追踪日志

### 2.2 核心场景

1. 运营同学上传最新服务器管理表和基础规则表
2. 输入目标核数，系统自动产出续保候选清单
3. 审批/管理层按排名与策略字段复核结果
4. 导出 CSV/XLSX 给下游执行或归档

---

## 3. 产品范围（Phase 1 功能）

### 3.1 数据导入与管理（Import）

系统提供 6 类数据导入入口（XLSX）：
1. 服务器管理（servers）
2. 主机套餐配置（host-packages）
3. 特殊名单（special-rules）
4. 型号故障率（failure-rates/model）
5. 套餐故障率（failure-rates/package）
6. 套餐型号故障率（failure-rates/package-model）

支持：
- 常见中英文列名/别名自动映射
- 行级校验与错误回传（第几行、失败原因）
- 导入后列表即时可见

### 3.2 续保方案生成（Plan）

输入：`target_cores`（目标核数，>0）

核心规则：
- 服务器与主机套餐按 `config_type` 关联
- 计算分数：`final_score = PSA × arch_standardized_factor`
- 排序优先级：
  1) final_score 降序
  2) cpu_logical_cores 降序
  3) SN 升序
- 特殊名单策略：
  - `whitelist`：强制入选（优先）
  - `blacklist`：强制排除
- 从高到低选取，直至达到/超过目标核数（加白机器先入选）

输出：
- `plan_id`
- 目标核数、已选核数、入选台数
- 入选明细（排名、SN、型号、配置类型、分数、策略等）

### 3.3 方案查询与导出（Result）

- 按 `plan_id` 查询历史方案
- 前端支持关键词过滤（SN/型号）
- 支持导出格式：XLSX、CSV

### 3.4 审计与可观测

- 请求级审计日志（JSON 行）
- 关键字段：request_id、操作路径、状态码、耗时、动作、结果、详情
- 错误响应统一携带 `request_id`，便于排障

### 3.5 部署方式

- 本地开发：
  - Backend：Go（默认 `:8080`）
  - Frontend：Vite（默认 `:5173`）
- 容器部署：`./deploy.sh`
  - 对外访问：`http://localhost:18080`
  - 审计日志：`./logs/audit.log`

---

## 4. 信息架构与页面

前端路由：
- `/import`：数据导入与列表管理
- `/plan`：生成续保方案
- `/result/:planId`：查询与查看方案结果

页面关系：
- 导入数据（Import）是生成方案（Plan）的前置
- 方案生成后跳转结果页（Result）查看明细并导出

---

## 5. API 概览（对齐当前实现）

基础前缀：`/api/v1`

- `GET /healthz`
- `POST /servers/import`, `GET /servers`
- `POST /host-packages/import`, `GET /host-packages`
- `POST /special-rules/import`, `GET /special-rules`
- `POST /failure-rates/model/import`, `GET /failure-rates/model`
- `POST /failure-rates/package/import`, `GET /failure-rates/package`
- `POST /failure-rates/package-model/import`, `GET /failure-rates/package-model`
- `POST /renewals/plan`
- `GET /renewals/plans/:plan_id`
- `GET /renewals/plans/:plan_id/export?format=xlsx|csv`

返回规范：
- 成功：`code=0`
- 失败：携带错误信息与 `request_id`

---

## 6. 数据与规则约束

### 6.1 必要数据前提

生成方案前必须至少满足：
- 已导入服务器列表
- 已导入主机套餐配置（且配置类型可关联）

### 6.2 关键校验

- `target_cores > 0`
- `cpu_logical_cores > 0`
- `SN`、`config_type` 等关键字段不能为空
- 特殊策略仅接受加白/加黑（含英文同义词）

### 6.3 当前存储形态

- 默认 `memory`（重启丢失）
- 预留 `mysql` 抽象与迁移脚本（`backend/migrations/mysql_v1.sql`）

---

## 7. 非功能需求（Phase 1）

1. **可用性**：单机部署可稳定提供导入/计算/导出能力
2. **可维护性**：前后端分离，API 与页面解耦
3. **可追踪性**：审计日志可关联请求链路
4. **可扩展性**：存储层接口化，后续切换 MySQL 成本可控

---

## 8. 已知限制与风险

1. 默认内存存储，服务重启后数据不保留
2. 方案算法是规则驱动，不是最优解引擎
3. 故障率数据当前不参与计算，仅作为后续版本入口
4. 大体量导入/并发场景尚未做专项压测
5. 当前仓库已临时忽略 `.github/workflows/ci.yml`（因推送权限策略），CI 需后续恢复

---

## 9. 交接说明（给后续接手同学）

### 9.1 最短上手路径

1. 阅读：`README.md` + `docs/openapi.yaml` + 本 PRD
2. 一键启动：`./deploy.sh`
3. 使用 `frontend/public/server_template.csv` 准备样例数据
4. 按顺序验证：导入 → 生成方案 → 查询/导出

### 9.2 代码主结构

- 后端入口：`backend/cmd/server/main.go`
- 路由：`backend/internal/http/router.go`
- 导入服务：`backend/internal/service/import_service.go`
- 续保服务：`backend/internal/service/renewal_service.go`
- 前端页面：`frontend/src/pages/{ImportPage,PlanPage,ResultPage}.tsx`

### 9.3 建议优先事项（下一阶段）

P0：
- 恢复 CI（workflow 权限后恢复 `.github/workflows/ci.yml`）
- 切 MySQL 持久化并补迁移说明

P1：
- 把故障率纳入评分模型（可配置权重）
- 增加方案比对与版本管理

P2：
- 对接 CMDB/资产系统自动同步
- 增加审批流与权限控制

---

## 10. 验收标准（Phase 1）

满足以下条件可判定交付可用：

1. 6 类数据均可导入并返回结构化结果
2. 输入目标核数可稳定生成 plan_id
3. 结果页可查询、过滤并导出 XLSX/CSV
4. 错误提示可定位（含 request_id）
5. `deploy.sh` 在目标机器可一键启动成功

---

## 11. 术语

- **PSA**：用于排序的业务评分字段（由业务侧提供）
- **架构标准化系数**：按配置类型定义的权重系数
- **加白/加黑**：强制入选 / 强制排除策略
- **目标核数**：本次续保计划的资源目标

---

> 备注：本文档是“当前代码实现”的产品化说明，后续如算法或数据模型变更，请同步更新本 PRD。
