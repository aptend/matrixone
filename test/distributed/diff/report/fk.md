# 批量Git Diff报告

生成时间: 2025-10-10 14:22:25.965181
比较commit: HEAD
总文件数: 4

## 1. fk_base.result

**文件路径:** fk_base.result

**统计信息:**
- 当前版本行数: 378
- Git版本行数: 382

**内容差异 (10 个):**

### 1. 文件2中独有的内容 (行 11-11)

**Git历史版本内容:**
```
+ a b
```

### 2. 文件2中独有的内容 (行 13-13)

**Git历史版本内容:**
```
+ a b
```

### 3. 内容被修改 (文件1行 129-136, 文件2行 131-138)

**当前版本内容:**
```
- 7521 ward null 1250
- 1234 meixi null 1250
- 7369 smith null 1300
- 7499 allen 30 1600
- 7698 blake 30 2850
- 7782 clark 10 2950
- 7566 jones 20 3475
- 7788 scott 20 3500
```

**Git历史版本内容:**
```
+ 7521 ward null 1250.0
+ 1234 meixi null 1250.0
+ 7369 smith null 1300.0
+ 7499 allen 30 1600.0
+ 7698 blake 30 2850.0
+ 7782 clark 10 2950.0
+ 7566 jones 20 3475.0
+ 7788 scott 20 3500.0
```

### 4. 内容被修改 (文件1行 182-189, 文件2行 184-191)

**当前版本内容:**
```
- 7369 smith 20 1300
- 7499 allen 30 1600
- 7521 ward 30 1250
- 7566 jones 20 3475
- 1234 meixi 30 1250
- 7698 blake 30 2850
- 7788 scott 20 3500
- 7782 clark 50 2950
```

**Git历史版本内容:**
```
+ 7369 smith 20 1300.0
+ 7499 allen 30 1600.0
+ 7521 ward 30 1250.0
+ 7566 jones 20 3475.0
+ 1234 meixi 30 1250.0
+ 7698 blake 30 2850.0
+ 7788 scott 20 3500.0
+ 7782 clark 50 2950.0
```

### 5. 内容被修改 (文件1行 198-204, 文件2行 200-206)

**当前版本内容:**
```
- 7369 smith 20 1300
- 7499 allen 30 1600
- 7521 ward 30 1250
- 7566 jones 20 3475
- 1234 meixi 30 1250
- 7698 blake 30 2850
- 7788 scott 20 3500
```

**Git历史版本内容:**
```
+ 7369 smith 20 1300.0
+ 7499 allen 30 1600.0
+ 7521 ward 30 1250.0
+ 7566 jones 20 3475.0
+ 1234 meixi 30 1250.0
+ 7698 blake 30 2850.0
+ 7788 scott 20 3500.0
```

### 6. 内容被修改 (文件1行 247-254, 文件2行 249-256)

**当前版本内容:**
```
- 7369 smith 20 1300
- 7499 allen 30 1600
- 7521 ward 30 1250
- 7566 jones 20 3475
- 1234 meixi 30 1250
- 7698 blake 30 2850
- 7788 scott 20 3500
- 7782 clark null 2950
```

**Git历史版本内容:**
```
+ 7369 smith 20 1300.0
+ 7499 allen 30 1600.0
+ 7521 ward 30 1250.0
+ 7566 jones 20 3475.0
+ 1234 meixi 30 1250.0
+ 7698 blake 30 2850.0
+ 7788 scott 20 3500.0
+ 7782 clark null 2950.0
```

### 7. 内容被修改 (文件1行 263-270, 文件2行 265-272)

**当前版本内容:**
```
- 7369 smith 20 1300
- 7499 allen 30 1600
- 7521 ward 30 1250
- 7566 jones 20 3475
- 1234 meixi 30 1250
- 7698 blake 30 2850
- 7788 scott 20 3500
- 7782 clark null 2950
```

**Git历史版本内容:**
```
+ 7369 smith 20 1300.0
+ 7499 allen 30 1600.0
+ 7521 ward 30 1250.0
+ 7566 jones 20 3475.0
+ 1234 meixi 30 1250.0
+ 7698 blake 30 2850.0
+ 7788 scott 20 3500.0
+ 7782 clark null 2950.0
```

### 8. 内容被修改 (文件1行 288-288, 文件2行 290-290)

**当前版本内容:**
```
- aaa bbb
```

**Git历史版本内容:**
```
+ aaa bbbb
```

### 9. 文件2中独有的内容 (行 356-356)

**Git历史版本内容:**
```
+ tables_in_db1
```

### 10. 文件2中独有的内容 (行 369-369)

**Git历史版本内容:**
```
+ tables_in_db2
```

---

## 2. fk_show_columns.result

**文件路径:** fk_show_columns.result

**统计信息:**
- 当前版本行数: 1211
- Git版本行数: 1259

**内容差异 (4 个):**

### 1. 文件2中独有的内容 (行 192-203)

**Git历史版本内容:**
```
+ desc mo_catalog.mo_table_partitions;
+ field type null key default extra comment
+ table_id bigint unsigned(64)no pri null
+ database_id bigint unsigned(64)no null
+ number smallint unsigned(16)no null
+ name varchar(64)no pri null
+ partition_type varchar(50)no null
+ partition_expression varchar(2048)yes null
+ description_utf8 text(0)yes null
+ comment varchar(2048)no null
+ options text(0)yes null
+ partition_table_name varchar(1024)no null
```

### 2. 文件2中独有的内容 (行 517-528)

**Git历史版本内容:**
```
+ show columns from mo_catalog.mo_table_partitions;
+ field type null key default extra comment
+ table_id bigint unsigned(64)no pri null
+ database_id bigint unsigned(64)no null
+ number smallint unsigned(16)no null
+ name varchar(64)no pri null
+ partition_type varchar(50)no null
+ partition_expression varchar(2048)yes null
+ description_utf8 text(0)yes null
+ comment varchar(2048)no null
+ options text(0)yes null
+ partition_table_name varchar(1024)no null
```

### 3. 文件2中独有的内容 (行 833-844)

**Git历史版本内容:**
```
+ desc mo_catalog.mo_table_partitions;
+ field type null key default extra comment
+ table_id bigint unsigned(64)no pri null
+ database_id bigint unsigned(64)no null
+ number smallint unsigned(16)no null
+ name varchar(64)no pri null
+ partition_type varchar(50)no null
+ partition_expression varchar(2048)yes null
+ description_utf8 text(0)yes null
+ comment varchar(2048)no null
+ options text(0)yes null
+ partition_table_name varchar(1024)no null
```

### 4. 文件2中独有的内容 (行 1136-1147)

**Git历史版本内容:**
```
+ show columns from mo_catalog.mo_table_partitions;
+ field type null key default extra comment
+ table_id bigint unsigned(64)no pri null
+ database_id bigint unsigned(64)no null
+ number smallint unsigned(16)no null
+ name varchar(64)no pri null
+ partition_type varchar(50)no null
+ partition_expression varchar(2048)yes null
+ description_utf8 text(0)yes null
+ comment varchar(2048)no null
+ options text(0)yes null
+ partition_table_name varchar(1024)no null
```

---

## 3. foreign_key.result

**文件路径:** foreign_key.result

**统计信息:**
- 当前版本行数: 665
- Git版本行数: 672

**内容差异 (53 个):**

### 1. 文件1中独有的内容 (行 15-15)

**当前版本内容:**
```
- 90 5983
```

### 2. 文件2中独有的内容 (行 17-17)

**Git历史版本内容:**
```
+ 90 5983
```

### 3. 文件2中独有的内容 (行 35-35)

**Git历史版本内容:**
```
+ 100 734
```

### 4. 文件1中独有的内容 (行 36-36)

**当前版本内容:**
```
- 100 734
```

### 5. 文件2中独有的内容 (行 47-47)

**Git历史版本内容:**
```
+ 100 500
```

### 6. 文件1中独有的内容 (行 48-48)

**当前版本内容:**
```
- 100 500
```

### 7. 文件1中独有的内容 (行 66-66)

**当前版本内容:**
```
- 90 5983
```

### 8. 文件2中独有的内容 (行 68-68)

**Git历史版本内容:**
```
+ 90 5983
```

### 9. 文件2中独有的内容 (行 100-100)

**Git历史版本内容:**
```
+ 10 student 4
```

### 10. 文件1中独有的内容 (行 102-102)

**当前版本内容:**
```
- 10 student 4
```

### 11. 文件2中独有的内容 (行 115-115)

**Git历史版本内容:**
```
+ 10 student 4
```

### 12. 文件1中独有的内容 (行 117-117)

**当前版本内容:**
```
- 10 student 4
```

### 13. 文件2中独有的内容 (行 125-125)

**Git历史版本内容:**
```
+ col1 col2 col3
```

### 14. 文件2中独有的内容 (行 138-139)

**Git历史版本内容:**
```
+ 1 opppo 51
+ 2 apple 50
```

### 15. 文件1中独有的内容 (行 138-139)

**当前版本内容:**
```
- 2 apple 50
- 1 opppo 51
```

### 16. 文件2中独有的内容 (行 156-157)

**Git历史版本内容:**
```
+ 1 opppo 51
+ 2 apple 50
```

### 17. 文件1中独有的内容 (行 156-157)

**当前版本内容:**
```
- 2 apple 50
- 1 opppo 51
```

### 18. 文件2中独有的内容 (行 169-170)

**Git历史版本内容:**
```
+ 1 opppo 51
+ 2 apple 50
```

### 19. 内容被修改 (文件1行 169-171, 文件2行 172-173)

**当前版本内容:**
```
- 2 apple 50
- 1 opppo 51
- select * from fk_02;
```

**Git历史版本内容:**
```
+ select * from fk_02;
+ col1 col2 col3
```

### 20. 文件2中独有的内容 (行 210-210)

**Git历史版本内容:**
```
+ col1 col2 col3
```

### 21. 文件2中独有的内容 (行 227-227)

**Git历史版本内容:**
```
+ 23.100000000000000000 a 20 2022-10-02
```

### 22. 文件1中独有的内容 (行 225-225)

**当前版本内容:**
```
- 23.100000000000000000 a 20 2022-10-02
```

### 23. 文件2中独有的内容 (行 242-242)

**Git历史版本内容:**
```
+ null a null null
```

### 24. 文件1中独有的内容 (行 240-240)

**当前版本内容:**
```
- null a null null
```

### 25. 文件2中独有的内容 (行 352-352)

**Git历史版本内容:**
```
+ 1 window deli
```

### 26. 文件1中独有的内容 (行 352-352)

**当前版本内容:**
```
- 1 window deli
```

### 27. 文件2中独有的内容 (行 413-413)

**Git历史版本内容:**
```
+ col1 col2 col3
```

### 28. 文件2中独有的内容 (行 433-433)

**Git历史版本内容:**
```
+ 2 3 2
```

### 29. 文件1中独有的内容 (行 430-430)

**当前版本内容:**
```
- 2 3 2
```

### 30. 文件2中独有的内容 (行 438-439)

**Git历史版本内容:**
```
+ 1 5 2
+ 2 3 2
```

### 31. 文件1中独有的内容 (行 435-436)

**当前版本内容:**
```
- 2 3 2
- 1 5 2
```

### 32. 文件2中独有的内容 (行 452-452)

**Git历史版本内容:**
```
+ 1 5 2
```

### 33. 文件1中独有的内容 (行 449-449)

**当前版本内容:**
```
- 1 5 2
```

### 34. 文件2中独有的内容 (行 458-458)

**Git历史版本内容:**
```
+ id book_id author_id
```

### 35. 文件2中独有的内容 (行 511-511)

**Git历史版本内容:**
```
+ a f_a f_b f_c f_d
```

### 36. 文件1中独有的内容 (行 523-523)

**当前版本内容:**
```
- 23.100 a 21 2022-10-01
```

### 37. 文件2中独有的内容 (行 530-530)

**Git历史版本内容:**
```
+ 23.100 a 21 2022-10-01
```

### 38. 文件1中独有的内容 (行 533-533)

**当前版本内容:**
```
- 23.100 a 21 2022-10-01
```

### 39. 文件2中独有的内容 (行 540-540)

**Git历史版本内容:**
```
+ 23.100 a 21 2022-10-01
```

### 40. 文件2中独有的内容 (行 548-548)

**Git历史版本内容:**
```
+ null a 19 2022-10-01
```

### 41. 文件1中独有的内容 (行 543-543)

**当前版本内容:**
```
- null a 19 2022-10-01
```

### 42. 文件2中独有的内容 (行 576-576)

**Git历史版本内容:**
```
+ a c_a c_b
```

### 43. 文件1中独有的内容 (行 583-583)

**当前版本内容:**
```
- 23.100 a 21 2022-10-01
```

### 44. 文件2中独有的内容 (行 591-591)

**Git历史版本内容:**
```
+ 23.100 a 21 2022-10-01
```

### 45. 文件1中独有的内容 (行 593-593)

**当前版本内容:**
```
- 23.100 a 21 2022-10-01
```

### 46. 文件2中独有的内容 (行 601-601)

**Git历史版本内容:**
```
+ 23.100 a 21 2022-10-01
```

### 47. 文件1中独有的内容 (行 600-600)

**当前版本内容:**
```
- 23.100 a 21 2022-10-01
```

### 48. 内容被修改 (文件1行 602-604, 文件2行 608-610)

**当前版本内容:**
```
- select * from fk_02;
- col1 col2 col3 col4
- 23.100 a 21 2022-10-01
```

**Git历史版本内容:**
```
+ 23.100 a 21 2022-10-01
+ select * from fk_02;
+ col1 col2 col3 col4
```

### 49. 文件2中独有的内容 (行 612-612)

**Git历史版本内容:**
```
+ 23.100 a 21 2022-10-01
```

### 50. 文件2中独有的内容 (行 643-643)

**Git历史版本内容:**
```
+ 6.000 a 21 2022-10-01
```

### 51. 文件1中独有的内容 (行 639-639)

**当前版本内容:**
```
- 6.000 a 21 2022-10-01
```

### 52. 文件2中独有的内容 (行 651-651)

**Git历史版本内容:**
```
+ 6.000 d 21 2022-10-01
```

### 53. 文件1中独有的内容 (行 645-645)

**当前版本内容:**
```
- 6.000 d 21 2022-10-01
```

---

## 4. foreign_key_multilayer.result

**文件路径:** foreign_key_multilayer.result

**统计信息:**
- 当前版本行数: 412
- Git版本行数: 428

**内容差异 (10 个):**

### 1. 文件2中独有的内容 (行 351-351)

**Git历史版本内容:**
```
+ s_suppkey s_name s_address s_nationkey s_phone s_acctbal s_comment
```

### 2. 内容被修改 (文件1行 352-352, 文件2行 353-355)

**当前版本内容:**
```
- select * from lineitem_fk;
```

**Git历史版本内容:**
```
+ c_custkey c_name c_address c_nationkey c_phone c_acctbal c_mktsegment c_comment
+ select * from lineitem_fk;
+ l_orderkey l_partkey l_suppkey l_linenumber l_quantity l_extendedprice l_discount l_tax l_returnflag l_linestatus l_shipdate l_commitdate l_receiptdate l_shipinstruct l_shipmode l_comment
```

### 3. 内容被修改 (文件1行 355-355, 文件2行 358-360)

**当前版本内容:**
```
- select * from orders_fk;
```

**Git历史版本内容:**
```
+ ps_partkey ps_suppkey ps_availqty ps_supplycost ps_comment
+ select * from orders_fk;
+ o_orderkey o_custkey o_orderstatus o_totalprice o_orderdate o_orderpriority o_clerk o_shippriority o_comment
```

### 4. 文件2中独有的内容 (行 368-368)

**Git历史版本内容:**
```
+ l_orderkey l_partkey l_suppkey l_linenumber l_quantity l_extendedprice l_discount l_tax l_returnflag l_linestatus l_shipdate l_commitdate l_receiptdate l_shipinstruct l_shipmode l_comment
```

### 5. 内容被修改 (文件1行 367-367, 文件2行 373-375)

**当前版本内容:**
```
- select * from lineitem_fk;
```

**Git历史版本内容:**
```
+ c_custkey c_name c_address c_nationkey c_phone c_acctbal c_mktsegment c_comment
+ select * from lineitem_fk;
+ l_orderkey l_partkey l_suppkey l_linenumber l_quantity l_extendedprice l_discount l_tax l_returnflag l_linestatus l_shipdate l_commitdate l_receiptdate l_shipinstruct l_shipmode l_comment
```

### 6. 文件2中独有的内容 (行 379-379)

**Git历史版本内容:**
```
+ o_orderkey o_custkey o_orderstatus o_totalprice o_orderdate o_orderpriority o_clerk o_shippriority o_comment
```

### 7. 内容被修改 (文件1行 373-374, 文件2行 382-386)

**当前版本内容:**
```
- select * from nation_fk;
- select * from supplier_fk;
```

**Git历史版本内容:**
```
+ ps_partkey ps_suppkey ps_availqty ps_supplycost ps_comment
+ select * from nation_fk;
+ n_nationkey n_name n_regionkey n_comment
+ select * from supplier_fk;
+ s_suppkey s_name s_address s_nationkey s_phone s_acctbal s_comment
```

### 8. 内容被修改 (文件1行 379-379, 文件2行 391-393)

**当前版本内容:**
```
- select * from orders_fk;
```

**Git历史版本内容:**
```
+ n_nationkey n_name n_regionkey n_comment
+ select * from orders_fk;
+ o_orderkey o_custkey o_orderstatus o_totalprice o_orderdate o_orderpriority o_clerk o_shippriority o_comment
```

### 9. 文件2中独有的内容 (行 398-398)

**Git历史版本内容:**
```
+ ps_partkey ps_suppkey ps_availqty ps_supplycost ps_comment
```

### 10. 文件2中独有的内容 (行 402-402)

**Git历史版本内容:**
```
+ n_nationkey n_name n_regionkey n_comment
```

---

