# Git Diff 工具更新说明

## 新增功能

### 1. 忽略大小写差异
现在 git_diff 工具在比较文件时会自动忽略大小写的差异。这意味着以下内容会被认为是相同的：
- `SELECT * FROM users` ≈ `select * from USERS`
- `INSERT INTO table1` ≈ `insert into TABLE1`
- `WHERE id = 100` ≈ `where ID = 100`

### 2. 忽略引号类型差异
工具现在会忽略字符串左右引号类型的差异，包括：
- 单引号 `'` 和双引号 `"` 的差异
- 中文引号 `"` `"` 和 `'` `'` 的差异
- 引号转义 `\'` 和 `\"` 的差异

例如，以下内容会被认为是相同的：
- `name = "John"` ≈ `name = 'John'`
- `SELECT "column"` ≈ `SELECT 'column'`
- `说明 = "测试"` ≈ `说明 = '测试'`

### 3. 组合情况
大小写和引号可以同时忽略：
- `SELECT "Name" FROM Users` ≈ `select 'name' from USERS`
- `WHERE id = "123"` ≈ `where ID = '123'`

## 修改的代码

### 修改1: SmartDiff.normalize_line() 方法
在行标准化过程中添加了：
```python
# 转换为小写以忽略大小写差异
normalized = normalized.lower()

# 标准化引号：将所有类型的引号统一处理
normalized = normalized.replace('"', "'")
# 也可以处理中文引号
normalized = normalized.replace('"', "'").replace('"', "'")
normalized = normalized.replace(''', "'").replace(''', "'")
```

### 修改2: SmartDiff.is_real_difference() 方法
添加了 `deep_normalize()` 函数，用于深度标准化比较：
```python
def deep_normalize(text: str) -> str:
    """深度标准化文本，用于最终比较"""
    # 转换为小写
    text = text.lower()
    # 统一所有引号类型
    text = text.replace('"', "'").replace('"', "'").replace('"', "'")
    text = text.replace(''', "'").replace(''', "'")
    # 移除引号转义
    text = text.replace("\\'", "'").replace('\\"', "'")
    # 移除分号等标点符号
    text = text.replace(';', '').strip()
    return text
```

## 使用方法

使用方法与之前完全相同，无需任何额外参数：

```bash
# 比较单个文件
python3 git_diff.py file.result --verbose

# 批量比较
python3 batch_git_diff.py --output report.md
```

## 测试验证

运行测试脚本验证功能：
```bash
python3 test_ignore_case_and_quotes.py
```

所有测试均已通过 ✅

## 影响范围

这些修改会影响：
1. 单文件比较工具 `git_diff.py`
2. 批量比较工具 `batch_git_diff.py`（因为它使用了 SmartDiff 类）

这些修改向后兼容，不会破坏现有的使用场景，只是减少了因大小写和引号类型不同而产生的误报。


