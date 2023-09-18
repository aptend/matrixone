


# 现状

## Flush

对一个表，满足两个条件会进行 flush：
1. 内存中有修改，但是超过 5 分钟没有做 flush
2. 内存中的修改超过 20 MB

目前这两个参数可以通过 mo-ctl 对表级别进行调整。但是实际生产不会使用，这里就不介绍。

一次 flush 会把内存中的 insert 按照主键排序，写成一个 object。同时也会把 deletes 收集起来，按照 rowid 排序，写成一个 object。

目前 flush 过程发生在一个事务中，具有原子性，而且不可能会产生 rollback。因为 flush 期间产生的新删除，都会延迟到下一次，或者转移到新产生的 object 上。



> 相比更早版本的 flush task，相同时间的内 tpcc，flush 数量下降了 1 ~ 2 个数量级

## Merge

定时对每个表做扫描，选择出满足如下条件的 object
1. object 中的行数小于 5 个 blk 的最大行数
2. object 的剩余行数小于初始行数的一半

在这些 object 中，选择剩余行数最小的两个 object 进行 merge。这种做法的缺点是，在资源足够的情况下，merge 的效率低。merge 出的 object 中最多也只有 9 个 blk，10 万行左右。另外也没有考虑主键的 overlap。

目前的 merge 过程也发生在一个事务中，而且过程同样不会 rollback，如果执行期间产生了删除，这些删除都会转移到新产生的 object 上。




# 目标


0. 实现 Tombstone
1. 减少 Row Tombstone Object 的数量
2. 提高 merge 效率，改善策略，减少 Row Object，同时也减少 object 之间的 overlap
4. 增加 metric，通过 mo-ctl 调整 merge 参数，实验出对 tp 结果更好的参数组合
5. 一个代价模拟器。给定模拟的输入参数，模拟不同策略的结果，方便直接对策略进行更高效率的迭代

# 方案

object 之后会分为 Row Object，Tombstone object，两种都需要 merge，但是相关的过程有区别。

## Row Object

### 选择与统计

目前默认是每个表周期性扫描一次，主要产出是一个用户表的 Row Object 列表。

大量表的时候，支持按照租户扫描。独立调度。

1. 单个 Object 基本信息，在创建时确定：
    - block 数量
    - 总行数
    - cluster key zonemap

2. 单个 Object 的可变化信息：
    - 剩余行数，随时间、删除操作而减少
    - 行数变化率
        - 5 s 变化率，可以反应热点 object
        - 3 min、10 min 变化率，反应静置的状态

3. Object 列表整体信息：
    - object cluster key 重叠情况，参考 snowflake
        - 重叠数量：有多少个 object 重叠。可能需要对行数取权重。
        - 重叠深度：对某一个主键，最多和多少个 object 重叠。越少越好，通过 zonemap 就能直接过滤掉

4. 历史信息：
    - 当前表的 flush 历史。辅助判断表的活跃程度。比如过去 1、3、10 分钟：
        - flush 次数
        - 新产生 object 的平均大小
        - 平均每次消耗资源
    - 当前表的 merge 历史。可以辅助判断是否真正执行 merge，因为可以从整体资源去考虑。过去 10 分钟：
        - merge 次数
        - 新产生 object 
        - 平均每次消耗资源
        - merge 前后的 object 的主键重叠情况




### 候选策略



```go
type Policy interface {
    Feed([]ObjectInfo)
    Revise(cpu, mem uint64) -> []ObjectInfo
}
```

后续 policy 都基于这个接口，输入是当前表的 object 列表，按照 cpu 和 mem 信息，输出需要 merge 的 object 列表。

比如目前会仿照 snowflake 制作一个基于 object 重叠的策略。针对每个表，选择出最有收益的 object 组合

#### Level

snowflake 按照 merge 的次数，把 partition 对应到不同层级。低层级通常具有更高的 merge 优先级，可以减少写放大。

当前方案中，level 按照剩余行数进行区分。因为 object 会产生 Tombstone，如果始终保持高 level，会增加无效缓存。

| level | blk数量 |
| ----- | -------- |
| 1     | 0~4      |
| 2     | 8~16     |
| 3     | 16~32    |
| 4     | 32~64    |
| 5     | 64~128   |
| 6     | 128~256  |

比如，层级 1 就是 flush 产生的新 object，或者是之前的高层级经过删除落回低层级的 object。



#### Zonemap

如果不同的 object 的重叠严重，会使查询时过滤效果变差，引入无效的 io。因此需要对重叠的 object 做 merge。通常的做法是在一个 level 上收集 object，估计哪一种对整体的重叠改善最大，从而选择需要做 merge 的 objects。

![width-depth.png](https://miro.medium.com/v2/resize:fit:500/format:webp/1*-W3o9AkxLJU6xPCocp-nJw.png)



#### 变化率

变化率越大，说明当前 object 的更新越频繁，做 merge 的收益可能更大。但是综合考虑之前的限制：
1. 考虑 Level。如果当前 object 本身就很大，变化率高就是自然的。
2. 考虑 zonemap。对于频繁更新的object，自然希望和其他 object 不要有重叠。通过搜索，如果发现和这个 object zonemap 重叠的比较多，就需要打包一起 merge，如果只有他一个，反而需要降低优先级。


变化率低，说明当前 object 基本没有更新，典型场景是 insert、ap 场景。同时需要考虑：
1. 发现长时间静置，则可以更加激进地去 merge 出更大的 object。
2. 充分预估 merge 的资源消耗。







### 执行

经过候选策略，每个表都呈现出了需要 merge 的 object。执行过程则需要：
1. 根据目前的资源情况，对各表的 merge 请求的资源预估以及优先级排序。
2. 记录正在执行的 merge 任务。避免重复


### 模拟器

对于一个固定的 object list、tombstone object list，给定如下的信息：
1. 基本负载模型(insert、update、delete)
    - 平均 tps
    - 修改分布情况
1. object 读写延时
2. cache 命中率
3. cpu 核数
4. 内存
5.  ...

按照时间顺序，可以模拟出代价，方便后续的策略迭代。需要 metric 的进一步迭代

## Tombstone

TableDataView
- In Mem Rows
- In Mem Rows Tombstone
- In Mem Objects Tombstone
- Row Objects
- Row Tombstone Objects


核心点，对一个表的插入和删除，可以看做是往两个独立表插入，通过 flush，形成各自的 objects。

### 选择与统计

对一个表而言，选择活跃的 tombstone object list

1. 单个 Object 基本信息，在创建时确定
    - block 数量
    - 总行数

2. Tombstone Object 上变化信息
    - 剩余的有效 tombstone 数量，随着 row object 的 merge 而减少。通过查 object tombstone 计算空洞

3. Tomestone Object 的整体信息
    - object 之间的 rowid 的重叠


### 候选策略

因为 deletes 的量数量通常不会太大，而且经过 merge 会被消费，所以这里主要的选择优先级是：
1. 减小数量
2. 减少空洞
3. 减少重叠


### 细节流程

#### flush - row object

1. 如果发现没有可以插入的 object，在事务里新建 object。满足当前 blk 可见性规则。createAt <= ts < deleteAt
3. 插入 row
4. 内存中的 object 依然以 In Memory Row 的结构对外提供查询
5. 当需要 flush 时，尝试 freeze object。成功后，新的 append 转 1 走新建流程
6. flush 时，收集全部 Appends，按照 rowid 排序，merge 成新的 objects，删除旧 object
8. flush row object commit 时，修改表的 snapshot 组成
    1. 建立被软删的 object 到新创建 object 的 transfer page。
    2. 把执行过程中的新 deletes 挑出来，按找行对照关系，写入新的 deletes
        > 同时也需要重新考虑这些转移的 deletes 的时间戳被抹平为 merge commit ts 的问题。所以 In Memory Rows Tombstone 需要增加列，记录类型，如果是 merge 产生的，额外记录 commits 时间
    3. 修改 snapshot：
        1. 清理 In Mem Rows
        2. 增加 Row Objects

#### merge - row object

1. 消费当前能看到的 deletes tombstone(包含 in mem 和 objects)，merge 后写成没有 deletes 的 row objects。删除被 merge 的 object
2. 在 merge commit 时
    1. 建立 block transfer page。同 flush row object
    2. 转移新产生的 deletes 到新 object。同 flush row object
    3. 修改 snapshot：
        1. 增加 Row Objects
        3. 增加 In Mem Objects Tombstone


#### flush - Tombstone object

flush tombstone 的过程和 flush row 基本一样。schema 变了而已。


#### merge - Tombstone object

随着 Row objects 的 merge，那些指向已经删除的 Row Object 的 deletes row 失去了意义，对后续查询而言，只是空洞。因此需要做 merge。

1. 开始时，**拿到此刻的 In Mem Objects Tombstone，凡是指向这些 objects 上的 tombstone，都不体现在 merge 之后产生的 object 中**
2. 如果执行过程中，有新的 Row Object 被删除，也不影响当前执行，所以 merge tombstone 产生的 object 中可能依然有空洞，没有关系，不读就行，所以也不会产生回滚。
3. merge tombstone commit 时，修改表 snapshot 组成
    1. 增加 Row Tombstone Objects



