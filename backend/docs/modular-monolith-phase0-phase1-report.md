# OPS 模块化单体重构报告（Phase 0 + Phase 1）

## Step 1（Phase 0）现状扫描报告

### 1) 现有目录结构与模块归类建议

**当前后端结构（重构前主干）**

- `internal/http/*`：API 路由与 Handler
- `internal/service/*`：业务编排逻辑（Import/Renewal/Contract）
- `internal/repository/*`：Repo 接口
- `internal/storage/mysql/*`：MySQL Repo 实现
- `internal/storage/memory/*`：内存 Repo 实现
- `internal/domain/*`：领域模型
- `migrations/*`：SQL 迁移脚本

**建议映射到目标模块（第一版）**

- `server + host-package + special-rule` => `asset-config`
- `contract_*` => `contract`
- `failure-rates + analyze` => `failure-analytics`
- `renewal_*` => `renewal`
- 新增（待建设）=> `replacement-planning / reconfig-planning / self-repair`
- `domain.Server` 的统一ID、型号、配置字典 => `master-data`
- 统一审计中间件与审批流支撑 => `workflow-audit`
- 报表聚合导出 => `reporting`

---

### 2) 跨模块 SQL/Repository 直接调用清单（风险点）

当前代码尚未模块化，主要表现为 **service 同时持有多个 repo**，未来会形成跨模块直连风险：

1. `internal/service/import_service.go`
   - `ImportService{serverRepo, datasetRepo}`
   - 同时读写服务器主数据与故障分析数据集
2. `internal/service/renewal_service.go`
   - `RenewalService{serverRepo, datasetRepo, renewalRepo}`
   - 续保规划直接跨读取服务器、套餐、特殊规则、故障率
3. `internal/service/contract_service.go`
   - `ContractService{contractRepo}`（当前相对独立）

**结论**：当前是“服务层聚合多仓储”，并非“模块边界内 application service 调用”。这是 Phase 2 必须切分的重点。

---

### 3) 旧 API 清单（路由 / 方法 / 请求响应形态）

统一响应形态（绝大多数接口）：
- 成功：`{"code":0,"message":"ok","data":...}`
- 失败：`{"code":<业务码>,"message":"...","data":{"request_id":"..."}}`

路由前缀：`/api/v1`

- GET `/healthz`
- POST `/servers/import`
- GET `/servers`
- GET `/servers/package-anomalies/export`
- POST `/host-packages/import`
- GET `/host-packages`
- POST `/special-rules/import`
- GET `/special-rules`
- POST `/failure-rates/model/import`
- GET `/failure-rates/model`
- POST `/failure-rates/package/import`
- GET `/failure-rates/package`
- POST `/failure-rates/package-model/import`
- GET `/failure-rates/package-model`
- GET `/failure-rates/overall`
- GET `/failure-rates/overview-cards`
- GET `/failure-rates/age-trend`
- GET `/failure-rates/features`
- GET `/failure-rates/storage-top-servers`
- GET `/failure-rates/storage-top-servers/export`
- POST `/failure-rates/analyze/import`
- POST `/failure-rates/year-fault-analysis/export`
- POST `/system/mysql/test`
- GET `/system/import-errors`
- POST `/contracts`
- GET `/contracts`
- GET `/contracts/:contract_id`
- PUT `/contracts/:contract_id`
- DELETE `/contracts/:contract_id`
- POST `/contracts/:contract_id/attachments`
- GET `/contracts/:contract_id/attachments/:attachment_id/download`
- DELETE `/contracts/:contract_id/attachments/:attachment_id`
- POST `/renewals/plan`
- GET `/renewals/plans`
- GET `/renewals/plans/:plan_id`
- DELETE `/renewals/plans/:plan_id`
- GET `/renewals/plans/:plan_id/export`
- GET `/renewals/plans/:plan_id/non-renewal/export`
- GET `/renewals/settings`
- PUT `/renewals/settings`
- GET `/renewals/unit-prices`
- PUT `/renewals/unit-prices`

---

### 4) 高风险点（兼容性 / 迁移顺序 / 回滚难点）

1. **兼容性风险**：前端依赖统一 `{code,message,data}` 结构，任何模块化重构都不能改。
2. **迁移顺序风险**：`renewal` 强依赖 `server + package + special-rule + failure-rates`，应在 `master-data + asset-config + failure-analytics` 之后迁。
3. **回滚难点**：若同一步混合“目录迁移 + SQL结构改动 + 业务逻辑调整”，回滚成本陡升。
4. **测试盲区**：当前以功能测试为主，模块边界/依赖规则缺少自动校验。

---

### 5) 分阶段迁移计划（Step 1 ~ Step 8）

- **Step 1**：搭模块骨架 + 边界规则检查（本次完成）
- **Step 2**：`master-data` 最小可用（资产统一视图）+ legacy adapter（本次完成）
- **Step 3**：`asset-config` 迁移接口壳，保持旧路由响应不变
- **Step 4**：`contract` 下沉到模块 application/domain
- **Step 5**：`failure-analytics` 从 ImportService 拆分
- **Step 6**：`renewal` 改为仅依赖 `master-data` 与 `failure-analytics` 应用接口
- **Step 7**：新增 `replacement-planning / reconfig-planning / self-repair`（规则先行）
- **Step 8**：清理 legacy adapter，完成一个兼容周期后收口

---

## Step 2（Phase 1）实施结果

### 变更说明

1. 新建 10 个目标模块目录骨架（含 `api/application/domain/infrastructure/tests`）。
2. 新建 shared 目录骨架（`kernel/db/eventbus/auth/utils`）。
3. 新增模块依赖规则自动检查测试（禁止跨模块基础设施层直接依赖）。
4. 新增 `master-data` 最小可用实现：
   - 资产领域对象
   - application service + repository interface
   - in-memory infrastructure repo
   - API handler（暂不挂路由）
   - legacy adapter（旧 `domain.Server` -> 新 `master-data.Asset`）
5. 未修改旧路由与旧响应结构，兼容性保持。

### 修改文件清单

- `internal/modules/**/.gitkeep`（模块骨架）
- `internal/shared/**/.gitkeep`（共享骨架）
- `internal/modules/master-data/domain/asset.go`
- `internal/modules/master-data/application/service.go`
- `internal/modules/master-data/infrastructure/memory_repo.go`
- `internal/modules/master-data/api/handler.go`
- `internal/modules/master-data/api/legacy_adapter.go`
- `internal/modules/master-data/tests/service_test.go`
- `internal/architecture/dependency_rules_test.go`
- `docs/modular-monolith-phase0-phase1-report.md`

### 关键 diff（摘要）

- 新增模块化目录：`internal/modules/<module>/{api,application,domain,infrastructure,tests}`
- 新增依赖规则测试：
  - `domain` 不得依赖非 `domain`
  - `application`/`api` 不得依赖 `infrastructure`
  - 跨模块不得 import 对方 `infrastructure`
- 新增 `master-data` 服务：`ListAssets(ctx)`，统一 `asset_id` 兜底规则

### 回归测试结果

- 命令：`cd backend && go test ./...`
- 结果：通过（见执行日志）

### 风险与回滚方案

- 风险：依赖规则测试当前为静态 import 规则，尚未覆盖运行时 SQL 调用路径。
- 风险：`master-data` API 尚未挂路由，Phase 2 接入时需保证旧接口 contract test 全通过。

**回滚方案（本步可回滚）**
1. `git revert <phase1_commit>` 一次性回退新增骨架和测试。
2. 无数据库变更，本步不涉及 migration 回滚。

### 下一步计划（仅 1 步）

- **Step 3**：迁移 `asset-config` 为模块化壳层（先搬接口与 application 边界，不改业务行为，不改路由响应）。

---

## Step 3（Phase 1 延续）实施结果

### 变更说明

1. 新增 `asset-config` 模块最小 application 壳层：
   - 聚合服务器、套餐、特殊规则查询能力
   - 输出模块内领域模型，形成后续从 legacy service 抽离的承接层
2. 新增 `asset-config` legacy 适配器占位（不改现有路由）
3. 新增基础单元测试，验证 `asset_id` 与查询输出
4. 未改动旧 API 路由与响应结构，兼容性保持

### 修改文件清单

- `internal/modules/asset-config/domain/models.go`
- `internal/modules/asset-config/application/service.go`
- `internal/modules/asset-config/api/legacy_adapter.go`
- `internal/modules/asset-config/tests/service_test.go`

### 关键 diff（摘要）

- 新增 `asset-config` 领域对象：`Server / HostPackage / SpecialRule`
- 新增应用服务：`ListServers / ListHostPackages / ListSpecialRules`
- 新增 legacy 适配器占位：支持增量替换旧 handler 的内部查询路径

### 回归测试结果

- 命令：`cd backend && go test ./...`
- 结果：通过（exit code 0）

### 风险与回滚方案

- 风险：当前 `asset-config` 仍通过旧 repo 接口读数据，Phase 2 需要继续下沉到模块 infrastructure repo 以完成边界闭环。
- 风险：尚未切换旧 import handler 到新 application service，当前为“并行壳层”阶段。

**回滚方案（本步可回滚）**
1. `git revert <step3_commit>` 回退新增的 asset-config 模块文件。
2. 本步无数据库变更，不需要 migration 回滚。

### 下一步计划（仅 1 步）

- **Step 4**：将 `contract` 迁移为模块 application/domain 壳层，并加 legacy adapter 占位（继续保持旧路由响应不变）。

---

## Step 4（Phase 1 延续）实施结果

### 变更说明

1. 新增 `contract` 模块最小 application 壳层：
   - 提供 `ListContracts / GetContract`
   - 内部完成 legacy domain 到模块 domain 的映射
2. 新增 `contract` 模块 legacy adapter（供后续旧 handler 增量切换）
3. 新增模块单测，验证查询与 contract_id trim 行为
4. 保持旧 API 路由与响应结构不变

### 修改文件清单

- `internal/modules/contract/domain/models.go`
- `internal/modules/contract/application/service.go`
- `internal/modules/contract/api/legacy_adapter.go`
- `internal/modules/contract/tests/service_test.go`

### 关键 diff（摘要）

- 新增 `contract` 领域模型（合同与附件）
- 新增应用服务：`ListContracts / GetContract`
- 新增 legacy adapter：为旧 handler 切换到模块服务提供兼容入口

### 回归测试结果

- 命令：`cd backend && go test ./...`
- 结果：通过（exit code 0）

### 风险与回滚方案

- 风险：当前仍复用旧 repository 接口，尚未形成 `contract` 模块内独立 infrastructure 实现。
- 风险：create/update/delete/upload/download 仍在 legacy service，下一步需按“只迁一类能力”的方式逐步下沉。

**回滚方案（本步可回滚）**
1. `git revert <step4_commit>` 回退本次 contract 模块新增文件。
2. 本步无数据库变更，不需要 migration 回滚。

### 下一步计划（仅 1 步）

- **Step 5**：拆分 `failure-analytics` 查询壳层（先读接口，不动导入分析写流程），继续保持旧路由响应不变。

---

## Step 5（Phase 1 延续）实施结果

### 变更说明

1. 新增 `failure-analytics` 模块读模型 domain（overall/overview/age-trend/features/storage-top）。
2. 新增 `failure-analytics` application service（仅查询能力）：
   - `ListOverallFailureRates`
   - `ListFailureOverviewCards`
   - `ListFailureAgeTrendPoints`
   - `ListFailureFeatureFacts`
   - `ListStorageTopServerRates`
3. 新增 legacy adapter 占位，供旧 handler 后续增量切换。
4. 新增模块单测，验证查询链路与字段映射。
5. 未改动旧导入分析写流程与旧 API 响应结构。

### 修改文件清单

- `internal/modules/failure-analytics/domain/models.go`
- `internal/modules/failure-analytics/application/service.go`
- `internal/modules/failure-analytics/api/legacy_adapter.go`
- `internal/modules/failure-analytics/tests/service_test.go`

### 关键 diff（摘要）

- 新增故障分析查询壳层，不触碰 `import_service` 写入流程。
- 采用 legacy domain -> module domain 的结构化映射，降低后续切路由风险。

### 回归测试结果

- 命令：`cd backend && go test ./...`
- 结果：通过（exit code 0）

### 风险与回滚方案

- 风险：当前查询仍依赖 legacy `DatasetRepo`，Phase 2 需迁移到模块内 infrastructure/repository。
- 风险：storage-top-servers 的 bucket 过滤逻辑尚在 legacy handler/service，未纳入模块应用层。

**回滚方案（本步可回滚）**
1. `git revert <step5_commit>` 回退本次新增 failure-analytics 模块文件。
2. 本步无数据库变更，不需要 migration 回滚。

### 下一步计划（仅 1 步）

- **Step 6**：将 `renewal` 查询能力下沉为模块 application 壳层（先 read path，保持旧 plan 生成与导出逻辑不变）。

---

## Step 6（Phase 1 延续）实施结果

### 变更说明

1. 新增 `renewal` 模块读模型 domain（plan/settings/unit-price 的最小查询视图）。
2. 新增 `renewal` application service（仅查询能力）：
   - `ListPlans`
   - `GetPlan`
   - `GetSettings`
   - `ListUnitPrices`
3. 新增 legacy adapter 占位，供旧 renewal handler 逐步切换 read path。
4. 新增模块单测，验证 `plan_id` trim 与查询结果映射。
5. 未修改旧的续保方案生成、导出、删除、写配置流程。

### 修改文件清单

- `internal/modules/renewal/domain/models.go`
- `internal/modules/renewal/application/service.go`
- `internal/modules/renewal/api/legacy_adapter.go`
- `internal/modules/renewal/tests/service_test.go`

### 关键 diff（摘要）

- 新增 `renewal` 查询壳层，不触碰 legacy `CreatePlan/Export/Delete/Update` 写路径。
- 通过 legacy repo 适配读取，先建立模块边界，再做后续下沉。

### 回归测试结果

- 命令：`cd backend && go test ./...`
- 结果：通过（exit code 0）

### 风险与回滚方案

- 风险：当前 renewal 查询仍依赖 legacy `RenewalPlanRepo`，模块 infrastructure 尚未独立。
- 风险：renewal 导出与非续保导出逻辑仍在 legacy handler/service，后续切换需补 contract test。

**回滚方案（本步可回滚）**
1. `git revert <step6_commit>` 回退本次 renewal 模块新增文件。
2. 本步无数据库变更，不需要 migration 回滚。

### 下一步计划（仅 1 步）

- **Step 7**：新增 `replacement-planning / reconfig-planning / self-repair` 三个决策模块的最小骨架与统一 decision DTO（先给建议读取接口，不接自动执行）。

---

## Step 7（Phase 1 延续）实施结果

### 变更说明

1. 新增统一决策 DTO（shared kernel）：`DecisionSuggestion / DecisionOption / RiskLevel`。
2. 新增三大决策模块最小可用壳层：
   - `replacement-planning`
   - `reconfig-planning`
   - `self-repair`
3. 每个模块均包含 `domain + application + api(handler placeholder) + tests`。
4. 当前仅提供“建议读取能力”，未接入自动执行链路，符合约束。

### 修改文件清单

- `internal/shared/kernel/decision.go`
- `internal/modules/replacement-planning/domain/models.go`
- `internal/modules/replacement-planning/application/service.go`
- `internal/modules/replacement-planning/api/handler.go`
- `internal/modules/replacement-planning/tests/service_test.go`
- `internal/modules/reconfig-planning/domain/models.go`
- `internal/modules/reconfig-planning/application/service.go`
- `internal/modules/reconfig-planning/api/handler.go`
- `internal/modules/reconfig-planning/tests/service_test.go`
- `internal/modules/self-repair/domain/models.go`
- `internal/modules/self-repair/application/service.go`
- `internal/modules/self-repair/api/handler.go`
- `internal/modules/self-repair/tests/service_test.go`

### 关键 diff（摘要）

- 引入统一决策结果结构，三个模块复用同一套建议输出格式。
- 三模块 application 先接 `Reader` 接口，便于后续挂接真实数据源与规则引擎。
- API handler 仅占位，未挂主路由，避免影响现网接口契约。

### 回归测试结果

- 命令：`cd backend && go test ./...`
- 结果：通过（exit code 0）

### 风险与回滚方案

- 风险：当前建议逻辑为模板化规则（MVP），真实业务权重与ROI公式待接入。
- 风险：三模块 API 尚未接入 router，下一步切路由需加 contract test 确保响应结构一致。

**回滚方案（本步可回滚）**
1. `git revert <step7_commit>` 回退本次新增 shared kernel 与三模块文件。
2. 本步无数据库变更，不需要 migration 回滚。

### 下一步计划（仅 1 步）

- **Step 8**：为三大决策模块补充 legacy adapter + 只读路由（`/api/v1/ops/decisions/*`）灰度入口，并增加 contract tests（保持 `{code,message,data}` 结构）。

---

## Step 8（Phase 1 延续）实施结果

### 变更说明

1. 三大决策模块补齐 legacy adapter：
   - `replacement-planning/api/legacy_adapter.go`
   - `reconfig-planning/api/legacy_adapter.go`
   - `self-repair/api/legacy_adapter.go`
2. 新增三模块只读灰度路由：
   - `GET /api/v1/ops/decisions/replacement`
   - `GET /api/v1/ops/decisions/reconfig`
   - `GET /api/v1/ops/decisions/self-repair`
3. 在 `app/bootstrap` 完成三模块服务与静态 reader 装配（单体内依赖注入）。
4. 增加路由契约测试，验证新路由保持 `{code,message,data}` 返回结构。

### 修改文件清单

- `internal/http/router.go`
- `internal/http/router_test.go`
- `internal/app/bootstrap.go`
- `internal/modules/replacement-planning/api/legacy_adapter.go`
- `internal/modules/reconfig-planning/api/legacy_adapter.go`
- `internal/modules/self-repair/api/legacy_adapter.go`
- `internal/modules/replacement-planning/infrastructure/static_reader.go`
- `internal/modules/reconfig-planning/infrastructure/static_reader.go`
- `internal/modules/self-repair/infrastructure/static_reader.go`
- `internal/modules/replacement-planning/api/handler.go`
- `internal/modules/reconfig-planning/api/handler.go`
- `internal/modules/self-repair/api/handler.go`

### 关键 diff（摘要）

- 新路由已挂载但仅只读，默认返回空列表，确保“先开接口契约、后接业务数据”。
- 新增 `TestNewRouter_DecisionRoutesContract`，逐一校验三条路由 200 + 统一 envelope。

### 回归测试结果

- 命令：`cd backend && go test ./...`
- 结果：通过（exit code 0）

### 风险与回滚方案

- 风险：当前静态 reader 返回空数据，需在后续步骤接入真实候选集与评分规则。
- 风险：新路由已公开，前端若提前接入会看到空列表（非错误）；需与前端同步“灰度阶段”预期。

**回滚方案（本步可回滚）**
1. `git revert <step8_commit>` 回退路由挂载与三模块 adapter/infrastructure 变更。
2. 本步无数据库变更，不需要 migration 回滚。

### 下一步计划（仅 1 步）

- **Step 9**：接入真实候选数据读取器（先 replacement），并落一版可配置评分规则（YAML）+ 单测基线。

---

## Step 9（Phase 2 起步）实施结果

### 变更说明

1. `replacement-planning` 接入真实候选读取器（LegacyReader）：
   - 从 `ServerRepo` + `DatasetRepo` 拉取服务器与套餐故障率数据
   - 生成候选项（年龄、AFR、维护成本、TCO估算）
2. 引入 YAML 可配置评分规则：
   - 新增 `ScoringRules` 与默认规则
   - 支持环境变量 `REPLACEMENT_RULES_FILE` 指向规则文件
   - 启动时加载规则并注入 `replacement` service
3. `replacement` 建议逻辑改为规则驱动：
   - 最小年龄/最小故障率门槛过滤
   - 成本与节省估算因子由规则控制
4. 补充单测：
   - YAML 规则加载测试
   - LegacyReader 候选生成测试
5. 新增示例配置文件：`config/replacement-rules.example.yml`

### 修改文件清单

- `internal/modules/replacement-planning/domain/rules.go`
- `internal/modules/replacement-planning/infrastructure/rules_loader.go`
- `internal/modules/replacement-planning/infrastructure/legacy_reader.go`
- `internal/modules/replacement-planning/infrastructure/rules_loader_test.go`
- `internal/modules/replacement-planning/infrastructure/legacy_reader_test.go`
- `internal/modules/replacement-planning/application/service.go`
- `internal/app/bootstrap.go`
- `config/replacement-rules.example.yml`

### 关键 diff（摘要）

- `bootstrap` 中 replacement 模块改为：
  - `LoadScoringRules(REPLACEMENT_RULES_FILE)`
  - `NewServiceWithRules(NewLegacyReader(serverRepo, datasetRepo), rules)`
- `ListSuggestions` 从硬编码阈值转为规则阈值。

### 回归测试结果

- 命令：`cd backend && go test ./...`
- 结果：通过（exit code 0）

### 风险与回滚方案

- 风险：当前维护成本/TCO 为启发式估算，需后续替换为真实成本口径。
- 风险：规则文件路径错误会导致启动失败（已保留默认值逻辑，但显式传错路径会报错）。

**回滚方案（本步可回滚）**
1. `git revert <step9_commit>` 回退 replacement 规则与 reader 相关变更。
2. 本步无数据库变更，不需要 migration 回滚。

### 下一步计划（仅 1 步）

- **Step 10**：将 `reconfig-planning` 接入真实候选读取（利用配置与利用率数据），并引入第二套 YAML 规则（改配阈值与收益模型）。
