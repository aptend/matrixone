# Session Stats Cache

## 核心逻辑

**有效 stats 在 3 秒内直接返回缓存，否则重新计算。**

```txt
Stats(tableID)
    │
    ▼
cache.Get(tableID)
    │
    ▼
w.Exists() && 3秒内访问 && AccurateObjectNumber > 0 ?
    │
    ├─ 是 → 返回缓存
    │
    └─ 否 → doStatsHeavyWork()
                │
                ▼
            cache.Set(tableID, result)
                │
                ▼
            返回 result
```

## doStatsHeavyWork 流程

```txt
doStatsHeavyWork(tableID)
    │
    ▼
ensureDatabaseIsNotEmpty()  ← 重操作
    │
    ▼
getRelation()               ← 重操作
    │
    ▼
ApproxObjectsNum()
    │
    ├─ == 0 → 返回 nil（空表）
    │
    ▼
缓存有效 && NeedUpdate() == false ?
    │
    ├─ 是 → 返回缓存
    │
    └─ 否 → table.Stats()
                │
                ▼
            返回 stats
```

## TODO: 潜在优化点

`ensureDatabaseIsNotEmpty` 和 `getRelation` 是当前主要阻塞点，每次进入 heavyWork 都会执行。后续可考虑优化这两个操作的开销。相反，table.Stats() 操作本身，除了第一次调用会比较重，其余时间都在读 GlobalStats 的内存。

## 关键设计

1. **有效性判断**：`AccurateObjectNumber > 0` 表示 stats 有效
2. **空表处理**：返回 `nil`，让调用方使用 `DefaultStats()`（Outcnt=1000）
3. **激进重试**：stats 无效时每次都重新获取，确保数据变化后能及时更新。这对 BVT 测试至关重要——测试中表刚创建/插入数据后立即查询，需要及时获取最新 stats 才能生成正确的执行计划（如 LEFT/RIGHT JOIN 选择）
4. **值类型缓存**：`map[uint64]StatsInfoWrapper` 使用值类型，通过 `lastVisit == 0` 判断是否存在，减少小对象分配

## 相关文件

- `pkg/sql/plan/stats.go` - StatsCache, StatsInfoWrapper 定义
- `pkg/frontend/compiler_context.go` - TxnCompilerContext.Stats() 实现
- `pkg/sql/compile/sql_executor_context.go` - compilerContext.Stats() 实现


# Global Stats Cache

GlobalStatsCache 是一个全局缓存，用于存储所有表的统计信息。它由 Logtail 事件驱动异步更新，Session Stats 读取时直接从内存获取。

```txt
Logtail 事件
    │
    ▼
tailC (chan, cap=10000) logtail 消费专用，最小化阻塞 logtail 消费
    │
    ▼
consumeWorker (1个 goroutine)
    │
    │ 判断入队条件（第一层）：
    │ - keyExists(): key 必须已存在
    │ - CkpLocation: checkpoint 时触发
    │ - MetaEntry: object 元数据变更时触发
    │
    ▼
updateC (chan, cap=3000)
    │
    ▼
updateWorker (16-27个 goroutine)
    │
    │ 判断执行条件（第二层）：便于统一 debounce force/normal update request
    │ - shouldUpdate(): 检查 inProgress 和 MinUpdateInterval (15s)
    │
    ▼
doUpdate()
    │
    ├─→ 订阅表获取 PartitionState
    ├─→ 从 CatalogCache 获取 TableDef
    └─→ collectTableStats()
            │
            ▼
        ForeachVisibleObjects()
            │ 并发遍历所有**已落盘的 Object** (concurrentExecutor)
            │ 注意：内存中的 dirty blocks 不参与统计
            ▼
        FastLoadObjectMeta() (S3 IO)
            │
            ▼
        累加统计信息 (ZoneMap, NDV, RowCount 等)
```

## 关键设计

1. **两级 channel 缓冲**：`tailC`(10000) 最小代价干扰 logtail 的消费流程
2. **两层过滤**：入队条件（keyExists + 事件类型）和执行条件（inProgress + MinUpdateInterval）减少更新频率。分离是为了满足 force 模式，可以直接跳过入队条件，但是需要受限于执行条件
4. **只统计已落盘数据**：内存中 logtail 不参与 stats 统计，所以入队条件只关注元数据类型 logtail entry，内存类型 entry 不关心
5. **并发遍历**：使用 concurrentExecutor 并发加载 Object 元数据

## 相关文件

- `pkg/vm/engine/disttae/stats.go` - GlobalStatsCache 定义和更新逻辑
- `pkg/vm/engine/disttae/txn_table.go` - table.Stats() 读取 GlobalStats
- `pkg/vm/engine/disttae/logtailreplay/partition_state.go` - PartitionState 和 ForeachVisibleObjects
