# 批量Git Diff报告

生成时间: 2025-10-09 16:44:59.152867
比较commit: HEAD
总文件数: 1

## 1. distinct.result

**文件路径:** distinct.result

**统计信息:**
- 当前版本行数: 258
- Git版本行数: 260

**内容差异 (10 个):**

### 1. 文件2中独有的内容 (行 70-71)


**Git历史版本内容:**

**SKIPPED**
```
+ select i+0.0 as i2,count(distinct j)from t3 group by i2;
+ Column 'i2' does not exist
```

### 2. 内容被修改 (文件1行 81-81, 文件2行 83-83)

**当前版本内容:**
```
- (select COUNT(distinct 12))
```

**Git历史版本内容:**
```
+ (select count(distinct 12))
```

### 3. 内容被修改 (文件1行 113-113, 文件2行 115-115)

**当前版本内容:**
```
- 330000000
```

**Git历史版本内容:**
```
+ 3.3E8
```

### 4. 内容被修改 (文件1行 116-116, 文件2行 118-118)

**当前版本内容:**
```
- 110000000.00000001
```

**Git历史版本内容:**
```
+ 1.1000000000000001E8
```

### 5. 内容被修改 (文件1行 133-133, 文件2行 135-135)

**当前版本内容:**
```
- COUNT(distinct a)
```

**Git历史版本内容:**
```
+ count(distinct a)
```

### 6. 内容被修改 (文件1行 173-173, 文件2行 175-175)

**当前版本内容:**
```
- AVG(distinct col_int_nokey)
```

**Git历史版本内容:**
```
+ avg(distinct col_int_nokey)
```

### 7. 内容被修改 (文件1行 177-177, 文件2行 179-179)

**当前版本内容:**
```
- AVG(distinct outr.col_int_nokey)
```

**Git历史版本内容:**
```
+ avg(distinct outr.col_int_nokey)
```

### 8. 内容被修改 (文件1行 224-225, 文件2行 226-227)

**当前版本内容:**
```
- AVG(2)BIT_AND(2)BIT_OR(2)BIT_XOR(2)
- 2 2 2 2
```

**Git历史版本内容:**
```
+ avg(2)bit_and(2)bit_or(2)bit_xor(2)
+ 2.0 2 2 2
```

### 9. 内容被修改 (文件1行 230-231, 文件2行 232-233)

**当前版本内容:**
```
- COUNT(12)COUNT(distinct 12)MIN(2)MAX(2)STD(2)VARIANCE(2)SUM(2)
- 1 1 2 2 0 0 2
```

**Git历史版本内容:**
```
+ count(12)count(distinct 12)min(2)max(2)std(2)variance(2)sum(2)
+ 1 1 2 2 0.0 0.0 2
```

### 10. 内容被修改 (文件1行 250-250, 文件2行 252-252)

**当前版本内容:**
```
- product country_id COUNT(*)COUNT(distinct year)
```

**Git历史版本内容:**
```
+ product country_id count(*)count(distinct year)
```

---

