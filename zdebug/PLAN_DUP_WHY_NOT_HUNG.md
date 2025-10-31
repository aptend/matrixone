# 为什么 plan-dup 不会 Hung？

## 问题

在 `plan-dup` 中，主键的 dedup 因为 `ON DUPLICATE KEY UPDATE` 语句变为 `DedupJoin`（而不是 `RightDedupJoin`）之后，就不发生 hung 了。为什么？

## 关键差异

### 1. Join[3] 的类型差异

**plan-hung**:
- Join[3]: `Join Type: RIGHT DEDUP (FAIL)` → `IsRightJoin = true` → 编译为 `RightDedupJoin`

**plan-dup**:
- Join[3]: `Join Type: DEDUP (UPDATE)` → `IsRightJoin = false` → 编译为 `DedupJoin`

**原因** (`pkg/sql/plan/stats.go:1570-1577`):
```go
case plan.Node_DEDUP:
    if node.OnDuplicateAction != plan.Node_FAIL || node.DedupJoinCtx != nil {
        node.IsRightJoin = false  // 如果是 UPDATE 或者有 DedupJoinCtx，设置为 false
    } else if builder.optimizerHints != nil && builder.optimizerHints.disableRightJoin != 0 {
        node.IsRightJoin = false
    } else if rightChild.Stats.Outcnt > 100 && leftChild.Stats.Outcnt < rightChild.Stats.Outcnt {
        node.IsRightJoin = true
    }
```

### 2. Broadcast Join 编译时的差异

**RightDedupJoin** (`pkg/sql/compile/compile.go:2743-2753`):
```go
if node.IsRightJoin {
    rs = c.newProbeScopeListForBroadcastJoin(probeScopes, true)
    currentFirstFlag := c.anal.isFirst
    for i := range rs {
        op := constructRightDedupJoin(node, leftTyps, rightTyps, c.proc)
        op.SetAnalyzeControl(c.anal.curNodeIdx, currentFirstFlag)
        rs[i].setRootOperator(op)
        rs[i].NodeInfo.Mcpu = 1  // ⚠️ 显式设置为 1
    }
    c.anal.isFirst = false
}
```

**DedupJoin** (`pkg/sql/compile/compile.go:2754-2763`):
```go
} else {
    rs = c.newProbeScopeListForBroadcastJoin(probeScopes, true)
    currentFirstFlag := c.anal.isFirst
    for i := range rs {
        op := constructDedupJoin(node, leftTyps, rightTyps, c.proc)
        op.SetAnalyzeControl(c.anal.curNodeIdx, currentFirstFlag)
        rs[i].setRootOperator(op)
        // ⚠️ 没有显式设置 rs[i].NodeInfo.Mcpu = 1
    }
    c.anal.isFirst = false
}
```

### 3. DedupJoin 的特殊机制

`DedupJoin` 有一个特殊的并行处理机制 (`pkg/sql/colexec/dedupjoin/types.go:84-86`):
```go
Channel  chan *bitmap.Bitmap
NumCPU   uint64
IsMerger bool
```

在 `dupOperator` 中 (`pkg/sql/compile/operator.go:647-674`):
```go
case vm.DedupJoin:
    t := sourceOp.(*dedupjoin.DedupJoin)
    op := dedupjoin.NewArgument()
    if t.Channel == nil {
        t.Channel = make(chan *bitmap.Bitmap, maxParallel)
    }
    op.Channel = t.Channel
    op.NumCPU = uint64(maxParallel)
    op.IsMerger = (index == 0)  // 只有第一个实例是 merger
    // ...
    op.ShuffleIdx = t.ShuffleIdx
    if t.ShuffleIdx == -1 { // shuffleV2
        op.ShuffleIdx = int32(index)  // 设置为 0, 1, 2, ..., N-1
    }
```

**关键点**:
- `DedupJoin` 的所有并行实例共享同一个 `Channel`
- 只有 `index == 0` 的实例是 `IsMerger = true`
- 非 merger 实例处理完数据后，将结果发送到 `Channel`
- Merger 实例从 `Channel` 接收所有结果并合并

### 4. 为什么 DedupJoin 不会 Hung？

虽然 `DedupJoin` 和 `RightDedupJoin` 都使用 `message.ReceiveJoinMap` 来接收消息，但它们的执行流程不同：

**RightDedupJoin**:
- 当 scope 被 merge 后，`Mcpu = 1`，不会被 `dupOperator` 复制
- `ShuffleIdx = -1`（shuffleV2 模式）
- `ShuffleBuild` 被 `dupOperator` 复制，`ShuffleIdx = 0-15`
- `ReceiveJoinMap(-1)` 无法匹配 `ShuffleBuild` 的消息（`ShuffleIdx = 0-15`）
- **Hung**

**DedupJoin**:
- 即使 scope 被 merge，`DedupJoin` 的 `NumCPU` 仍然可能大于 1（取决于 `newProbeScopeListForBroadcastJoin` 的 merge 行为）
- 更重要的是，`DedupJoin` 有一个 `Channel` 机制，可以在多个实例之间协调
- 即使 `Mcpu = 1`，`DedupJoin` 也可能通过其他机制避免 hung

**但实际原因可能更简单**：

在 plan-dup 中，如果 Join[3] 是独立的 shuffle join（不是作为 Join[5] 的 probe side），那么：
- Join[3] 的 `DedupJoin` 不会被 merge
- `DedupJoin` 的 `Mcpu > 1`，会被 `dupOperator` 复制
- `DedupJoin` 的 `ShuffleIdx` 从 `-1` 被设置为 `0-15`
- `ShuffleBuild` 的 `ShuffleIdx` 也从 `-1` 被设置为 `0-15`
- **可以正确匹配，不会 hung**

## 总结

**关键差异**：在 `compileProbeSideForBroadcastJoin` 中，`RightDedupJoin` 会**显式设置** `rs[i].NodeInfo.Mcpu = 1`（第 2751 行），而 `DedupJoin` **不会**显式设置（第 2760 行没有设置）。

这意味着：
1. **plan-hung**: Join[3] 是 `RightDedupJoin`，当作为 Join[5] 的 probe side 时：
   - `newProbeScopeListForBroadcastJoin` 会 merge scope，创建新的 scope，`Mcpu = 1`
   - `RightDedupJoin` 又**显式设置** `rs[i].NodeInfo.Mcpu = 1`（双重保证）
   - 运行时 `newParallelScope` 检查到 `Mcpu == 1`，不调用 `dupOperatorRecursively`
   - `RightDedupJoin` 的 `ShuffleIdx = -1`（shuffleV2 模式）
   - `ShuffleBuild` 被 `dupOperator` 复制，`ShuffleIdx = 0-15`
   - `ReceiveJoinMap(-1)` 无法匹配 `ShuffleBuild` 的消息（`ShuffleIdx = 0-15`）
   - **Hung**

2. **plan-dup**: Join[3] 是 `DedupJoin`，当作为 Join[5] 的 probe side 时：
   - `newProbeScopeListForBroadcastJoin` 会 merge scope，创建新的 scope，`Mcpu = 1`
   - 但 `DedupJoin` **不会**显式设置 `rs[i].NodeInfo.Mcpu = 1`
   - **关键点**：即使 merge 后的 scope 的 `Mcpu = 1`，`DedupJoin` 的 `NumCPU` 字段仍然可能大于 1（`NumCPU` 是在 `dupOperator` 中设置的，基于 `maxParallel`）
   - 或者，`DedupJoin` 可能有其他机制来处理这种情况

**实际原因**：虽然 `DedupJoin` 和 `RightDedupJoin` 都会在 merge 后 `Mcpu = 1`，但 `DedupJoin` 有一个 `Channel` 机制（`pkg/sql/colexec/dedupjoin/join.go:182-197`），可以在多个并行实例之间协调。即使只有一个实例，`DedupJoin` 也可能通过这个机制避免 hung。

**但更可能的原因是**：`DedupJoin` 没有显式设置 `rs[i].NodeInfo.Mcpu = 1`，这意味着即使 merge 后的 scope 的 `Mcpu = 1`，`DedupJoin` 本身可能仍然保持某种并行性，或者 `ShuffleIdx` 可以通过其他方式被正确设置。

**最终答案**：`RightDedupJoin` 在 broadcast join 编译时**显式设置** `rs[i].NodeInfo.Mcpu = 1`，导致它不会被 `dupOperator` 复制，`ShuffleIdx` 保持为 `-1`，无法匹配 `ShuffleBuild` 的消息（`ShuffleIdx = 0-15`）。而 `DedupJoin` 没有显式设置，可能仍然保持某种并行性，或者通过 `Channel` 机制避免了这个问题。

