# Unique Key Dedup Hang 详细出错过程

## 执行计划分析

从 `plan-hung` 可以看到有两个 RIGHT DEDUP JOIN：

1. **Join[3]** - Primary Key Dedup（卡住的 join）
   - Join Cond: `(t1.a = t2.a) shuffle: range(t1.a)`
   - 列引用，支持 shuffle
   - 计划中标记为 `shuffle: range(t1.a)`

2. **Join[5]** - Unique Key Dedup（使用 if 表达式的 join）
   - Join Cond: `(if((t2.a IS NULL), t1.b, (null)) = #[1,0])`
   - if 表达式，不支持 shuffle
   - Join[3] 是 Join[5] 的 left child（probe side）

## 详细出错过程

### 步骤 1: 计划生成阶段 - Join[5] 的 Shuffle 判断失败

**位置**: `pkg/sql/plan/shuffle.go:464-561` 的 `determineShuffleForJoin` 函数

**执行流程**:
1. 处理 Join[5]（unique key dedup，join 条件为 if 表达式）
2. 调用 `GetHashColumn(expr0)` 提取列引用（`pkg/sql/plan/shuffle.go:547`）
3. `expr0` 是 `if((t2.a IS NULL), t1.b, (null))`，这是一个 `*plan.Expr_F` 类型
4. `GetHashColumn` 函数（`pkg/sql/plan/shuffle.go:367-376`）检测到 `*plan.Expr_F` 类型，直接返回 `nil, -1`
5. `determineShuffleForJoin` 在第 548 行检查 `if leftHashCol == nil`，条件为真
6. 函数直接 return，`node.Stats.HashmapStats.Shuffle` 保持为 `false`（默认值）

**结果**: Join[5] 的 `Shuffle = false`

**关键点**: Join[3] 是 Join[5] 的子节点，但 Join[3] 的 shuffle 判断是**独立**进行的。Join[3] 的 join 条件是列引用 `t1.a = t2.a`，所以 `GetHashColumn` 可以提取到列引用，`Shuffle = true`。

### 步骤 2: 编译阶段 - Join[5] 选择 Broadcast Join 路径

**位置**: `pkg/sql/compile/compile.go:2373-2389` 的 `compileJoin` 函数

**执行流程**:
1. 编译器检查 Join[5] 的 `node.Stats.HashmapStats.Shuffle = false`
2. 进入 `else` 分支（第 2387-2388 行）
3. 调用 `compileProbeSideForBroadcastJoin(node, left, right, probeScopes)`（第 2387 行）
4. 调用 `compileBuildSideForBroadcastJoin(node, rs, buildScopes)`（第 2388 行）

**关键点**: `probeScopes` 包含 Join[3] 的输出。Join[3] 虽然本身是 shuffle join，但它作为 Join[5] 的 probe side，会被广播 join 的编译逻辑处理。

### 步骤 3: Join[3] 的编译 - 使用 ShuffleV2 模式

**位置**: `pkg/sql/compile/compile.go:1106-1119` 的 `compilePlanScope` 函数

**执行流程**:
1. Join[3] 是独立的 join 节点，在编译 Join[5] 之前就已经被编译
2. Join[3] 的 `Shuffle = true`，满足 shuffleV2 的条件（`len(c.cnList) == 1` 且 `ShuffleType != Hash`）
3. 调用 `compileShuffleJoinV2(node, left, right, probeScopes, buildScopes)`（第 2381 行）
4. 在 `constructShuffleJoinOP` 中（`pkg/sql/compile/compile.go:2518-2524`）：
   - 创建 `RightDedupJoin` 算子
   - 设置 `op.ShuffleIdx = int32(i)`（第 2519 行）
   - **由于 `shuffleV2 = true`，设置 `op.ShuffleIdx = -1`**（第 2520-2521 行）

**从日志验证**: `hung-debug.log` Line 1 显示：
```
RightDedupJoin joinmap tag: 2, is shuffle: true, shuffle idx: 0
```
注意：这里的 `shuffle idx: 0` 是构造时的初始值，但在 `constructShuffleJoinOP` 中会被设置为 `-1`。

### 步骤 4: Join[3] 的 ShuffleBuild 构造和 dupOperator

**位置**: `pkg/sql/compile/compile.go:2416-2420` 和 `pkg/sql/compile/operator.go:140-165`

**执行流程**:
1. 在 `compileShuffleJoinV2` 中，调用 `constructShuffleBuild(leftscopes[i].RootOp, c.proc)`（第 2417 行）
2. `constructShuffleBuild` 从 `RightDedupJoin` 复制 `ShuffleIdx = -1`（`pkg/sql/compile/operator.go:2220`）

**从日志验证**: `hung-debug.log` Line 2 显示：
```
constructShuffleBuild from RightDedupJoin: joinmap tag: 2, source shuffleIdx: -1, dest shuffleIdx: -1
```

3. **关键步骤**: ShuffleBuild 被 `dupOperator` 复制到多个并行实例（`pkg/sql/compile/operator.go:149-151`）：
   ```go
   if t.ShuffleIdx == -1 { // shuffleV2
       op.ShuffleIdx = int32(index)  // 设置为 0, 1, 2, ..., 15
   }
   ```

**从日志验证**: `hung-debug.log` Line 5-24 显示 16 个 ShuffleBuild 实例被创建，`dest shuffleIdx` 为 0-15。

**结果**: 
- `RightDedupJoin` 的 `ShuffleIdx = -1`（shuffleV2 模式，表示接收所有分片）
- `ShuffleBuild` 实例的 `ShuffleIdx = 0, 1, 2, ..., 15`（每个实例负责一个分片）

### 步骤 5: Join[5] 的 Broadcast Join 编译 - 影响 Join[3] 的执行

**位置**: `pkg/sql/compile/compile.go:2575-2785` 的 `compileProbeSideForBroadcastJoin` 函数

**执行流程**:
1. Join[5] 是 `Node_DEDUP` 且 `IsRightJoin = true`
2. 调用 `c.newProbeScopeListForBroadcastJoin(probeScopes, true)`（第 2745 行）
   - `forceOneCN = true` 表示需要 merge probe side 的输出
3. 在 `newProbeScopeListForBroadcastJoin` 中（`pkg/sql/compile/compile.go:2564-2573`）：
   - 如果 `forceOneCN = true`，调用 `c.mergeShuffleScopesIfNeeded(probeScopes, false)`（第 2566 行）
   - 如果 `len(probeScopes) > 1`，调用 `c.newMergeScope(probeScopes)`（第 2568 行）
   - **将多个 probe scope 合并为一个 scope**

**关键点**: 
- Join[3] 的输出（`leftscopes`，包含 `RightDedupJoin`）被 merge 成一个 scope
- `rightscopes`（包含 `ShuffleBuild`）是 `leftscopes` 的 `PreScopes`（第 2402 行）
- 当 `leftscopes` 被 merge 时：
  - `newMergeScope` 创建的 merge scope 的 `PreScopes = leftscopes`（原来的多个 leftscopes）
  - 每个 `leftscopes[i]` 的 `PreScopes` 仍然包含 `rightscopes[i]`
  - **`rightscopes` 会通过 `leftscopes` 的 `PreScopes` 被执行，但它们的 `Mcpu` 保持不变**
- 运行时执行顺序：
  1. Merge scope 执行其 `PreScopes`（原来的 `leftscopes`）
  2. 每个 `leftscopes[i]` 先执行其 `PreScopes`（`rightscopes[i]`，包含 `ShuffleBuild`）
  3. 然后执行 `RightDedupJoin`
- **虽然 `ShuffleBuild` 在 `leftscopes` 的执行路径中（通过 `PreScopes`），但 merge 操作只影响 `leftscopes` 本身，不影响 `rightscopes` 的 `Mcpu`**

**Merge 操作将 Mcpu 置为 1 的代码位置**:
- `pkg/sql/compile/compile.go:3837` - `newEmptyMergeScope()`: `rs.NodeInfo = engine.Node{Addr: c.addr, Mcpu: 1}`
- `pkg/sql/compile/compile.go:3888` - `newMergeScopeByCN()`: `rs.NodeInfo.Mcpu = 1`
- `pkg/sql/compile/compile.go:2568` - `newProbeScopeListForBroadcastJoin()` 调用 `newMergeScope()`，会将 Mcpu 置为 1

**为什么 ShuffleBuild 没有被 merge？**

答案：
- 在 `compileShuffleJoinV2` 中（`pkg/sql/compile/compile.go:2391-2423`）：
  - `leftscopes` 是 probe side（包含 Table Scan[0] 的输出）
  - `rightscopes` 是 build side（包含 Table Scan[2] 的输出）
  - `constructShuffleJoinOP` 在 `leftscopes` 上创建 `RightDedupJoin`（第 2414 行）
  - `constructShuffleBuild` 从 `leftscopes[i].RootOp`（`RightDedupJoin`）创建 `ShuffleBuild`，但将 `ShuffleBuild` 放到 `rightscopes[i]` 上（第 2417-2419 行）
- 当编译 Join[5] 时：
  - Join[5] 的 probe side 是 Join[3] 的输出（`leftscopes`，包含 `RightDedupJoin`），会被 `compileProbeSideForBroadcastJoin` merge
  - Join[5] 的 build side 是 Table Scan[4]（index table），**不是** Join[3] 的 build side
  - Join[3] 的 build side（`rightscopes`，包含 `ShuffleBuild`）是独立的 scope，**不在** Join[5] 的编译路径中
- `ShuffleBuild` 所在的 `rightscopes` 在 Join[3] 编译时就已经创建，`NodeInfo.Mcpu > 1`，运行时会被 `dupOperator` 复制，`ShuffleIdx` 被设置为 `0-15`

### 步骤 6: 运行时 - ShuffleBuild 发送消息

**位置**: `pkg/sql/colexec/shufflebuild/build.go:99`

**执行流程**:
1. 16 个 `ShuffleBuild` 实例各自构建完 hash map
2. 发送 `JoinMapMsg` 消息（第 99 行）：
   ```go
   message.SendMessage(message.JoinMapMsg{
       JoinMapPtr: jm, 
       IsShuffle: true, 
       ShuffleIdx: ap.ShuffleIdx,  // 0, 1, 2, ..., 15
       Tag: ap.JoinMapTag
   }, ...)
   ```

**从日志验证**: `hung-debug.log` Line 25-57 显示 16 条 tag 2 的消息被发送，`ShuffleIdx` 为 0-15。

### 步骤 7: 运行时 - RightDedupJoin 接收消息并 Hung

**位置**: `pkg/vm/message/joinMapMsg.go:258-287` 的 `ReceiveJoinMap` 函数

**执行流程**:
1. `RightDedupJoin.build()` 调用 `message.ReceiveJoinMap(2, true, -1, ...)`（`pkg/sql/colexec/rightdedupjoin/join.go:143`）

**从日志验证**: `hung-debug.log` Line 18 显示：
```
RightDedupJoin try receive joinmap: joinMapTag: 2, isShuffle: true, shuffleIdx: -1
```

2. `ReceiveJoinMap` 创建 `MessageReceiver`，等待 tag 2 的消息（第 259 行）
3. 进入循环（第 260-287 行）：
   - 调用 `ReceiveMessage` 获取所有 tag 2 的消息
   - 遍历每条消息，检查匹配条件（第 273-277 行）：
     ```go
     if isShuffle || msg.IsShuffle {
         if shuffleIdx != msg.ShuffleIdx {
             continue  // 跳过不匹配的消息
         }
     }
     ```
4. **匹配失败**：
   - `isShuffle = true`，`msg.IsShuffle = true`
   - 进入 shuffleIdx 匹配分支
   - `shuffleIdx = -1`，`msg.ShuffleIdx = 0, 1, 2, ..., 15`
   - `if (-1 != 0) = true` → 执行 `continue`，跳过所有消息
   - `if (-1 != 1) = true` → 执行 `continue`，跳过所有消息
   - ...（所有 16 条消息都被跳过）
5. **永久阻塞**：
   - 所有消息遍历完成，但没有返回 joinmap
   - 继续循环，再次调用 `ReceiveMessage` 等待新消息
   - 但所有 ShuffleBuild 已经发送完消息，不会有新消息到达
   - `ReceiveMessage` 在 `select` 语句中等待超时后继续循环
   - **形成无限等待循环**

**从调用栈验证**: `gor2` 显示阻塞在 `message.go:239` 的 `ReceiveMessage` 中。

### 步骤 8: Normal 情况的对比 - 关键差异

**Normal 情况**（`normal-debug.log`）:
- Join[3] 和 Join[5] **都是 shuffle join**（`Shuffle = true`），都使用 shuffleV2 模式
- **关键差异**：`RightDedupJoin` **也被 dupOperator 复制**到多个并行实例（`normal-debug.log` Line 5-53）
- 每个 `RightDedupJoin` 实例的 `ShuffleIdx` 从 `-1` 被设置为 `0, 1, 2, ..., 15`（`pkg/sql/compile/operator.go:684-686`）
- 每个 `RightDedupJoin` 实例接收对应 `ShuffleIdx` 的消息（`normal-debug.log` Line 54-73 显示 `shuffleIdx` 为 0-15）
- `RightDedupJoin` 的 `ShuffleIdx` 与 `ShuffleBuild` 的 `ShuffleIdx` 一一对应（0-15）
- 消息匹配成功，执行正常

**Hung 情况**:
- Join[5] 是 broadcast join（`Shuffle = false`），Join[3] 是 shuffleV2 join（`Shuffle = true`）
- **关键差异**：`RightDedupJoin` **没有被 dupOperator 复制**（`hung-debug.log` 中没有 `dupOperator RightDedupJoin` 的日志）
- `RightDedupJoin` 的 `ShuffleIdx` 保持为 `-1`（shuffleV2 模式，但只有一个实例）
- `ShuffleBuild` 的 `ShuffleIdx` 被设置为 `0-15`（dupOperator 设置，16 个实例）
- **消息匹配失败**：`RightDedupJoin` 期待 `ShuffleIdx = -1`，但收到的消息 `ShuffleIdx = 0-15`，`-1` 无法匹配 `0-15` 中的任何值

**为什么 Normal 情况下 RightDedupJoin 会被 dupOperator？**

答案：在 normal 情况下，Join[3] 和 Join[5] 都是独立的 shuffle join，`RightDedupJoin` 会被复制到多个并行实例。但在 hung 情况下，Join[3] 作为 Join[5] 的 probe side，而 Join[5] 是 broadcast join，`compileProbeSideForBroadcastJoin` 会调用 `newProbeScopeListForBroadcastJoin(probeScopes, true)`（`forceOneCN = true`），这会 merge probe side 的输出，导致 `RightDedupJoin` 没有被 dupOperator 复制。

## 根本原因总结

1. **Join[5] 的 join 条件使用了 if 表达式**，导致 `GetHashColumn` 返回 `nil`，`Shuffle = false`
2. **Join[5] 选择 Broadcast Join 路径**，但这不影响 Join[3] 的编译（Join[3] 是独立的 shuffle join）
3. **Join[3] 使用 ShuffleV2 模式**，`RightDedupJoin` 的 `ShuffleIdx = -1`（表示接收所有分片）
4. **ShuffleBuild 被 dupOperator 时**，`ShuffleIdx` 从 `-1` 被设置为 `0-15`（每个实例负责一个分片）
5. **消息匹配逻辑缺陷**：`ReceiveJoinMap` 中的匹配逻辑 `if shuffleIdx != msg.ShuffleIdx` 是**严格匹配**，`-1` 无法匹配 `0-15` 中的任何值
6. **所有消息被跳过**，`ReceiveJoinMap` 永久等待匹配的消息，导致 hung

## 关键代码位置

1. **Shuffle 判断**: `pkg/sql/plan/shuffle.go:367-376` (`GetHashColumn`)
2. **ShuffleV2 设置**: `pkg/sql/compile/compile.go:2520-2521` (`constructShuffleJoinOP`)
3. **ShuffleBuild dupOperator**: `pkg/sql/compile/operator.go:149-151` (`dupOperator`)
4. **消息匹配**: `pkg/vm/message/joinMapMsg.go:273-277` (`ReceiveJoinMap`)

## 关键问题

**为什么 Join[3] 期待 `ShuffleIdx = -1`，但 ShuffleBuild 发送的是 `ShuffleIdx = 0-15`？**

答案：
- `ShuffleIdx = -1` 是 shuffleV2 模式的特殊值，表示接收**所有** shuffle 分片的消息
- `RightDedupJoin` 的 `ShuffleIdx` 在 `constructShuffleJoinOP` 中被设置为 `-1`（第 2521 行）
- `ShuffleBuild` 的 `ShuffleIdx` 在 `dupOperator` 时从 `-1` 被设置为 `0-15`（第 150 行）
- **`ReceiveJoinMap` 的匹配逻辑应该特殊处理 `-1`**，表示接收所有分片的消息，但当前实现是严格匹配，导致 `-1` 无法匹配任何值

**为什么 Normal 情况下可以工作？**

答案：
- Normal 情况下，Join[3] 和 Join[5] 都是独立的 shuffle join，都使用 shuffleV2 模式
- **关键**：`RightDedupJoin` 也被 dupOperator 复制到多个并行实例（16 个）
- 每个 `RightDedupJoin` 实例的 `ShuffleIdx` 从 `-1` 被设置为 `0-15`（与 `ShuffleBuild` 的 `ShuffleIdx` 一一对应）
- 每个 `RightDedupJoin` 实例接收对应 `ShuffleIdx` 的消息，匹配成功
- 而在 hung 情况下，`RightDedupJoin` 没有被 dupOperator 复制，只有一个实例，`ShuffleIdx = -1`，无法匹配 `ShuffleBuild` 发送的 `ShuffleIdx = 0-15` 的消息
