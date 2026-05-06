# 价值评分 3.1.1 成本参数配置口径说明

> 适用范围：OPS 系统「价值评分配置底座」中的 **3.1.1 成本参数配置**。
> 目标：统一配置与计算口径，避免月TCO解释偏差。

## 1. 成本参数项（3.1.1）

当前可配置参数：

1. **折旧月数（depreciation_months）**
   - 类型：整数
   - 约束：`> 0`
   - 默认值：`60`
   - 说明：用于约束折旧周期口径（当前月TCO中折旧金额来源于主机套餐表字段，折旧月数用于口径管理与展示）。

2. **网络机柜分摊（network_cabinet_share_cny）**
   - 类型：数值（CNY/月）
   - 约束：`>= 0`
   - 默认值：`0`
   - 说明：按配置类型统一计入的网络/机柜类分摊固定成本。

3. **其他固定成本（other_fixed_cost_cny）**
   - 类型：数值（CNY/月）
   - 约束：`>= 0`
   - 默认值：`0`
   - 说明：除机柜费、折旧、网络机柜分摊外的其他固定月成本。

---

## 2. 月TCO计算口径

系统当前统一采用：

`月TCO = 机柜费 + 折旧 + 网络机柜分摊 + 其他固定成本`

其中：
- **机柜费**：由机柜配置基线与功率计算；
- **折旧**：当前来自主机套餐中的 `monthly_depreciation_cny`；
- **网络机柜分摊**：来自 3.1.1 参数 `network_cabinet_share_cny`；
- **其他固定成本**：来自 3.1.1 参数 `other_fixed_cost_cny`。

---

## 3. 页面与接口

### 页面入口
- 路径：`导入配置 > 价值评分配置底座 > 3.1.1 成本参数配置`
- 操作：修改参数后点击「保存参数」
- 保存后行为：自动刷新「服务器月TCO试算」

### 接口
- 查询：`GET /api/v1/value-score/cost-params`
- 更新：`PUT /api/v1/value-score/cost-params`

请求/响应示例：

```json
{
  "depreciation_months": 60,
  "network_cabinet_share_cny": 120.5,
  "other_fixed_cost_cny": 300
}
```

---

## 4. 数据库存储

- 表：`ops_value_score_cost_params`
- 主键：固定 `id=1`
- 字段：
  - `depreciation_months`
  - `network_cabinet_share_cny`
  - `other_fixed_cost_cny`
  - `updated_at`

迁移文件：`backend/migrations/mysql_v19_value_score_cost_params.sql`

---

## 5. 填报建议（运营侧）

1. 折旧月数统一维护为当前组织口径（默认 60）。
2. 网络机柜分摊与其他固定成本建议以月度复核节奏维护，避免跨月沿用失真。
3. 若口径发生策略变化，优先更新 3.1.1 参数，再刷新月TCO导出。
