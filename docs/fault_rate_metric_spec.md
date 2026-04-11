# 故障率口径说明（computility-ops）

## 目标
- 概览卡片：`XX年故障率` vs `历史平均故障率`
- 趋势图：机龄 1~10 年，按 `storage` / `non_storage` 分开

## 分子
- 机龄 k（1~10）分子：故障清单中 `created_at` 落入服务器机龄第 k 年区间的故障数。

## 分母
- non_storage：每台服务器贡献 1
- storage：每台服务器贡献 `1 + data_disk_count`（沿用历史“存储加盘数”口径）

## 机龄桶
- 1年：`[purchase_date, purchase_date+1y)`
- 2年：`[purchase_date+1y, purchase_date+2y)`
- ...
- 10年：`[purchase_date+9y, purchase_date+10y)`

## 观察窗口
- 历史趋势窗口：`[故障清单最早 created_at, now]`
- 当年卡片窗口：`[当年1月1日, now]`
- 分母计入规则：服务器机龄桶与窗口有重叠即参与该桶分母。

## 公式
- 趋势点：`fault_rate(segment, age_k) = numerator_k / denominator_k`
- 历史平均：`sum(numerator_1..10) / sum(denominator_1..10)`
- 当年故障率：`当年故障数 / 当年分母暴露`

## 输出接口
- `GET /api/v1/failure-rates/overview-cards`
- `GET /api/v1/failure-rates/age-trend`
- 触发计算：`POST /api/v1/failure-rates/analyze/import`
