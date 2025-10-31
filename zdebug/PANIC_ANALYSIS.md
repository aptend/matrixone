# Panic 问题传导路径分析

## 问题现象

在 `pkg/sql/compile/operator.go:1710` 的 `constructShuffleOperatorForJoin` 函数中发生 nil pointer dereference：
```go
hashCol, typ := plan2.GetHashColumn(expr)
arg.ShuffleColIdx = hashCol.ColPos  // panic: hashCol is nil
```

## 传导路径

### 步骤 1: 用户的修改
**位置**: `pkg/sql/plan/bind_insert.go`

用户将 unique key 的 dedup 条件改为 `if` 表达式：
```go
idxIndexColExpr := createIfExpr(appendedUniqueProjs[idxIdxColName])
appendedUniqueProjs[idxIdxColName] = idxIndexColExpr
```

这导致 Join[5] 的 join 条件变为：`if((t2.a IS NULL), t1.b, (null)) = #[1,0]`

### 步骤 2: 优化器阶段 - `determineShuffleForJoin`

**位置**: `pkg/sql/plan/shuffle.go:465-561`

对于 Join[5]（`Node_DEDUP` 且 `IsRightJoin = true`）：

1. **第 477-481 行**: 检查 `leftChild.Stats.Outcnt > 200000`
   - 如果满足，继续执行
   - 如果不满足，直接返回（不设置 `Shuffle = true`）

2. **第 505-523 行**: 找到第一个等值条件
   ```go
   for i := range node.OnList {
       if isEquiCond(node.OnList[i], leftTags, rightTags) {
           idx = i
           break
       }
   }
   ```

3. **`isEquiCond` 函数** (`pkg/sql/plan/join_order.go:137-154`):
   - 检查是否是 `Expr_F`（函数表达式）且是等号函数 ✓
   - 调用 `getJoinSide(e.F.Args[0], ...)` 和 `getJoinSide(e.F.Args[1], ...)`
   - **关键问题**: `getJoinSide` 对于 `if` 表达式会递归检查所有参数
   - 对于 `if((t2.a IS NULL), t1.b, (null))`:
     - `t2.a` → `JoinSideRight`
     - `t1.b` → `JoinSideLeft`
     - `(null)` → `JoinSideNone`
     - 结果：`JoinSideLeft | JoinSideRight = JoinSideBoth`
   - 对于 `#[1,0]`:
     - 这是一个列引用，来自 right side → `JoinSideRight`
   - **如果 `getJoinSide` 正确返回**，`isEquiCond` 应该返回 `false`（因为 `JoinSideBoth != JoinSideLeft` 且 `JoinSideBoth != JoinSideRight`）

4. **但是**，如果 `isEquiCond` 错误地返回 `true`（可能因为 `getJoinSide` 的处理问题），那么：
   - `idx = 0` 被设置
   - 继续执行到第 538-545 行

5. **第 538-545 行**: 提取表达式
   ```go
   cond := node.OnList[idx]  // idx = 0
   switch condImpl := cond.Expr.(type) {
   case *plan.Expr_F:
       expr0 = condImpl.F.Args[0]  // if((t2.a IS NULL), t1.b, (null))
       expr1 = condImpl.F.Args[1]  // #[1,0]
   }
   ```

6. **第 547-550 行**: 调用 `GetHashColumn(expr0)`
   ```go
   leftHashCol, typ := GetHashColumn(expr0)  // expr0 是 if 表达式
   if leftHashCol == nil {
       return  // 应该在这里返回，不设置 Shuffle = true
   }
   ```

**预期行为**: `GetHashColumn` 返回 `nil`，函数在第 549 行返回，不设置 `Shuffle = true`。

### 步骤 3: 编译阶段 - `compileJoin`

**位置**: `pkg/sql/compile/compile.go:2375-2391`

```go
func (c *Compile) compileJoin(node, left, right *plan.Node, probeScopes, buildScopes []*Scope) []*Scope {
    if node.Stats.HashmapStats.Shuffle {  // 检查 Shuffle 标志
        // ...
        return c.compileShuffleJoin(node, left, right, probeScopes, buildScopes)
    }
    // ...
}
```

**问题**: 如果 `node.Stats.HashmapStats.Shuffle` 被错误地设置为 `true`（可能在某个地方被错误设置），那么会进入 `compileShuffleJoin`。

### 步骤 4: 编译阶段 - `compileShuffleJoin`

**位置**: `pkg/sql/compile/compile.go:2546-2563`

```go
func (c *Compile) compileShuffleJoin(node, left, right *plan.Node, lefts, rights []*Scope) []*Scope {
    shuffleJoins := c.newShuffleJoinScopeList(lefts, rights, node)
    constructShuffleJoinOP(c, shuffleJoins, node, left, right, false)
    // ...
}
```

### 步骤 5: 编译阶段 - `newShuffleJoinScopeList`

**位置**: `pkg/sql/compile/compile.go:4040-4158`

```go
func (c *Compile) newShuffleJoinScopeList(probeScopes, buildScopes []*Scope, n *plan.Node) []*Scope {
    // ...
    if !reuse {
        for i := range probeScopes {
            shuffleProbeOp := constructShuffleOperatorForJoin(int32(bucketNum), n, true)  // ⚠️ 这里！
            // ...
        }
    }
    // ...
    for i := range buildScopes {
        shuffleBuildOp := constructShuffleOperatorForJoin(int32(bucketNum), n, false)  // ⚠️ 这里！
        // ...
    }
}
```

### 步骤 6: Panic 发生

**位置**: `pkg/sql/compile/operator.go:1696-1725`

```go
func constructShuffleOperatorForJoin(bucketNum int32, node *plan.Node, left bool) *shuffle.Shuffle {
    // ...
    cond := node.OnList[node.Stats.HashmapStats.ShuffleColIdx]  // idx = 0
    switch condImpl := cond.Expr.(type) {
    case *plan.Expr_F:
        if left {
            expr = condImpl.F.Args[0]  // if((t2.a IS NULL), t1.b, (null))
        } else {
            expr = condImpl.F.Args[1]  // #[1,0]
        }
    }
    
    hashCol, typ := plan2.GetHashColumn(expr)  // 返回 nil（因为 if 表达式）
    arg.ShuffleColIdx = hashCol.ColPos  // ⚠️ PANIC: hashCol is nil
}
```

## 根本原因

**问题在于**: `determineShuffleForJoin` 在某些情况下可能错误地设置了 `Shuffle = true`，即使 `GetHashColumn` 返回 `nil`。

可能的原因：
1. **`isEquiCond` 错误地返回 `true`**: 
   - `getJoinSide` 对于 `if` 表达式可能没有正确处理
   - 或者 `isEquiCond` 的逻辑有 bug

2. **`Shuffle` 标志在其他地方被设置**:
   - 可能在优化器的其他阶段被错误设置
   - 或者有某些代码路径绕过了 `GetHashColumn` 的检查

3. **`ShuffleColIdx` 的设置问题**:
   - 即使 `GetHashColumn` 返回 `nil`，`ShuffleColIdx` 可能仍然被设置为 `0`
   - 然后在编译时，`constructShuffleOperatorForJoin` 使用这个 `idx` 访问 `OnList[idx]`，但无法从表达式中提取 hash column

## 关键发现

看 `determineShuffleForJoin` 的逻辑，特别是第 518-523 行：

```go
// for now ,only support the first join condition
for i := range node.OnList {
    if isEquiCond(node.OnList[i], leftTags, rightTags) {
        idx = i
        break
    }
}
```

**问题**: `idx` 在这里被设置为 `0`（如果 `isEquiCond` 返回 `true`），但**没有检查 `GetHashColumn` 是否返回 `nil`**。

然后，即使 `GetHashColumn` 返回 `nil` 并在第 549 行返回，**`ShuffleColIdx` 可能已经被设置为 `idx`**（虽然 `Shuffle` 没有被设置为 `true`）。

但是，如果 `Shuffle` 在某个地方被错误设置为 `true`（可能是之前的状态或者其他代码路径），那么：
- `node.Stats.HashmapStats.Shuffle = true`
- `node.Stats.HashmapStats.ShuffleColIdx = 0`（或者之前的值）
- 在编译时，`constructShuffleOperatorForJoin` 使用 `ShuffleColIdx` 访问 `OnList[ShuffleColIdx]`
- 但 `GetHashColumn` 返回 `nil`，导致 panic

## 需要验证的点

1. **`isEquiCond` 对于 `if` 表达式的处理**:
   - `getJoinSide(if((t2.a IS NULL), t1.b, (null)), ...)` 返回什么？
   - 它应该返回 `JoinSideBoth`（因为包含 left 和 right side 的列）
   - 但 `isEquiCond` 需要 `lside == JoinSideLeft && rside == JoinSideRight`
   - 如果 `getJoinSide` 返回 `JoinSideBoth`，`isEquiCond` 应该返回 `false`
   - **但如果 `isEquiCond` 错误地返回 `true`，`idx` 会被设置，但 `GetHashColumn` 会返回 `nil`**

2. **`Shuffle` 标志的设置时机**:
   - 在什么情况下，`Shuffle = true` 会被设置，即使 `GetHashColumn` 返回 `nil`？
   - 是否有其他代码路径设置了 `Shuffle = true`？
   - **第 578-580 行的 recheck 逻辑**应该会将 `Shuffle` 设置为 `false`（对于 `Node_DEDUP` 且 `IsRightJoin = true` 且 `ShuffleType = Hash`）

3. **`ShuffleColIdx` 的设置**:
   - `ShuffleColIdx` 在第 518-523 行被设置（如果 `isEquiCond` 返回 `true`）
   - 但即使 `GetHashColumn` 返回 `nil`，`ShuffleColIdx` 可能仍然保持这个值
   - 如果在某个地方 `Shuffle` 被错误设置为 `true`，编译时会使用这个 `ShuffleColIdx`

4. **可能的问题场景**:
   - `isEquiCond` 错误地返回 `true` → `idx = 0`
   - `GetHashColumn` 返回 `nil` → 在第 549 行返回，`Shuffle` 没有被设置为 `true`
   - 但 `ShuffleColIdx` 可能已经在之前的某个地方被设置为 `0`
   - 如果在某个地方 `Shuffle` 被错误设置为 `true`（可能是之前的优化阶段或者其他代码路径）
   - 编译时，`constructShuffleOperatorForJoin` 使用 `ShuffleColIdx = 0` 访问 `OnList[0]`
   - `GetHashColumn` 返回 `nil` → panic

