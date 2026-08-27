# TPC-DS 1TB 查询优化记录

- 最后更新：2026-08-28
- 数据库：`tpcds_1t`，TPC-DS SF1000
- 测试目录：`/d/mo-worktrees/tpcds-lab`
- 目的：逐条记录每个查询的问题、通用修改和 1TB 实测结果，避免把查询特判混入优化器

## 测试约束

- 单机 16 核、31 GiB 内存，一次只运行一条查询。
- 使用普通 `EXPLAIN` 检查计划；性能数据来自 `system.statement_info`。
- 运行查询前关闭自动 merge，并等待已经启动的 merge 任务退出。
- 当前实例统一设置：

  ```sql
  set global agg_spill_mem = 134217728;
  set global join_spill_mem = 134217728;
  set global sort_spill_mem = 134217728;
  ```

- Q2 基线使用 `e866c535f5`；Q4 计划基线使用 `484d9815eb`。耗时受版本、缓存和对象布局影响，计划结构、扫描计数和结果一致性是主要验收依据。
- 当前工作树下 Q1–Q10 的确切 SQL、普通 `EXPLAIN` 和输入 stats 保存在 [`tpcds-1t-q01-q10-baseline`](tpcds-1t-q01-q10-baseline/README.md)；Q11 以后按完成顺序保存在 [`tpcds-1t-q11-plus-baseline`](tpcds-1t-q11-plus-baseline/README.md)。
- 目前 Q1–Q66 已实跑完成；Q67 在 20 分钟预算处取消，Q68–Q99 待继续。

## 进度摘要

| 查询 | 问题 | 修改 | 1TB 结果 | 状态 |
|---|---|---|---|---|
| Q1 | 聚合需要更早、可预测地 spill | 统一将 spill 阈值设为 128 MiB；无代码特判 | 21.411 秒，100 行 | 完成 |
| Q2 | 外层多引用 CTE 因内部引用另一个 CTE 而被整体 inline 两次 | 放宽外层 CTE reuse；保留全部安全、成本和内存保护 | 1980.838 → 855.284 秒，结果仍为 2513 行 | 完成 |
| Q3 | 未发现明显危险计划 | 不修改 | 17.214 秒，100 行 | 完成 |
| Q5 | `ROLLUP` 将完整输入复制三份，但原 SQL 可在 spill 保护下完成 | 记录；共享 grouping-set 输入需要独立设计，不在此处做大改 | 250.914 秒，100 行 | 完成 |
| Q6 | 无相关标量等值条件位于大连接之后，月份选择性未进入事实表路径 | 将确定性的 `FILTER + SINGLE JOIN` 整体移到最小相关输入；不依赖 stats | 原计划 813.478 秒；新计划已命中，1TB 待复测 | 计划修复完成 |
| Q7 | 维表过滤和 PK join 均合理 | 不修改 | 122.506 秒，100 行 | 完成 |
| Q8 | 日期过滤、地址集合缩减和小 build side 均合理 | 不修改 | 32.712 秒，11 行 | 完成 |
| Q9 | 15 个同源标量聚合分别全扫 `store_sales` | 记录；正确方向是通用子查询 CSE/聚合融合 | 253.277 秒，1 行 | 完成 |
| Q10 | `EXISTS` 等式原先退化为 loop join；修复后 `OR` 下仍保留两个大 MARK build | 保留可哈希裸等式；再将同键正向 `EXISTS OR` 合并为 `UNION ALL` key 输入上的 SEMI join | 超过 26 分钟 → 52.226 → 29.271 秒，100 行 | 完成；Q35 规则的独立正向样本 |
| Q11 | Q4 同类多消费者 CTE；当前共享和 partial SUM 均已命中 | 无新增修改 | 254.265 秒，100 行 | 完成 |
| Q12 | 日期、品类过滤下推；窗口位于聚合之后 | 不修改 | 10.827 秒，100 行 | 完成 |
| Q13 | OR 条件的单表候选过滤已下推，跨表相关条件保留在 join | 不修改 | 169.422 秒，1 行 | 完成 |
| Q14 | Q14a grouping sets 重复公共输入；Q14b 标量周条件位于大连接之后 | grouping-set 共享公共输入；通用 `SINGLE JOIN` 条件提前；`cross_items` 共享只记录 | Q14a 2097.510 秒；Q14b 超过 2932 秒未完成 → 1311.300 秒；均为 100 行 | 完成；`cross_items` 待优化 |
| Q15 | 单次事实表扫描；日期过滤和 PK join 均合理 | 不修改 | 14.664 秒，100 行 | 完成 |
| Q16 | `EXISTS` 自连接和 `NOT EXISTS` 反连接均保留可哈希等值键 | 不修改 | 29.091 秒，1 行 | 完成 |
| Q17 | 三张事实表的季度过滤均在可哈希连接前生效 | 不修改 | 63.533 秒，100 行 | 完成 |
| Q18 | 四列 `ROLLUP` 原本会把完整公共输入复制五份 | 已命中 grouping-set 公共输入共享，`catalog_sales` 只扫描一次 | 72.750 秒，100 行 | 完成；Q14 修复的独立正向样本 |
| Q19 | 单次事实表扫描；日期和 item 过滤已提前 | 不修改 | 38.028 秒，100 行 | 完成 |
| Q20 | 事实侧按 item 的 partial SUM 已在维表连接前生效 | 不修改 | 14.344 秒，100 行 | 完成 |
| Q21 | 日期范围和 item 价格过滤已下推，单次 inventory 扫描 | 不修改 | 5.713 秒，100 行 | 完成 |
| Q22 | grouping-set partial 的哈希域被错误地从单批 sentinel 内容推断 | 从 producer 计划显式传递 grouping-aware 哈希域；partial 只验证、不决定 | 首次 10.626 秒报错 → 78.422 秒成功，100 行 | 完成；覆盖多 partial 与 spill merge |
| Q23 | `frequent_ss_items`、`best_ss_customer` 及其标量上界在 catalog/web 分支间重复构造 | 共享可证明完整消费的 hash-SEMI build-side CTE；物化仅保留消费键；固定物理 build side | Q23a 1502.293 秒失败 → 1134.065 秒成功，1 行；Q23b 1185.441 秒，59 行 | 完成；`store_sales` 扫描 6 → 3 |
| Q24 | 同一 `ssales` CTE 在颜色分支和全量平均分支各计算一次 | 不共享：全量物化会丢失颜色下推并写出宽、高基数聚合结果 | Q24a 139.424 秒，179 行；Q24b 132.679 秒，57 行 | 完成；无修改 |
| Q25 | 三张事实表复合键连接 | 日期过滤均下推，小表/build side 和最终聚合合理；无修改 | 62.629 秒，100 行 | 完成 |
| Q26 | 单次 `catalog_sales` 扫描与四个维表过滤 | 年、人口属性和 promotion 条件均提前；无修改 | 44.976 秒，100 行 | 完成 |
| Q27 | `ROLLUP(i_item_id,s_state)` 传统展开会重复扫描三次 | 已命中 grouping-set 公共输入共享，`store_sales` 只扫描一次 | 111.151 秒，100 行 | 完成；Q14 修复的独立正向样本 |
| Q28 | 六个同表标量聚合独立扫描 `store_sales` | 记录；正确方向是通用的 scalar aggregate fusion，不做 SQL 形状特判 | 269.515 秒，1 行；读取 172.80 亿行 | 完成；后续机制优化 |
| Q29 | Q25 同类三事实表复合键连接 | 日期过滤和高选择性 returns 路径均合理；无修改 | 61.434 秒，100 行 | 完成 |
| Q30 | 相关标量平均已去相关，但 CTE 仍 inline 两次 | 记录；`web_returns` 规模小且整体仅 6.134 秒，不扩大 CTE 共享边界 | 6.134 秒，100 行 | 完成；无修改 |
| Q31 | `ss`/`ws` 两个 CTE 各被三个季度分支消费 | 已命中 predicate-aware CTE reuse，两张事实表各扫描一次 | 58.389 秒，297 行 | 完成；CTE reuse 独立正向样本 |
| Q32 | 按 item 相关的平均折扣 | 已去相关，并将 manufacturer/date 条件传入两个分支；无修改 | 15.357 秒，1 行 | 完成 |
| Q33 | store/catalog/web 三渠道独立月聚合 | 日期、时区和 Books manufacturer SEMI 过滤均正确；无修改 | 45.544 秒，100 行 | 完成 |
| Q34 | 票据/客户分组后的客户筛选 | 日期、家庭人口和门店过滤均提前；先聚合、HAVING，再关联 customer | 53.446 秒，133518 行 | 完成；无修改 |
| Q35 | `OR` 下两个 `EXISTS` 保留为并行大事实表 MARK build，总内存峰值失控 | 同一外层等值键的正向过滤 `EXISTS OR` 合并为 `UNION ALL` key 输入上的一次 SEMI join | 原 SQL OOM → 66.883 秒，100 行 | 完成；RSS 约 6.3 GiB |
| Q36 | 两级 `ROLLUP` 与窗口排名 | 已命中 grouping-set 公共输入共享；事实表只扫描一次，无修改 | 99.953 秒，100 行 | 完成；无 OOM |
| Q37 | item 与 inventory/catalog 两条事实路径会合 | item 高选择性过滤先进入 catalog 路径，日期/库存过滤先进入 inventory 路径；无修改 | 15.368 秒，7 行 | 完成；无 OOM |
| Q38 | `INTERSECT` 为每个 distinct key 申请一整块 selection buffer，却只用一个标志位 | 改为每个 key 一个布尔状态；不改计划和语义 | 原执行 OOM → 206.927 秒，1 行 | 完成；5.5 GiB 无效堆分配消失 |
| Q39 | 两次月份自连接引用同一聚合 CTE | 已命中 predicate-aware CTE reuse，只计算 4/5 月并扫描一次 inventory；无修改 | Q39a 15.270 秒，10190 行；Q39b 15.311 秒，251 行 | 完成；无 OOM |
| Q40 | preserved-side 维表过滤被挡在 LEFT JOIN 之后 | 将唯一键 INNER JOIN 重排到 LEFT JOIN preserved side；语义安全不依赖 stats | 230.962 → 17.643 秒，100 行 | 完成；输出不变 |
| Q41 | item 自相关标量 COUNT 与复杂 OR | 已去相关为按 manufacturer/条件分组的小表 LEFT JOIN；无修改 | 0.030 秒，0 行 | 完成 |
| Q42 | 单次事实扫描与月份/manager 过滤 | 两个维表条件均在聚合前通过主键 join 生效；无修改 | 36.919 秒，12 行 | 完成 |
| Q43 | 七个 weekday 条件聚合 | 已在 store join 前按 store 做 partial aggregate；无修改 | 89.009 秒，100 行 | 完成 |
| Q44 | 升序/降序排名复制同一事实聚合与标量平均 | 记录为通用子计划 CSE/多窗口共享候选；当前不特判 | 140.807 秒，10 行 | 完成；四次事实扫描 |
| Q45 | ZIP 条件与小 item 集合 OR | item 子查询为 10-key MARK build，日期过滤提前；无修改 | 9.079 秒，100 行 | 完成 |
| Q46 | 票据聚合后关联客户当前地址 | 日期/门店/人口过滤在票据聚合前，客户宽列在聚合后；无修改 | 102.353 秒，100 行 | 完成 |
| Q47 | 同一聚合窗口 CTE 在 lag/current/lead 三个位置完整展开 | 对三消费者、完整消费且有 2 倍成本安全余量的 producer 开放有界 spill reuse；支持安全嵌套物化 | 844.957 → 250.868 秒，100 行 | 完成；扫描量降为 1/3 |
| Q48 | DNF 被误判为单表条件，公共人口属性等值键未被提出，退化为非等值 join | 完整收集 DNF 中引用的 relation；仅对真正单表 DNF 保留原形，跨表 DNF 正常提出公共条件 | 1GB 227.546 → 0.231 秒；1TB 86.605 秒，1 行 | 完成；不依赖 stats |
| Q49 | 三渠道 sales/returns 复合键连接与双排名 | nullable-side return amount 条件已将 LEFT JOIN 安全转为 INNER JOIN；无修改 | 92.743 秒，100 行 | 完成；无 OOM |
| Q50 | 销售和退货按票据/item/customer 做复合键连接 | 一个月退货日期过滤先进入连接，最终只在门店粒度聚合；无修改 | 57.744 秒，100 行 | 完成；无 OOM |
| Q51 | 累计 MIN/MAX 因缺少 source-preserving merge 契约而为每行重扫 partition prefix | 为 MIN/MAX 声明并验证只读 source 的 merge 契约，复用窗口已有 running aggregate 快路径 | 旧执行 655.073 秒时仍未完成并取消；修复后 733.427 秒完成，100 行 | 完成；复杂度由二次降为线性 |
| Q52 | 单次事实扫描与月份/manager 过滤 | 两个维表过滤均通过主键 join 在聚合前生效；无修改 | 20.023 秒，100 行 | 完成 |
| Q53 | 事实过滤后按季度聚合并计算 manufacturer 窗口平均 | window 位于低基数聚合后；无修改 | 30.222 秒，100 行 | 完成 |
| Q54 | catalog/web 客户集合驱动后续 store revenue 分段 | 日期/item 先过滤客户集合，后续事实连接和聚合资源安全；无修改 | 55.165 秒，100 行 | 完成 |
| Q55 | Q52 同类单次事实扫描 | 日期和 manager 过滤均提前；无修改 | 23.541 秒，100 行 | 完成 |
| Q56 | store/catalog/web 三渠道月聚合 | 日期、时区和 item 集合均在渠道聚合前生效；无修改 | 43.369 秒，100 行 | 完成 |
| Q57 | 同一窗口 CTE 被 lag/current/lead 三次引用 | 已命中 Q47 的三消费者 CTE 共享，只构造一次 producer | 156.025 秒，100 行 | 完成；CTE reuse 独立正向样本 |
| Q58 | 三渠道单周 item revenue 聚合 | 每个渠道只扫描一次，周集合在事实连接前构造；无修改 | 30.534 秒，100 行 | 完成 |
| Q59 | 全历史周聚合后，两个消费者只取相邻两年 | 记录可传播 week domain 的后续方向；当前一次事实扫描且资源安全，不扩大修改 | 330.993 秒，100 行 | 完成；后续机制优化 |
| Q60 | Q56 同类三渠道月聚合 | 日期、时区和 category 条件均提前；无修改 | 53.487 秒，100 行 | 完成 |
| Q61 | promotional/total 两个标量聚合重复相同事实路径 | 记录为通用 conditional aggregate/subplan CSE 候选；当前资源安全 | 73.292 秒，1 行 | 完成；后续机制优化 |
| Q62 | web shipping 时延条件聚合 | 已在 web_site join 前做 partial aggregate；无修改 | 45.151 秒，100 行 | 完成 |
| Q63 | Q53 同类聚合后窗口 | window 输入已缩到 manager/month 粒度；无修改 | 35.099 秒，100 行 | 完成 |
| Q64 | `cross_sales` 两次引用且内部依赖 `cs_ui` | 已命中 Q47 的安全嵌套 CTE 共享；两个 CTE 各构造一次，年份上界进入 producer | 145.852 秒，7094 行 | 完成；无 OOM |
| Q65 | 相同 store/item 月聚合被两个派生表各执行一次 | 记录通用非 CTE 子计划共享候选；当前两路 spill 仍在资源预算内 | 272.469 秒，100 行 | 完成；后续机制优化 |
| Q66 | web/catalog 两渠道 24 个条件聚合 | 两个事实分支均先按 year/warehouse 聚合，再连接宽 warehouse；无修改 | 81.115 秒，20 行 | 完成 |
| Q67 | 9 级 ROLLUP 将每条明细扩成 9 个 grouping row，再对宽字符串 key 做大聚合 | 现有 grouping-set 共享只消除了重复扫描，未消除明细级 9 倍 hash/spill 放大 | 1202.352 秒时仍未完成并取消；spill 峰值约 112 GB | 阻塞；需要通用分层 rollup/聚合机制 |
| Q4 | 共享前重复扫描；共享后仍把大量事实行直接送入 customer join 和最终聚合 | 复用一个可 spill producer；在有主键证明且成本显著下降时加入事实侧 partial SUM | 1176.530 → 588.892 秒，100 行 | 两阶段结构优化完成；spill 效率待优化 |

## Q1：统一 spill 阈值

### 问题

Q1 的主要风险是聚合在单机内存预算内不能足够早地 spill。系统变量默认值 `0` 表示按机器内存、文件缓存和 `GOMAXPROCS` 自动计算；这使不同机器上的实际阈值不够直观。没有发现需要为 Q1 增加专用 planner rewrite 的证据。

### 修改

不修改 SQL、执行计划或生产代码。将 `agg_spill_mem`、`join_spill_mem` 和 `sort_spill_mem` 统一持久化为 128 MiB，作为当前 1TB 单机测试的通用资源策略。Q1 实际使用的是 aggregate spill；另外两个变量保持相同值，避免后续查询逐条设置 session 参数。

### 阈值实验

所有成功运行的 `rows_read`、`bytes_scan` 和结果行数相同。

| 阈值 | statement ID | 耗时 | 结论 |
|---:|---|---:|---|
| 64 MiB | `01a03ce9-805e-7ae7-b0be-f5c588263c83` | 31.734 秒 | spill 更频繁，耗时明显增加 |
| 128 MiB | `01a03d2a-99f1-70d2-9939-bdf553eb7078` | 21.411 秒 | 最快，保留足够内存余量 |
| 172 MiB | `01a03d2e-30e6-727a-93b3-83011f896667` | 21.667 秒 | 无收益；hash 容量按离散档位增长 |
| 256 MiB | `01a03d3a-0aa0-7683-925f-d7b642234411` | 21.868 秒 | 无收益，增加资源峰值风险 |

### 结论

128 MiB 是当前机器上的合理通用阈值，不是 Q1 特判。后续查询若证明该阈值系统性不合适，应调整资源策略或自动阈值公式，而不是为单条 SQL 覆盖变量。

## Q2：复用包含普通 CTE 的外层 CTE

### 问题

Q2 中：

1. `wscs` 合并 `web_sales` 和 `catalog_sales`；
2. `wswscs` 引用 `wscs`，完成 date join 和按周聚合；
3. 主查询两次引用 `wswscs`。

已有 CTE reuse 因 `wswscs` 内部引用了另一个 CTE，直接用 `hasNestedRef` 拒绝复用。结果是整个 `web_sales UNION ALL catalog_sales -> date_dim join -> aggregate` 生产者被 inline 两次。这不是 stats 或 join order 误差，而是一个过度保守的 reuse 准入条件。

### 修改

- 删除“外层 CTE 内部引用另一个 CTE 就禁止复用”的 blanket guard。
- 外层 CTE 仍必须满足现有条件：多次引用、确定性、非相关、所有消费者完整消费、输出一致、物化结果在内存上限内，并且物化成本优于重复计算。
- 被绑定在另一个 CTE 内部的消费者只有在最终 rewrite root 可达时才允许复用；独立 step 中的 occurrence 继续 fail closed。
- 已受保护的内层物化可作为确定性外层 producer 的输入，因此内外两层可同时共享，不会因先选择内层而迫使更大的外层重复执行。
- 递归 CTE、非确定表达式、`LIMIT`/`EXISTS` 等部分消费场景仍拒绝复用。

修改是基于 CTE 语义和成本的不变量，不包含 Q2、表名、年份或 TPC-DS 特判。

### 计划变化

修复前有两份相同的大型生产者，每份都扫描一次 `web_sales` 和 `catalog_sales`。修复后：

- Plan 0 只执行一次 `web_sales UNION ALL catalog_sales -> date_dim join -> aggregate`；
- Plan 1 的两个 `wswscs` 消费者通过两个 `Sink Scan` 读取同一个 Plan 0；
- 两张大表都从扫描两次降为一次。

### 1TB 实测

| 指标 | 修复前 | 修复后 | 变化 |
|---|---:|---:|---:|
| statement ID | `01a03cea-dd7f-74a1-98cc-033cfe1e66bd` | `01a03d4f-43f5-7f24-a5cc-4823fb8792b3` | — |
| 耗时 | 1980.838 秒 | 855.284 秒 | -56.8% |
| `rows_read` | 4,320,253,780 | 2,160,199,939 | 约 -50% |
| `bytes_scan` | 51,845,382,928 | 25,923,275,856 | 约 -50% |
| `result_count` | 2513 | 2513 | 一致 |

总扫描量不是数学上的精确二分之一，因为共享生产者之外的 `date_dim` 消费仍各自存在；两张大型事实表的重复扫描已经消除。运行期间 RSS 约为 3.5–5.0 GiB，没有 OOM 趋势。

### 回归保护

- 白盒 planner 测试：包含普通 CTE 的昂贵外层 CTE 只产生一个 producer 和两个 `Sink Scan`。
- 最近反例：绑定在另一个 CTE 内的消费者仍 inline；包含递归 CTE 的 producer 不使用普通 CTE materialization。
- 黑盒执行测试：检查计划只有一次基础表扫描、两个共享消费者，并验证查询结果。
- 完整 `pkg/sql/plan` 测试、定向执行测试和 `go vet` 均通过。

### 剩余问题

Q2 仍需约 14 分钟，执行期间只使用约 1.2 个 CPU 核。当前提交解决的是重复执行；扫描和表达式计算吞吐属于下一阶段问题，不应与本次 CTE reuse 修复混在一起。

## Q3：无需修改

普通 `EXPLAIN` 未发现重复大型子计划、危险 build side 或明显缺失的选择性条件。1TB 实跑成功：statement ID `01a03d70-373c-7c7b-9c3d-645e8a7f49c8`，耗时 17.214 秒，读取 2,880,361,048 行、扫描 46,091,484,572 字节，返回 100 行。运行期间无 spill、无 OOM，RSS 增量约 565 MiB。

## Q4：谓词感知的多消费者 CTE 复用

### 问题

`year_total` 由 `store_sales`、`catalog_sales` 和 `web_sales` 三个分支组成，主查询按渠道和年份引用六次。原机制只能在“完整物化”与“完整 inline”之间二选一：变量宽度输出和偏大的物化估计使 reuse 被拒绝，最终生成六份 producer。虽然常量折叠会在每份 producer 内裁掉两个无关渠道，三张事实表仍各扫描六次。

这不是 Q4 的 stats 或 join order 特例，而是缺少通用的参数化/谓词感知共享机制。

### 修改

- 从每个消费者收集只依赖该 CTE occurrence、确定性且可安全下推的局部谓词。
- 只选择“每个消费者都约束”的输出列，避免把仅部分消费者使用的 `year_total > 0` 提到 producer 后阻塞更深的日期下推。
- 将 occurrence tag 映射回 producer tag，以 `P1 OR ... OR Pn` 限定共享 producer；消费者保留原谓词，因此不会扩大任何消费者的结果。
- 只穿过普通 `FILTER` 和 `INNER JOIN` 收集谓词；外连接、相关子查询、volatile 表达式和部分消费路径继续 fail closed。
- blocking `SORT`/`AGG` 即使自身带 `LIMIT` 也会完整读取输入，因此允许共享；在 blocking operator 之前的 `LIMIT/OFFSET` 仍拒绝。
- 小型定宽输出沿用原 32 MiB 内存准入。谓词感知 producer 可使用现有的 64 MiB 内存有界 materialization 并 spill；planner 仍以 8 GiB 估算上限拒绝明显失控的 spool，运行时每个 spill 字节和 FD 继续受 statement/CN resource budget 控制。

修改不检查 query 编号、表名、渠道值或年份。

### 计划变化

同一份 1TB stats、同一普通 `EXPLAIN`，基线为父提交 `484d9815eb`：

| 指标 | 修复前 | 修复后 |
|---|---:|---:|
| 总 `Table Scan` | 54 | 9 |
| `store_sales` 扫描 | 6 | 1 |
| `catalog_sales` 扫描 | 6 | 1 |
| `web_sales` 扫描 | 6 | 1 |
| `Sink Scan` | 0 | 6 |

修复后 Plan 0 只包含一份三渠道 producer，`d_year IN (2001, 2002)` 的等价 OR 条件下推到三个 `date_dim` 扫描；Plan 1 的六个消费者保留各自渠道、年份和 `year_total` 条件。

相同机制还将 Q11、Q74 的总表扫描分别从 24 降到 6，`store_sales`/`web_sales` 都从各扫描四次降为一次，并各生成四个 `Sink Scan`。Q2 的既有计划保持不变（5 个表扫描、2 个 `Sink Scan`）。这是跨查询结构命中，不是 Q4 特判；Q11/Q74 尚未实跑，本文不声称其运行时间已经改善。

### 1TB 实测

| 指标 | 结果 |
|---|---:|
| statement ID | `01a03d97-e50e-7567-a18d-2ff45248bf01` |
| 状态 | Success |
| 耗时 | 1176.530 秒 |
| `rows_read` | 5,076,187,938 |
| `bytes_scan` | 208,566,344,217 |
| `result_count` | 100 |
| 观察到的 MO RSS 峰值 | 约 8.9 GiB |
| 观察到的临时磁盘增长峰值 | 约 74 GiB |

在 31 GiB 单机上没有 OOM，spill FD 从峰值 3072 逐步归零，临时文件随查询结束回收。当前结构性重复扫描已消除；19 分钟耗时和较大的 join spill IO 放大仍需在后续执行器/物理计划工作中处理，不能把“成功跑完”误写为性能问题全部解决。

### 第二阶段：主键维表连接前的 partial SUM

CTE 复用后，每个渠道分支仍先把日期过滤后的事实行连接 `customer`，再按年份和多个 customer 展示列聚合。这里可以安全地先按“事实侧分组列 + customer 连接键”计算 partial SUM，再连接 customer，最后保留原聚合合并展示列相同但主键不同的维表行。

改写只在以下条件同时满足时发生：

- `INNER JOIN` 的维表侧是直接 `TABLE_SCAN`，等值条件完整覆盖真实主键；
- 所有聚合都是非 `DISTINCT SUM`，参数只依赖事实侧且不含 volatile 函数；
- grouping sets 未启用；分组表达式不会跨连接两侧；
- 所有 NDV 已知，并且 partial aggregate 的估算输出小于原 join 输出的一半。

正确性由连接约束和保留最终聚合证明，stats 只决定是否值得改写。成本估算也不直接用维表行数截断事实键 NDV：匹配键最多为维表行数，未匹配事实行仍作为可能的不同键计入上界，避免在维表高选择性时错误低估 partial aggregate。

普通 `EXPLAIN` 显示 `store_sales` 和 `catalog_sales` 分支各增加一个事实侧 partial aggregate，`web_sales` 因估算缩减不足而保持原计划。相同规则也命中 Q11 的 `store_sales` 分支；Q74 使用 `MAX`，按设计不改写。

| 指标 | 第一阶段 | 第二阶段 | 变化 |
|---|---:|---:|---:|
| statement ID | `01a03d97-e50e-7567-a18d-2ff45248bf01` | `01a04185-2928-7e7a-b301-3dcf9f0e0e9e` | — |
| 耗时 | 1176.530 秒 | 588.892 秒 | -49.9%，约 2.00 倍加速 |
| `rows_read` | 5,076,187,938 | 5,076,187,938 | 不变 |
| `bytes_scan` | 208,566,344,217 | 208,566,344,217 | 不变 |
| `result_count` | 100 | 100 | 一致 |

回归测试覆盖真实/伪造/复合主键、外连接、未知 NDV、未匹配事实键、低收益、`DISTINCT SUM` 和非可分解聚合等反例。嵌入式执行测试使用两个主键不同但展示属性相同的维表行，确认最终聚合仍产生精确结果，防止把“展示列看似唯一”误当成约束。

最终性能复测在启动后显式执行 `merge switch off`，并确认 merge 队列为空。另一次 823.273 秒的运行与多轮事实表 auto merge 重叠（包括 7.1 GiB 的 `catalog_sales` merge），因此只作为干扰样本，不用于比较优化收益。

### 回归保护

- 白盒计划测试覆盖：变量宽度 producer、两个不同且重叠的消费者谓词、Top-N `SORT LIMIT`，断言基础表只扫描一次且有两个消费者。
- 黑盒执行测试同时断言 producer 上的谓词并集、单次基础表扫描、两个 `Sink Scan` 和精确结果。
- 反例覆盖：存在无谓词消费者、volatile 消费者谓词、提前 `LIMIT/OFFSET`、半连接/相关消费等场景继续 inline。

## Q5：ROLLUP 重复完整输入，暂不修改

普通 `EXPLAIN` 中，`ROLLUP(channel, id)` 被展开为三个 `UNION ALL` 分支；每个分支都重新执行 `ssr`、`csr`、`wsr`，因此六张大事实表以及 `web_sales`–`web_returns` 连接均重复三次。这是 grouping sets 缺少共享输入的通用机制问题，不是 join order 或单表 stats 特例。

安全修复需要让多个 grouping-set aggregate 共享一次确定性输入，同时增加物化收益、输出大小、spill 和非确定性表达式的准入证明。它会跨越 binder、计划 DAG 和执行期物化，不属于本轮可直接落地的小范围修改，因此只记录，不为 Q5 增加 SQL、表名或 `ROLLUP(channel,id)` 特判。

原 SQL 在 `agg_spill_mem`、`join_spill_mem`、`sort_spill_mem` 均为 128 MiB、auto merge 关闭的隔离实例上成功完成：

| 指标 | 结果 |
|---|---:|
| statement ID | `01a0419a-f807-7fa3-ad75-1ff878db55b1` |
| 耗时 | 250.914 秒 |
| `rows_read` | 18,792,056,523 |
| `bytes_scan` | 425,952,506,304 |
| `result_count` | 100 |
| 观察到的 MO RSS 峰值 | 约 16.7 GiB |

执行早期 RSS 从约 5.0 GiB 升至 16.7 GiB，随后回落并稳定完成，未发生 OOM。扫描放大和较高内存峰值保留为 grouping-set 输入共享的后续设计依据。

## Q6：无相关标量条件提前放置

普通 `EXPLAIN` 将 `d.d_month_seq = (select distinct d_month_seq ...)` 保留为大连接之上的 `FILTER + SINGLE JOIN`。因此 `date_dim` 连接 `store_sales` 时没有月份谓词，后续 customer/address/item 连接也先处理了本可提前裁掉的行。

标量子查询实际返回 `1201`；仅用于诊断的普通 `EXPLAIN` 将它替换为字面量后，`d.d_month_seq = 1201` 会立即进入 `date_dim` 的 filter 和 block filter，证明瓶颈来自谓词位置，而不是 stats。不能据此在产品代码中常量化查询结果。

通用修复将“原谓词 `FILTER` + 无相关 `SINGLE JOIN`”作为整体，穿过只依赖一侧的 `INNER JOIN`，以及 `SEMI/ANTI JOIN` 的保留左侧，直到最小相关输入。`SINGLE JOIN` 仍负责标量子查询多行报错，原 `FILTER` 仍负责 SQL 三值逻辑；相关、非确定、right-SINGLE、outer join、`LIMIT/OFFSET` 等边界全部拒绝改写。

改写是否发生只由表达式依赖和语义边界决定，不使用 stats。stats 只在改写后参与 build/probe、shuffle 和 runtime filter 等物理选择。当前普通 `EXPLAIN` 已把月份标量条件放到 `date_dim` 输入，再连接 `store_sales`；Q6 的 1TB 性能尚未复测，因此保留原运行结果作为修复前基线。

| 指标 | 结果 |
|---|---:|
| statement ID | `01a0419f-dede-7881-8204-966b6076b522` |
| 耗时 | 813.478 秒 |
| `rows_read` | 2,898,734,097 |
| `bytes_scan` | 34,845,716,968 |
| `result_count` | 52 |
| 观察到的 MO RSS 峰值 | 约 10.3 GiB |

原 SQL 成功完成且没有 OOM，但持续约 13.6 分钟的 CPU 处理说明该放置规则值得后续独立设计。

## Q7：计划合理，无需修改

普通 `EXPLAIN` 中，年份、人口属性和 promotion 渠道条件都位于维表扫描侧；`store_sales` 依次通过四个维表主键连接过滤，再按 item 聚合。没有重复大型子计划、错误 build side 或无界聚合。

1TB 原 SQL 实跑成功：statement ID `01a041ad-edea-7835-91ed-1108e0315cf6`，耗时 122.506 秒，读取 2,882,283,348 行、扫描 126,874,515,148 字节，返回 100 行。观察到的 MO RSS 峰值约 8.8 GiB，没有 OOM。虽然 `AVG` 可分解为 partial sum/count，但当前计划没有足够证据支持为它扩大本轮改动范围。

## Q8：计划合理，无需修改

`store_sales` 通过 2002 年第一季度的 `date_dim` 过滤；customer/address 子计划先完成 IN 过滤、聚合和 INTERSECT，再作为小 build side 参与邮编前缀连接。最终只按 store name 聚合，没有大型高 NDV 状态或重复子计划。

1TB 原 SQL实跑成功：statement ID `01a041b0-77a7-714a-92d5-2ef1b1fb7eaf`，耗时 32.712 秒，读取 2,904,062,050 行、扫描 46,728,736,676 字节，返回 11 行；没有 OOM。无需修改。

## Q9：同源标量聚合扫描放大，暂不修改

五个数量区间各有 `count(*)`、`avg(ss_ext_list_price)`、`avg(ss_net_paid_inc_tax)` 三个无相关标量子查询。普通 `EXPLAIN` 为它们生成 15 个独立 `store_sales` 扫描，未合并相同 FROM/filter 的聚合，也未进一步用条件聚合共享全表扫描。

正确的通用修复应在无相关标量子查询层做语义等价识别和聚合融合，并处理不同类型、空输入、错误/volatile 表达式和标量 cardinality；按五个 bucket 或当前 CASE 文本改写会过拟合。这不是小范围修复，因此只记录。

1TB 原 SQL 成功完成：statement ID `01a041b1-e0e7-723a-81d2-2a97ef479f48`，耗时 253.277 秒，读取 43,199,819,986 行、扫描 403,198,319,864 字节，返回 1 行。MO RSS 峰值约 10.5 GiB，没有 OOM；约 15 倍 `rows_read` 直接量化了扫描放大。

## Q10：EXISTS 的等值 MARK join 恢复 hash 路径

普通 `EXPLAIN` 的 join order、日期过滤和维表过滤看起来合理，但原计划把两个 `EXISTS` 对应的 MARK join 条件生成成 `(probe_key = build_key) IS TRUE`。物理编译器只识别顶层等式为 hash key，因此退化为 loop join。原查询运行超过 26 分钟仍未完成；20 秒 CPU profile 中 `loopjoin.(*container).scanProbeMatches` 累计占 94.43%，证明瓶颈不是对象扫描、stats 或 spill，而是等值条件的表达式形态。

修复仅针对 `EXISTS`/`NOT EXISTS` 的二值语义：MARK join 保留裸等式，使物理层可选 hash；在 marker 输出处应用 `IS TRUE`，`NOT EXISTS` 再取反。`IN`/`NOT IN`/`ANY`/`ALL` 继续保留原有三值逻辑和 nullable marker，不共享这条改写。filter pushdown 同时识别 `IS TRUE(marker)` 和 `NOT(IS TRUE(marker))`，分别转换为 SEMI/ANTI join。

修复后的普通 `EXPLAIN` 显示三个存在性 join 条件均为裸等式。10 秒 CPU profile 中旧 `loopjoin` 热点完全消失，主要路径变为 `hashjoin.(*container).probe`。1TB 原 SQL成功完成：statement ID `01a041dc-ec42-7dbd-b1c3-e17184ba5bf8`，耗时 52.226 秒，读取 5,060,108,738 行、扫描 40,857,666,557 字节，返回 100 行；运行时 RSS 约 4.5 GiB。

Q35 引入的通用 `OR-of-EXISTS` 规则也独立命中 Q10：web/catalog 两个 MARK build 被 `Union All -> RIGHT SEMI` 替代，store 分支仍保持 SEMI。1TB 原 SQL statement ID `01a043f1-4302-7544-87de-f34823cc8933`，成功耗时 29.271 秒；`rows_read=5,060,108,738`、`bytes_scan=40,857,666,557`、结果 100 行均与 52.226 秒版本一致。这说明新规则不是 Q35 的文本特判，并在已有可运行计划上继续获得 44.0% 的耗时下降。

回归测试同时覆盖 projected `EXISTS`、两个 `EXISTS` 的 OR、`NOT EXISTS`、nullable probe/build key，并断言精确结果和 EXPLAIN 的 MARK join 条件不含 `IS TRUE`。既有 `IN`/`NOT IN` 三值逻辑测试作为反例保持通过；`pkg/sql/plan` 全量测试和 `go vet` 通过。

## Q11：共享 CTE 和 partial SUM 已命中

Q11 与 Q4 同属多消费者渠道年度汇总。普通 `EXPLAIN` 中 `year_total` 只生成一次，`store_sales`、`web_sales` 各扫描一次，四个引用均为 `Sink Scan`；`store_sales` 分支还命中事实侧 partial SUM，`web_sales` 因收益门槛未命中。没有重复大型 producer、危险 build side 或无界内存状态。

1TB 原 SQL成功完成：statement ID `01a041e5-aeff-7478-b595-7fe670a3ebf4`，耗时 254.265 秒，读取 3,624,134,473 行、扫描 91,044,782,718 字节，返回 100 行；观察到的 MO RSS 峰值约 8.8 GiB。无需新增修改。

## Q12：计划合理，无需修改

`d_date` 的 30 天范围和 item category 条件均已下推到维表扫描；事实表依次通过主键连接过滤，窗口函数发生在按 item 聚合之后，最终 Top-N 只处理聚合结果。1TB 原 SQL成功完成：statement ID `01a041e9-f9a7-7dfd-856e-e26d64749231`，耗时 10.827 秒，读取 720,308,568 行、扫描 11,582,154,968 字节，返回 100 行。无需修改。

## Q13：OR 候选过滤已下推，无需修改

planner 从两个跨表 OR 组中提取了安全的单表候选条件：年份、销售价格和净利润范围进入 `store_sales`/`date_dim`，州、婚姻教育状态和家庭依赖人数分别进入维表扫描；必须保持分支相关性的条件仍在 join 上求值。该分解不会把一个分支的条件误配到另一个分支。

1TB 原 SQL成功完成：statement ID `01a041ea-88d1-7c5b-9a98-be5a864cbb71`，耗时 169.422 秒，读取 2,886,537,174 行、扫描 161,616,305,992 字节，返回 1 行；RSS 稳定约 6 GiB。无需修改。

## Q14：grouping-set 公共输入与标量条件放置

Q14 文件包含 Q14a、Q14b 两条独立 SQL，暴露两个正交问题。

### Q14a：共享 grouping-set 公共输入

原 binder 将四列 `ROLLUP` 展开成五个完整 `UNION ALL` 分支，使 `store_sales`、`catalog_sales`、`web_sales` 各扫描 21 次，共 63 次事实表扫描。通用修复在展开前只绑定一次确定性公共输入，由 grouping-aware aggregate 消费；correlated、非确定性和不支持的 distinct/order 形状 fail closed。普通 `EXPLAIN` 中事实表扫描从 63 次降到 15 次。

1TB 实跑成功：statement ID `01a04264-0247-746e-b769-eb71830d5701`，耗时 2097.510 秒，读取 25,205,439,690 行、扫描 302,479,769,928 字节，返回 100 行。运行无 OOM；前约 10 分钟与一项重启前已排队的后台 merge 重叠，因此耗时不是干净性能基线，正确性和完成性有效。merge 结束后的 CPU profile 主要热点为 `INTERSECT`、grouping key 填充和 hash join，没有 grouping expansion 状态泄漏。

### Q14b：把标量周条件移到日期输入

Q14b 没有 grouping sets，因此不受上一项修改影响。原计划把：

```sql
d_week_seq = (select d_week_seq from date_dim where ...)
```

flatten 成整个事实表/item/`cross_items` 连接之上的 `FILTER + SINGLE JOIN`。两个分支都先让整张 `store_sales` 穿过多层 join，最后才缩到一周。20 秒 CPU profile 显示 15 个核满载在 hash/loop join 的逐行结果复制；它不是 SSD 等待或 spill 卡住。

通用修复保持 `FILTER + SINGLE JOIN` 为一个语义单元，把它移到只提供 `d_week_seq` 的 `date_dim` 输入。触发不读取 stats；正确性仍由原 `SINGLE JOIN` 的多行检查和原过滤谓词保证。Q6 独立命中同一规则，证明机制不含 Q14、表名、年份或 TPC-DS 特判。

| 指标 | 修复前 | 修复后 |
|---|---:|---:|
| statement ID | `01a04284-81f1-77a2-82df-3193788401e6` | `01a042ba-8b43-7137-9d60-0c6d8c4af30d` |
| 状态 | 运行 2932.157 秒仍未完成，手动取消 | Success |
| 耗时 | >2932.157 秒 | 1311.300 秒 |
| `rows_read` | 取消样本未记账 | 20,883,832,008 |
| `bytes_scan` | 取消样本未记账 | 276,535,287,152 |
| `result_count` | 0 | 100 |

新计划至少缩短 55.3%，并在 RSS 约 6.1 GiB、auto merge 关闭、统一 128 MiB spill 阈值下正常完成。1GB 的完整 Q14a+Q14b 输出与父版本基线逐字一致，SHA-256 均为 `a581011fcc880666be34f0dda56a8836443b0d71573701a7c7f9c3ca06228d8e`。

### 剩余问题：只记录，不在本轮修改

`cross_items` 仍在 Q14a 的三个渠道和 Q14b 的两个年度分支中重复计算。当前事实表扫描计数如下：

| 计划 | `store_sales` | `catalog_sales` | `web_sales` | 事实表扫描合计 | 共享 `cross_items` 后目标 |
|---|---:|---:|---:|---:|---:|
| Q14a | 5 | 5 | 5 | 15 | 9 |
| Q14b | 5 | 3 | 3 | 11 | 8 |

因此 Q14b 修复后仍读取 208.84 亿行、扫描 276.54 GB；剩余耗时主要是九次 `cross_items + avg_sales` 大扫描。后续应单独以确定性、多消费者完整读取、物化成本和 spill 预算为准入条件复用 `cross_items`，不能混入标量条件放置规则。

完整 TPC-DS 查询集共有 11 条 `ROLLUP` 查询，没有额外的 `CUBE`/显式 `GROUPING SETS`。除 Q5、Q14a 外，后续覆盖如下；“预计”表示公共输入只执行一次时的结构上界，最终仍由 stats 成本门槛决定是否启用：

| 查询 | 当前重复 | 共享公共输入后的直接变化 | 收益判断 |
|---|---:|---:|---|
| Q18 | `catalog_sales` 5 次 | 1 次 | 高 |
| Q22 | `inventory` 5 次 | 1 次 | 高 |
| Q27 | `store_sales` 3 次 | 1 次 | 中 |
| Q36 | `store_sales` 3 次 | 1 次 | 中 |
| Q67 | `store_sales` 9 次 | 1 次 | 很高；最强的后续独立样本 |
| Q70 | 外层和 IN 子查询随 3 个 grouping sets 各重复一次，`store_sales` 共 6 次 | 2 次 | 高 |
| Q77 | 六张事实表已由 CTE reuse 各扫描一次，但有 18 个重复 CTE `Sink Scan` | 基础表扫描不变，公共渠道输入只构造一次 | 较低但正向 |
| Q80 | 六张 sales/returns 事实表各 3 次，共 18 次 | 各 1 次，共 6 次 | 很高 |
| Q86 | `web_sales` 3 次 | 1 次 | 中 |

因此回归集不应只使用 Q14：Q5/Q80 覆盖多渠道输入，Q67 覆盖九级 ROLLUP，Q18/Q22 覆盖五级单事实输入，Q77 是“已有 CTE 共享、基础扫描不应变化”的控制组。

## Q15：计划合理，无需修改

普通 `EXPLAIN` 只有一次 `catalog_sales` 扫描；`d_year = 2000` 和 `d_qoy = 2` 已下推到 `date_dim`，`customer`、`customer_address` 和 `date_dim` 均通过主键等值连接。跨 `customer_address` 与 `catalog_sales` 的 OR 条件必须在两侧连接后求值，位置正确。计划没有重复大型子计划、危险 build side 或无界高基数状态。

1GB 输出与既有基线逐字一致，SHA-256 均为 `203de72dc18d063edfdbf40e06fe23e67c98a5faf6ab7170a41dda4e4dbf1014`。1TB 原 SQL 成功完成：statement ID `01a042dc-6514-7c25-9f96-a5feb3a39f8a`，耗时 14.664 秒，读取 1,458,053,465 行、扫描 23,448,563,244 字节，返回 100 行；没有 OOM。无需新增修改。

## Q16：SEMI/ANTI 连接计划合理，无需修改

普通 `EXPLAIN` 将相关 `EXISTS` 转为以 `cs_order_number` 为哈希键、仓库不等式为附加条件的 RIGHT SEMI JOIN，将 `NOT EXISTS` 转为同一订单号上的 RIGHT ANTI JOIN。两次 `catalog_sales` 扫描是原查询自连接语义所需；日期、州和 call center county 过滤均在事实表进入相关连接前生效。计划没有 loop join 退化，128 MiB join spill 阈值覆盖全量订单号 hash 的资源风险。

1GB 输出与既有基线逐字一致，SHA-256 均为 `36c779ee1a0c493e5be592e714ebcc3aec35d43a1db41e0a34b75905151881b4`。1TB 原 SQL 成功完成：statement ID `01a042dd-c39f-7940-aa7c-fffc6960ecf0`，耗时 29.091 秒，读取 3,029,965,822 行、扫描 64,103,192,040 字节，返回 1 行；没有 OOM。无需新增修改。

## Q17：多事实表连接计划合理，无需修改

普通 `EXPLAIN` 中三张事实表各扫描一次，三个季度条件均下推到对应的 `date_dim`；`store_sales` 与 `store_returns` 使用 customer/item/ticket 三列等值键，`store_returns` 与 `catalog_sales` 使用 customer/item 两列等值键，均为 hash join，没有笛卡尔积或 loop join。最终聚合前的多事实表连接是原查询语义所需，不能通过局部改写消除。

1GB 输出与既有基线逐字一致，SHA-256 均为 `71bcc2b88ee9dac50b8f1c68b97374c974e60d5e2f19b980d3be7b268fc262be`。1TB 原 SQL 成功完成：statement ID `01a042df-16a0-7900-95e3-35ac48b02085`，耗时 63.533 秒，读取 4,608,423,471 行、扫描 97,969,025,504 字节，返回 100 行；运行期间 RSS 稳定在约 5.7 GiB，没有 OOM。无需新增修改。

## Q18：grouping-set 公共输入共享的独立验证

Q18 是四列 `ROLLUP`，旧展开方式会产生五份完整的 `catalog_sales + 六张维表` 输入。当前普通 `EXPLAIN` 只有一个公共 producer、一处 grouping-aware aggregate 和五个按 grouping id 过滤的 `Sink Scan`；`catalog_sales` 从结构上的五次扫描降为一次。这是 Q14 之外的独立命中，说明修改针对 grouping-set 机制，而非查询特判。

1GB 输出与既有基线逐字一致，SHA-256 均为 `f79d3e2bfb029173b213ae18a275713257eaeacd11c33efb4fa35b02e3968d7e`。1TB 原 SQL 成功完成：statement ID `01a042e0-bc2f-72c0-a586-e28a57159f58`，耗时 72.750 秒，读取 1,462,195,065 行、扫描 75,699,687,689 字节，返回 100 行；运行期间 RSS 约 5.9 GiB，没有 OOM。

## Q19：计划合理，无需修改

普通 `EXPLAIN` 只有一次 `store_sales` 扫描；年月条件下推到 `date_dim`，manager 条件下推到 `item`，其余维表通过主键等值连接。邮编不等式同时依赖 customer address 和 store，保留在两侧连接完成后的 filter 上是正确位置。没有重复大型子计划或危险 build side。

1GB 输出与既有基线逐字一致，SHA-256 均为 `aae52e6fa8aac378e242ff4b33a0306411ff7088369523e1a62b8d24f2f61de8`。1TB 原 SQL 成功完成：statement ID `01a042e2-7279-7a64-85f5-3c7f4f6fd0f6`，耗时 38.028 秒，读取 2,898,362,050 行、扫描 69,403,816,620 字节，返回 100 行；没有 OOM。无需新增修改。

## Q20：事实侧预聚合已命中，无需修改

普通 `EXPLAIN` 中日期范围已下推；`catalog_sales` 在连接 `item` 前先按 `cs_item_sk` 计算 partial SUM，将事实行缩成 item 粒度，之后才完成分组和窗口计算。窗口位于聚合结果之上，只有一次事实表扫描，没有危险的明细级窗口或重复子计划。

1GB 输出与既有基线逐字一致，SHA-256 均为 `5428553bcf2115f6f7722340a6790db8513197cc30f395671ba3051f5ad94124`。1TB 原 SQL 成功完成：statement ID `01a042e3-9e22-76c1-891a-4cd62a49d9b7`，耗时 14.344 秒，读取 1,440,288,608 行、扫描 23,101,835,608 字节，返回 100 行；没有 OOM。无需新增修改。

## Q21：计划合理，无需修改

普通 `EXPLAIN` 只有一次 `inventory` 扫描；61 天日期范围和 item 价格范围分别下推到维表，warehouse/item/date 均使用主键 hash join。before/after 两个条件和比例过滤在一次分组聚合内完成，没有重复扫描。

1GB 输出与既有基线逐字一致，SHA-256 均为 `b19e3baf8ec40371d15ae6d10757d9da7007803b4d6636325c4513ddeec0a317`。1TB 原 SQL 成功完成：statement ID `01a042e4-7c3f-7abe-899a-46ae511d75d8`，耗时 5.713 秒，读取 783,308,212 行、扫描 12,538,866,096 字节，返回 100 行；没有 OOM。无需新增修改。

## Q22：修复 grouping-set partial 哈希域推断

普通 `EXPLAIN` 已命中 grouping-set 公共输入共享：五级 `ROLLUP` 只有一个 producer，`inventory` 从结构上的五次扫描降为一次。1GB 输出与既有基线逐字一致，SHA-256 均为 `c1346ca9ed630fb072b1107e0c9b0c2398d5e8e9820a319d07a58ae77c096da8`。

首次 1TB 执行在 10.626 秒失败：statement ID `01a042e5-4a96-7a18-adc5-be2299c57e1d`，错误为 `inconsistent merge-group partial metadata`。根因不是 SQL、stats 或内存不足，而是 MergeGroup 从每个 partial 的实际 grouping sentinel 位推断哈希域：完整 grouping set 的 partial 没有 sentinel，后续 rollup partial 才有，合法的同一数据流因此被误判为元数据变化。小数据若所有 grouping set 落在一个 partial 中不会触发，所以 1GB 未暴露。

修复由 producer 的静态 `GroupingFlag`/`DynamicGrouping` 语义显式设置 MergeGroup 的 grouping-aware 哈希域，并通过本地构造和远程 pipeline 编解码传递。partial 中的 sentinel 现在只用于验证：首批存在 sentinel 时仍可兼容性升级；一旦哈希表以普通域建立，后来无声明地出现 sentinel 仍 fail closed。定向单测覆盖“首 partial 无 sentinel、后 partial 有 sentinel”、普通 grouping partial、篡改 hash metadata 和跨远程编解码。

修复后 1TB 原 SQL 成功完成：statement ID `01a042ea-6d3b-78d7-98f0-ca76367ccbdf`，耗时 78.422 秒，读取 783,373,049 行、扫描 9,430,212,887 字节，返回 100 行；期间发生约 2.7 GB spill，RSS 峰值约 3.4 GiB，没有 OOM。这补齐了 grouping-set 共享在多 partial、并行 MergeGroup 和 spill 下的执行闭环。

## Q23：共享 hash-SEMI build side 上的多消费者 CTE

Q23 文件包含 Q23a、Q23b 两条独立 SQL，必须分别 EXPLAIN 和计时。两条计划具有同一结构：`frequent_ss_items`、`best_ss_customer` 及其依赖的 `max_store_sales` 在 catalog/web 两个 `UNION ALL` 分支中分别 inline；每条 SQL 因此包含六次 `store_sales` 扫描，而不是三个公共 producer 各执行一次。已有局部 `shuffle: REUSE` 只能复用单分支内部的 partial aggregate，无法跨 union 分支消除 producer。

1GB 完整 Q23 输出与既有基线逐字一致，SHA-256 均为 `0632e0054fbaf8210bd12b126867ff4cc9cc8e3a563380f297d923da5edde006`。1TB Q23a 受控运行期间 RSS 稳定在约 10–13 GiB，没有 OOM 或泄漏趋势，但重复的高基数聚合持续产生 spill。statement ID `01a042ee-446f-7620-894b-e783e2e9e251` 在 1502.293 秒失败，错误为：

```text
resource exhausted: hash build spill disk budget exceeded
(requested=65536, used=213587460096, limit=213587512528)
```

失败前 spill 已达 213.59 GB。Q23b 曾因最初把含两条 SQL 的文件只加一个 `EXPLAIN` 而被意外启动，56.117 秒即取消，statement ID `01a042ec-8692-7598-8d54-e6c9689e2672`；该次不作为执行结果。

修复扩展了 CTE reuse 的消费证明：仅当正向、非相关的等值 `IN`/`EXISTS` 必然转成 hash SEMI join，且 CTE reader 位于必须完整读取的 build side 时允许共享。该 reader 带有 planner 内部契约标记，后续成本选择不得将它翻到 probe side；物化容量估算使用最终消费者列的并集，Q23 因此只写入 `item_sk` 和 `c_customer_sk`，不把未消费的 varchar payload 算入共享 spool。`NOT IN`、非等值 `ANY`、probe-side 消费、`LIMIT`/`OFFSET`、相关子查询和非确定 producer 仍 fail closed。这是通用物理消费契约，没有 Q23 查询或表名特判。

修复后 Q23a/Q23b 的普通 `EXPLAIN` 均出现两个共享 producer：`frequent_ss_items` 一次计算、两次 sink scan，`best_ss_customer` 及其 `max_store_sales` 一次计算、两次 sink scan；`store_sales` 扫描数从 6 降为 3。1GB 完整 Q23 输出 SHA-256 仍为 `0632e0054fbaf8210bd12b126867ff4cc9cc8e3a563380f297d923da5edde006`。

1TB 实测均成功：

| 查询 | Statement ID | 耗时 | 读取行 | 扫描字节 | 结果行 |
|---|---|---:|---:|---:|---:|
| Q23a | `01a04380-6554-7093-ae2b-4ef6b8312e8e` | 1134.065 秒 | 10,824,536,985 | 178,696,308,536 | 1 |
| Q23b | `01a04392-4b86-7f4a-9b3e-6d502da78d95` | 1185.441 秒 | 10,824,537,044 | 178,696,311,604 | 59 |

Q23a 从 1502.293 秒后耗尽 213.59 GB 预算改为 1134.065 秒成功。两次运行均只观察到约 100 GB 的峰值临时磁盘增量，查询结束后回收到基线；RSS 低于约 12.5 GiB，没有 OOM。

## Q24：重复扫描是合理的计划取舍

Q24a/Q24b 只有颜色常量不同。普通 `EXPLAIN` 中 `ssales` 被 inline 两次：外层分支把 `i_color` 下推到 item 扫描，标量 `avg(netpaid)` 分支则必须读取全部颜色。因此 `store_sales` 和 `store_returns` 均扫描两次。

这里不应强制 CTE reuse：共享 producer 必须按姓名、商店、地址、颜色和多个 item 属性生成宽、高基数的全量聚合结果，再为外层分支过滤一种颜色。这既失去了现有颜色下推，又会增加大量 spool I/O；当前安全门槛拒绝该共享是正确的。

1TB 实测不需要修改：

| 查询 | Statement ID | 耗时 | 读取行 | 扫描字节 | 结果行 |
|---|---|---:|---:|---:|---:|
| Q24a | `01a043b2-2572-722c-a4b7-9e6722b6fc45` | 139.424 秒 | 6,372,577,530 | 145,732,380,496 | 179 |
| Q24b | `01a043b6-03a7-7c2f-94de-e2f2a96bee04` | 132.679 秒 | 6,372,577,530 | 145,732,380,496 | 57 |

执行期间没有观察到临时磁盘增长，RSS 约 6.5–10 GiB，无 OOM。当前 2.2–2.3 分钟主要是两次 1TB 事实表扫描和返回复合键连接，不存在需要小范围修复的危险计划。

## Q25：三事实表连接计划合理

普通 `EXPLAIN` 中 `store_sales`、`store_returns`、`catalog_sales` 均只扫描一次；三个 `date_dim` 的年/月条件分别在对应事实表连接前生效。`store_returns` 与 `catalog_sales` 先按 customer/item 缩减，再通过 customer/item/ticket 复合键与四月的 `store_sales` 连接，最后才按 item/store 聚合并做 Top-100。没有重复扫描、过晚过滤或明显危险的 build side。

1TB 原 SQL 成功：statement ID `01a043ba-19fd-7c5b-a607-b8fe2d6ce4e8`，耗时 62.629 秒，读取 4,608,488,328 行、扫描 116,399,231,912 字节，返回 100 行。查询初期临时磁盘增长约 5 GB 后立即回收，RSS 约 6.5–6.9 GiB，无 OOM。无需修改。

## Q26：单扫描维表过滤计划合理

普通 `EXPLAIN` 只扫描一次 `catalog_sales`，并依次与已过滤的 `date_dim`、`customer_demographics`、`promotion` 和 `item` 做主键等值连接。年份、人口属性和渠道条件都在事实行进入最终 item 聚合前生效，没有扫描或 join 放大。

1TB 原 SQL 成功：statement ID `01a043bc-4420-7aea-831c-70d4f12cd3ce`，耗时 44.976 秒，读取 1,442,275,765 行、扫描 63,514,181,496 字节，返回 100 行；无 OOM，无需修改。

## Q27：grouping-set 共享再次命中

Q27 的 `ROLLUP(i_item_id,s_state)` 需要三个 grouping level。普通 `EXPLAIN` 已命中 Q14 引入的 grouping-set 公共输入：Plan 0 扫描、连接和聚合一次，Plan 1 通过三个 sink scan 拆出对应 grouping level。`store_sales` 从传统展开的三次扫描降为一次，年、人口属性和州过滤均保持下推。

1TB 原 SQL 成功：statement ID `01a043bd-dd48-7222-833e-3666c4dbc44a`，耗时 111.151 秒，读取 2,882,282,850 行、扫描 126,874,465,204 字节，返回 100 行；无 OOM。这是 grouping-set 共享在 Q18 之后的又一个独立黑盒正向样本，无需新增修改。

## Q28：六次同表标量聚合扫描

普通 `EXPLAIN` 包含六个独立的 `store_sales -> scalar aggregate`，每个分支只是 quantity 区间和价格/coupon/cost 条件不同。计划安全：每个分支都是无 group key 的小聚合，没有 OOM 或 spill 风险；但结构上明确不经济，`store_sales` 被完整扫描六次。

1TB 原 SQL 成功：statement ID `01a043c0-aa3d-7c87-92a9-03823825d95c`，耗时 269.515 秒，读取 17,279,927,994 行、扫描 483,837,983,832 字节，返回 1 行；无临时磁盘增长，无 OOM。

当前不做小修：正确的通用机制是识别同一输入上的多个独立 scalar aggregate，用条件聚合状态在一次扫描中同时计算 `avg/count/count(distinct)`，并保留 NULL、distinct 和空输入语义。这与 Q9 的多标量同表扫描是同一类问题，需独立设计和反例矩阵，不应对 Q28 的六个范围做特判。

## Q29：Q25 同类三事实表连接

Q29 与 Q25 结构相同，只是日期范围不同。普通 `EXPLAIN` 中三张事实表均只扫描一次；四月 `store_sales`、四至七月 `store_returns` 和三年 `catalog_sales` 的路径均先利用更高选择性的 date/returns 条件，再通过 customer/item/ticket 复合键会合。

1TB 原 SQL 成功：statement ID `01a043c6-2aaf-73b5-a12d-c9c4c6b75595`，耗时 61.434 秒，读取 4,608,488,328 行、扫描 97,967,067,000 字节，返回 100 行；无 OOM，无需修改。

## Q30：去相关正确，重复扫描不值得扩大机制

普通 `EXPLAIN` 已将按 state 相关的 scalar `avg` 去相关为一个按 `ctr_state` 聚合的 LEFT JOIN，没有 per-row loop 子查询。`customer_total_return` 仍 inline 两次，因为其中一个 binder occurrence 原本是相关消费者，当前 CTE reuse 安全证明 fail closed。这使 `web_returns` 扫描两次，但没有形成危险计划。

1TB 原 SQL 成功：statement ID `01a043c8-8103-75bf-8ddf-6c83a1884c64`，耗时 6.134 秒，读取 174,141,142 行、扫描 6,235,016,131 字节，返回 100 行。若未来要复用这类 CTE，应在去相关后重新证明消费契约，而不是放宽原始相关 occurrence；当前 6 秒不值得扩大修改范围。

## Q31：predicate-aware CTE reuse 命中

`ss` 和 `ws` 均被三个季度分支引用。普通 `EXPLAIN` 对每个 CTE 各生成一个共享 producer，将三个消费者条件合并为 `2000 + Q1/Q2/Q3` 的安全上界，再由三个 sink scan 恢复各自季度。`store_sales` 和 `web_sales` 均只扫描一次；`store_sales` 路径还在 address join 前命中了 partial SUM。

1TB 原 SQL 成功：statement ID `01a043c9-e188-7f76-8ce1-93973ef65a5a`，耗时 58.389 秒，读取 3,612,134,473 行、扫描 57,938,514,506 字节，返回 297 行；无 OOM。这是当前 CTE 共享和 Q4 partial aggregate 两项机制的独立正向样本，无需修改。

## Q32：相关平均已去相关并传播过滤

普通 `EXPLAIN` 将按 `i_item_sk` 相关的 `avg(cs_ext_discount_amt)` 去相关为按 item 分组聚合的 LEFT JOIN。外层的 `i_manufact_id=269` 被传播为内层 catalog 路径的 item SEMI 过滤，90 天日期区间也在两个 `catalog_sales` 分支中均提前生效。虽然事实表扫描两次，但没有 per-item loop 执行或未约束分支。

1TB 原 SQL 成功：statement ID `01a043cb-f5e0-7730-9c44-f50af386b7d9`，耗时 15.357 秒，读取 2,880,577,216 行、扫描 46,084,304,384 字节，返回 1 行；无 OOM，无需修改。

## Q33：三渠道计划合理

普通 `EXPLAIN` 中 store/catalog/web 三个分支均只扫描各自事实表一次。`1999-03`、address GMT offset 和 Books manufacturer 条件均在渠道聚合前生效；manufacturer 子查询被转换为可哈希 SEMI join，没有 loop 执行。三个小聚合结果通过 `UNION ALL` 合并后再按 manufacturer 汇总。

1TB 原 SQL 成功：statement ID `01a043cd-44ef-7a77-844c-d7c7a5116dc2`，耗时 45.544 秒，读取 5,059,987,938 行、扫描 101,050,405,584 字节，返回 100 行；无 OOM，无需修改。

## Q34：先聚合再关联客户，计划合理

普通 `EXPLAIN` 中 `store_sales` 只扫描一次；日期、家庭人口和门店县过滤都在票据聚合前生效。计划先按 ticket/customer 聚合并执行 `HAVING count(*) between 15 and 20`，再通过 runtime filter 关联 `customer`，没有把客户宽列带入大聚合。

1TB 原 SQL 成功：statement ID `01a043cf-0242-70fc-927f-ebaaa424957f`，耗时 53.446 秒，读取 2,880,201,966 行、扫描 57,614,195,424 字节，返回 133,518 行；无 OOM，无需修改。

## Q35：`OR` 下的多路 EXISTS 形成危险 MARK build

普通 `EXPLAIN` 将第一个正向 `EXISTS(store_sales)` 正确转成 SEMI join，但括号内的 `EXISTS(web_sales) OR EXISTS(catalog_sales)` 仍是两个左深 MARK join，最终在两个 marker 上做 OR。两个 MARK 的 build 输入都是 1999 年前三季度的大事实表路径。

1TB 原 SQL 首次运行时服务被系统 OOM kill。重启后的受控复现中，RSS 在约 12 秒内从 8.0 GiB 连续升至 22.1 GiB；达到安全线后主动 `KILL QUERY`，statement ID `01a043d6-01fd-7b78-a566-44c8dd6ca693`，37.619 秒后以 `context canceled` 结束，未再次 OOM。128 MiB 的单 join spill 阈值没有约束住两个并行 MARK build 的查询级总峰值。

使用关系代数等价形式验证方向：把 web/catalog 的 customer key 用 `UNION ALL` 合并，再对 customer 做一次 `EXISTS`。计划变成 `Union All -> RIGHT SEMI`，以客户侧小输入建表并流式读取两个渠道；1GB 与原 SQL 输出哈希完全一致。1TB statement ID `01a043d9-9388-79fa-8e93-cc01473bff94`，成功耗时 66.893 秒，读取 5,060,108,738 行、扫描 40,757,311,292 字节，返回 100 行，RSS 稳定在约 6.0–7.6 GiB。

已在 filter pushdown 的语义边界实现通用规则：仅当整个 OR 由正向 `EXISTS` marker 构成、MARK 节点形成连续前缀、每个分支具有相同且确定性的外层等值键时，投影各分支的内层 key，以 `UNION ALL` 合并并生成一次 SEMI join。重复 key 不影响存在性；NULL 继续不与普通等式匹配。不同外层键、`NOT EXISTS`、`IN`/`ANY` 三值 marker、非等值相关条件、volatile 分支和投影 marker 全部 fail closed。

新二进制直接执行未改写的 Q35 原 SQL：普通 `EXPLAIN` 已自动产生 `Union All -> RIGHT SEMI`。1GB 输出哈希与修复前原 SQL相同；1TB statement ID `01a043ef-47c1-7d6b-b822-66422836c09c`，成功耗时 66.883 秒，读取 5,060,108,738 行、扫描 40,757,311,292 字节，返回 100 行，输出哈希与手工等价 SQL完全一致，运行中 RSS 约 3.2–6.3 GiB，OOM 计数未增加。

白盒测试覆盖不同内层关系、复合键、三分支和两个独立 OR 组；反例覆盖不同外层键、`NOT EXISTS`、非等值相关、投影 marker、`IN` 三值逻辑和 volatile 分支。小表黑盒同时覆盖重复 key、NULL 与复合键，原始 OR-of-EXISTS 和显式 UNION 参考查询均只返回 id 1、2。`pkg/sql/plan`、`pkg/sql/compile` 的 build、vet 和全量测试通过。

## Q36：grouping-set 共享与聚合后窗口计划合理

普通 `EXPLAIN` 只有一次 `store_sales` 扫描。1999 年日期和八个州的过滤先进入事实表连接，随后统一按 item category/class 和 grouping id 聚合；三个 `ROLLUP` level 通过同一 producer 上的三个 `Sink Scan` 拆分。窗口排名位于聚合和 `UNION ALL` 之后，只处理低基数分组结果，没有明细级窗口或重复事实表扫描。

1TB 原 SQL成功：statement ID `01a043fa-174c-77d7-b887-37d9d3b2fdba`，耗时 99.953 秒，读取 2,880,362,050 行、扫描 80,655,876,420 字节，返回 100 行；输出 SHA-256 为 `9a4d031574d63002a23f314b3608e4aae87ebb3ab2f80aeb8a187997c2a356b5`，没有 OOM。无需修改。

## Q37：两条事实路径会合但选择性充分

普通 `EXPLAIN` 中 `catalog_sales` 和 `inventory` 各扫描一次。manufacturer 和价格条件先把 item 缩小，再进入 catalog 路径；60 天日期范围和库存量条件先进入 inventory 路径，最后两侧按 item key 做 hash join 并去重。虽然最终分组使事实表 multiplicity 不影响语义，但实测选择性充分，未形成危险中间结果，不值得扩大为 aggregate 下的通用 SEMI 化改写。

1TB 原 SQL成功：statement ID `01a043fc-7502-70b5-b88b-3dcd548322eb`，耗时 15.368 秒，读取 2,223,288,608 行、扫描 15,204,870,616 字节，返回 7 行；输出 SHA-256 为 `b64ad8ec15d9147055505366411fe3148b44a79f8702968b9f0c91ea456d80b6`，没有 OOM。无需修改。

## Q38：修复 `INTERSECT` 每个 key 的巨额无效状态

普通 `EXPLAIN` 的关系结构合理：store/catalog/web 三个渠道各扫描一次，12 个月日期条件均下推，每个分支先按姓名和日期去重，再执行两级 `INTERSECT`。风险位于执行器而不是 planner 或 stats。

首次 1TB 运行中，RSS 在 166.751 秒升至约 23.4 GiB，达到安全线后取消，但进程仍被系统 OOM kill；statement ID `01a043fd-b0c8-7c01-a1a4-2f0649c31abf`。heap profile 明确显示 `intersect.(*Intersect).buildHashTable` 保留约 5.54 GiB，其中约 5.02 GiB 来自 `vector.GetSels`。

根因是 `INTERSECT` 对每个 build-side distinct key 调用一次 `vector.GetSels()`，分配整块 selection slice，实际却始终只读写第一个 `int64`，用它表示“该 key 是否已输出”。修复将嵌套的 `[][]int64` 改为扁平 `[]bool`：每个 key 只保留一个状态，Reset/Free 直接释放该扁平状态。集合语义、NULL 处理、重复行抑制和 hash key 均未改变。

修复后 1GB 输出 SHA-256 仍为 `44c2578ef3fd0e46738dd05186abf22fe67b7be58382288ec0ba8eba5f72a5aa`。1TB 原 SQL成功：statement ID `01a04408-784a-7575-8d39-f334904e538d`，耗时 206.927 秒，读取 5,076,187,938 行、扫描 42,194,380,092 字节，返回 1 行，输出 SHA-256 为 `a0195aea96b0c01a02b547cb2ccafb871f6974d7c006ae8b02a87d74ba648def`。运行中 RSS 短暂达到约 15.7 GiB 后回落；同阶段 heap profile 中不再出现 `INTERSECT`/`vector.GetSels` 保留项，OOM 计数未增加。

## Q39：predicate-aware CTE reuse 再次命中

Q39 文件包含 Q39a、Q39b 两条 SQL。两条普通 `EXPLAIN` 均把 `inv` 物化为一个公共 producer，合并两个消费者的月份条件，只读取 1998 年 4/5 月；两个 sink scan 分别恢复 `d_moy=4` 和 `d_moy=5`，Q39b 额外在 inv1 侧恢复 `cov>1.5`。`inventory` 因而只扫描一次，不会为自连接重复完成全年聚合。

1TB 原 SQL均成功：Q39a statement ID `01a0440e-3f4f-764c-abd2-69db3b59c1d7`，15.270 秒，读取 783,373,069 行、扫描 12,530,077,148 字节，返回 10,190 行，输出 SHA-256 `24408c07bc6debdbac00defc8d2f77cb0f3d1e3515ba54eae61cb81ade650d16`；Q39b statement ID `01a0440e-8b8e-73c8-91be-96c2fb7e58c6`，15.311 秒，扫描量相同，返回 251 行，输出 SHA-256 `fb5dfc8596a6aeaf6ebde9f1f37cef4285e7030cb4a0c22d9a4d470498a30fc4`。无 OOM，无需修改。

## Q40：将 preserved-side 唯一维表连接移到 LEFT JOIN 之前

旧普通 `EXPLAIN` 先对全量 `catalog_sales` 和 `catalog_returns` 按 order/item 做 LEFT JOIN，之后才连接带 60 天日期过滤的 `date_dim`、带价格过滤的 `item` 和 `warehouse`。LEFT JOIN 不能减少 preserved-side 的 sales 行，却把本可提前淘汰的行带入大事实表连接和 spill。原 SQL statement ID `01a0440f-9865-7f3a-9df4-b57b5919f964`，成功但耗时 230.962 秒，读取 1,584,285,384 行、扫描 36,874,344,176 字节，返回 100 行；中间临时磁盘一度增长约 21 GiB。

实现通用结合律：`(A LEFT JOIN B) INNER JOIN C` 在上层条件只引用 A/C、C 在连接键上由主键证明唯一、两层 join 没有 limit/projection/filter 等局部语义边界且条件确定时，改写为 `(A INNER JOIN C) LEFT JOIN B`。C 唯一意味着该 INNER JOIN 对每个 A 最多产生一行，因此提前执行不会增加 LEFT JOIN 输入；这是约束驱动的安全证明，不依赖 stats 的基数或选择率猜测。nullable side 条件、非唯一 C、volatile 条件和局部 limit 等边界全部 fail closed。

新普通 `EXPLAIN` 自动与手工等价关系式一致：date/item/warehouse 三个主键 INNER JOIN 全部位于 catalog returns 外连接的 preserved side。1GB 输出 SHA-256 保持 `814dc097c0df408a9c0ab9fc1494edb7ddc26b9d1310c116c286db6191508d74`；1TB 原 SQL statement ID `01a0441b-f823-7c05-a051-db16ed6b7ab0`，耗时 17.643 秒，扫描计数保持不变，返回 100 行，输出 SHA-256 仍为 `e0bda22955241f30339bafb61647fc0d4b0de9b8cd4d9d87624de96116d464c9`。

白盒测试覆盖正常和上层 INNER JOIN 交换输入的两种形态，反例覆盖非唯一输入、引用 nullable side 和局部 limit。`pkg/sql/plan`、`pkg/sql/compile` 的全量测试、vet 和完整 mo-service build 通过；TPC-DS 1GB Q1–Q99 全部成功，除无 ORDER BY 的 Q31/Q71 行序变化外结果集合一致，Q48 得到既有重试基线值 19869。

## Q41：小表相关标量聚合已正确去相关

普通 `EXPLAIN` 将按 manufacturer 相关的 `count(*)` 去相关为一个按 manufacturer 和两组布尔条件分组的 item 聚合，再通过 LEFT JOIN 恢复空输入为 0；没有 per-row 子查询执行。外层 manufacturer id 条件下推到独立 item 扫描，两次 item 扫描规模都很小。

1TB 原 SQL成功：statement ID `01a0441e-2aa2-7b8b-ab23-a86c1bde465c`，耗时 0.030 秒，读取 36,000 行、扫描 3,107,049 字节，返回 0 行，输出 SHA-256 `1aaf2750eecf84d6cc6a4cfaa2e5224e4199fb097651f3745b32f4f6e3aee138`。无需修改。

## Q42：单事实扫描计划合理

普通 `EXPLAIN` 只扫描一次 `store_sales`；1998 年 12 月和 manager 条件分别下推到 date/item，两者都通过主键 hash join 在 category 聚合前生效。没有重复扫描或高基数中间状态。

1TB 原 SQL成功：statement ID `01a04424-b38c-7121-b007-875fee49a035`，耗时 36.919 秒，读取 2,880,361,048 行、扫描 46,091,484,572 字节，返回 12 行，输出 SHA-256 `8a276f666d0f8615e015116bd75318ceebfc1a62d3ebbf65b113a598e2e57573`。无 OOM，无需修改。

## Q43：条件聚合已在维表连接前缩减

普通 `EXPLAIN` 只有一次 `store_sales` 扫描。1998 年日期条件先进入事实路径，随后在连接 store 前按 `ss_store_sk` 同时计算七个 weekday SUM；GMT offset 过滤位于小 store 表。没有七次扫描或将明细行带入最终 store 名称聚合。

1TB 原 SQL成功：statement ID `01a04425-dca4-78d0-911a-906e4bc39bdb`，耗时 89.009 秒，读取 2,880,062,050 行、扫描 46,082,205,672 字节，返回 100 行，输出 SHA-256 `379d5b872e03c65f2c7d5c553fd446339441210cf1ff190a763443633cf78e47`。无 OOM，无需修改。

## Q44：相同排名输入被计算两次

普通 `EXPLAIN` 的升序和降序分支完全复制：每个分支一次按 item 的 store 410 平均利润扫描，以及一次 `ss_hdemo_sk IS NULL` 的标量平均扫描，总计四次 `store_sales`。两个分支只在窗口排序方向上不同，正确的通用方向是共享聚合输入并在其上计算两个 window，而不是针对查询文本合并。

当前计划资源安全且低于 20 分钟目标，先记录、不扩大修改：1TB statement ID `01a04428-181b-7e36-a62b-d5ba60ac66eb`，成功耗时 140.807 秒，读取 11,520,551,996 行、扫描 184,343,288,926 字节，返回 10 行，输出 SHA-256 `926b0674f737d815af107cfc6b4f73d79720439947c67e310793aaf8b96ec641`；无 OOM、无 spill。后续与 Q9/Q28 一并作为通用 CSE/聚合融合路线处理。

## Q45：OR 下的小集合 MARK join 安全

普通 `EXPLAIN` 将 item 子查询构造成仅含 10 个 item key 的 MARK build，再与 ZIP 条件执行 OR；这与 Q35 的两个大事实 MARK build 不同，状态规模有严格小上界。2000 Q2 日期过滤在 web_sales 路径最前端生效，其余 customer/address/item 都是等值 hash join。

1TB 原 SQL成功：statement ID `01a0442a-fb1a-776c-9ca2-7a946fc50bd7`，耗时 9.079 秒，读取 738,373,435 行、扫描 14,817,758,053 字节，返回 100 行，输出 SHA-256 `3e2024a457942d11bf70b61afa60a04313450114f3a327432f9bf78de7977e44`。无 OOM，无需修改。

## Q46：先票据聚合再关联客户

普通 `EXPLAIN` 中三年周末日期、五个门店城市和家庭人口条件都在事实路径的票据聚合前生效；计划按 ticket/customer/address 聚合 coupon/profit 后才关联 customer 当前地址并比较城市，没有把姓名和当前地址宽列带入大聚合。

1TB 原 SQL成功：statement ID `01a0442b-c7a9-7a44-9245-1a72001a6c87`，耗时 102.353 秒，读取 2,904,069,250 行、扫描 116,208,511,004 字节，返回 100 行，输出 SHA-256 `12bb3a2cea57e487baa9713557ab0bc3d6594ec5e4a44f406b506bb75fff0542`。无 OOM，无需修改。

## Q47：三消费者高收益 CTE 共享

旧普通 `EXPLAIN` 将 `v1` 在 lag、current、lead 三个位置完整展开；每份都重复执行 `store_sales -> date/item/store join -> group -> avg/rank window`。1TB 基线 statement ID `01a04431-bee4-798c-a5d4-cee990e9484c`，耗时 844.957 秒，读取 8,641,086,150 行、扫描 172,848,866,016 字节，返回 100 行。资源没有失控，但距离 20 分钟目标余量很小，重复工作明确。

修复没有匹配 Q47 文本，而是补齐两条通用 CTE reuse 规则：

1. 对至少三个完整消费的 occurrence，只有当估算的共享成本在乘 2 安全因子后仍低于 inline 成本时，才允许在 8 GiB planner 估算上限内使用已有有界 spill materialization；两消费者的保守边界保持不变。
2. 最终 rewrite root 可覆盖另一个 CTE 内部的 occurrence 时允许复用；已验证的 materialized source 可继续作为外层确定性 producer 的输入，使嵌套 CTE 内外层均可共享。不可达 step、递归/相关/volatile producer 和部分消费路径继续拒绝。

新计划只有一个 `v1` producer 和三个 `Sink Scan`，`store_sales`、`date_dim`、`item`、`store` 都从三次扫描降为一次。1TB 原 SQL statement ID `01a04446-7db1-7e94-afe8-0f26c9ed5bb5`，耗时 250.868 秒，读取 2,880,362,050 行、扫描 57,616,288,672 字节，返回 100 行；输出与基线逐字节一致，SHA-256 均为 `1d9a2e75d57bbcb1837be3f2aa6dfb63ae06b476753ce317c9a273d8302c2507`，OOM 计数未增加。

1GB Q47 从 20.556 秒降至 14.045 秒，结果哈希保持 `6d3b4cd5b5b8be8528ebfe5dc56fad647a359c28db09859c4bffb986c39718f3`。全 99 条 1GB 回归均成功；相对上一版只有 Q5/Q23/Q47/Q57/Q64/Q80 的计划变化，其中 Q23/Q64 验证了嵌套内外层同时共享，六条变化查询结果均逐字节一致，其余 93 条计划完全相同。planner/compile 全量测试、定向 race、vet 和完整 build 均通过。

## Q48：跨表 DNF 中的公共等值键必须可见

旧计划没有把三个 OR 分支都包含的 `cd_demo_sk = ss_cdemo_sk` 提到 DNF 外，因此 customer demographics 与事实路径之间没有独立的 hash key，只能把完整 OR 当作非等值 join 条件。1GB 原 SQL虽然只扫描一次 `store_sales`，仍耗时 227.546 秒；手工额外写入同一个冗余等式后变为 hash join，结果逐字节一致且只需约 0.1 秒，确认瓶颈不是 stats 或扫描 IO。

根因在既有 distributivity 保护：为保留 composite-key range folding，它会让单表 DNF 保持原形；但旧实现只检查每个二元谓词的第一个参数，忽略 `BETWEEN`/`IN` 内的列和等式另一侧的 relation。Q48 因此被错误归类为单表 DNF。修复改为递归收集完整表达式中的 relation：仅当 DNF 确实只引用一个 relation 时保持原形；引用多个 relation 时按既有严格结构相等规则提出公共条件。没有增加 SQL 模式特判，也不使用基数或选择率估计。

新普通 `EXPLAIN` 形成独立的 `cd_demo_sk = ss_cdemo_sk` hash join，原 OR 只保留三组人口属性和价格残余条件。1GB 原 SQL耗时 0.231 秒，输出 SHA-256 与旧计划完全一致，均为 `3d67bb2ebc3f3adac66cb2fb117e0d1ec51049176b67bb5622013bd8f4439726`。1TB 原 SQL statement ID `01a04460-31d9-7f63-b52b-7569d2948d16`，成功耗时 86.605 秒，读取 2,887,982,850 行、扫描 104,092,037,964 字节，返回 1 行，输出 SHA-256 `7e484e91fcf078f7f393beaa9d23544491acac9f06eab1d52b6561d3ee72d797`，无 OOM。

全 99 条 1GB 普通计划相对 Q47 版本只有 Q48/Q85 变化。Q85 同样提出人口属性分支共有的等值键，原 SQL结果逐字节一致；其余 97 条计划完全相同。白盒覆盖“跨 relation 等式 + 三元 BETWEEN”正例和必须保留的单表 DNF 反例；planner/compile 全量测试、定向 race、vet 和完整 build 均通过。

## Q49：三渠道复合键连接计划合理

普通 `EXPLAIN` 中 web、catalog、store 三个分支各只扫描对应 sales/returns 一次。虽然 SQL 写为 LEFT JOIN，但 `wr_return_amt/cr_return_amount/sr_return_amt > 10000` 会拒绝 NULL，计划已安全转为 INNER JOIN；年月与 sales 的利润、实付和数量条件都在复合键连接前生效。每个分支先按 item 聚合，再计算两个 rank，窗口不会处理事实明细。

1TB 原 SQL成功，客户端实测 92.743 秒，返回 100 行，输出 SHA-256 `c0c56b88c4a27a39dc4453a9297c81d8781d9dfeaa10104710fbe24954df427f`。RSS 保持在约 5–6 GiB，无 OOM；无需修改。

## Q50：退货日期先过滤，复合键连接计划合理

普通 `EXPLAIN` 只扫描一次 `store_sales` 和一次 `store_returns`。2000 年 9 月条件先通过 `date_dim` 缩小 returns，再按 ticket/item/customer 复合键连接 sales；sales sold date 和 store 都是主键等值连接。最终只按门店聚合五个退货时延区间，没有重复事实扫描、危险 build side 或明细级窗口。

1TB 原 SQL成功：statement ID `01a04469-57e0-71bd-b459-d8593ad3aaf6`，耗时 57.744 秒，读取 3,168,134,863 行、扫描 62,209,149,436 字节，返回 100 行；输出 SHA-256 `1545ba0fb06614fb307673994ce98e5db9e9e37af14bcfc4997058a64137a463`。无 OOM，无需修改。

## Q51：累计 MIN/MAX 使用线性 running aggregate

Q51 先分别计算 web/store 的按 item、date 累计销售额，FULL OUTER JOIN 后再计算两个累计 MAX。计划的表扫描、日期过滤和连接关系合理，但旧执行器只允许声明了 source-preserving merge 的聚合进入累计窗口快路径；SUM 已声明，MIN/MAX 没有。因此外层 MAX 对每个输出行重新扫描 partition prefix，CPU profile 中 `minMaxExecFixed.BatchFill` 一度占 66%–73%，总复杂度为各 partition 长度平方和。旧 1TB 执行在 655.073 秒时仍处于该阶段，主动取消以避免无收益地继续运行。

MIN/MAX 的现有 `Merge` 实际只读取 source，并将值复制或比较到 destination，不转移所有权、也不修改 source。修复为定宽和 bytes 两种 MIN/MAX executor 声明这一既有契约，使累计窗口复用通用 running aggregate 路径：每个输入行只 Fill 一次，每个输出行只 snapshot 当前状态一次，复杂度降为线性。修改不检查查询、函数参数或 stats；所有累计 MIN/MAX 窗口均可受益。契约测试覆盖定宽/变长 source 被重复 merge 后仍不变，窗口测试覆盖跨 output chunk 保留 running state。

1GB 修复前后输出逐字节一致，耗时 28.881 → 27.548 秒；该规模的固定扫描、排序和 partition 成本占主导。1TB 修复后 statement ID `01a04479-0600-7fb1-a881-c824e650b2e9`，733.427 秒成功完成，读取 3,600,134,473 行、扫描 57,601,567,176 字节，返回 100 行，输出 SHA-256 `b3969a77ba03b764105f5add1dae566fdae7701f50d0591df02b7eabaf4d389c`。后期 profile 中 MIN/MAX `BatchFill` 仅占约 6%–12%，RSS 峰值约 12 GiB 后回落，无 OOM。aggexec/window/compile 全量测试、定向 race 和 vet 均通过。

## Q52：单次事实扫描，计划合理

普通 `EXPLAIN` 只扫描一次 `store_sales`。1998 年 12 月日期和 manager 条件分别在 date/item 主键 join 前过滤，随后按 brand 聚合；没有重复子计划或高基数窗口。1TB 原 SQL成功：statement ID `01a04487-c08d-759b-9c9c-67fc0a4d297e`，耗时 20.023 秒，读取 2,880,361,048 行、扫描 46,091,484,572 字节，返回 100 行；输出 SHA-256 `bd8b35e9ecc42ca0d27abe3652314e78f3667d990bfff65326011c52fe9a09bf`。无需修改。

## Q53：窗口位于聚合后，计划合理

日期 12 个月和两组 item 属性 OR 条件均先进入事实路径，计划按 manufacturer/quarter 聚合后才计算 manufacturer 内平均值。窗口输入已缩到低基数季度结果，不会对事实明细执行。1TB 原 SQL成功：statement ID `01a04488-6e49-7ed0-89c1-afdbc48791a6`，耗时 30.222 秒，读取 2,880,362,050 行、扫描 57,624,640,576 字节，返回 100 行；输出 SHA-256 `303091c3d0fec5d1b9cebd17d299c4eed2a299a2be28ffb795ce330ab9cca8fd`。无需修改。

## Q54：客户集合先缩减，后续 revenue 聚合安全

catalog/web 两个渠道先由 1999 年 3 月和 Jewelry/consignment item 条件缩减，再与 customer 去重形成 `my_customers`。后续三个月 `store_sales` 与该客户集合连接并按 customer 聚合，county/state 与小 store 维表的连接虽会保留 SQL 本身要求的 multiplicity，但没有危险 build side。两个无相关月份标量已成为只扫描小 `date_dim` 的 SINGLE join，不构成主要成本。

1TB 原 SQL成功：statement ID `01a04489-283e-7d77-9953-ca327cee1f14`，耗时 55.165 秒，读取 5,052,682,926 行、扫描 72,120,737,698 字节，返回 100 行；输出 SHA-256 `f713c36af53d45933cfbe5b89589ba9a6adec8bceaca39b0ab795571cce5d451`。无 OOM，无需修改。

## Q55：Q52 同类计划合理

普通 `EXPLAIN` 只扫描一次 `store_sales`，2001 年 12 月和 manager 条件均通过主键维表 join 在 brand 聚合前生效。1TB 原 SQL成功：statement ID `01a0448a-3bc3-7b5b-a457-b0be00a048cb`，耗时 23.541 秒，读取 2,880,361,048 行、扫描 46,091,484,572 字节，返回 100 行；输出 SHA-256 `806a555bbb4ea6f54745f4deb4cd03e5d32828ae423bc597a9176553b8f13dba`。无需修改。

## Q56：三渠道过滤与聚合均合理

store/catalog/web 三个分支各扫描对应事实表一次。2000 年 1 月、GMT -8 和三种颜色 item 集合均在各自渠道聚合前生效；item SEMI 输入很小，没有 Q35 类型的大型 MARK build。1TB 原 SQL成功：statement ID `01a0448b-8c21-77c3-b9e4-dde817c19850`，耗时 43.369 秒，读取 5,059,987,938 行、扫描 101,086,405,584 字节，返回 100 行；输出 SHA-256 `0147a02bef91dfc2e3cff95a6bdb93d1ca243dd05bf13136542d83be498e8e08`。无需修改。

## Q57：Q47 的三消费者 CTE 共享再次命中

`v1` 同时被 lag/current/lead 三个位置完整消费。普通 `EXPLAIN` 只生成一个 `catalog_sales -> aggregate -> avg/rank window` producer 和三个 `Sink Scan`，事实表及维表均只扫描一次；这是 Q47 通用机制的独立正向样本。窗口位于按 category/brand/call-center/month 聚合后，状态规模受控。

1TB 原 SQL成功：statement ID `01a0448c-6310-7a7b-b0d1-36ab8e4695ba`，耗时 156.025 秒，读取 1,440,353,507 行、扫描 28,816,086,084 字节，返回 100 行；输出 SHA-256 `a2ae8668c02732f95e6085888f093ec51de7d415cdf42f74ff2433f0bd55abd9`。无 OOM，无新增修改。

## Q58：三渠道单周聚合计划合理

store/catalog/web 每个事实表各扫描一次，并先与目标周的 `date_dim` 集合连接，再按 item 聚合。求目标 week 的标量子查询在三个分支内有少量重复，但只涉及小 date_dim，不值得扩大 CTE/CSE 边界。1TB 原 SQL成功：statement ID `01a0448e-f3cb-730d-b631-7a298b9302c3`，耗时 30.534 秒，读取 5,041,137,090 行、扫描 80,666,847,048 字节，返回 100 行；输出 SHA-256 `144b7f31f3c0b48617ca02eef4d1cc660f9cfd4f37e8c9d8cac388a4d1349d81`。无需修改。

## Q59：安全完成，记录跨消费者 week domain 传播

`wss` 被两个年份消费者复用，当前计划正确地只生成一个 producer 和两个 `Sink Scan`；但 producer 在按 `d_week_seq, ss_store_sk` 聚合前扫描全历史 `store_sales`，两个消费者的 month range 只在聚合后通过 `date_dim.d_week_seq` join 生效。可进一步做的通用方向，是把消费者从唯一维表得到的 week key domain 合并为 SEMI filter 再传回 producer，同时保留消费者 join 以维持边界周 multiplicity。该证明跨 join 和聚合，不能用简单 range 猜测替代，因此当前不为 Q59 扩大 rewrite。

现计划只有一次事实扫描，聚合 key 规模受 week/store 上界约束，资源安全。1TB 原 SQL成功：statement ID `01a0448f-a15b-7da2-b9c0-5efe54c1eb00`，耗时 330.993 秒，读取 2,880,209,150 行、扫描 46,083,394,496 字节，返回 100 行；输出 SHA-256 `37f0c0ddd7a5bb3b8909535c4a03af22b45cae100b7c3c999e572d170c539488`。无需当前修改。

## Q60：Q56 同类计划合理

三渠道各扫描一次事实表，1999 年 9 月、GMT -6 和 Children category 都在聚合前生效。1TB 原 SQL成功：statement ID `01a04494-f0f3-7c46-81cf-0ad35d7cc7d5`，耗时 53.487 秒，读取 5,059,987,938 行、扫描 101,086,405,584 字节，返回 100 行；输出 SHA-256 `0d22a648ce620f92efca80cbbd54f6a292e941f43aa993d05ecf8909cd5ed9e2`。无 OOM，无需修改。

## Q61：重复标量聚合安全完成，归入通用融合方向

promotional 和 total 两个无相关标量分支重复扫描相同月份的 `store_sales` 及五张公共维表，前者仅多一个 promotion 过滤。正确的通用方向是识别共享事实路径，并将子集分支改成 conditional aggregate 或共享 producer；这与 Q9/Q28 的 scalar aggregate fusion 属于同一类问题。当前只有两路、日期选择性明确且资源安全，不为 Q61 单独实现 SQL 形状 rewrite。

1TB 原 SQL成功：statement ID `01a04496-e118-74f0-a04d-7b1276e7c5f7`，耗时 73.292 秒，读取 5,796,725,600 行、扫描 150,114,067,172 字节，返回 1 行；输出 SHA-256 `7a80f2d76904863ae81a91380d4f01d62c4625d1c1de594a0be0e7a178f6a0b5`。无需当前修改。

## Q62：partial aggregate 已提前

目标 12 个月通过 ship date 主键 join 先过滤 `web_sales`；warehouse 和 ship mode join 后，计划先按 warehouse/type/web-site-key 计算五个时延条件 SUM，再连接带名称的 web_site 并完成最终聚合。宽展示列没有进入事实级聚合。1TB 原 SQL成功：statement ID `01a04498-44c4-7ab1-a001-3412acb53c32`，耗时 45.151 秒，读取 720,073,519 行、扫描 14,400,594,544 字节，返回 100 行；输出 SHA-256 `d9e99a1c1ba1d91890ad390fb437dbf54791ef633873229a4dc4f4bdb159a502`。无需修改。

## Q63：Q53 同类聚合后窗口

12 个月日期和 item 属性 OR 条件均先进入 `store_sales` 路径，计划按 manager/month 聚合后才计算 manager 内平均值。1TB 原 SQL成功：statement ID `01a04499-2191-71fe-9da1-485fc3de4caf`，耗时 35.099 秒，读取 2,880,362,050 行、扫描 57,624,640,576 字节，返回 100 行；输出 SHA-256 `70cb25329e061997d8a7f2d39ead42746c6500e6478be41e15dad1ae5f3bac4a`。无需修改。

## Q64：安全嵌套 CTE 共享后 1TB 实测通过

`cross_sales` 被 2000/2001 两个消费者完整引用，并在内部依赖 `cs_ui`。当前普通 `EXPLAIN` 命中 Q47 补齐的嵌套共享：`cs_ui` 只有一个 catalog sales/returns producer，`cross_sales` 也只有一个 producer，两个最终消费者通过 `Sink Scan` 读取；消费者年份条件合并为 `d1.d_year IN (2000, 2001)` 并进入 producer。事实/退货复合键、选择性 item 条件和维表主键 join 都在宽聚合前生效。

1TB 原 SQL成功：statement ID `01a0449a-0ae5-7d26-81bb-fb412c0db568`，耗时 145.852 秒，读取 4,780,342,624 行、扫描 204,369,980,835 字节，返回 7,094 行；输出 SHA-256 `ec356895fe68f449532e5224ba4fcb988252a0c0fdb9175f16317056e71ef61a`。RSS 约 8 GiB 且稳定，无 OOM。该结果说明当前 Q64 已不再是危险计划，但不等于 stats 的系统性误差已经解决。

## Q65：重复非 CTE 聚合安全完成

同一个“12 个月 store_sales 按 store/item 聚合”子计划在 `sb` 和 `sc` 中各执行一次：一份继续按 store 求平均，另一份保留 item revenue。两路大聚合都发生 spill，CPU profile 显示正常的 group spill write/merge，没有无界状态；代价是事实扫描、聚合和 spill IO 全部重复。正确的后续方向是基于逻辑表达式和消费列证明的非 CTE common-subplan reuse，并让两个消费者共享一次 store/item producer，而不是匹配 Q65 文本。

1TB 原 SQL成功：statement ID `01a0449c-8524-721b-8d6c-5ef800719fe3`，耗时 272.469 秒，读取 5,760,422,597 行、扫描 115,250,786,188 字节，返回 100 行；输出 SHA-256 `72828662ee045f652ae50853315b695940ce159f2811a89fe8bc6e8e94279a3d`。RSS 约 8–9 GiB，无 OOM；当前不扩大修改。

## Q66：宽条件聚合已在维表展示列前完成

web/catalog 两个分支各扫描对应事实表一次。2002 年、时间区间和两种 carrier 条件均先进入事实路径；计划先按 year/warehouse key 计算 24 个按月份的 sales/net SUM，再连接 warehouse 的名称、地址和面积列，最后合并两渠道。没有把宽 warehouse 列带入事实级 hash 状态。

1TB 原 SQL成功：statement ID `01a044a1-de8b-7726-ba7e-500130466751`，耗时 81.115 秒，读取 2,160,299,770 行、扫描 77,762,450,328 字节，返回 20 行；输出 SHA-256 `ff002f9f4e471fee528ac1847158aa9faa0a24bfb438dfa2ba4c0c49f3cbb15f`。无需修改。

## Q67：20 分钟目标的首个新阻塞点

普通 `EXPLAIN` 已命中 grouping-set 共享：`store_sales` 只扫描一次，一个 Plan 0 producer 计算全部 ROLLUP level，Plan 1 的九个 `Sink Scan` 只按 grouping id 拆分结果。因此问题不再是重复事实扫描。

剩余瓶颈在 grouping-set 的物理算法。当前 producer 在聚合前把每条明细扩成 9 个 grouping row，再以 category/class/brand/product/year/quarter/month/store 和 grouping id 的宽复合 key 做 hash aggregate。CPU profile 持续落在 `fillGroupingAwareStr`、hash rehash、`UnionOne` 和 `groupSpillWriter.Write`；运行到 20 分钟仍在 producer 聚合/spill 阶段，尚未进入后续 rank。statement ID `01a044a3-72ed-7807-910a-bf0d7485ae13`，1202.352 秒后因客户端预算取消，未产生结果；RSS 约 8–10 GiB，临时磁盘相对查询前峰值增长约 112 GB，取消后已回收。

这不能再靠提高 spill 阈值或调整 join order解决：内存受控，但明细级 9 倍 hash、宽 key 和 spill IO 是算法性放大。通用修复方向应是分层 ROLLUP：先计算最细 level，再由其结果逐级聚合父 level；或者引入等价的 rollup-aware aggregate state，使每条事实不必进入九套宽 key 状态。必须同时验证 grouping id/NULL 语义、distinct/非可合并聚合反例、分布式 partial/final 合并、spill 和排序窗口消费。按既定规则，Q67 需要先确定这一机制设计，再继续 Q68。

## 后续查询记录模板

每条查询都应补齐以下证据后再标记为完成：

1. 原始计划和主要运行时瓶颈；
2. 修改属于配置、stats、逻辑优化、物理计划还是执行器；
3. 为什么修改具有通用性，以及保留了哪些反例；
4. 修复前后的 statement ID、耗时、扫描量、结果行数和资源峰值；
5. 白盒计划测试、黑盒结果测试和 1TB 实测。
