早前的 TAE 分析文章(TAE-MatrixOne云原生事务与分析引擎)提到 TN 目前拥有三个职责：
1. 处理提交的事务
2. 为 CN 提供 Logtail 服务
3. 转存最新的事务数据至对象存储中，并且推动日志窗口

其中 1、3 完成时，都会产生状态变化，比如数据成功写入内存，或者成功写入对象存储，而 Logtail 是一种日志同步协议，面向 CN 以低成本的方式同步 TN 的部分状态。CN 通过获取 Logtail，在本地重建出必要的，可读的数据。作为 MatrixOne 存算分离架构的关键协议，Logtail 有以下特点：
- 串联协同 CN 和 TN，本质是一个日志复制状态机，让 CN 同步 TN 部分状态
- Logtail 支持以 pull 和 push 两种模式获取
    - push 有更强的实时性，不断地从 TN 同步增量日志到 CN
    - pull 支持指定任意时间区间，同步表的快照，也可以按需同步后续产生的增量日志信息
- Logtail 支持按照表级别进行订阅、收集，在多 CN 支持方面更加灵活，更好地实现 CN 负载的均衡


## Logtail 协议内容

Logtail 协议的目前主体内容用简单的语言来说，分为两个部分，内存数据和元数据，核心区别在于是否已经转存到对象存储中。

一次事务 commit 产生的更新，在转存至对象存储前，其日志在 logtail 中会以内存数据的形式呈现。任何数据上的修改都可以归结到插入和删除两种形式：对于插入，logtail 信息包含该数据行的 row-id，commit-timestamp 以及表定义中的列；对于删除，logtail 信息则包含 row-id，commit-timestamp 以及主键列。这样的内存数据传递到 CN 后，会在内存中以 Btree 的形式组织，对上层提供查询。

显然内存数据不能一直保留在内存中增加内存压力，通过时间或者容量的限制，内存数据会被 flush 到对象存储上形成一个 object。object 包含一个或者多个 block。block 是表数据的最小存储单位，一个 block 包含的行数不超过固定上限，目前默认是 8192 行。**当 flush 完成时，logtail 把 block 元数据传递给 CN，CN 根据事务时间戳，过滤出可见的 block list，读取出 block 中的内容，再综合内存数据，得到某个时刻数据的完整视图。**


上面仅仅是最基础的过程，随着一些性能优化的引入，会呈现更多的细节，比如:

### checkpoint

当 TN 运行一段时间，在某个时刻做 checkpoint，这个时刻前的全部数据都已经转存到对象存储，因此把这些元数据统一收集并精简，得到一个"压缩包"，当一个新启动的 CN 链接到 TN，获取第一次 logtail，如果订阅时间戳大于 checkpoint 的时间戳，就可以把 checkpoint 的元数据(一个字符串)通过 logtail 传递，直接在 CN 读取 checkpoint 时刻之前产生的 block 信息，避免了在网络上从零传递 block 元数据，增加 TN 的 IO 压力。


### 清理内存

当 TN 产生的 block 元数据给到 CN 时，会按照 block-id 去清理之前送达的内存数据。但是在 TN 侧 flush 事务进行过程中，可能同时发生了数据的更新，比如被 flush 的 block 上新产生了删除。如果此时的策略是回滚重试，那么已经写好的数据就完全无效，在更新密集的负载下，会产生大量的回滚，浪费 TN 资源。为了避免这种情况，TN 会继续 commit，导致这部分 flush 事务开始后产生的内存数据就不能从 CN 中删除，于是在 logtail 的 block 元信息中会传递一个时间戳，在这个时间戳之前，属于这个 block 的内存数据才可以从内存中清理。这些未清除的更新，会在下一次 flush 中的异步地刷盘，并且推送的到 CN 进行删除。

### 更快读取

已经转存到对象存储上的 block，可能继续产生删除，读取这些 block 时需要综合内存里的删除数据。为了更快地明确哪些 block 需要结合内存数据，CN 额外维护一个 block 的 Btree 索引，在应用 logtail 的时候就需要更小心地修改这个索引：处理内存数据时增加索引条目，处理 block 元数据时减少索引条目。只能在这个索引中的 block，才有去检查内存数据，在 block 数量比较多的情况下收益比较大。

### More

关于 logtail 协议还有更多的改进空间：
- 后续会向 Logtail 中添加更多的控制信息，让 CN 中的过期数据得到更及时的清理，控制内存使用。
- 尽管经过 TN 后台 merge，多个 block 可能在一个 object 中，但是目前 CN 中仍然按照 block list 组织数据，之后会让元数据在 object 层面传递，减少 block list 的规模。

## Logtail 的产生

如前所述，可以通过 pull 和 push 两种方式进行 logtail 获取。这两种模式的特点不同，接下来分别介绍。

### Pull

如前所述，pull 实际上是同步表的快照，以及后续产生的增量日志信息。

为了达成这个目的，TN 维护了一个按照事务 prepare 时间排序的 txn handle 列表，logtail table，给定任意时刻，通过二分查找得到范围内的 txn handle，再由 txn handle 得到该事务在哪些 block 上做过更新，通过遍历这些 block，就能得到完整的日志。为了加快查找速度，对 txn handle 做了分页，一个页的 bornTs 就是当前页中的 txn handle 的最小 prepare 时间，第一层二分的对象就是这些页。

基于 logtail table，从接收到 pull 请求，主要工作路径如下：
1. 根据已有的 checkpoints，调整请求的时间范围，更早的可以通过 checkpoint 给出
2. 取 logtail table 的一份快照，基于访问者模式用 `RespBuilder` 去迭代这份快照中相关的 txn handle，收集已提交的日志信息
3. 将收集到的日志信息，按照 logtail 协议格式转换，作为响应返回给 CN

![image|800](https://github.com/matrixorigin/matrixone/assets/49832303/7fc68f41-2f33-4e58-8576-aa2e6bf660e9)



```go
type RespBuilder interface {
	OnDatabase(database *DBEntry) error
	OnPostDatabase(database *DBEntry) error
	OnTable(table *TableEntry) error
	OnPostTable(table *TableEntry) error
	OnPostSegment(segment *SegmentEntry) error
	OnSegment(segment *SegmentEntry) error
	OnBlock(block *BlockEntry) error
	BuildResp() (api.SyncLogTailResp, error)
	Close()
}
```

这个过程依然有非常多的优化可以做，这里举几个例子
- 如果请求的时间范围不是最新，在请求范围右侧范围内发生了数据刷盘，那么请求范围内的数据更改应该怎么给呢？我们认为再从对象存储中读出这个区间的更改是不必要的，直接把 block 元数据给到 CN，在 CN 利用时间戳过滤去指定范围的数据是更节省的做法。
- 如果一个时间范围内的内存更改是在太多，通过内存发送也是不实际的，所以会在收集时检查响应大小，如果过大，可以强制刷盘，重新收集。
- Logtail table 也需要根据 checkpoint 做定时清理，避免内存膨胀。另外也对更早的 txn handle 可以做一些数据上的聚合，避免每次都从 txn handle 层面迭代


### Push

Push 的主要目的是更实时地 TN 同步增量日志到 CN。整体流程分为订阅、收集、推送三个阶段。

订阅。一个新 CN 启动后的必要流程，就是作为客户端，和服务端 TN 建立一个 RPC stream，并且订阅 catalog 相关表，当 database、table、column 这些基本信息同步完成后，CN 才能对外提供服务。当 TN 收到订阅一个表的请求时，其实先走一遍 pull 流程，会把 **截止到上次 push 时间戳前** 的所有 logtail 都包含在订阅响应中。目前对一个 CN，logtail 的订阅、取消订阅、数据发送，都发生在一条 RPC stream 链接上，如果它有任何异常，CN 会进入重连流程，直到恢复。一旦订阅成功，后续的 logtail 就是推送增量日志。

收集。在 TN，一个事务完成 WAL 写入后，触发回调执行，在当前事务中收集 logtail。主要流程是遍历 workspace 中的 TxnEntry(一种事务更新的基本容器，直接参与到 commit pipeline 中)，依据其类型，取对应的日志信息转换为 logtail 协议的数据格式。这个收集过程通过 pipeline，和 WAL 的 fysnc 并发执行，减少阻塞。

推送。推送阶段主要做一次过滤，如果发现某个 CN 没有订阅该表，就跳过该 CN，避免推送。

按事务收集有一个比较明显的问题，如果一个表长时间没有更新怎么办？CN 怎么知道是没有更新，还是没有收到呢？这里就加入了心跳机制，默认是 2 ms，TN 的 commit 队列中会放入一个 heartbeat 的空事务，不做任何实质性工作，只消耗时间戳，从而触发一次心跳 logtail 发送，告知 CN 此前的所有表数据已经发送过更新，推动 CN 侧的时间戳水位更新。

![image|800](https://github.com/matrixorigin/matrixone/assets/49832303/4a9fdbcd-a12b-49e6-a23d-5cee5ad6e366)

## 总结

至此只是从总体上介绍了 logtail，篇幅有限，挂一漏万，并不能展示 logtail 的全部细节。而恰好 logtail 是对细节要求非常高的模块，充分考验了协议在两个系统中的适用性。如果对细节的把控不足，则会导致数据的漏读、重复读，再加上时间戳的这个变量，经常引发偶现的错误。今后我们会继续迭代 logtail，提升其稳定性，规范流程，力求更加透明可理解。
