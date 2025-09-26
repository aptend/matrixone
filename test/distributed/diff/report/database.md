# 批量Git Diff报告

生成时间: 2025-09-26 15:10:57.913039
比较commit: HEAD~1
总文件数: 2

## 1. create_table_like.result

**文件路径:** create_table_like.result

**统计信息:**
- 当前版本行数: 37
- Git版本行数: 37

**内容差异 (1 个):**

### 1. 内容被修改 (文件1行 34-34, 文件2行 34-34)

**当前版本内容:**
```
- view1 create view view1 as select * from test utf8mb4 utf8mb4_general_ci
```

**Git历史版本内容:**
```
+ view1 create view view1 as select * from test;utf8mb4 utf8mb4_general_ci
```

---

## 2. system_table.result

**文件路径:** system_table.result

**统计信息:**
- 当前版本行数: 256
- Git版本行数: 293

**内容差异 (11 个):**

### 1. 文件2中独有的内容 (行 5-5)

**Git历史版本内容:**
```
+ character_set_name default_collate_name description maxlen
```

### 2. 文件2中独有的内容 (行 11-11)

**Git历史版本内容:**
```
+ constraint_catalog constraint_schema constraint_name table_catalog table_schema table_name column_name ordinal_position position_in_unique_constraint referenced_table_schema referenced_table_name referenced_column_name
```

### 3. 文件2中独有的内容 (行 13-13)

**Git历史版本内容:**
```
+ query_id seq state duration cpu_user cpu_system context_voluntary context_involuntary block_ops_in block_ops_out messages_sent messages_received page_faults_major page_faults_minor swaps source_function source_file source_line
```

### 4. 文件2中独有的内容 (行 18-18)

**Git历史版本内容:**
```
+ trigger_catalog trigger_schema trigger_name event_manipulation event_object_catalog event_object_schema event_object_table action_order action_condition action_statement action_orientation action_timing action_reference_old_table action_reference_new_table action_reference_old_row action_reference_new_row created sql_mode definer character_set_client collation_connection database_collation
```

### 5. 文件2中独有的内容 (行 20-20)

**Git历史版本内容:**
```
+ grantee table_catalog privilege_type is_grantable
```

### 6. 文件2中独有的内容 (行 25-25)

**Git历史版本内容:**
```
+ host db user table_name column_name timestamp column_priv
```

### 7. 文件2中独有的内容 (行 27-27)

**Git历史版本内容:**
```
+ host db user select_priv insert_priv update_priv delete_priv create_priv drop_priv grant_priv references_priv index_priv alter_priv create_tmp_table_priv lock_tables_priv create_view_priv show_view_priv create_routine_priv alter_routine_priv execute_priv event_priv trigger_priv
```

### 8. 文件2中独有的内容 (行 29-29)

**Git历史版本内容:**
```
+ host db user routine_name routine_type grantor proc_priv timestamp
```

### 9. 文件2中独有的内容 (行 31-31)

**Git历史版本内容:**
```
+ host db user table_name grantor timestamp table_priv column_priv
```

### 10. 文件2中独有的内容 (行 33-33)

**Git历史版本内容:**
```
+ host user select_priv insert_priv update_priv delete_priv create_priv drop_priv reload_priv shutdown_priv process_priv file_priv grant_priv references_priv index_priv alter_priv show_db_priv super_priv create_tmp_table_priv lock_tables_priv execute_priv repl_slave_priv repl_client_priv create_view_priv show_view_priv create_routine_priv alter_routine_priv create_user_priv event_priv trigger_priv create_tablespace_priv ssl_type ssl_cipher x509_issuer x509_subject max_questions max_updates max_connections max_user_connections plugin authentication_string password_expired password_last_changed password_lifetime account_locked create_role_priv drop_role_priv password_reuse_history password_reuse_time password_require_current user_attributes
```

### 11. 文件2中独有的内容 (行 248-274)

**Git历史版本内容:**
```
+ show columns from `PARTITIONS`;
+ Field Type Null Key Default Extra Comment
+ table_catalog VARCHAR(3)NO null
+ table_schema VARCHAR(5000)YES null
+ table_name VARCHAR(5000)YES null
+ partition_name VARCHAR(64)YES null
+ subpartition_name TEXT(4)YES null
+ partition_ordinal_position SMALLINT UNSIGNED(16)YES null
+ subpartition_ordinal_position TEXT(4)YES null
+ partition_method VARCHAR(13)YES null
+ subpartition_method TEXT(4)YES null
+ partition_expression VARCHAR(2048)YES null
+ subpartition_expression TEXT(4)YES null
+ partition_description TEXT(0)YES null
+ table_rows BIGINT(0)YES null
+ avg_row_length BIGINT(0)NO null
+ data_length BIGINT(0)YES null
+ max_data_length BIGINT(0)NO null
+ index_length BIGINT(0)NO null
+ data_free BIGINT(0)NO null
+ create_time TIMESTAMP(0)YES null
+ update_time TEXT(4)YES null
+ check_time TEXT(4)YES null
+ checksum TEXT(4)YES null
+ partition_comment VARCHAR(2048)NO null
+ nodegroup VARCHAR(7)NO null
+ tablespace_name TEXT(4)YES null
```

---

