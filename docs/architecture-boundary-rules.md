# Architecture Boundary Rules (Ops Backend)

目的：防止模块化架构回流成“大泥球”，把边界纪律变成可执行约束。

## Layer Rules

以 `backend/internal/modules/<module>/` 为基本单元，推荐分层：

- `api`：HTTP 输入输出与协议适配
- `application`：用例编排、流程 orchestration
- `domain`：纯业务规则与领域模型
- `infrastructure`：存储/外部依赖实现

强制规则：

1. `api` 只能依赖本模块 `application`（或 shared 的横切能力），不得直接访问 repository/sql。
2. `application` 不得依赖 `api`。
3. `domain` 不得依赖 `api` / `application` / `infrastructure`。
4. `infrastructure` 不得依赖 `api`。
5. `shared` 仅允许横切能力（错误、日志、ID、时间、通用工具），禁止业务语义对象进入 shared。

## Shared Layer Guardrail

`backend/internal/shared/` 目录下新增代码时，需满足：

- 与具体业务模块无强耦合。
- 可被多个模块复用。
- 通过架构评审（PR 描述写明“为何必须 shared”）。

## PR Checklist (Required)

每个涉及 backend 的 PR 必须回答：

- [ ] 是否新增跨层依赖？若有，是否符合以上规则？
- [ ] 是否引入 shared 新对象？为什么不能放在模块内？
- [ ] 是否补充了对应测试（单测/契约/E2E）？

## Enforcement

- 架构规则测试：`go test ./internal/architecture/...`
- 全量后端测试：`go test ./...`
- E2E 决策链路测试：`scripts/check_e2e_ops.sh`

以上检查由 CI workflow 自动执行，失败则阻断合并。
