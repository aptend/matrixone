# Git版本Diff工具

## 概述

这个工具套件专门用于比较SQL测试结果文件的当前版本和git历史版本之间的差异，忽略格式变化，专注于实际的数据差异。

## 工具说明

### 1. `git_diff.py` - 单个文件Git比较工具
比较单个文件的当前版本和git历史版本。

**使用方法:**
```bash
python3 git_diff.py file.result --verbose
```

**参数:**
- `file`: 要比较的文件路径
- `--commit, -c`: 要比较的git commit (默认: HEAD~1)
- `--verbose, -v`: 详细输出
- `--history`: 显示文件的git提交历史

### 2. `batch_git_diff.py` - 批量Git比较工具
自动比较所有修改过的.result文件与git历史版本。

**使用方法:**
```bash
python3 batch_git_diff.py --output report.md
```

**参数:**
- `--commit, -c`: 要比较的git commit (默认: HEAD~1)
- `--output, -o`: 输出报告文件路径
- `--verbose, -v`: 详细输出

## 功能特性

### 忽略的格式变化
1. **SQL元数据header**: 忽略 `#SQL[@5,N20]Result[]` 格式的元数据
2. **分隔符变化**: 将 `  ¦  ` 分隔符标准化为空格
3. **多行数据格式**: 处理转义的 `\n` 和实际换行符
4. **整体删除**: 识别被整体删除的语句和输出

### 识别的实际差异
- 数据值的变化
- SQL语句内容的修改
- 新增或删除的SQL语句
- 查询结果的数据差异

## 使用示例

### 比较单个文件
```bash
# 比较文件与上一个commit
python3 git_diff.py ../cases/ddl/alter.result --verbose

# 比较与指定commit
python3 git_diff.py ../cases/ddl/alter.result --commit HEAD~2 --verbose
```

### 批量比较所有修改的文件
```bash
# 比较所有修改过的.result文件
python3 batch_git_diff.py

# 生成详细报告
python3 batch_git_diff.py --output git_diff_report.md
```

## 输出说明

### 成功情况
当没有发现实际数据差异时，工具会显示：
```
✅ 没有发现实际的数据差异（忽略格式变化）
```

### 发现差异时
工具会详细显示：
1. **内容差异**: 标准化后的行级差异
2. **SQL语句差异**: SQL语句级别的差异
3. **统计信息**: 文件行数、SQL语句数等

## 注意事项

- 工具需要在git仓库中运行
- 文件必须在git版本控制中
- 建议使用 `--verbose` 参数获取详细信息
- 可以使用 `--output` 参数生成详细报告
