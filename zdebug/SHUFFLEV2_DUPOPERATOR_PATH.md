# ShuffleV2 模式下 dupOperator 的调用路径

## 概述

在 shuffleV2 模式下，`dupOperator` 用于将算子复制到多个并行实例。对于 `RightDedupJoin` 和 `ShuffleBuild`，它们的 `ShuffleIdx` 从 `-1` 被设置为 `0, 1, 2, ..., N-1`（N 是并行度）。

## 调用路径

### 路径 1: 运行时并行化（Runtime Parallelization）

**调用链**:
```
Scope.ParallelRun()
  └─> buildScanParallelRun() / buildLoadParallelRun()
       └─> newParallelScope()
            └─> dupOperatorRecursively()
                 └─> dupOperator()  [对每个算子类型]
```

**位置**:
- `pkg/sql/compile/scope.go:438` - `Scope.ParallelRun()`
- `pkg/sql/compile/scope.go:517` - `buildScanParallelRun()`
- `pkg/sql/compile/scope.go:500` - `buildLoadParallelRun()`
- `pkg/sql/compile/scope.go:753` - `newParallelScope()`
- `pkg/sql/compile/scope.go:775` - `dupOperatorRecursively()`
- `pkg/sql/compile/operator.go:128` - `dupOperator()`

**触发条件**:
- Scope 的 `NodeInfo.Mcpu > 1`
- Scope 是 TableScan 或 Load 类型
- 在运行时，通过 `ParallelRun()` 触发并行化

**代码片段**:
```go
// pkg/sql/compile/scope.go:753-787
func newParallelScope(s *Scope) (*Scope, []*Scope) {
    if s.NodeInfo.Mcpu == 1 {
        return s, nil
    }
    
    parallelScopes := make([]*Scope, s.NodeInfo.Mcpu)
    for i := 0; i < s.NodeInfo.Mcpu; i++ {
        parallelScopes[i] = newScope(Normal)
        parallelScopes[i].setRootOperator(dupOperatorRecursively(s.RootOp, i, s.NodeInfo.Mcpu))
    }
    // ...
}
```

### 路径 2: ShuffleV2 Join 编译阶段

**对于 Normal 情况**（Join[3] 和 Join[5] 都是独立的 shuffle join）:

**调用链**:
```
compilePlanScope()
  └─> compileJoin()
       └─> compileShuffleJoinV2()
            └─> constructShuffleJoinOP()  [创建 RightDedupJoin, ShuffleIdx = -1]
            └─> constructShuffleBuild()  [创建 ShuffleBuild, ShuffleIdx = -1]
                 
运行时:
Scope.ParallelRun()
  └─> newParallelScope()  [如果 Scope.NodeInfo.Mcpu > 1]
       └─> dupOperatorRecursively()
            └─> dupOperator()  [复制 RightDedupJoin 和 ShuffleBuild]
                 └─> 对于 ShuffleIdx == -1 的情况:
                      op.ShuffleIdx = int32(index)  [设置为 0-15]
```

**位置**:
- `pkg/sql/compile/compile.go:1106` - `compilePlanScope()` 处理 `Node_JOIN`
- `pkg/sql/compile/compile.go:2373` - `compileJoin()` 判断是否使用 shuffleV2
- `pkg/sql/compile/compile.go:2391` - `compileShuffleJoinV2()` 编译 shuffleV2 join
- `pkg/sql/compile/compile.go:2414` - `constructShuffleJoinOP()` 创建 join 算子
- `pkg/sql/compile/compile.go:2417` - `constructShuffleBuild()` 创建 build 算子

**代码片段**:
```go
// pkg/sql/compile/compile.go:2518-2524
case plan.Node_DEDUP:
    if node.IsRightJoin {
        for i := range shuffleJoins {
            op := constructRightDedupJoin(node, leftTyps, rightTyps, c.proc)
            op.ShuffleIdx = int32(i)
            if shuffleV2 {
                op.ShuffleIdx = -1  // shuffleV2 模式，初始值为 -1
            }
            shuffleJoins[i].setRootOperator(op)
        }
    }
```

**dupOperator 处理**:
```go
// pkg/sql/compile/operator.go:675-704
case vm.RightDedupJoin:
    t := sourceOp.(*rightdedupjoin.RightDedupJoin)
    op := rightdedupjoin.NewArgument()
    op.ShuffleIdx = t.ShuffleIdx
    if t.ShuffleIdx == -1 { // shuffleV2
        op.ShuffleIdx = int32(index)  // 设置为 0, 1, 2, ..., N-1
    }
    // ...
```

### 路径 3: Hung 情况 - Broadcast Join 不触发 dupOperator

**对于 Hung 情况**（Join[5] 是 broadcast join，Join[3] 是 shuffleV2 join）:

**调用链**:
```
compilePlanScope()
  └─> compileJoin()
       └─> compileProbeSideForBroadcastJoin()  [Join[5] 的编译]
            └─> newProbeScopeListForBroadcastJoin(probeScopes, true)  [forceOneCN = true]
                 └─> mergeShuffleScopesIfNeeded()  [合并 Join[3] 的输出]
                      └─> newMergeScope()  [合并多个 scope]
                           
运行时:
Scope.ParallelRun()
  └─> 由于 Scope.NodeInfo.Mcpu = 1（被 merge 后）
       └─> newParallelScope() 不会创建并行实例
            └─> 直接返回原 scope，不调用 dupOperatorRecursively()
```

**关键差异**:
- `newProbeScopeListForBroadcastJoin(probeScopes, true)` 中的 `forceOneCN = true`
- 这会调用 `mergeShuffleScopesIfNeeded()`，将多个 shuffle scope 合并成一个
- 合并后的 scope 的 `NodeInfo.Mcpu = 1`
- `newParallelScope()` 检查到 `Mcpu == 1`，直接返回，不调用 `dupOperatorRecursively()`

**代码片段**:
```go
// pkg/sql/compile/compile.go:2564-2573
func (c *Compile) newProbeScopeListForBroadcastJoin(probeScopes []*Scope, forceOneCN bool) []*Scope {
    if forceOneCN { // for right join, we have to merge these input for now
        probeScopes = c.mergeShuffleScopesIfNeeded(probeScopes, false)
        if len(probeScopes) > 1 {
            probeScopes = []*Scope{c.newMergeScope(probeScopes)}
        }
    }
    return probeScopes
}
```

```go
// pkg/sql/compile/scope.go:753-756
func newParallelScope(s *Scope) (*Scope, []*Scope) {
    if s.NodeInfo.Mcpu == 1 {
        return s, nil  // 直接返回，不并行化
    }
    // ...
}
```

## 总结

1. **Normal 情况**: 
   - Join[3] 和 Join[5] 都是独立的 shuffle join
   - 每个 join 的 scope 都有 `NodeInfo.Mcpu > 1`
   - 运行时 `ParallelRun()` 会调用 `newParallelScope()`
   - `newParallelScope()` 调用 `dupOperatorRecursively()` 复制所有算子
   - `RightDedupJoin` 和 `ShuffleBuild` 的 `ShuffleIdx` 从 `-1` 被设置为 `0-15`

2. **Hung 情况**:
   - Join[5] 是 broadcast join，Join[3] 是 shuffleV2 join
   - Join[3] 作为 Join[5] 的 probe side，被 `newProbeScopeListForBroadcastJoin()` merge
   - Merge 后的 scope 的 `NodeInfo.Mcpu = 1`
   - `newParallelScope()` 检查到 `Mcpu == 1`，直接返回，不调用 `dupOperatorRecursively()`
   - `RightDedupJoin` 的 `ShuffleIdx` 保持为 `-1`，无法匹配 `ShuffleBuild` 发送的 `ShuffleIdx = 0-15` 的消息

