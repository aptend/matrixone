# 批量Git Diff报告

生成时间: 2025-10-10 19:01:55.453883
比较commit: HEAD
总文件数: 5

## 1. load_data.result

**文件路径:** load_data.result

**统计信息:**
- 当前版本行数: 1081
- Git版本行数: 1094

**内容差异 (24 个):**

### 1. 文件2中独有的内容 (行 15-15)

**Git历史版本内容:**
```
+ col1 col2 col3 col4 col5 col6 col7 col8
```

### 2. 文件2中独有的内容 (行 117-117)

**Git历史版本内容:**
```
+ col1 col2 col3 col4
```

### 3. 内容被修改 (文件1行 180-180, 文件2行 182-182)

**当前版本内容:**
```
- 1 1 1.00 1.00000
```

**Git历史版本内容:**
```
+ 1.0 1.0 1.00 1.00000
```

### 4. 内容被修改 (文件1行 185-185, 文件2行 187-187)

**当前版本内容:**
```
- 0.0000000001 0.0000000001 0.00 0.00000
```

**Git历史版本内容:**
```
+ 1.0e-10 1.0e-10 0.00 0.00000
```

### 5. 内容被修改 (文件1行 191-191, 文件2行 193-193)

**当前版本内容:**
```
- 1 1 1.00 1.00000
```

**Git历史版本内容:**
```
+ 1.0 1.0 1.00 1.00000
```

### 6. 内容被修改 (文件1行 196-196, 文件2行 198-198)

**当前版本内容:**
```
- 0.0000000001 0.0000000001 0.00 0.00000
```

**Git历史版本内容:**
```
+ 1.0e-10 1.0e-10 0.00 0.00000
```

### 7. 文件2中独有的内容 (行 271-271)

**Git历史版本内容:**
```
+ col1 col2 col3 col4
```

### 8. 文件2中独有的内容 (行 291-294)

**Git历史版本内容:**
```
+ load data infile '$resources/load_data/auto_increment_2.csv' into table t6 fields terminated by ',';
+ duplicate entry '4' for key 'col1'
+ select * from t6;
+ col1 col2 col3
```

### 9. 文件2中独有的内容 (行 317-317)

**Git历史版本内容:**
```
+ a b c d
```

### 10. 文件2中独有的内容 (行 322-322)

**Git历史版本内容:**
```
+ a b c d
```

### 11. 内容被修改 (文件1行 644-644, 文件2行 653-654)

**当前版本内容:**
```
- load data local infile '$resources/load_data/test_enclosed_by01.csv' into table test06 fields terminated by ',' enclosed by ''' escaped by '\\' lines terminated by '
```

**Git历史版本内容:**
```
+ col1 col2
+ load data local infile '$resources_local/load_data/test_enclosed_by01.csv' into table test06 fields terminated by ',' enclosed by ''' escaped by '\\' lines terminated by '
```

### 12. 内容被修改 (文件1行 651-651, 文件2行 661-661)

**当前版本内容:**
```
- load data local infile {'filepath'='$resources/load_data/test_enclosed_by01.csv','compression'='','format'='csv'} into table test06 fields terminated by ',' enclosed by ''' escaped by '\\' lines terminated by '
```

**Git历史版本内容:**
```
+ load data local infile {'filepath'='$resources_local/load_data/test_enclosed_by01.csv','compression'='','format'='csv'} into table test06 fields terminated by ',' enclosed by ''' escaped by '\\' lines terminated by '
```

### 13. 文件2中独有的内容 (行 767-767)

**Git历史版本内容:**
```
+ id name age
```

### 14. 文件2中独有的内容 (行 776-776)

**Git历史版本内容:**
```
+ id name age
```

### 15. 内容被修改 (文件1行 820-820, 文件2行 832-832)

**当前版本内容:**
```
- load data local infile '$resources/load_data/test_columnlist_04.csv' into table load_data_t5 fields terminated by ',' lines terminated by '
```

**Git历史版本内容:**
```
+ load data local infile '$resources_local/load_data/test_columnlist_04.csv' into table load_data_t5 fields terminated by ',' lines terminated by '
```

### 16. 内容被修改 (文件1行 827-827, 文件2行 839-839)

**当前版本内容:**
```
- load data local infile '$resources/load_data/test_columnlist_04.csv' into table load_data_t5 fields terminated by ',' lines terminated by '
```

**Git历史版本内容:**
```
+ load data local infile '$resources_local/load_data/test_columnlist_04.csv' into table load_data_t5 fields terminated by ',' lines terminated by '
```

### 17. 内容被修改 (文件1行 834-834, 文件2行 846-846)

**当前版本内容:**
```
- load data local infile '$resources/load_data/test_columnlist_04.csv' into table load_data_t5 fields terminated by ',' lines terminated by '
```

**Git历史版本内容:**
```
+ load data local infile '$resources_local/load_data/test_columnlist_04.csv' into table load_data_t5 fields terminated by ',' lines terminated by '
```

### 18. 内容被修改 (文件1行 841-841, 文件2行 853-853)

**当前版本内容:**
```
- load data local infile '$resources/load_data/test_columnlist_04.csv' into table load_data_t5 fields terminated by ',' lines terminated by '
```

**Git历史版本内容:**
```
+ load data local infile '$resources_local/load_data/test_columnlist_04.csv' into table load_data_t5 fields terminated by ',' lines terminated by '
```

### 19. 内容被修改 (文件1行 848-848, 文件2行 860-860)

**当前版本内容:**
```
- load data local infile '$resources/load_data/test_columnlist_03.csv' into table load_data_t5 fields terminated by ',' lines terminated by '
```

**Git历史版本内容:**
```
+ load data local infile '$resources_local/load_data/test_columnlist_03.csv' into table load_data_t5 fields terminated by ',' lines terminated by '
```

### 20. 内容被修改 (文件1行 855-855, 文件2行 867-867)

**当前版本内容:**
```
- load data local infile '$resources/load_data/test_columnlist_03.csv' into table load_data_t5 fields terminated by ',' lines terminated by '
```

**Git历史版本内容:**
```
+ load data local infile '$resources_local/load_data/test_columnlist_03.csv' into table load_data_t5 fields terminated by ',' lines terminated by '
```

### 21. 内容被修改 (文件1行 864-864, 文件2行 876-876)

**当前版本内容:**
```
- load data local infile '$resources/load_data/test_columnlist_03.csv' into table load_data_t6 fields terminated by ',' lines terminated by '
```

**Git历史版本内容:**
```
+ load data local infile '$resources_local/load_data/test_columnlist_03.csv' into table load_data_t6 fields terminated by ',' lines terminated by '
```

### 22. 内容被修改 (文件1行 873-873, 文件2行 885-885)

**当前版本内容:**
```
- load data local infile '$resources/load_data/test_escaped_by03.csv' into table load_data_t7 fields terminated by ',' enclosed by ''' escaped by '\\' lines terminated by '
```

**Git历史版本内容:**
```
+ load data local infile '$resources_local/load_data/test_escaped_by03.csv' into table load_data_t7 fields terminated by ',' enclosed by ''' escaped by '\\' lines terminated by '
```

### 23. 文件2中独有的内容 (行 895-895)

**Git历史版本内容:**
```
+ col1 col2 col3
```

### 24. 内容被修改 (文件1行 888-888, 文件2行 901-901)

**当前版本内容:**
```
- sql_mode error_for_division_by_zero,no_engine_substitution,no_zero_date,no_zero_in_date,only_full_group_by,pipes_as_concat,strict_trans_tables
```

**Git历史版本内容:**
```
+ sql_mode error_for_division_by_zero,no_engine_substitution,no_zero_date,no_zero_in_date,only_full_group_by,strict_trans_tables
```

---

## 2. load_data_csv_values.result

**文件路径:** load_data_csv_values.result

**统计信息:**
- 当前版本行数: 112
- Git版本行数: 112

**内容差异 (4 个):**

### 1. 内容被修改 (文件1行 44-45, 文件2行 44-45)

**当前版本内容:**
```
- 1000-01-01 0001-01-01 00:00:00 1969-12-31 16:00:01 false
- 9999-12-31 9999-12-31 00:00:00 2038-01-18 16:00:00 true
```

**Git历史版本内容:**
```
+ 1000-01-01 0001-01-01 00:00:00 1970-01-01 00:00:01 false
+ 9999-12-31 9999-12-31 00:00:00 2038-01-19 00:00:00 true
```

### 2. 内容被修改 (文件1行 70-72, 文件2行 70-72)

**当前版本内容:**
```
- row5 1 1.1 1 2023-05-16 16:06:18.070277 {'key1':'insert charactor \''}
- row5 1 1.1 1 2023-05-16 16:06:18.070277 {'key1':'insert charactor \''}
- row7 1 1.1 1 2023-05-16 16:06:18.070277 {'key1':'newline:
```

**Git历史版本内容:**
```
+ row5 1 1.1 1 2023-05-16 16:06:18.070277000 {'key1':'insert charactor \''}
+ row5 1 1.1 1 2023-05-16 16:06:18.070277000 {'key1':'insert charactor \''}
+ row7 1 1.1 1 2023-05-16 16:06:18.070277000 {'key1':'newline:
```

### 3. 内容被修改 (文件1行 74-74, 文件2行 74-74)

**当前版本内容:**
```
- row7 1 1.1 1 2023-05-16 16:06:18.070277 {'key1':'newline:
```

**Git历史版本内容:**
```
+ row7 1 1.1 1 2023-05-16 16:06:18.070277000 {'key1':'newline:
```

### 4. 内容被修改 (文件1行 76-77, 文件2行 76-77)

**当前版本内容:**
```
- row\8 1 1.1 1 2023-05-16 16:06:18.070277 {'key1':'newline'}
- row\8 1 1.1 1 2023-05-16 16:06:18.070277 {'key1':'newline'}
```

**Git历史版本内容:**
```
+ row\8 1 1.1 1 2023-05-16 16:06:18.070277000 {'key1':'newline'}
+ row\8 1 1.1 1 2023-05-16 16:06:18.070277000 {'key1':'newline'}
```

---

## 3. load_data_jsonline.result

**文件路径:** load_data_jsonline.result

**统计信息:**
- 当前版本行数: 296
- Git版本行数: 297

**内容差异 (10 个):**

### 1. 内容被修改 (文件1行 132-133, 文件2行 132-133)

**当前版本内容:**
```
- 1.3 5 2.0000000000000000 0.4000000000000000
- 1 null 635437923742.3333333330000000 1.0000000000000000
```

**Git历史版本内容:**
```
+ 1.3 5.0 2.0000000000000000 0.4000000000000000
+ 1.0 null 635437923742.3333333330000000 1.0000000000000000
```

### 2. 内容被修改 (文件1行 135-136, 文件2行 135-136)

**当前版本内容:**
```
- -1763835000000000000000 -0.00000000000001 1.2345600000000000 3.9800000000000000
- -345.34244 -11118772349834299000000000000000000000000000000000000000000000000000000000 8349538974359357.0000000000000000 3.9484359854839584
```

**Git历史版本内容:**
```
+ -1.763835e21 -1.0e-14 1.2345600000000000 3.9800000000000000
+ -345.34244 -1.1118772349834299e73 8349538974359357.0000000000000000 3.9484359854839584
```

### 3. 内容被修改 (文件1行 141-142, 文件2行 141-142)

**当前版本内容:**
```
- 1.3 5 2.0000000000000000 0.4000000000000000
- 1 null 635437923742.3333333330000000 1.0000000000000000
```

**Git历史版本内容:**
```
+ 1.3 5.0 2.0000000000000000 0.4000000000000000
+ 1.0 null 635437923742.3333333330000000 1.0000000000000000
```

### 4. 内容被修改 (文件1行 144-144, 文件2行 144-144)

**当前版本内容:**
```
- -17638359000000000000000000 -1.9348593579835793 1.2345600000000000 3.9800000000000000
```

**Git历史版本内容:**
```
+ -1.7638359e25 -1.9348593579835793 1.2345600000000000 3.9800000000000000
```

### 5. 内容被修改 (文件1行 157-164, 文件2行 157-164)

**当前版本内容:**
```
- 1000-01-01 0001-01-01 00:00:00.000000 1970-01-01 00:00:01.000 false
- 9999-12-31 9999-12-31 00:00:00.000000 2038-01-19 00:00:00.000 true
- 1000-01-01 0001-01-01 00:00:00.000000 null false
- 1000-01-01 0001-01-01 00:00:00.000000 null true
- 1000-01-01 0001-01-01 00:00:00.000001 null false
- null null null null
- null null null null
- 9999-12-31 9999-12-30 23:59:59.999999 null false
```

**Git历史版本内容:**
```
+ 1000-01-01 0001-01-01 00:00:00 1970-01-01 00:00:01 false
+ 9999-12-31 9999-12-31 00:00:00 2038-01-19 00:00:00 true
+ 1000-01-01 0001-01-01 00:00:00 null false
+ 1000-01-01 0001-01-01 00:00:00 null true
+ 1000-01-01 0001-01-01 00:00:00.000001000 null false
+ null null null null
+ null null null null
+ 9999-12-31 9999-12-30 23:59:59.999999000 null false
```

### 6. 内容被修改 (文件1行 169-184, 文件2行 169-184)

**当前版本内容:**
```
- 1000-01-01 0001-01-01 00:00:00.000000 1970-01-01 00:00:01.000 false
- 9999-12-31 9999-12-31 00:00:00.000000 2038-01-19 00:00:00.000 true
- 1000-01-01 0001-01-01 00:00:00.000000 null false
- 1000-01-01 0001-01-01 00:00:00.000000 null true
- 1000-01-01 0001-01-01 00:00:00.000001 null false
- null null null null
- null null null null
- 9999-12-31 9999-12-30 23:59:59.999999 null false
- 1000-01-01 0001-01-01 00:00:00.000000 1970-01-01 00:00:01.000 false
- 9999-12-31 9999-12-31 00:00:00.000000 2038-01-19 00:00:00.000 true
- 1000-01-01 0001-01-01 00:00:00.000000 null false
- 1000-01-01 0001-01-01 00:00:00.000000 null true
- 1000-01-01 0001-01-01 00:00:00.000001 null false
- null null null null
- null null null null
- 9999-12-31 9999-12-30 23:59:59.999999 null false
```

**Git历史版本内容:**
```
+ 1000-01-01 0001-01-01 00:00:00 1970-01-01 00:00:01 false
+ 9999-12-31 9999-12-31 00:00:00 2038-01-19 00:00:00 true
+ 1000-01-01 0001-01-01 00:00:00 null false
+ 1000-01-01 0001-01-01 00:00:00 null true
+ 1000-01-01 0001-01-01 00:00:00.000001000 null false
+ null null null null
+ null null null null
+ 9999-12-31 9999-12-30 23:59:59.999999000 null false
+ 1000-01-01 0001-01-01 00:00:00 1970-01-01 00:00:01 false
+ 9999-12-31 9999-12-31 00:00:00 2038-01-19 00:00:00 true
+ 1000-01-01 0001-01-01 00:00:00 null false
+ 1000-01-01 0001-01-01 00:00:00 null true
+ 1000-01-01 0001-01-01 00:00:00.000001000 null false
+ null null null null
+ null null null null
+ 9999-12-31 9999-12-30 23:59:59.999999000 null false
```

### 7. 内容被修改 (文件1行 189-192, 文件2行 189-192)

**当前版本内容:**
```
- 1000-01-01 0001-01-01 00:00:00.000000 null false
- 1000-01-01 0001-01-01 00:00:00.000000 2023-01-12 10:02:34.000 true
- 1000-01-01 0001-01-01 00:00:00.000000 2022-09-10 00:00:00.000 true
- 9999-12-31 9999-12-30 23:59:59.999999 2023-01-12 10:02:34.093 false
```

**Git历史版本内容:**
```
+ 1000-01-01 0001-01-01 00:00:00 null false
+ 1000-01-01 0001-01-01 00:00:00 2023-01-12 10:02:34 true
+ 1000-01-01 0001-01-01 00:00:00 2022-09-10 00:00:00 true
+ 9999-12-31 9999-12-30 23:59:59.999999000 2023-01-12 10:02:34.093000000 false
```

### 8. 文件2中独有的内容 (行 198-198)

**Git历史版本内容:**
```
+ a b c d
```

### 9. 内容被修改 (文件1行 287-288, 文件2行 288-289)

**当前版本内容:**
```
- 1.3 5 2.0000000000000000 0.4000000000000000
- 1 null 635437923742.3333333330000000 1.0000000000000000
```

**Git历史版本内容:**
```
+ 1.3 5.0 2.0000000000000000 0.4000000000000000
+ 1.0 null 635437923742.3333333330000000 1.0000000000000000
```

### 10. 内容被修改 (文件1行 290-291, 文件2行 291-292)

**当前版本内容:**
```
- -1763835000000000000000 -0.00000000000001 1.2345600000000000 3.9800000000000000
- -345.34244 -11118772349834299000000000000000000000000000000000000000000000000000000000 8349538974359357.0000000000000000 3.9484359854839584
```

**Git历史版本内容:**
```
+ -1.763835e21 -1.0e-14 1.2345600000000000 3.9800000000000000
+ -345.34244 -1.1118772349834299e73 8349538974359357.0000000000000000 3.9484359854839584
```

---

## 4. load_data_parquet.result

**文件路径:** load_data_parquet.result

**统计信息:**
- 当前版本行数: 59
- Git版本行数: 59

**内容差异 (7 个):**

### 1. 文件2中独有的内容 (行 8-13)

**Git历史版本内容:**
```
+ 1 user1
+ 2 user2
+ 7 user7
+ 8 user8
+ 10 user10
+ 12 user12
```

### 2. 文件1中独有的内容 (行 10-15)

**当前版本内容:**
```
- 1 user1
- 2 user2
- 8 user8
- 10 user10
- 7 user7
- 12 user12
```

### 3. 内容被修改 (文件1行 20-20, 文件2行 20-22)

**当前版本内容:**
```
- 7 user7 false 7
```

**Git历史版本内容:**
```
+ 1 user1 false 1.0
+ 2 user2 false null
+ 7 user7 false 7.0
```

### 4. 内容被修改 (文件1行 22-22, 文件2行 24-24)

**当前版本内容:**
```
- 10 user10 false 10
```

**Git历史版本内容:**
```
+ 10 user10 false 10.0
```

### 5. 文件1中独有的内容 (行 26-27)

**当前版本内容:**
```
- 1 user1 false 1
- 2 user2 false null
```

### 6. 内容被修改 (文件1行 50-50, 文件2行 50-55)

**当前版本内容:**
```
- 7 user7 false 7
```

**Git历史版本内容:**
```
+ 1 user1 false 1.0
+ 2 user2 false null
+ 7 user7 false 7.0
+ 8 user8 true null
+ 10 user10 false 10.0
+ 12 user12 false null
```

### 7. 文件1中独有的内容 (行 53-57)

**当前版本内容:**
```
- 12 user12 false null
- 8 user8 true null
- 10 user10 false 10
- 1 user1 false 1
- 2 user2 false null
```

---

## 5. load_data_set_null.result

**文件路径:** load_data_set_null.result

**统计信息:**
- 当前版本行数: 106
- Git版本行数: 106

**内容差异 (8 个):**

### 1. 内容被修改 (文件1行 18-18, 文件2行 18-18)

**当前版本内容:**
```
- null 1 1 1111-11-11 1
```

**Git历史版本内容:**
```
+ null 1 1.0 1111-11-11 1
```

### 2. 内容被修改 (文件1行 22-23, 文件2行 22-23)

**当前版本内容:**
```
- null 1 1 1111-11-11 1
- 1 null 1 1111-11-11 1
```

**Git历史版本内容:**
```
+ null 1 1.0 1111-11-11 1
+ 1 null 1.0 1111-11-11 1
```

### 3. 内容被修改 (文件1行 28-29, 文件2行 28-29)

**当前版本内容:**
```
- null 1 1 1111-11-11 1
- 1 null 1 1111-11-11 1
```

**Git历史版本内容:**
```
+ null 1 1.0 1111-11-11 1
+ 1 null 1.0 1111-11-11 1
```

### 4. 内容被修改 (文件1行 33-35, 文件2行 33-35)

**当前版本内容:**
```
- null 1 1 1111-11-11 1
- 1 null 1 1111-11-11 1
- 1 1 1 1111-11-11 1
```

**Git历史版本内容:**
```
+ null 1 1.0 1111-11-11 1
+ 1 null 1.0 1111-11-11 1
+ 1 1 1.0 1111-11-11 1
```

### 5. 内容被修改 (文件1行 39-41, 文件2行 39-41)

**当前版本内容:**
```
- null 1 1 1111-11-11 1
- 1 null 1 1111-11-11 1
- 1 1 1 1111-11-11 1
```

**Git历史版本内容:**
```
+ null 1 1.0 1111-11-11 1
+ 1 null 1.0 1111-11-11 1
+ 1 1 1.0 1111-11-11 1
```

### 6. 内容被修改 (文件1行 46-48, 文件2行 46-48)

**当前版本内容:**
```
- null 1 1 1111-11-11 1
- 1 null 1 1111-11-11 1
- 1 1 1 1111-11-11 1
```

**Git历史版本内容:**
```
+ null 1 1.0 1111-11-11 1
+ 1 null 1.0 1111-11-11 1
+ 1 1 1.0 1111-11-11 1
```

### 7. 内容被修改 (文件1行 50-50, 文件2行 50-50)

**当前版本内容:**
```
- null 1 1 1111-11-11 1
```

**Git历史版本内容:**
```
+ null 1 1.0 1111-11-11 1
```

### 8. 内容被修改 (文件1行 52-59, 文件2行 52-59)

**当前版本内容:**
```
- 3 3 3 1111-03-11 3
- 4 4 4 null 4
- 5 5 5 1111-05-11 null
- 6 6 6 1111-06-11 6
- 7 7 7 1111-07-11 7
- 8 8 8 1111-08-11 8
- 9 9 9 1111-09-11 9
- 10 10 10 1111-10-11 10
```

**Git历史版本内容:**
```
+ 3 3 3.0 1111-03-11 3
+ 4 4 4.0 null 4
+ 5 5 5.0 1111-05-11 null
+ 6 6 6.0 1111-06-11 6
+ 7 7 7.0 1111-07-11 7
+ 8 8 8.0 1111-08-11 8
+ 9 9 9.0 1111-09-11 9
+ 10 10 10.0 1111-10-11 10
```

---

