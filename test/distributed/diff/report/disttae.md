# 批量Git Diff报告

生成时间: 2025-10-09 16:50:07.186788
比较commit: HEAD
总文件数: 5

## 1. ranges_filters.result

**文件路径:** disttae_filters/ranges_filters/ranges_filters.result

**统计信息:**
- 当前版本行数: 32
- Git版本行数: 32

**内容差异 (1 个):**

### 1. 内容被修改 (文件1行 15-15, 文件2行 15-15)

**当前版本内容:**
```
- add_fault_point(fj/cn/flush_small_objs,:::,echo,40,testdb.t2)
```

**Git历史版本内容:**
```
+ add_fault_point('fj/cn/flush_small_objs',':::','echo',40,'testdb.t2');
```

---

## 2. block_reader_filter.result

**文件路径:** disttae_filters/reader_filters/block_reader/block_reader_filter.result

**统计信息:**
- 当前版本行数: 132
- Git版本行数: 133

**内容差异 (6 个):**

### 1. 内容被修改 (文件1行 28-28, 文件2行 28-28)

**当前版本内容:**
```
- add_fault_point(fj/cn/flush_small_objs,:::,echo,40,testdb.t2)
```

**Git历史版本内容:**
```
+ add_fault_point(fj/cn/flush_small_objs,:::,echo,40,tesdb.t2)
```

### 2. 内容被修改 (文件1行 48-48, 文件2行 48-48)

**当前版本内容:**
```
- a
```

**Git历史版本内容:**
```
+ distinct a
```

### 3. 内容被修改 (文件1行 51-51, 文件2行 51-51)

**当前版本内容:**
```
- a
```

**Git历史版本内容:**
```
+ distinct a
```

### 4. 内容被修改 (文件1行 56-56, 文件2行 56-56)

**当前版本内容:**
```
- a
```

**Git历史版本内容:**
```
+ distinct a
```

### 5. 内容被修改 (文件1行 63-63, 文件2行 63-63)

**当前版本内容:**
```
- add_fault_point(fj/cn/flush_small_objs,:::,echo,40,testdb.t3)
```

**Git历史版本内容:**
```
+ add_fault_point(fj/cn/flush_small_objs,:::,echo,40,testdb.t1)
```

### 6. 文件2中独有的内容 (行 111-111)

**Git历史版本内容:**
```
+ b
```

---

## 3. partition_state_primary_key_index_filter.result

**文件路径:** disttae_filters/reader_filters/partition_reader/partition_state_primary_key_index_filter.result

**统计信息:**
- 当前版本行数: 86
- Git版本行数: 86

**内容差异 (6 个):**

### 1. 文件1中独有的内容 (行 36-36)

**当前版本内容:**
```
- 3
```

### 2. 文件2中独有的内容 (行 38-38)

**Git历史版本内容:**
```
+ 3
```

### 3. 文件1中独有的内容 (行 41-41)

**当前版本内容:**
```
- 3
```

### 4. 文件2中独有的内容 (行 43-43)

**Git历史版本内容:**
```
+ 3
```

### 5. 文件2中独有的内容 (行 80-83)

**Git历史版本内容:**
```
+ -1
+ -2
+ -3
+ -4
```

### 6. 文件1中独有的内容 (行 81-84)

**当前版本内容:**
```
- -4
- -3
- -2
- -1
```

---

## 4. mo_table_stats1.result

**文件路径:** mo_table_stats/mo_table_stats1.result

**统计信息:**
- 当前版本行数: 107
- Git版本行数: 107

**内容差异 (7 个):**

### 1. 内容被修改 (文件1行 17-17, 文件2行 17-17)

**当前版本内容:**
```
- "dd1dccb4-4d3c-41f8-b482-5251dc7a41bf": "move on: true to false"
```

**Git历史版本内容:**
```
+ "dd1dccb4-4d3c-41f8-b482-5251dc7a41bf": "move on state,true to false"
```

### 2. 内容被修改 (文件1行 30-30, 文件2行 30-30)

**当前版本内容:**
```
- "dd1dccb4-4d3c-41f8-b482-5251dc7a41bf": "force update: false to true"
```

**Git历史版本内容:**
```
+ "dd1dccb4-4d3c-41f8-b482-5251dc7a41bf": "force update,false to true"
```

### 3. 内容被修改 (文件1行 38-38, 文件2行 38-38)

**当前版本内容:**
```
- mo_ctl(cn,MoTableStats,force_update:false)
```

**Git历史版本内容:**
```
+ mo_ctl(cn,MoTableStats,force_update:true)
```

### 4. 内容被修改 (文件1行 42-42, 文件2行 42-42)

**当前版本内容:**
```
- "dd1dccb4-4d3c-41f8-b482-5251dc7a41bf": "force update: true to false"
```

**Git历史版本内容:**
```
+ "dd1dccb4-4d3c-41f8-b482-5251dc7a41bf": "force update,true to false"
```

### 5. 内容被修改 (文件1行 53-53, 文件2行 53-53)

**当前版本内容:**
```
- "dd1dccb4-4d3c-41f8-b482-5251dc7a41bf": "force update: false to true"
```

**Git历史版本内容:**
```
+ "dd1dccb4-4d3c-41f8-b482-5251dc7a41bf": "force update,false to true"
```

### 6. 内容被修改 (文件1行 60-60, 文件2行 60-60)

**当前版本内容:**
```
- mo_ctl(cn,MoTableStats,force_update:false)
```

**Git历史版本内容:**
```
+ mo_ctl(cn,MoTableStats,force_update:true)
```

### 7. 内容被修改 (文件1行 64-64, 文件2行 64-64)

**当前版本内容:**
```
- "dd1dccb4-4d3c-41f8-b482-5251dc7a41bf": "force update: true to false"
```

**Git历史版本内容:**
```
+ "dd1dccb4-4d3c-41f8-b482-5251dc7a41bf": "force update,true to false"
```

---

## 5. mo_table_stats3.result

**文件路径:** mo_table_stats/mo_table_stats3.result

**统计信息:**
- 当前版本行数: 99
- Git版本行数: 99

**内容差异 (2 个):**

### 1. 内容被修改 (文件1行 4-42, 文件2行 4-42)

**当前版本内容:**
```
- mo_database Tae Dynamic 8 0 2109 0 0 NULL 0 2025-10-09 08:42:56 NULL NULL utf8mb4_bin NULL 0 moadmin
- mo_tables Tae Dynamic 129 0 55570 0 0 NULL 0 2025-10-09 08:42:56 NULL NULL utf8mb4_bin NULL 0 moadmin
- mo_columns Tae Dynamic 1384 0 72360 0 0 NULL 0 2025-10-09 08:42:56 NULL NULL utf8mb4_bin NULL 0 moadmin
- mo_account Tae Dynamic 1 0 1394 0 0 NULL 0 2025-10-09 08:42:58 NULL NULL utf8mb4_bin NULL 0 moadmin
- mo_cache null null null null null null null null null 2025-10-09 08:42:58 null null null null null VIEW 0 moadmin
- mo_cdc_task Tae Dynamic 0 0 0 0 0 NULL 0 2025-10-09 08:42:58 NULL NULL utf8mb4_bin NULL 0 moadmin
- mo_cdc_watermark Tae Dynamic 0 0 0 0 0 NULL 0 2025-10-09 08:42:58 NULL NULL utf8mb4_bin NULL 0 moadmin
- mo_configurations null null null null null null null null null 2025-10-09 08:42:58 null null null null null VIEW 0 moadmin
- mo_data_key Tae Dynamic 1 0 1218 0 0 NULL 0 2025-10-09 08:42:58 NULL NULL utf8mb4_bin NULL 0 moadmin
- mo_foreign_keys Tae Dynamic 0 0 0 0 0 NULL 0 2025-10-09 08:42:58 NULL NULL utf8mb4_bin NULL 0 moadmin
- mo_indexes Tae Dynamic 116 0 6764 0 0 NULL 0 2025-10-09 08:42:57 NULL NULL utf8mb4_bin NULL 0 moadmin
- mo_iscp_log Tae Dynamic 0 0 0 0 0 NULL 0 2025-10-09 08:42:58 NULL NULL utf8mb4_bin NULL 0 moadmin
- mo_locks null null null null null null null null null 2025-10-09 08:42:58 null null null null null VIEW 0 moadmin
- mo_merge_settings Tae Dynamic 0 0 0 0 0 NULL 0 2025-10-09 08:42:58 NULL NULL utf8mb4_bin NULL 0 moadmin
- mo_mysql_compatibility_mode Tae Dynamic 4 0 1273 0 0 NULL 0 2025-10-09 08:42:58 NULL NULL utf8mb4_bin NULL 0 moadmin
- mo_partition_metadata Tae Dynamic 0 0 0 0 0 NULL 0 2025-10-09 08:42:58 NULL NULL utf8mb4_bin NULL 0 moadmin
- mo_partition_tables Tae Dynamic 0 0 0 0 0 NULL 0 2025-10-09 08:42:58 NULL NULL utf8mb4_bin NULL 0 moadmin
- mo_pitr Tae Dynamic 0 0 0 0 0 NULL 0 2025-10-09 08:42:58 NULL NULL utf8mb4_bin NULL 0 moadmin
- mo_pubs Tae Dynamic 0 0 0 0 0 NULL 0 2025-10-09 08:42:58 NULL NULL utf8mb4_bin NULL 0 moadmin
- mo_role Tae Dynamic 2 0 1081 0 0 NULL 0 2025-10-09 08:42:58 NULL NULL utf8mb4_bin NULL 0 moadmin
- mo_role_grant Tae Dynamic 0 0 0 0 0 NULL 0 2025-10-09 08:42:58 NULL NULL utf8mb4_bin NULL 0 moadmin
- mo_role_privs Tae Dynamic 35 0 2713 0 0 NULL 0 2025-10-09 08:42:58 NULL NULL utf8mb4_bin NULL 0 moadmin
- mo_sessions null null null null null null null null null 2025-10-09 08:42:58 null null null null null VIEW 0 moadmin
- mo_shards Tae Dynamic 0 0 0 0 0 NULL 0 2025-10-09 08:42:58 NULL NULL utf8mb4_bin NULL 0 moadmin
- mo_shards_metadata Tae Dynamic 0 0 0 0 0 NULL 0 2025-10-09 08:42:58 NULL NULL utf8mb4_bin NULL 0 moadmin
- mo_snapshots Tae Dynamic 0 0 0 0 0 NULL 0 2025-10-09 08:42:58 NULL NULL utf8mb4_bin NULL 0 moadmin
- mo_stages Tae Dynamic 0 0 0 0 0 NULL 0 2025-10-09 08:42:58 NULL NULL utf8mb4_bin NULL 0 moadmin
- mo_stored_procedure Tae Dynamic 0 0 0 0 0 NULL 0 2025-10-09 08:42:58 NULL NULL utf8mb4_bin NULL 0 moadmin
- mo_subs Tae Dynamic 0 0 0 0 0 NULL 0 2025-10-09 08:42:58 NULL NULL utf8mb4_bin NULL 0 moadmin
- mo_table_partitions Tae Dynamic 0 0 0 0 0 NULL 0 2025-10-09 08:42:57 NULL NULL utf8mb4_bin NULL 0 moadmin
- mo_table_stats_alpha Tae Dynamic 34 0 6522 0 0 NULL 0 2025-10-09 08:42:58 NULL NULL utf8mb4_bin NULL 0 moadmin
- mo_transactions null null null null null null null null null 2025-10-09 08:42:58 null null null null null VIEW 0 moadmin
- mo_upgrade Tae Dynamic 0 0 0 0 0 NULL 0 2025-10-09 08:42:58 NULL NULL utf8mb4_bin NULL 0 moadmin
- mo_upgrade_tenant Tae Dynamic 0 0 0 0 0 NULL 0 2025-10-09 08:42:58 NULL NULL utf8mb4_bin NULL 0 moadmin
- mo_user Tae Dynamic 2 0 2156 0 0 NULL 0 2025-10-09 08:42:58 NULL NULL utf8mb4_bin NULL 0 moadmin
- mo_user_defined_function Tae Dynamic 0 0 0 0 0 NULL 0 2025-10-09 08:42:58 NULL NULL utf8mb4_bin NULL 0 moadmin
- mo_user_grant Tae Dynamic 4 0 1037 0 0 NULL 0 2025-10-09 08:42:58 NULL NULL utf8mb4_bin NULL 0 moadmin
- mo_variables null null null null null null null null null 2025-10-09 08:42:58 null null null null null VIEW 0 moadmin
- mo_version Tae Dynamic 1 0 1074 0 0 NULL 0 2025-10-09 08:42:58 NULL NULL utf8mb4_bin NULL 0 moadmin
```

**Git历史版本内容:**
```
+ mo_database Tae Dynamic 8 0 2158 0 0 NULL 0 2025-09-01 10:22:32 NULL NULL utf8mb4_bin NULL 0 moadmin
+ mo_tables Tae Dynamic 129 0 52461 0 0 NULL 0 2025-09-01 10:22:32 NULL NULL utf8mb4_bin NULL 0 moadmin
+ mo_columns Tae Dynamic 1384 0 71602 0 0 NULL 0 2025-09-01 10:22:32 NULL NULL utf8mb4_bin NULL 0 moadmin
+ mo_account Tae Dynamic 1 0 1399 0 0 NULL 0 2025-09-01 10:22:32 NULL NULL utf8mb4_bin NULL 0 moadmin
+ mo_cache null null null null null null null null null 2025-09-01 10:22:32 null null null null null VIEW 0 moadmin
+ mo_cdc_task Tae Dynamic 0 0 0 0 0 NULL 0 2025-09-01 10:22:32 NULL NULL utf8mb4_bin NULL 0 moadmin
+ mo_cdc_watermark Tae Dynamic 0 0 0 0 0 NULL 0 2025-09-01 10:22:32 NULL NULL utf8mb4_bin NULL 0 moadmin
+ mo_configurations null null null null null null null null null 2025-09-01 10:22:32 null null null null null VIEW 0 moadmin
+ mo_data_key Tae Dynamic 1 0 1218 0 0 NULL 0 2025-09-01 10:22:32 NULL NULL utf8mb4_bin NULL 0 moadmin
+ mo_foreign_keys Tae Dynamic 0 0 0 0 0 NULL 0 2025-09-01 10:22:32 NULL NULL utf8mb4_bin NULL 0 moadmin
+ mo_indexes Tae Dynamic 116 0 6496 0 0 NULL 0 2025-09-01 10:22:32 NULL NULL utf8mb4_bin NULL 0 moadmin
+ mo_iscp_log Tae Dynamic 0 0 0 0 0 NULL 0 2025-09-01 10:22:32 NULL NULL utf8mb4_bin NULL 0 moadmin
+ mo_locks null null null null null null null null null 2025-09-01 10:22:32 null null null null null VIEW 0 moadmin
+ mo_merge_settings Tae Dynamic 1 0 1504 0 0 NULL 0 2025-09-01 10:22:32 NULL NULL utf8mb4_bin NULL 0 moadmin
+ mo_mysql_compatibility_mode Tae Dynamic 4 0 1451 0 0 NULL 0 2025-09-01 10:22:32 NULL NULL utf8mb4_bin NULL 0 moadmin
+ mo_partition_metadata Tae Dynamic 0 0 0 0 0 NULL 0 2025-09-01 10:22:32 NULL NULL utf8mb4_bin NULL 0 moadmin
+ mo_partition_tables Tae Dynamic 0 0 0 0 0 NULL 0 2025-09-01 10:22:32 NULL NULL utf8mb4_bin NULL 0 moadmin
+ mo_pitr Tae Dynamic 0 0 0 0 0 NULL 0 2025-09-01 10:22:32 NULL NULL utf8mb4_bin NULL 0 moadmin
+ mo_pubs Tae Dynamic 0 0 0 0 0 NULL 0 2025-09-01 10:22:32 NULL NULL utf8mb4_bin NULL 0 moadmin
+ mo_role Tae Dynamic 2 0 1086 0 0 NULL 0 2025-09-01 10:22:32 NULL NULL utf8mb4_bin NULL 0 moadmin
+ mo_role_grant Tae Dynamic 0 0 0 0 0 NULL 0 2025-09-01 10:22:32 NULL NULL utf8mb4_bin NULL 0 moadmin
+ mo_role_privs Tae Dynamic 35 0 2713 0 0 NULL 0 2025-09-01 10:22:32 NULL NULL utf8mb4_bin NULL 0 moadmin
+ mo_sessions null null null null null null null null null 2025-09-01 10:22:32 null null null null null VIEW 0 moadmin
+ mo_shards Tae Dynamic 0 0 0 0 0 NULL 0 2025-09-01 10:22:32 NULL NULL utf8mb4_bin NULL 0 moadmin
+ mo_shards_metadata Tae Dynamic 0 0 0 0 0 NULL 0 2025-09-01 10:22:32 NULL NULL utf8mb4_bin NULL 0 moadmin
+ mo_snapshots Tae Dynamic 0 0 0 0 0 NULL 0 2025-09-01 10:22:32 NULL NULL utf8mb4_bin NULL 0 moadmin
+ mo_stages Tae Dynamic 0 0 0 0 0 NULL 0 2025-09-01 10:22:32 NULL NULL utf8mb4_bin NULL 0 moadmin
+ mo_stored_procedure Tae Dynamic 0 0 0 0 0 NULL 0 2025-09-01 10:22:32 NULL NULL utf8mb4_bin NULL 0 moadmin
+ mo_subs Tae Dynamic 0 0 0 0 0 NULL 0 2025-09-01 10:22:32 NULL NULL utf8mb4_bin NULL 0 moadmin
+ mo_table_partitions Tae Dynamic 0 0 0 0 0 NULL 0 2025-09-01 10:22:32 NULL NULL utf8mb4_bin NULL 0 moadmin
+ mo_table_stats_alpha Tae Dynamic 103 0 9921 0 0 NULL 0 2025-09-01 10:22:32 NULL NULL utf8mb4_bin NULL 0 moadmin
+ mo_transactions null null null null null null null null null 2025-09-01 10:22:32 null null null null null VIEW 0 moadmin
+ mo_upgrade Tae Dynamic 0 0 0 0 0 NULL 0 2025-09-01 10:22:32 NULL NULL utf8mb4_bin NULL 0 moadmin
+ mo_upgrade_tenant Tae Dynamic 0 0 0 0 0 NULL 0 2025-09-01 10:22:32 NULL NULL utf8mb4_bin NULL 0 moadmin
+ mo_user Tae Dynamic 2 0 2159 0 0 NULL 0 2025-09-01 10:22:32 NULL NULL utf8mb4_bin NULL 0 moadmin
+ mo_user_defined_function Tae Dynamic 0 0 0 0 0 NULL 0 2025-09-01 10:22:32 NULL NULL utf8mb4_bin NULL 0 moadmin
+ mo_user_grant Tae Dynamic 4 0 1038 0 0 NULL 0 2025-09-01 10:22:32 NULL NULL utf8mb4_bin NULL 0 moadmin
+ mo_variables null null null null null null null null null 2025-09-01 10:22:32 null null null null null VIEW 0 moadmin
+ mo_version Tae Dynamic 1 0 1073 0 0 NULL 0 2025-09-01 10:22:32 NULL NULL utf8mb4_bin NULL 0 moadmin
```

### 2. 内容被修改 (文件1行 50-50, 文件2行 50-50)

**当前版本内容:**
```
- t1 Tae Dynamic 100000 0 804690 0 0 NULL 0 2025-10-09 08:44:26 NULL NULL utf8mb4_bin NULL 0 moadmin
```

**Git历史版本内容:**
```
+ t1 Tae Dynamic 100000 0 804690 0 0 NULL 0 2025-09-02 15:36:09 NULL NULL utf8mb4_bin NULL 0 moadmin
```

---

