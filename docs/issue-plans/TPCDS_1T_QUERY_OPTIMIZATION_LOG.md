# TPC-DS 1TB 查询优化记录

- 最后更新：2026-08-26
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
- 目前已检查 Q1–Q4。Q5–Q99 尚未逐条诊断，不在本文中声称已优化。

## 进度摘要

| 查询 | 问题 | 修改 | 1TB 结果 | 状态 |
|---|---|---|---|---|
| Q1 | 聚合需要更早、可预测地 spill | 统一将 spill 阈值设为 128 MiB；无代码特判 | 21.411 秒，100 行 | 完成 |
| Q2 | 外层多引用 CTE 因内部引用另一个 CTE 而被整体 inline 两次 | 放宽外层 CTE reuse；保留全部安全、成本和内存保护 | 1980.838 → 855.284 秒，结果仍为 2513 行 | 完成 |
| Q3 | 未发现明显危险计划 | 不修改 | 17.214 秒，100 行 | 完成 |
| Q4 | 六个消费者把同一 CTE 及三张事实表各 inline 六次 | 合并消费者局部谓词后复用一个可 spill producer | 1176.530 秒，100 行，无 OOM | 结构问题完成；spill 效率待优化 |

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
- 被绑定在另一个 CTE 内部的 CTE 自身仍保持 inline，避免当前 rewrite root 无法覆盖独立 step graph。
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

### 回归保护

- 白盒计划测试覆盖：变量宽度 producer、两个不同且重叠的消费者谓词、Top-N `SORT LIMIT`，断言基础表只扫描一次且有两个消费者。
- 黑盒执行测试同时断言 producer 上的谓词并集、单次基础表扫描、两个 `Sink Scan` 和精确结果。
- 反例覆盖：存在无谓词消费者、volatile 消费者谓词、提前 `LIMIT/OFFSET`、半连接/相关消费等场景继续 inline。

## 后续查询记录模板

每条查询都应补齐以下证据后再标记为完成：

1. 原始计划和主要运行时瓶颈；
2. 修改属于配置、stats、逻辑优化、物理计划还是执行器；
3. 为什么修改具有通用性，以及保留了哪些反例；
4. 修复前后的 statement ID、耗时、扫描量、结果行数和资源峰值；
5. 白盒计划测试、黑盒结果测试和 1TB 实测。
