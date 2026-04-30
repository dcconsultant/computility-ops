# Computility Ops 变更归档（2026-04-30）

## 基本信息

- 日期：2026-04-30
- 分支：`main`
- 归档提交：`ce3d13f`
- 提交主题：`feat(ops): wire decision modules and optimize frontend chunk loading`
- 变更规模：110 files，+2447 / -30

---

## 本次目标

1. 在后端接入并打通 ops 决策模块路由（replacement / reconfig / self-repair）。
2. 推进模块化目录与分层结构落地（modules/shared/architecture）。
3. 优化前端首屏加载（路由懒加载 + 构建分包策略）。
4. 保持可构建、可测试、可回退。

---

## 关键改动摘要

### 1) 后端应用装配与路由

- `backend/internal/app/bootstrap.go`
  - 注入并装配：
    - replacement-planning
    - reconfig-planning
    - self-repair
  - replacement 支持通过 `REPLACEMENT_RULES_FILE` 加载规则。

- `backend/internal/http/router.go`
  - 新增决策接口：
    - `GET /api/v1/ops/decisions/replacement`
    - `GET /api/v1/ops/decisions/reconfig`
    - `GET /api/v1/ops/decisions/self-repair`

- `backend/internal/http/router_test.go`
  - 增加对应路由契约测试（200 & envelope 校验）。

### 2) 模块化结构落地

新增/补全目录（含 domain/application/api/infrastructure/tests 骨架与实现）：

- `backend/internal/modules/asset-config`
- `backend/internal/modules/contract`
- `backend/internal/modules/failure-analytics`
- `backend/internal/modules/master-data`
- `backend/internal/modules/renewal`
- `backend/internal/modules/replacement-planning`
- `backend/internal/modules/reconfig-planning`
- `backend/internal/modules/self-repair`
- `backend/internal/modules/reporting`（骨架）
- `backend/internal/modules/workflow-audit`（骨架）

共享层与架构约束：

- `backend/internal/shared/*`
- `backend/internal/architecture/dependency_rules_test.go`

### 3) 前端性能优化

- `frontend/src/App.tsx`
  - 页面改为 `React.lazy + Suspense`（路由级懒加载）。

- `frontend/vite.config.ts`
  - 新增 `manualChunks` 分包策略（react / antd / utils / vendor）。

构建结果：

- 首包显著缩小，页面按需加载。
- 仍存在 `vendor-antd` 大包 warning（当前轮次可接受，不阻断发布）。

### 4) 配置与文档

- `backend/config/replacement-rules.example.yml`
- `backend/docs/modular-monolith-phase0-phase1-report.md`

---

## 质量验证记录

已通过：

- `backend`: `go test -race ./...`
- `backend`: `go build ./cmd/server`
- `frontend`: `npm run build`

说明：

- 前端构建存在 chunk size warning（非阻断项）。

---

## 发布与回退预案

### 已发布

- `git push` 成功：`main -> origin/main`

### 回退预案（仅备忘）

- 推荐（保留审计）：
  - `git revert ce3d13f`
- 临时查看旧版本：
  - `git checkout ce3d13f^`

---

## 后续建议（可选）

1. 前端第二轮瘦身（按组件按需策略、进一步拆重页面）。
2. 持续补齐 `reporting/workflow-audit` 模块从骨架到可用实现。
3. 将决策接口加入端到端联调脚本，固化发布前检查。

---

## 备注

本归档用于记录本次“功能接线 + 架构落地 + 前端减负”阶段性成果，供后续版本评审、回溯和回滚参考。
