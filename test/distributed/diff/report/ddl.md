# 批量Git Diff报告

生成时间: 2025-09-26 18:54:31.783380
比较commit: HEAD
总文件数: 19

## 1. json_type.result

**文件路径:** json_type.result

**统计信息:**
- 当前版本行数: 389
- Git版本行数: 399

**内容差异 (34 个):**

### 1. 文件2中独有的内容 (行 107-110)

**Git历史版本内容:**
```
+ create table json_table_5(j1 json primary key,j2 json default '{"x": 17,"x": "red"}',j3 json not null);
+ not supported: JSON column 'j1' cannot be in primary key
+ create table json_table_5(j1 json)partition by hash(j1);
+ Field 'j1' is of a not allowed type for this type of partitioning
```

### 2. 文件2中独有的内容 (行 131-131)

**Git历史版本内容:**
```
+ j1 a b
```

### 3. 文件2中独有的内容 (行 143-145)

**Git历史版本内容:**
```
+ 34 {"key10": "value1","key2": "value2"}
+ 501 {"123456": "中文mo","key1": "@#$_%^&*()!@"}
+ 1111 {"123456": "中文mo","芝士面包": "12abc"}
```

### 4. 文件1中独有的内容 (行 140-141)

**当前版本内容:**
```
- 501 {"123456": "中文mo","key1": "@#$_%^&*()!@"}
- 1111 {"123456": "中文mo","芝士面包": "12abc"}
```

### 5. 文件1中独有的内容 (行 143-143)

**当前版本内容:**
```
- 34 {"key10": "value1","key2": "value2"}
```

### 6. 内容被修改 (文件1行 146-146, 文件2行 151-153)

**当前版本内容:**
```
- {"": "","123456": "中文mo"}
```

**Git历史版本内容:**
```
+ {"key10": "value1","key2": "value2"}
+ {"123456": "中文mo","key1": "@#$_%^&*()!@"}
+ {"123456": "中文mo","芝士面包": "12abc"}
```

### 7. 内容被修改 (文件1行 149-149, 文件2行 156-157)

**当前版本内容:**
```
- {"123456": "中文mo","key1": "@#$_%^&*()!@"}
```

**Git历史版本内容:**
```
+ {"d1": "2020-10-09","d2": "2019-08-20 12:30:00"}
+ {"key10": "value1","key2": "value2"}
```

### 8. 内容被修改 (文件1行 152-155, 文件2行 160-160)

**当前版本内容:**
```
- {"123456": "中文mo","芝士面包": "12abc"}
- {"d1": "2020-10-09","d2": "2019-08-20 12:30:00"}
- {"key10": "value1","key2": "value2"}
- {"key10": "value1","key2": "value2"}
```

**Git历史版本内容:**
```
+ {"": "","123456": "中文mo"}
```

### 9. 文件2中独有的内容 (行 165-165)

**Git历史版本内容:**
```
+ id j1
```

### 10. 文件2中独有的内容 (行 200-208)

**Git历史版本内容:**
```
+ {}
+ {"": "","12_key": "中文mo"}
+ {"d1": [true,false]}
+ {"eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee": "1234567890000000000000000000000000000000000000000000000","uuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuu": ["aaaaaaaaaaaaaaaaaaaaaaa11111111111111111111111111111111111111"]}
+ {"key_56": 78.9,"芝士面包": "12abc"}
+ {"key10": "value1","key2": "value2"}
+ {"key1": "@#$_%^&*()!@","key123456": 223}
+ {"13key4": "中文mo","a 1": "b 1"}
+ {"d1": "2020-10-09","d2": "2019-08-20 12:30:00"}
```

### 11. 文件2中独有的内容 (行 210-218)

**Git历史版本内容:**
```
+ select j1 from json_table_1 intersect select j1 from json_table_3;
+ j1
+ select j1 from json_table_1 minus select j1 from json_table_3;
+ j1
+ {}
+ {"": "","12_key": "中文mo"}
+ {"d1": [true,false]}
+ {"eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee": "1234567890000000000000000000000000000000000000000000000","uuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuu": ["aaaaaaaaaaaaaaaaaaaaaaa11111111111111111111111111111111111111"]}
+ {"key_56": 78.9,"芝士面包": "12abc"}
```

### 12. 文件1中独有的内容 (行 197-198)

**当前版本内容:**
```
- {"key_56": 78.9,"芝士面包": "12abc"}
- {"": "","12_key": "中文mo"}
```

### 13. 文件1中独有的内容 (行 201-215)

**当前版本内容:**
```
- {"d1": [true,false]}
- {}
- {"eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee": "1234567890000000000000000000000000000000000000000000000","uuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuu": ["aaaaaaaaaaaaaaaaaaaaaaa11111111111111111111111111111111111111"]}
- select j1 from json_table_1 intersect select j1 from json_table_3;
- select j1 from json_table_1 minus select j1 from json_table_3;
- j1
- {"key10": "value1","key2": "value2"}
- {"key1": "@#$_%^&*()!@","key123456": 223}
- {"key_56": 78.9,"芝士面包": "12abc"}
- {"": "","12_key": "中文mo"}
- {"13key4": "中文mo","a 1": "b 1"}
- {"d1": "2020-10-09","d2": "2019-08-20 12:30:00"}
- {"d1": [true,false]}
- {}
- {"eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee": "1234567890000000000000000000000000000000000000000000000","uuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuuu": ["aaaaaaaaaaaaaaaaaaaaaaa11111111111111111111111111111111111111"]}
```

### 14. 文件2中独有的内容 (行 288-289)

**Git历史版本内容:**
```
+ {"a": [1,"2",{"aa": "bb"}]} json_table_6.j1 0 key1 $.key1 null "value1" {"key1": "value1","key2": "value2"}
+ {"a": [1,"2",{"aa": "bb"}]} json_table_6.j1 1 key2 $.key2 null "value2" {"key1": "value1","key2": "value2"}
```

### 15. 文件1中独有的内容 (行 282-282)

**当前版本内容:**
```
- {"a": [1,"2",{"aa": "bb"}]} json_table_6.j1 0 key1 $.key1 null "value1" {"key1": "value1","key2": "value2"}
```

### 16. 文件1中独有的内容 (行 284-284)

**当前版本内容:**
```
- {"a": [1,"2",{"aa": "bb"}]} json_table_6.j1 1 key2 $.key2 null "value2" {"key1": "value1","key2": "value2"}
```

### 17. 文件2中独有的内容 (行 296-297)

**Git历史版本内容:**
```
+ {"a": [1,"2",{"aa": "bb"}]} json_table_6.j1 0 key1 $.key1 null "value1" {"key1": "value1","key2": "value2"}
+ {"a": [1,"2",{"aa": "bb"}]} json_table_6.j1 1 key2 $.key2 null "value2" {"key1": "value1","key2": "value2"}
```

### 18. 文件1中独有的内容 (行 290-290)

**当前版本内容:**
```
- {"a": [1,"2",{"aa": "bb"}]} json_table_6.j1 0 key1 $.key1 null "value1" {"key1": "value1","key2": "value2"}
```

### 19. 文件1中独有的内容 (行 292-292)

**当前版本内容:**
```
- {"a": [1,"2",{"aa": "bb"}]} json_table_6.j1 1 key2 $.key2 null "value2" {"key1": "value1","key2": "value2"}
```

### 20. 文件2中独有的内容 (行 304-305)

**Git历史版本内容:**
```
+ {"a": [1,"2",{"aa": "bb"}]} json_table_6.j1 1 null $.a[1] 1 "2" [1,"2",{"aa": "bb"}]
+ {"a": [1,"2",{"aa": "bb"}]} json_table_6.j1 2 null $.a[2] 2 {"aa": "bb"} [1,"2",{"aa": "bb"}]
```

### 21. 文件1中独有的内容 (行 298-298)

**当前版本内容:**
```
- {"a": [1,"2",{"aa": "bb"}]} json_table_6.j1 1 null $.a[1] 1 "2" [1,"2",{"aa": "bb"}]
```

### 22. 文件1中独有的内容 (行 300-300)

**当前版本内容:**
```
- {"a": [1,"2",{"aa": "bb"}]} json_table_6.j1 2 null $.a[2] 2 {"aa": "bb"} [1,"2",{"aa": "bb"}]
```

### 23. 文件2中独有的内容 (行 321-321)

**Git历史版本内容:**
```
+ col seq key path index value this
```

### 24. 文件2中独有的内容 (行 338-338)

**Git历史版本内容:**
```
+ {"a": [1,"2",{"aa": "bb"}]} json_table_6.j1 0 null $.key1 null null "value1"
```

### 25. 文件1中独有的内容 (行 331-331)

**当前版本内容:**
```
- {"a": [1,"2",{"aa": "bb"}]} json_table_6.j1 0 null $.key1 null null "value1"
```

### 26. 文件2中独有的内容 (行 344-346)

**Git历史版本内容:**
```
+ {"a": [1,"2",{"aa": "bb"}]} json_table_6.j1 1 null $.a[1] 1 "2" [1,"2",{"aa": "bb"}]
+ {"a": [1,"2",{"aa": "bb"}]} json_table_6.j1 2 null $.a[2] 2 {"aa": "bb"} [1,"2",{"aa": "bb"}]
+ {"a": [1,"2",{"aa": "bb"}]} json_table_6.j1 0 null $.a null null null
```

### 27. 文件1中独有的内容 (行 337-337)

**当前版本内容:**
```
- {"a": [1,"2",{"aa": "bb"}]} json_table_6.j1 1 null $.a[1] 1 "2" [1,"2",{"aa": "bb"}]
```

### 28. 文件1中独有的内容 (行 339-339)

**当前版本内容:**
```
- {"a": [1,"2",{"aa": "bb"}]} json_table_6.j1 2 null $.a[2] 2 {"aa": "bb"} [1,"2",{"aa": "bb"}]
```

### 29. 文件1中独有的内容 (行 341-341)

**当前版本内容:**
```
- {"a": [1,"2",{"aa": "bb"}]} json_table_6.j1 0 null $.a null null null
```

### 30. 文件2中独有的内容 (行 352-352)

**Git历史版本内容:**
```
+ j1 col seq key path index value this
```

### 31. 文件2中独有的内容 (行 358-358)

**Git历史版本内容:**
```
+ seq value
```

### 32. 文件2中独有的内容 (行 364-365)

**Git历史版本内容:**
```
+ {"a": [1,"2",{"aa": "bb"}]} json_table_6.j1 1 null $.a[1] 1 "2" [1,"2",{"aa": "bb"}]
+ {"a": [1,"2",{"aa": "bb"}]} json_table_6.j1 2 null $.a[2] 2 {"aa": "bb"} [1,"2",{"aa": "bb"}]
```

### 33. 文件1中独有的内容 (行 355-355)

**当前版本内容:**
```
- {"a": [1,"2",{"aa": "bb"}]} json_table_6.j1 1 null $.a[1] 1 "2" [1,"2",{"aa": "bb"}]
```

### 34. 文件1中独有的内容 (行 357-357)

**当前版本内容:**
```
- {"a": [1,"2",{"aa": "bb"}]} json_table_6.j1 2 null $.a[2] 2 {"aa": "bb"} [1,"2",{"aa": "bb"}]
```

---

## 2. lowercase.result

**文件路径:** lowercase.result

**统计信息:**
- 当前版本行数: 707
- Git版本行数: 710

**内容差异 (18 个):**

### 1. 文件2中独有的内容 (行 169-169)

**Git历史版本内容:**
```
+ id word id word
```

### 2. 文件2中独有的内容 (行 202-202)

**Git历史版本内容:**
```
+ user_host user_name status
```

### 3. 文件1中独有的内容 (行 203-203)

**当前版本内容:**
```
- localhost u_name unlock
```

### 4. 内容被修改 (文件1行 206-207, 文件2行 207-207)

**当前版本内容:**
```
- localhost user1 unlock
- localhost user2 unlock
```

**Git历史版本内容:**
```
+ localhost u_name unlock
```

### 5. 文件2中独有的内容 (行 325-325)

**Git历史版本内容:**
```
+ col1 col2 col3
```

### 6. 文件1中独有的内容 (行 352-352)

**当前版本内容:**
```
- rolex
```

### 7. 文件2中独有的内容 (行 355-355)

**Git历史版本内容:**
```
+ rolex
```

### 8. 文件2中独有的内容 (行 357-357)

**Git历史版本内容:**
```
+ role_name COmments
```

### 9. 内容被修改 (文件1行 387-387, 文件2行 389-389)

**当前版本内容:**
```
- vie CREATE VIEW ViE AS SELECT * FROM TAb utf8mb4 utf8mb4_general_ci
```

**Git历史版本内容:**
```
+ vie CREATE VIEW ViE AS SELECT * FROM TAb;utf8mb4 utf8mb4_general_ci
```

### 10. 内容被修改 (文件1行 398-398, 文件2行 400-400)

**当前版本内容:**
```
- view01 CREATE VIEW VIEW01 AS SELECT * FROM t13 WHERE a=1 or a = 2 utf8mb4 utf8mb4_general_ci
```

**Git历史版本内容:**
```
+ view01 CREATE VIEW VIEW01 AS SELECT * FROM t13 WHERE a=1 or a = 2;utf8mb4 utf8mb4_general_ci
```

### 11. 文件2中独有的内容 (行 565-565)

**Git历史版本内容:**
```
+ col1 col2 col3
```

### 12. 文件1中独有的内容 (行 590-590)

**当前版本内容:**
```
- rolex
```

### 13. 文件2中独有的内容 (行 595-595)

**Git历史版本内容:**
```
+ rolex
```

### 14. 文件2中独有的内容 (行 597-597)

**Git历史版本内容:**
```
+ role_name COmments
```

### 15. 文件1中独有的内容 (行 603-603)

**当前版本内容:**
```
- localhost user_name unlock
```

### 16. 内容被修改 (文件1行 606-607, 文件2行 609-609)

**当前版本内容:**
```
- localhost user1 unlock
- localhost user2 unlock
```

**Git历史版本内容:**
```
+ localhost user_name unlock
```

### 17. 文件2中独有的内容 (行 611-611)

**Git历史版本内容:**
```
+ user_host user_name status
```

### 18. 内容被修改 (文件1行 638-638, 文件2行 641-641)

**当前版本内容:**
```
- vie CREATE VIEW `ViE` AS SELECT * FROM `TAb` utf8mb4 utf8mb4_general_ci
```

**Git历史版本内容:**
```
+ vie CREATE VIEW `ViE` AS SELECT * FROM `TAb`;utf8mb4 utf8mb4_general_ci
```

---

## 3. mysql_ddl_1.result

**文件路径:** mysql_ddl_1.result

**统计信息:**
- 当前版本行数: 106
- Git版本行数: 113

**内容差异 (10 个):**

### 1. 文件2中独有的内容 (行 5-5)

**Git历史版本内容:**
```
+ Tables_in_mysql_ddl_test_db_5
```

### 2. 内容被修改 (文件1行 21-21, 文件2行 22-22)

**当前版本内容:**
```
- /*!40101 use mysql_ddl_test_db_3;
```

**Git历史版本内容:**
```
+ /*!40101 use mysql_ddl_test_db_3;*/
```

### 3. 文件1中独有的内容 (行 23-23)

**当前版本内容:**
```
- */
```

### 4. 文件2中独有的内容 (行 32-37)

**Git历史版本内容:**
```
+ CREATE DATABASE /*!32312 IF NOT EXISTS*/ `mysql_ddl_test_db_4` /*!40100 DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci */ /*!80016 DEFAULT ENCRYPTION='N' */;
+ [unknown result because it is related to issue#moc 1231]
+ /*!40101 use mysql_ddl_test_db_4;*/
+ [unknown result because it is related to issue#moc 1231]
+ select database();
+ [unknown result because it is related to issue#moc 1231]
```

### 5. 内容被修改 (文件1行 39-39, 文件2行 45-45)

**当前版本内容:**
```
- @@sql_log_bin
```

**Git历史版本内容:**
```
+ @@SQL_LOG_BIN
```

### 6. 内容被修改 (文件1行 43-43, 文件2行 49-49)

**当前版本内容:**
```
- @@gtid_purged
```

**Git历史版本内容:**
```
+ @@GTID_PURGED
```

### 7. 内容被修改 (文件1行 47-47, 文件2行 53-53)

**当前版本内容:**
```
- @@time_zone
```

**Git历史版本内容:**
```
+ @@TIME_ZONE
```

### 8. 内容被修改 (文件1行 51-51, 文件2行 57-57)

**当前版本内容:**
```
- @@time_zone
```

**Git历史版本内容:**
```
+ @@TIME_ZONE
```

### 9. 内容被修改 (文件1行 55-55, 文件2行 61-61)

**当前版本内容:**
```
- @@time_zone
```

**Git历史版本内容:**
```
+ @@TIME_ZONE
```

### 10. 文件2中独有的内容 (行 112-112)

**Git历史版本内容:**
```
+ config_id config_type config_detail create_time description
```

---

## 4. mysql_ddl_2.result

**文件路径:** mysql_ddl_2.result

**统计信息:**
- 当前版本行数: 380
- Git版本行数: 378

**内容差异 (2 个):**

### 1. 文件1中独有的内容 (行 116-116)

**当前版本内容:**
```
- -- create table
```

### 2. 内容被修改 (文件1行 119-120, 文件2行 118-118)

**当前版本内容:**
```
- -- create table
- DROP TABLE IF EXISTS `projects`";
```

**Git历史版本内容:**
```
+ DROP TABLE IF EXISTS `projects`;";
```

---

## 5. mysql_ddl_3.result

**文件路径:** mysql_ddl_3.result

**统计信息:**
- 当前版本行数: 331
- Git版本行数: 331

**内容差异 (2 个):**

### 1. 内容被修改 (文件1行 8-8, 文件2行 8-8)

**当前版本内容:**
```
- mysql_ddl_test_v31 /*!50001 CREATE DEFINER = `root`@`%` VIEW mysql_ddl_test_v31 AS Select id from mysql_ddl_test_t31 */ utf8mb4 utf8mb4_general_ci
```

**Git历史版本内容:**
```
+ mysql_ddl_test_v31 /*!50001 CREATE DEFINER = `root`@`%` VIEW mysql_ddl_test_v31 AS Select id from mysql_ddl_test_t31 */;utf8mb4 utf8mb4_general_ci
```

### 2. 内容被修改 (文件1行 316-316, 文件2行 316-316)

**当前版本内容:**
```
- mysql_ddl_test_v32 CREATE VIEW mysql_ddl_test_v32 AS select convert(`mysql_ddl_test_t39`.`name` using utf8mb4)as converted_name from mysql_ddl_test_t39 utf8mb4 utf8mb4_general_ci
```

**Git历史版本内容:**
```
+ mysql_ddl_test_v32 CREATE VIEW mysql_ddl_test_v32 AS select convert(`mysql_ddl_test_t39`.`name` using utf8mb4)as converted_name from mysql_ddl_test_t39;utf8mb4 utf8mb4_general_ci
```

---

## 6. mysql_ddl_4.result

**文件路径:** mysql_ddl_4.result

**统计信息:**
- 当前版本行数: 209
- Git版本行数: 217

**内容差异 (1 个):**

### 1. 内容被修改 (文件1行 208-208, 文件2行 208-216)

**当前版本内容:**
```
- where(now()between `b`.`CONTRACT_EFFECTIVE_START` and `b`.`CONTRACT_EFFECTIVE_END`)utf8mb4 utf8mb4_general_ci
```

**Git历史版本内容:**
```
+ where(now()between `b`.`CONTRACT_EFFECTIVE_START` and `b`.`CONTRACT_EFFECTIVE_END`);utf8mb4 utf8mb4_general_ci
+ DROP VIEW IF EXISTS mysql_ddl_test_v42;
+ [unknown result because it is related to issue#moc 1229]
+ /*!50001 CREATE ALGORITHM=UNDEFINED */
+ /*!50013 DEFINER=`root`@`%` SQL SECURITY DEFINER */
+ /*!50001 VIEW mysql_ddl_test_v41 AS select `a`.`QUERYID` AS `QUERYID`,floor(`a`.`ORGANIZATION_ID`)AS `ORGANIZATION_ID`,floor(`a`.`WAREHOUSE_ID`)AS `WAREHOUSE_ID`,`d`.`WAREHOUSE_CODE` AS `WAREHOUSE_CODE`,`d`.`WAREHOUSE_NAME` AS `WAREHOUSE_NAME`,`d`.`WAREHOUSE_PID` AS `WAREHOUSE_PID`,`h`.`WAREHOUSE_NAME` AS `WAREHOUSE_PNAME`,floor(`a`.`ITEM_ID`)AS `ITEM_ID`,`b`.`ITEM_CLASS3` AS `ITEM_CLASS3`,`b`.`ITEM_CODE` AS `ITNBR`,`b`.`ITEM_NAME` AS `ITDSC`,`a`.`COLORNUMBER` AS `COLORNUMBER`,`a`.`BASE_LEVEL_ID` AS `BASE_LEVEL_ID`,`e`.`LEVELNAME` AS `LEVELNAME`,`a`.`WIDTH` AS `PRDWID`,`a`.`LENGTH` AS `PRDLNG`,floor(`a`.`BASE_PADDR_ID`)AS `BASE_PADDR_ID`,`f`.`BASE_PADDR` AS `BASE_PADDR`,floor(`a`.`ASSISTANT_UOM_ID`)AS `ASSISTANT_UOM_ID`,floor(`c`.`UOM_ID`)AS `UOM_ID`,`c`.`UOM_NAME` AS `UOM_NAME`,floor(`a`.`INV_BATCH_ID`)AS `INV_BATCH_ID`,`a`.`QTY_CANDRAFIT` AS `QTY_CANDRAFIT`,`a`.`QTY_SUBSISTENCE` AS `QTY_SUBSISTENCE`,`a`.`QTY_PRESHIP` AS `QTY_PRESHIP`,`a`.`QTY_FREEZE` AS `QTY_FREEZE`,`a`.`QTY_ONHAND` AS `QTY_ONHAND`,`a`.`QTY_ONHAND2` AS `QTY_ONHAND2`,`a`.`INV_ITEM_ID` AS `INV_ITEM_ID`,`a`.`IS_PALLET` AS `IS_PALLET`,`a`.`INV_BATCH_CODE` AS `INV_BATCH_CODE`,`a`.`CUSTOMER_ID` AS `CUSTOMER_ID`,`a`.`CUSTOMER_CODE` AS `CUSTOMER_CODE`,`a`.`CUSTOMER_NAME` AS `CUSTOMER_NAME`,(`a`.`QTY_CANDRAFIT` - ifnull(`g`.`LOCKQTY`,0))AS `USEFULQTY`,ifnull(`g`.`LOCKQTY`,0)AS `LOCKQTY`,`a`.`KULINGDAY` AS `KULINGDAY` from(((((((`skim`.`ces0020m` `a` join `skim`.`kaf_item` `b`)join `skim`.`kaf_uom` `c`)join `skim`.`kaf_warehouse` `d`)join `skim`.`kaf_base_level` `e`)left join `skim`.`kaf_base_paddr` `f` on((`a`.`BASE_PADDR_ID` = `f`.`BASE_PADDR_ID`)))left join(select sum(`skim`.`ces0021`.`LOCKQTY`)AS `LOCKQTY`,`skim`.`ces0021`.`ITNBR` AS `itnbr`,`skim`.`ces0021`.`WAREHOUSE_ID` AS `WAREHOUSE_ID`,`skim`.`ces0021`.`UOMID` AS `UOMID`,`skim`.`ces0021`.`COLORNUMBER` AS `COLORNUMBER`,`skim`.`ces0021`.`BASE_LEVEL_ID` AS `base_level_id`,`skim`.`ces0021`.`LENGTH` AS `LENGTH`,`skim`.`ces0021`.`WIDTH` AS `WIDTH`,`skim`.`ces0021`.`IS_PALLET` AS `IS_PALLET` from `skim`.`ces0021` where((`skim`.`ces0021`.`VALIDYN` = 'Y')and(`skim`.`ces0021`.`LOCKTYPE` = '1'))group by `skim`.`ces0021`.`ITNBR`,`skim`.`ces0021`.`WAREHOUSE_ID`,`skim`.`ces0021`.`UOMID`,`skim`.`ces0021`.`COLORNUMBER`,`skim`.`ces0021`.`BASE_LEVEL_ID`,`skim`.`ces0021`.`LENGTH`,`skim`.`ces0021`.`WIDTH`,`skim`.`ces0021`.`IS_PALLET`)`g` on(((`b`.`ITEM_CODE` = `g`.`itnbr`)and(`a`.`WAREHOUSE_ID` = `g`.`WAREHOUSE_ID`)and(`a`.`ASSISTANT_UOM_ID` = `g`.`UOMID`)and(`a`.`COLORNUMBER` = `g`.`COLORNUMBER`)and(`a`.`BASE_LEVEL_ID` = `g`.`base_level_id`)and(`a`.`LENGTH` = `g`.`LENGTH`)and(`a`.`WIDTH` = `g`.`WIDTH`)and(`a`.`IS_PALLET` = `g`.`IS_PALLET`))))left join `skim`.`kaf_warehouse` `h` on(((`d`.`WAREHOUSE_PID` = `h`.`WAREHOUSE_ID`)and(`d`.`ORGANIZATION_ID` = `h`.`ORGANIZATION_ID`))))where((`a`.`ITEM_ID` = `b`.`ITEM_ID`)and(`a`.`ASSISTANT_UOM_ID` = `c`.`UOM_ID`)and(`a`.`WAREHOUSE_ID` = `d`.`WAREHOUSE_ID`)and(`a`.`BASE_LEVEL_ID` = `e`.`BASE_LEVEL_ID`)and(`d`.`ACCOUNT_FLAG` = '1'))*/;
+ [unknown result because it is related to issue#moc 1229]
+ SHOW CREATE VIEW mysql_ddl_test_v42;
+ [unknown result because it is related to issue#moc 1229]
```

---

## 7. partition.result

**文件路径:** partition.result

**统计信息:**
- 当前版本行数: 0
- Git版本行数: 389

**内容差异 (1 个):**

### 1. 文件2中独有的内容 (行 1-389)

**Git历史版本内容:**
```
+ drop table if exists pt_table_1;
+ drop table if exists pt_table_2;
+ drop table if exists pt_table_3;
+ drop table if exists pt_table_5;
+ drop table if exists pt_table_6;
+ drop table if exists pt_table_21;
+ drop table if exists pt_table_22;
+ drop table if exists pt_table_23;
+ drop table if exists pt_table_24;
+ drop table if exists pt_table_31;
+ drop table if exists pt_table_32;
+ drop table if exists pt_table_33;
+ drop table if exists pt_table_34;
+ drop table if exists pt_table_35;
+ drop table if exists pt_table_36;
+ drop table if exists pt_table_37;
+ drop table if exists pt_table_41;
+ drop table if exists pt_table_42;
+ drop table if exists pt_table_43;
+ drop table if exists pt_table_44;
+ drop table if exists pt_table_45;
+ create table pt_table_1(col1 tinyint,col2 smallint,col3 int,clo4 bigint,col5 tinyint unsigned,col6 smallint unsigned,col7 int unsigned,col8 bigint unsigned,col9 float,col10 double,col11 varchar(255),col12 Date,col13 DateTime,col14 timestamp,col15 bool,col16 decimal(5,2),col17 text,col18 varchar(255),col19 varchar(255),col20 text,primary key(col1))partition by hash(col1)partitions 4;
+ load data infile '$resources/external_table_file/pt_table_data.csv' into table pt_table_1 fields terminated by ',';
+ select col1 from pt_table_1;
+ col1
+ -8
+ 21
+ -62
+ 91
+ -93
+ 33
+ 122
+ 121
+ 40
+ -75
+ 110
+ create table pt_table_2(col1 tinyint,col2 smallint,col3 int,clo4 bigint,col5 tinyint unsigned,col6 smallint unsigned,col7 int unsigned,col8 bigint unsigned,col9 float,col10 double,col11 varchar(255),col12 Date,col13 DateTime,col14 timestamp,col15 bool,col16 decimal(5,2),col17 text,col18 varchar(255),col19 varchar(255),col20 text,primary key(col5))partition by hash(col5);
+ load data infile '$resources/external_table_file/pt_table_data.csv' into table pt_table_2 fields terminated by ',';
+ select col5 from pt_table_2;
+ col5
+ 154
+ 122
+ 104
+ 141
+ 79
+ 82
+ 234
+ 28
+ 89
+ 98
+ 56
+ create table pt_table_3(col1 tinyint not null,col2 smallint,col3 int,clo4 bigint,col5 tinyint unsigned,col6 smallint unsigned,col7 int unsigned,col8 bigint unsigned,col9 float,col10 double,col11 varchar(255),col12 Date,col13 DateTime,col14 timestamp,col15 bool,col16 decimal(5,2),col17 text,col18 varchar(255),col19 varchar(255),col20 char(255)default 'style nine',primary key(col1,col20))partition by hash(col1)partitions 4;
+ load data infile '$resources/external_table_file/pt_table_data.csv' into table pt_table_3 fields terminated by ',';
+ select col1 from pt_table_3;
+ col1
+ -8
+ 21
+ -62
+ 91
+ -93
+ 33
+ 122
+ 121
+ 40
+ -75
+ 110
+ create table pt_table_5(col1 tinyint,col2 smallint,col3 int,clo4 bigint,col5 tinyint unsigned,col6 smallint unsigned,col7 int unsigned,col8 bigint unsigned,col9 float,col10 double,col11 varchar(255),col12 Date,col13 DateTime,col14 timestamp,col15 bool,col16 decimal(5,2),col17 text,col18 varchar(255),col19 varchar(255),col20 char(255))partition by hash(year(col12));
+ load data infile '$resources/external_table_file/pt_table_data.csv' into table pt_table_5 fields terminated by ',';
+ select col12 from pt_table_5;
+ col12
+ 4149-04-30
+ 2865-02-22
+ 6316-02-16
+ 9948-05-08
+ 7854-05-11
+ 2316-05-27
+ 8499-01-03
+ 9687-10-15
+ 1295-04-12
+ 1619-06-04
+ 3674-02-25
+ show create table pt_table_5;
+ Table Create Table
+ pt_table_5 CREATE TABLE `pt_table_5`(
+ `col1` tinyint DEFAULT NULL,
+ `col2` smallint DEFAULT NULL,
+ `col3` int DEFAULT NULL,
+ `clo4` bigint DEFAULT NULL,
+ `col5` tinyint unsigned DEFAULT NULL,
+ `col6` smallint unsigned DEFAULT NULL,
+ `col7` int unsigned DEFAULT NULL,
+ `col8` bigint unsigned DEFAULT NULL,
+ `col9` float DEFAULT NULL,
+ `col10` double DEFAULT NULL,
+ `col11` varchar(255)DEFAULT NULL,
+ `col12` date DEFAULT NULL,
+ `col13` datetime DEFAULT NULL,
+ `col14` timestamp NULL DEFAULT NULL,
+ `col15` bool DEFAULT NULL,
+ `col16` decimal(5,2)DEFAULT NULL,
+ `col17` text DEFAULT NULL,
+ `col18` varchar(255)DEFAULT NULL,
+ `col19` varchar(255)DEFAULT NULL,
+ `col20` char(255)DEFAULT NULL
+ )partition by hash(year(col12))
+ create table pt_table_6(col1 tinyint,col2 smallint,col3 int,clo4 bigint,col5 tinyint unsigned,col6 smallint unsigned,col7 int unsigned,col8 bigint unsigned,col9 float,col10 double,col11 varchar(255),col12 Date,col13 DateTime,col14 timestamp,col15 bool,col16 decimal(5,2),col17 text,col18 varchar(255),col19 varchar(255),col20 text)partition by LINEAR hash(col2)partitions 10;
+ load data infile '$resources/external_table_file/pt_table_data.csv' into table pt_table_6 fields terminated by ',';
+ select col2 from pt_table_6;
+ col2
+ 5807
+ 4300
+ 30792
+ -30001
+ 19053
+ 775
+ -23777
+ 19514
+ -22564
+ 11896
+ -18596
+ create table pt_table_10(col1 tinyint,col2 smallint,col3 int,primary key(col1))partition by hash(col2);
+ A PRIMARY KEY must include all columns in the table's partitioning function
+ create table pt_table_11(col1 tinyint,col2 smallint,col3 int,clo4 bigint,col5 tinyint unsigned,col6 smallint unsigned,col7 int unsigned,col8 bigint unsigned,col9 float,col10 double,col11 varchar(255),col12 Date,col13 DateTime,col14 timestamp,col15 bool,col16 decimal(5,2),col17 text,col18 varchar(255),col19 varchar(255),col20 text)partition by hash(col9)partitions 6;
+ Field 'col9' is of a not allowed type for this type of partitioning
+ create table pt_table_12(col1 tinyint,col2 smallint,col3 int,clo4 bigint,col5 tinyint unsigned,col6 smallint unsigned,col7 int unsigned,col8 bigint unsigned,col9 float,col10 double,col11 varchar(255),col12 Date,col13 DateTime,col14 timestamp,col15 bool,col16 decimal(5,2),col17 text,col18 varchar(255),col19 varchar(255),col20 char(255))partition by hash(col20);
+ Field 'col20' is of a not allowed type for this type of partitioning
+ create table pt_table_13(col1 tinyint,col2 smallint,col3 int,clo4 bigint,col5 tinyint unsigned,col6 smallint unsigned,col7 int unsigned,col8 bigint unsigned,col9 float,col10 double,col11 varchar(255),col12 Date,col13 DateTime,col14 timestamp,col15 bool,col16 decimal(5,2),col17 text,col18 varchar(255),col19 varchar(255),col20 char(255))partition by hash(col12);
+ Field 'col12' is of a not allowed type for this type of partitioning
+ create table pt_table_13(col1 tinyint,col2 smallint,col3 int,clo4 bigint,col5 tinyint unsigned,col6 smallint unsigned,col7 int unsigned,col8 bigint unsigned,col9 float,col10 double,col11 varchar(255),col12 Date,col13 DateTime,col14 timestamp,col15 bool,col16 decimal(5,2),col17 text,col18 varchar(255),col19 varchar(255),col20 char(255))partition by(col12);
+ SQL parser error: You have an error in your SQL syntax;check the manual that corresponds to your MatrixOne server version for the right syntax to use. syntax error at line 1 column 351 near "(col12);";
+ create table pt_table_21(col1 tinyint,col2 smallint,col3 int,clo4 bigint,col5 tinyint unsigned,col6 smallint unsigned,col7 int unsigned,col8 bigint unsigned,col9 float,col10 double,col11 varchar(255),col12 Date,col13 DateTime,col14 timestamp,col15 bool,col16 decimal(5,2),col17 text,col18 varchar(255),col19 varchar(255),col20 text,primary key(col2))partition by key(col2)partitions 4;
+ load data infile '$resources/external_table_file/pt_table_data.csv' into table pt_table_21 fields terminated by ',';
+ select col2 from pt_table_21;
+ col2
+ 5807
+ -30001
+ 11896
+ -23777
+ 19053
+ 4300
+ -22564
+ 30792
+ -18596
+ 19514
+ 775
+ create table pt_table_22(col1 tinyint,col2 smallint,col3 int,clo4 bigint,col5 tinyint unsigned,col6 smallint unsigned,col7 int unsigned,col8 bigint unsigned,col9 float,col10 double,col11 varchar(255),col12 Date,col13 DateTime,col14 timestamp,col15 bool,col16 decimal(5,2),col17 text,col18 varchar(255),col19 varchar(255),col20 text,primary key(col6,col18))partition by key(col6,col18)partitions 4;
+ load data infile '$resources/external_table_file/pt_table_data.csv' into table pt_table_22 fields terminated by ',';
+ select col2 from pt_table_22;
+ col2
+ -18596
+ 5807
+ 30792
+ 11896
+ -23777
+ 19053
+ -22564
+ 19514
+ 4300
+ -30001
+ 775
+ create table pt_table_23(col1 tinyint,col2 smallint,col3 int,clo4 bigint,col5 tinyint unsigned,col6 smallint unsigned,col7 int unsigned,col8 bigint unsigned,col9 float,col10 double,col11 varchar(255),col12 Date,col13 DateTime,col14 timestamp,col15 bool,col16 decimal(5,2),col17 text,col18 varchar(255),col19 varchar(255),col20 text,primary key(col19))partition by key(col19)partitions 4;
+ load data infile '$resources/external_table_file/pt_table_data.csv' into table pt_table_23 fields terminated by ',';
+ select col19 from pt_table_23;
+ col19
+ +=Zab
+ R.STU-_+=Zabcdefghigklmnopqrstuvw
+ U-_+=Zabcdefghigklmno
+ ;KL/MN?OPQR.STU-_+=Zabcdefghigklmno
+ -_+=Zabcdefghigklmnopqr
+ /MN?OPQR.STU-_+=Zabcdefghigklmnopqrstuvw
+ R.STU-_+=Zabcdefghigklmnopqrstuvwxyz0123456
+ TU-_+=Zabcdefghigklmnopqrstuvwxyz01234567
+ STU-_+=Zabcdefghigklmnopqrstuvwxyz01
+ I,G;KL/MN?OPQR.STU-_+=Zabcdefghigklmnopq
+ DEF,GHI,G;KL/MN?OPQR.STU-_+=Zabcdefghigklmnopqrstuvwxyz0123456
+ create table pt_table_24(col1 tinyint,col2 smallint,col3 int,clo4 bigint,col5 tinyint unsigned,col6 smallint unsigned,col7 int unsigned,col8 bigint unsigned,col9 float,col10 double,col11 varchar(255),col12 Date,col13 DateTime,col14 timestamp,col15 bool,col16 decimal(5,2),col17 text,col18 varchar(255),col19 varchar(255),col20 text)partition by key(col13)partitions 10;
+ load data infile '$resources/external_table_file/pt_table_data.csv' into table pt_table_24 fields terminated by ',';
+ select col13 from pt_table_24;
+ col13
+ 3114-10-05 23:59:59
+ 3647-01-21 23:59:59
+ 4023-04-27 23:59:59
+ 1014-07-01 23:59:59
+ 7031-10-23 00:00:00
+ 5732-08-07 00:00:00
+ 6216-12-30 00:00:00
+ 6868-02-03 00:00:00
+ 4844-01-09 23:59:59
+ 9976-06-04 00:00:00
+ 6438-11-29 00:00:00
+ create table pt_table_31(col1 tinyint,col2 smallint,col3 int,col4 bigint,col5 tinyint unsigned,col6 smallint unsigned,col7 int unsigned,col8 bigint unsigned,col9 float,col10 double,col11 varchar(255),col12 Date,col13 DateTime,col14 timestamp,col15 bool,col16 decimal(5,2),col17 text,col18 varchar(255),col19 varchar(255),col20 text,primary key(col3))partition by range(col3)(PARTITION p0 VALUES LESS THAN(100),PARTITION p1 VALUES LESS THAN(2000),PARTITION p2 VALUES LESS THAN(4000),PARTITION p3 VALUES LESS THAN(6000),PARTITION p5 VALUES LESS THAN MAXVALUE);
+ load data infile '$resources/external_table_file/pt_table_data.csv' into table pt_table_31 fields terminated by ',';
+ select col2 from pt_table_31;
+ col2
+ 5807
+ 19514
+ -30001
+ 11896
+ -23777
+ 4300
+ -22564
+ 30792
+ 775
+ -18596
+ 19053
+ create table pt_table_32(col1 tinyint,col2 smallint,col3 int,col4 bigint,col5 tinyint unsigned,col6 smallint unsigned,col7 int unsigned,col8 bigint unsigned,col9 float,col10 double,col11 varchar(255),col12 Date,col13 DateTime,col14 timestamp,col15 bool,col16 decimal(5,2),col17 text,col18 varchar(255),col19 varchar(255),col20 text)partition by range(col7)(PARTITION p0 VALUES LESS THAN(100),PARTITION p1 VALUES LESS THAN(2000),PARTITION p2 VALUES LESS THAN(4000),PARTITION p3 VALUES LESS THAN(6000),PARTITION p5 VALUES LESS THAN MAXVALUE);
+ load data infile '$resources/external_table_file/pt_table_data.csv' into table pt_table_32 fields terminated by ',';
+ select col2 from pt_table_32;
+ col2
+ 5807
+ 19514
+ 4300
+ -22564
+ 30792
+ -30001
+ 11896
+ 775
+ -18596
+ -23777
+ 19053
+ create table pt_table_33(col1 tinyint,col2 smallint,col3 int,col4 bigint,col5 tinyint unsigned,col6 smallint unsigned,col7 int unsigned,col8 bigint unsigned,col9 float,col10 double,col11 varchar(255),col12 Date,col13 DateTime,col14 timestamp,col15 bool,col16 decimal(5,2),col17 text,col18 varchar(255),col19 varchar(255),col20 char(255),primary key(col3,col7))partition by range(col7)(PARTITION p0 VALUES LESS THAN(100),PARTITION p1 VALUES LESS THAN(2000),PARTITION p2 VALUES LESS THAN(4000),PARTITION p3 VALUES LESS THAN(6000),PARTITION p5 VALUES LESS THAN MAXVALUE);
+ load data infile '$resources/external_table_file/pt_table_data.csv' into table pt_table_33 fields terminated by ',';
+ select col2 from pt_table_33;
+ col2
+ 5807
+ 19514
+ 4300
+ -22564
+ 30792
+ -30001
+ 11896
+ 775
+ -18596
+ -23777
+ 19053
+ create table pt_table_34(col1 tinyint,col2 smallint,col3 int,col4 bigint,col5 tinyint unsigned,col6 smallint unsigned,col7 int unsigned,col8 bigint unsigned,col9 float,col10 double,col11 varchar(255),col12 Date,col13 DateTime,col14 timestamp,col15 bool,col16 decimal(5,2),col17 text,col18 varchar(255),col19 varchar(255),col20 text)partition by range(year(col14))(PARTITION p0 VALUES LESS THAN(1991)comment ='expression range',PARTITION p1 VALUES LESS THAN(2000),PARTITION p2 VALUES LESS THAN(2009)comment ='range',PARTITION p3 VALUES LESS THAN(2010),PARTITION p5 VALUES LESS THAN MAXVALUE);
+ Constant,random or timezone-dependent expressions in(sub)partitioning function are not allowed
+ create table pt_table_35(col1 int not null,col2 smallint,col3 int not null,col4 bigint,col5 tinyint unsigned,col6 smallint unsigned,col7 int unsigned,col8 bigint unsigned,col9 float,col10 double,col11 varchar(255),col12 Date,col13 DateTime,col14 timestamp,col15 bool,col16 decimal(5,2),col17 text,col18 varchar(255),col19 varchar(255),col20 text,primary key(col3,col1))partition by range columns(col1,col3)(PARTITION p0 VALUES LESS THAN(100,300),PARTITION p1 VALUES LESS THAN(300,500),PARTITION p2 VALUES LESS THAN(500,MAXVALUE),PARTITION p3 VALUES LESS THAN(6000,MAXVALUE),PARTITION p4 VALUES LESS THAN(MAXVALUE,MAXVALUE));
+ load data infile '$resources/external_table_file/pt_table_data.csv' into table pt_table_35 fields terminated by ',';
+ select col14 from pt_table_35;
+ col14
+ 1975-09-09 23:59:59
+ 1985-01-12 23:59:59
+ 2034-02-10 00:00:00
+ 1977-03-18 23:59:59
+ 2036-08-23 00:00:00
+ 2037-12-04 00:00:00
+ 2035-05-25 00:00:00
+ 2014-09-26 00:00:00
+ 2011-03-10 00:00:00
+ 1996-08-27 23:59:59
+ 2011-10-04 00:00:00
+ create table pt_table_36(col1 tinyint,col2 smallint,col3 int,col4 bigint,col5 tinyint unsigned,col6 smallint unsigned,col7 int unsigned,col8 bigint unsigned,col9 float,col10 double,col11 varchar(255),col12 Date,col13 DateTime,col14 timestamp,col15 bool,col16 decimal(5,2),col17 text,col18 varchar(255),col19 varchar(255),col20 text,primary key(col3))partition by range(col3)(PARTITION p0 VALUES LESS THAN(100+50),PARTITION p1 VALUES LESS THAN(2000+100),PARTITION p2 VALUES LESS THAN(4000),PARTITION p3 VALUES LESS THAN(6000),PARTITION p5 VALUES LESS THAN MAXVALUE);
+ load data infile '$resources/external_table_file/pt_table_data.csv' into table pt_table_36 fields terminated by ',';
+ select col2 from pt_table_36;
+ col2
+ 5807
+ 19514
+ -30001
+ 11896
+ -23777
+ 4300
+ -22564
+ 30792
+ 775
+ -18596
+ 19053
+ create table pt_table_37(col1 tinyint,col2 smallint,col3 int,col4 bigint)partition by range(col90)(PARTITION p0 VALUES LESS THAN(100),PARTITION p1 VALUES LESS THAN(2000),PARTITION p2 VALUES LESS THAN(4000),PARTITION p3 VALUES LESS THAN(6000),PARTITION p5 VALUES LESS THAN MAXVALUE);
+ Unknown column 'col90' in 'partition function'
+ create table pt_table_37(col1 tinyint,col11 varchar(255),col12 Date)partition by range(col11)(PARTITION p0 VALUES LESS THAN(100),PARTITION p1 VALUES LESS THAN(2000),PARTITION p2 VALUES LESS THAN(4000),PARTITION p3 VALUES LESS THAN(6000),PARTITION p5 VALUES LESS THAN MAXVALUE);
+ Field 'col11' is of a not allowed type for this type of partitioning
+ create table pt_table_37(col1 tinyint,col11 varchar(255),col12 timestamp)partition by range(col12)(PARTITION p0 VALUES LESS THAN(100),PARTITION p1 VALUES LESS THAN(2000),PARTITION p2 VALUES LESS THAN(4000),PARTITION p3 VALUES LESS THAN(6000),PARTITION p5 VALUES LESS THAN MAXVALUE);
+ Field 'col12' is of a not allowed type for this type of partitioning
+ create table pt_table_37(col1 tinyint,col11 float,col12 timestamp)partition by range(col11)(PARTITION p0 VALUES LESS THAN(1991),PARTITION p1 VALUES LESS THAN(2000),PARTITION p2 VALUES LESS THAN(2009),PARTITION p3 VALUES LESS THAN(2010),PARTITION p5 VALUES LESS THAN MAXVALUE);
+ Field 'col11' is of a not allowed type for this type of partitioning
+ create table pt_table_37(col1 tinyint,col11 float,col12 timestamp)partition by range(ceil(col11))(PARTITION p0 VALUES LESS THAN(100),PARTITION p1 VALUES LESS THAN(2000),PARTITION p2 VALUES LESS THAN(4000),PARTITION p3 VALUES LESS THAN(6000),PARTITION p5 VALUES LESS THAN MAXVALUE);
+ Field 'ceil(col11)' is of a not allowed type for this type of partitioning
+ create table pt_table_37(col1 tinyint,col11 float,col12 timestamp)partition by range(col1);
+ For RANGE partitions each partition must be defined
+ create table pt_table_37(col1 tinyint not null,col2 smallint,col3 int not null,col4 bigint,col5 tinyint unsigned,col6 smallint unsigned,col7 int unsigned,col8 bigint unsigned,col9 float,col10 double,col11 varchar(255),col12 Date,col13 DateTime,col14 timestamp,col15 bool,col16 decimal(5,2),col17 text,col18 varchar(255),col19 varchar(255),col20 text,primary key(col3,col1))partition by range columns(col1,col3)(PARTITION p0 VALUES LESS THAN(100,300),PARTITION p1 VALUES LESS THAN(300,500),PARTITION p2 VALUES LESS THAN(500,MAXVALUE),PARTITION p3 VALUES LESS THAN(6000,MAXVALUE),PARTITION p4 VALUES LESS THAN(MAXVALUE,MAXVALUE));
+ Data truncation: data out of range: data type int8,value '300'
+ create table pt_table_41(col1 tinyint,col2 smallint,col3 int,col4 bigint,col5 tinyint unsigned,col6 smallint unsigned,col7 int unsigned,col8 bigint unsigned,col9 float,col10 double,col11 varchar(255),col12 Date,col13 DateTime,col14 timestamp,col15 bool,col16 decimal(5,2),col17 text,col18 varchar(255),col19 varchar(255),col20 text,primary key(col4))partition by list(col4)(PARTITION r0 VALUES IN(-6041648745842399623,2267877015687134490,7769629822818484334),PARTITION r1 VALUES IN(1234138289513302348,-3038428195984464330,-1681456935776973509),PARTITION r2 VALUES IN(-484407619835391694,-5246968895134993792,-3237107390156157130),PARTITION r3 VALUES IN(-2998549470145089608,6123486173032718578,6123486173032718570));
+ load data infile '$resources/external_table_file/pt_table_data.csv' into table pt_table_41 fields terminated by ',';
+ select col8 from pt_table_41 order by col8;
+ col8
+ 3143191107533743301
+ 4029688785176298663
+ 6204822205090614210
+ 6625004793680807495
+ 7094376021034692269
+ 8740918055557791046
+ 13381191796017069332
+ 14999475422109240954
+ 16635491969502097586
+ 17397115807377870895
+ 18225693328091251880
+ create table pt_table_42(col1 tinyint,col2 smallint,col3 int,col4 bigint,col5 tinyint unsigned,col6 smallint unsigned,col7 int unsigned,col8 bigint unsigned,col9 float,col10 double,col11 varchar(255),col12 Date,col13 DateTime,col14 timestamp,col15 bool,col16 decimal(5,2),col17 text,col18 varchar(255),col19 varchar(255),col20 text)partition by list(col8)(PARTITION r0 VALUES IN(14999475422109240954,6204822205090614210,6625004793680807490),PARTITION r1 VALUES IN(17397115807377870895,3143191107533743301,13381191796017069332),PARTITION r2 VALUES IN(8740918055557791046,4029688785176298663,6625004793680807495),PARTITION r3 VALUES IN(16635491969502097586,7094376021034692269,18225693328091251880));
+ load data infile '$resources/external_table_file/pt_table_data.csv' into table pt_table_42 fields terminated by ',';
+ select col8 from pt_table_42 order by col8;
+ col8
+ 3143191107533743301
+ 4029688785176298663
+ 6204822205090614210
+ 6625004793680807495
+ 7094376021034692269
+ 8740918055557791046
+ 13381191796017069332
+ 14999475422109240954
+ 16635491969502097586
+ 17397115807377870895
+ 18225693328091251880
+ create table pt_table_43(col1 tinyint,col2 smallint,col3 int,col4 bigint,col5 tinyint unsigned,col6 smallint unsigned,col7 int unsigned,col8 bigint unsigned,col9 float,col10 double,col11 varchar(255),col12 Date,col13 DateTime,col14 timestamp,col15 bool,col16 decimal(5,2),col17 text,col18 varchar(255),col19 varchar(255),col20 text)partition by list(col10)(PARTITION r0 VALUES IN(1,5,9,13,17,21),PARTITION r1 VALUES IN(2,6,10,14,18,22),PARTITION r2 VALUES IN(3,7,11,15,19,23),PARTITION r3 VALUES IN(4,8,12,16,20,24));
+ Field 'col10' is of a not allowed type for this type of partitioning
+ create table pt_table_44(col1 tinyint,col2 smallint,col3 int,col4 bigint,col5 tinyint unsigned,col6 smallint unsigned,col7 int unsigned,col8 bigint unsigned,col9 float,col10 double,col11 varchar(255),col12 Date,col13 DateTime,col14 timestamp,col15 bool,col16 decimal(5,2),col17 text,col18 varchar(255),col19 varchar(255),col20 text,primary key(col3,col4))partition by list columns(col3,col4)(PARTITION p0 VALUES IN((-1889972806,7769629822818484334),(NULL,NULL)),PARTITION p1 VALUES IN((-1030254547,-5246968895134993792),(-1006909301,-6041648745842399623),(-232972021,-3237107390156157130))comment='list column comment',PARTITION p2 VALUES IN((-179559641,1234138289513302348),(330484802,-2998549470145089608),(476482983,-484407619835391694)),PARTITION p3 VALUES IN((837702822,6123486173032718578),(1124555433,-1681456935776973509),(1287532466,-3038428195984464330),(1449911253,2267877015687134490)));
+ load data infile '$resources/external_table_file/pt_table_data.csv' into table pt_table_44 fields terminated by ',';
+ select col3,col4 from pt_table_44 order by col3,col4;
+ col3 col4
+ -1889972806 7769629822818484334
+ -1030254547 -5246968895134993792
+ -1006909301 -6041648745842399623
+ -232972021 -3237107390156157130
+ -179559641 1234138289513302348
+ 330484802 -2998549470145089608
+ 476482983 -484407619835391694
+ 837702822 6123486173032718578
+ 1124555433 -1681456935776973509
+ 1287532466 -3038428195984464330
+ 1449911253 2267877015687134490
+ create table pt_table_45(col1 tinyint,col2 smallint,col3 int,col4 bigint,col5 tinyint unsigned,col6 smallint unsigned,col7 int unsigned,col8 bigint unsigned,col9 float,col10 double,col11 varchar(255),col12 Date,col13 DateTime,col14 timestamp,col15 bool,col16 decimal(5,2),col17 text,col18 varchar(255),col19 varchar(255),col20 text)partition by list(year(col13))(PARTITION r0 VALUES IN(5732,9976,3647,6216),PARTITION r1 VALUES IN(7031,6868,4844,6438),PARTITION r2 VALUES IN(3114,1014,4023,2008));
+ load data infile '$resources/external_table_file/pt_table_data.csv' into table pt_table_45 fields terminated by ',';
+ select col3,col4 from pt_table_45 order by col3,col4;
+ col3 col4
+ -1889972806 7769629822818484334
+ -1030254547 -5246968895134993792
+ -1006909301 -6041648745842399623
+ -232972021 -3237107390156157130
+ -179559641 1234138289513302348
+ 330484802 -2998549470145089608
+ 476482983 -484407619835391694
+ 837702822 6123486173032718578
+ 1124555433 -1681456935776973509
+ 1287532466 -3038428195984464330
+ 1449911253 2267877015687134490
+ show create table pt_table_45;
+ Table Create Table
+ pt_table_45 CREATE TABLE `pt_table_45`(
+ `col1` tinyint DEFAULT NULL,
+ `col2` smallint DEFAULT NULL,
+ `col3` int DEFAULT NULL,
+ `col4` bigint DEFAULT NULL,
+ `col5` tinyint unsigned DEFAULT NULL,
+ `col6` smallint unsigned DEFAULT NULL,
+ `col7` int unsigned DEFAULT NULL,
+ `col8` bigint unsigned DEFAULT NULL,
+ `col9` float DEFAULT NULL,
+ `col10` double DEFAULT NULL,
+ `col11` varchar(255)DEFAULT NULL,
+ `col12` date DEFAULT NULL,
+ `col13` datetime DEFAULT NULL,
+ `col14` timestamp NULL DEFAULT NULL,
+ `col15` bool DEFAULT NULL,
+ `col16` decimal(5,2)DEFAULT NULL,
+ `col17` text DEFAULT NULL,
+ `col18` varchar(255)DEFAULT NULL,
+ `col19` varchar(255)DEFAULT NULL,
+ `col20` text DEFAULT NULL
+ )partition by list(year(col13))(partition r0 values in(5732,9976,3647,6216),partition r1 values in(7031,6868,4844,6438),partition r2 values in(3114,1014,4023,2008))
+ create table pt_table_46(col1 tinyint,col2 smallint,col3 int,col4 bigint,col5 tinyint unsigned,col6 smallint unsigned,col7 int unsigned,col8 bigint unsigned,col9 float,col10 double,col11 varchar(255),col12 Date,col13 DateTime,col14 timestamp,col15 bool,col16 decimal(5,2),col17 text,col18 varchar(255),col19 varchar(255),col20 text)partition by list(col20)(PARTITION r0 VALUES IN(1,5,9,13,17,21),PARTITION r1 VALUES IN(2,6,10,14,18,22),PARTITION r2 VALUES IN(3,7,11,15,19,23),PARTITION r3 VALUES IN(4,8,12,16,20,24));
+ Field 'col20' is of a not allowed type for this type of partitioning
+ create table pt_table_47(col13 DateTime,col14 timestamp,col15 bool,partition by list(col13)(PARTITION r0 VALUES IN(1,5,9,13,17,21),PARTITION r1 VALUES IN(2,6,10,14,18,22),PARTITION r2 VALUES IN(3,7,11,15,19,23),PARTITION r3 VALUES IN(4,8,12,16,20,24));
+ SQL parser error: You have an error in your SQL syntax;check the manual that corresponds to your MatrixOne server version for the right syntax to use. syntax error at line 1 column 76 near "partition by list(col13)(PARTITION r0 VALUES IN(1,5,9,13,17,21),PARTITION r1 VALUES IN(2,6,10,14,18,22),PARTITION r2 VALUES IN(3,7,11,15,19,23),PARTITION r3 VALUES IN(4,8,12,16,20,24));";
+ create table pt_table_48(col1 tinyint,col2 smallint,col10 decimal)partition by list(col10)(PARTITION r0 VALUES IN(1,5,9,13,17,21),PARTITION r1 VALUES IN(2,6,10,14,18,22));
+ Field 'col10' is of a not allowed type for this type of partitioning
+ create table pt_table_49(col1 tinyint,col2 smallint,col15 bool)partition by list(col15)(PARTITION r0 VALUES IN(1,5,9,13,17,21),PARTITION r1 VALUES IN(2,6,10,14,18,22));
+ Field 'col15' is of a not allowed type for this type of partitioning
+ create table pt_table_50(col1 tinyint,col2 smallint,col3 int,col4 bigint,col5 tinyint unsigned,col6 smallint unsigned,col7 int unsigned,col8 bigint unsigned,col9 float,col10 double,col11 varchar(255),col12 Date,col13 DateTime,col14 timestamp,col15 bool,col16 decimal(5,2),col17 text,col18 varchar(255),col19 varchar(255),col20 text,primary key(col4,col3,col11))partition by list(col3)(PARTITION r0 VALUES IN(1,5*2,9,13,17-20,21),PARTITION r1 VALUES IN(2,6,10,14/2,18,22),PARTITION r2 VALUES IN(3,7,11+6,15,19,23),PARTITION r3 VALUES IN(4,8,12,16,20,24));
+ invalid input: operator / is not allowed in the partition expression
+ create table pt_table_51(col1 tinyint,col2 smallint,col3 int,col4 bigint,col5 tinyint unsigned,col6 smallint unsigned,col7 int unsigned,col8 bigint unsigned,col9 float,col10 double,col11 varchar(255),col12 Date,col13 DateTime,col14 timestamp,col15 bool,col16 decimal(5,2),col17 text,col18 varchar(255),col19 varchar(255),col20 text)partition by list(year(col13))(PARTITION r0 VALUES IN(1999,2001,2003),PARTITION r1 VALUES IN(1999,2001,2003),PARTITION r2 VALUES IN(1999,2001,2003));
+ Multiple definition of same constant in list partitioning
+ create table pt_table_52(col1 tinyint,col2 smallint,col3 int,col4 bigint,col11 varchar(255),col12 Date,col13 DateTime,primary key(col4,col3,col11))partition by list(col2)(PARTITION r0 VALUES IN(1,5,9,13,17,21),PARTITION r1 VALUES IN(2,6,10,14,18,22),PARTITION r2 VALUES IN(3,7,11,15,19,23),PARTITION r3 VALUES IN(4,8,12,16,20,24));
+ A PRIMARY KEY must include all columns in the table's partitioning function
+ create table pt_table_53(col1 tinyint,col2 smallint,col3 int,col4 bigint,col5 tinyint unsigned,col6 smallint unsigned,col7 int unsigned,col8 bigint unsigned,col9 float,col10 double,col11 varchar(255),col12 Date,col13 DateTime,col14 timestamp,col15 bool,col16 decimal(5,2),col17 text,col18 varchar(255),col19 varchar(255),col20 text,primary key(col4,col3,col11))partition by list(col3)(PARTITION r0 VALUES IN(1,5*2,9,13,17-20,21),PARTITION r1 VALUES IN(2,6,10,14*2,18,22),PARTITION r2 VALUES IN(3,7,11+6,15,19,23),PARTITION r3 VALUES IN(4,8,12,16,20,24));
+ Multiple definition of same constant in list partitioning
+ create table pt_table_54(col1 tinyint,col2 smallint,col3 int,col4 bigint,col5 tinyint unsigned,col6 smallint unsigned,col7 int unsigned,col8 bigint unsigned,col9 float,col10 double,col11 varchar(255),col12 Date,col13 DateTime,col14 timestamp,col15 bool,col16 decimal(5,2),col17 text,col18 varchar(255),col19 varchar(255),col20 text,primary key(col4,col3,col11))partition by list(col3)(PARTITION r0 VALUES IN(1,5*2,9,13,17-20,21),PARTITION r1 VALUES IN(2,6,11,14*2,18,22),PARTITION r2 VALUES IN(3,7,11+6,15,19,23),PARTITION r3 VALUES IN(4,8,12,16,20,24));
+ create table dept(deptno int unsigned auto_increment,dname varchar(15),loc varchar(50),primary key(deptno));
+ create table emp(empno int unsigned auto_increment,ename varchar(15),job varchar(10),mgr int unsigned,hiredate date,sal decimal(7,2),comm decimal(7,2),deptno int unsigned,primary key(empno),foreign key(deptno)references dept(deptno))partition by key(empno)partitions 2;
+ Foreign keys are not yet supported in conjunction with partitioning
+ create table p_hash_table_test(col1 tinyint,col2 varchar(30),col3 decimal(6,3))partition by hash(ceil(col3))partitions 2;
+ The PARTITION function returns the wrong type
```

---

## 8. partition2.result

**文件路径:** partition2.result

**统计信息:**
- 当前版本行数: 0
- Git版本行数: 46

**内容差异 (1 个):**

### 1. 文件2中独有的内容 (行 1-46)

**Git历史版本内容:**
```
+ drop table if exists t1;
+ drop table if exists t2;
+ drop table if exists t3;
+ drop table if exists t4;
+ drop table if exists t5;
+ drop table if exists t6;
+ drop table if exists t7;
+ drop table if exists t8;
+ drop table if exists t9;
+ drop table if exists t10;
+ drop table if exists t11;
+ create table t1(a int,b int)partition by hash(a)partitions 2;
+ create table t2(a int,b int)partition by hash(a)partitions 2(partition x,partition y);
+ create table t3(a int,b int)partition by hash(a)partitions 3(partition x,partition y);
+ invalid input: Wrong number of partitions defined
+ create table t4(a int,b int)partition by key(a)partitions 2;
+ create table t5(a int,b int)partition by key()partitions 2;
+ invalid input: Field in list of fields for partition function not found in table
+ create table t6(a int primary key,b int)partition by key(a)partitions 2;
+ create table t7(a int primary key,b int)partition by key(b)partitions 2;
+ A PRIMARY KEY must include all columns in the table's partitioning function
+ create table t8(a int,b int,primary key(a,b))partition by key(b)partitions 2;
+ create table t9(a int unique key,b int)partition by key(a)partitions 2;
+ create table t10(a int,b int)partition by key(a)partitions 2(partition x,partition y);
+ create table t11(a int,b int)partition by key(a)partitions 3(partition x,partition y);
+ invalid input: Wrong number of partitions defined
+ show tables;
+ Tables_in_partition2
+ t1
+ t2
+ t4
+ t6
+ t8
+ t9
+ t10
+ drop table if exists t1;
+ drop table if exists t2;
+ drop table if exists t3;
+ drop table if exists t4;
+ drop table if exists t5;
+ drop table if exists t6;
+ drop table if exists t7;
+ drop table if exists t8;
+ drop table if exists t9;
+ drop table if exists t10;
+ drop table if exists t11;
```

---

## 9. partition3.result

**文件路径:** partition3.result

**统计信息:**
- 当前版本行数: 1264
- Git版本行数: 1266

**内容差异 (2 个):**

### 1. 文件2中独有的内容 (行 731-731)

**Git历史版本内容:**
```
+ emp_no title from_date to_date
```

### 2. 文件2中独有的内容 (行 733-733)

**Git历史版本内容:**
```
+ emp_no title from_date to_date
```

---

## 10. partition4.result

**文件路径:** partition4.result

**统计信息:**
- 当前版本行数: 120
- Git版本行数: 126

**内容差异 (6 个):**

### 1. 文件2中独有的内容 (行 78-78)

**Git历史版本内容:**
```
+ col1 col2 col3
```

### 2. 文件2中独有的内容 (行 80-80)

**Git历史版本内容:**
```
+ col1 col2 col3
```

### 3. 文件2中独有的内容 (行 82-82)

**Git历史版本内容:**
```
+ col1 col2 col3
```

### 4. 文件2中独有的内容 (行 84-84)

**Git历史版本内容:**
```
+ col1 col2 col3
```

### 5. 文件2中独有的内容 (行 86-86)

**Git历史版本内容:**
```
+ col1 col2 col3
```

### 6. 文件2中独有的内容 (行 116-116)

**Git历史版本内容:**
```
+ sale_id product_id sale_amount sale_date
```

---

## 11. partition_metadata.result

**文件路径:** partition_metadata.result

**统计信息:**
- 当前版本行数: 0
- Git版本行数: 167

**内容差异 (1 个):**

### 1. 文件2中独有的内容 (行 1-167)

**Git历史版本内容:**
```
+ drop database if exists db1;
+ create database db1;
+ use db1;
+ drop table if exists client_firms;
+ CREATE TABLE client_firms(
+ id INT,
+ name VARCHAR(35)
+ )
+ PARTITION BY LIST(id)(
+ PARTITION r0 VALUES IN(1,5,9,13,17,21),
+ PARTITION r1 VALUES IN(2,6,10,14,18,22),
+ PARTITION r2 VALUES IN(3,7,11,15,19,23),
+ PARTITION r3 VALUES IN(4,8,12,16,20,24)
+ );
+ select
+ table_catalog,
+ table_schema,
+ table_name,
+ partition_name,
+ partition_ordinal_position,
+ partition_method,
+ partition_expression,
+ partition_description,
+ table_rows,
+ avg_row_length,
+ data_length,
+ max_data_length,
+ partition_comment
+ from information_schema.partitions
+ where table_name = 'client_firms' and table_schema = 'db1';
+ table_catalog table_schema table_name partition_name partition_ordinal_position partition_method partition_expression partition_description table_rows avg_row_length data_length max_data_length partition_comment
+ def db1 client_firms r0 1 LIST id values in(1,5,9,13,17,21)0 0 0 0
+ def db1 client_firms r1 2 LIST id values in(2,6,10,14,18,22)0 0 0 0
+ def db1 client_firms r2 3 LIST id values in(3,7,11,15,19,23)0 0 0 0
+ def db1 client_firms r3 4 LIST id values in(4,8,12,16,20,24)0 0 0 0
+ drop table client_firms;
+ drop table if exists tk;
+ CREATE TABLE tk(col1 INT,col2 CHAR(5),col3 DATE)PARTITION BY KEY(col1,col2)PARTITIONS 4;
+ select
+ table_catalog,
+ table_schema,
+ table_name,
+ partition_name,
+ partition_ordinal_position,
+ partition_method,
+ partition_expression,
+ partition_description,
+ table_rows,
+ avg_row_length,
+ data_length,
+ max_data_length,
+ partition_comment
+ from information_schema.partitions
+ where table_name = 'tk' and table_schema = 'db1';
+ table_catalog table_schema table_name partition_name partition_ordinal_position partition_method partition_expression partition_description table_rows avg_row_length data_length max_data_length partition_comment
+ def db1 tk p0 1 null col1,col2 0 0 0 0
+ def db1 tk p1 2 null col1,col2 0 0 0 0
+ def db1 tk p2 3 null col1,col2 0 0 0 0
+ def db1 tk p3 4 null col1,col2 0 0 0 0
+ drop table tk;
+ drop table if exists t1;
+ CREATE TABLE t1(col1 INT,col2 CHAR(5),col3 DATE)PARTITION BY LINEAR HASH(YEAR(col3))PARTITIONS 6;
+ select
+ table_catalog,
+ table_schema,
+ table_name,
+ partition_name,
+ partition_ordinal_position,
+ partition_method,
+ partition_expression,
+ partition_description,
+ table_rows,
+ avg_row_length,
+ data_length,
+ max_data_length,
+ partition_comment
+ from information_schema.partitions
+ where table_name = 't1' and table_schema = 'db1';
+ table_catalog table_schema table_name partition_name partition_ordinal_position partition_method partition_expression partition_description table_rows avg_row_length data_length max_data_length partition_comment
+ def db1 t1 p0 1 LINEAR HASH YEAR(col3)0 0 0 0
+ def db1 t1 p1 2 LINEAR HASH YEAR(col3)0 0 0 0
+ def db1 t1 p2 3 LINEAR HASH YEAR(col3)0 0 0 0
+ def db1 t1 p3 4 LINEAR HASH YEAR(col3)0 0 0 0
+ def db1 t1 p4 5 LINEAR HASH YEAR(col3)0 0 0 0
+ def db1 t1 p5 6 LINEAR HASH YEAR(col3)0 0 0 0
+ drop table t1;
+ drop table if exists employees;
+ CREATE TABLE employees(
+ emp_no INT NOT NULL,
+ birth_date DATE NOT NULL,
+ first_name VARCHAR(14)NOT NULL,
+ last_name VARCHAR(16)NOT NULL,
+ gender varchar(5)NOT NULL,
+ hire_date DATE NOT NULL,
+ PRIMARY KEY(emp_no)
+ )
+ partition by range columns(emp_no)
+ (
+ partition p01 values less than(100001),
+ partition p02 values less than(270001),
+ partition p03 values less than(450001),
+ partition p04 values less than(530001),
+ partition p05 values less than(610001),
+ partition p06 values less than(MAXVALUE)
+ );
+ select
+ table_catalog,
+ table_schema,
+ table_name,
+ partition_name,
+ partition_ordinal_position,
+ partition_method,
+ partition_expression,
+ partition_description,
+ table_rows,
+ avg_row_length,
+ data_length,
+ max_data_length,
+ partition_comment
+ from information_schema.partitions
+ where table_name = 'employees' and table_schema = 'db1';
+ table_catalog table_schema table_name partition_name partition_ordinal_position partition_method partition_expression partition_description table_rows avg_row_length data_length max_data_length partition_comment
+ def db1 employees p01 1 RANGE COLUMNS emp_no 100001 0 0 0 0
+ def db1 employees p02 2 RANGE COLUMNS emp_no 270001 0 0 0 0
+ def db1 employees p03 3 RANGE COLUMNS emp_no 450001 0 0 0 0
+ def db1 employees p04 4 RANGE COLUMNS emp_no 530001 0 0 0 0
+ def db1 employees p05 5 RANGE COLUMNS emp_no 610001 0 0 0 0
+ def db1 employees p06 6 RANGE COLUMNS emp_no MAXVALUE 0 0 0 0
+ drop table employees;
+ drop table if exists trp;
+ CREATE TABLE trp(
+ id INT NOT NULL,
+ fname VARCHAR(30),
+ lname VARCHAR(30),
+ hired DATE NOT NULL DEFAULT '1970-01-01',
+ separated DATE NOT NULL DEFAULT '9999-12-31',
+ job_code INT,
+ store_id INT
+ )
+ PARTITION BY RANGE(YEAR(separated))(
+ PARTITION p0 VALUES LESS THAN(1991),
+ PARTITION p1 VALUES LESS THAN(1996),
+ PARTITION p2 VALUES LESS THAN(2001),
+ PARTITION p3 VALUES LESS THAN MAXVALUE
+ );
+ select
+ table_catalog,
+ table_schema,
+ table_name,
+ partition_name,
+ partition_ordinal_position,
+ partition_method,
+ partition_expression,
+ partition_description,
+ table_rows,
+ avg_row_length,
+ data_length,
+ max_data_length,
+ partition_comment
+ from information_schema.partitions
+ where table_name = 'trp' and table_schema = 'db1';
+ table_catalog table_schema table_name partition_name partition_ordinal_position partition_method partition_expression partition_description table_rows avg_row_length data_length max_data_length partition_comment
+ def db1 trp p0 1 RANGE YEAR(separated)1991 0 0 0 0
+ def db1 trp p1 2 RANGE YEAR(separated)1996 0 0 0 0
+ def db1 trp p2 3 RANGE YEAR(separated)2001 0 0 0 0
+ def db1 trp p3 4 RANGE YEAR(separated)MAXVALUE 0 0 0 0
+ drop table trp;
```

---

## 12. partition_prepare.result

**文件路径:** partition_prepare.result

**统计信息:**
- 当前版本行数: 813
- Git版本行数: 814

**内容差异 (1 个):**

### 1. 文件2中独有的内容 (行 581-581)

**Git历史版本内容:**
```
+ o_id o_entry_d o_carrier_id
```

---

## 13. partition_prune.result

**文件路径:** partition_prune.result

**统计信息:**
- 当前版本行数: 0
- Git版本行数: 425

**内容差异 (1 个):**

### 1. 文件2中独有的内容 (行 1-425)

**Git历史版本内容:**
```
+ drop database if exists db1;
+ create database db1;
+ use db1;
+ drop table if exists t1;
+ CREATE TABLE t1(
+ col1 INT NOT NULL,
+ col2 DATE NOT NULL,
+ col3 INT PRIMARY KEY
+ )PARTITION BY KEY(col3)PARTITIONS 4;
+ insert into `t1` values
+ (1,'1980-12-17',7369),
+ (2,'1981-02-20',7499),
+ (3,'1981-02-22',7521),
+ (4,'1981-04-02',7566),
+ (5,'1981-09-28',7654),
+ (6,'1981-05-01',7698),
+ (7,'1981-06-09',7782),
+ (8,'0087-07-13',7788),
+ (9,'1981-11-17',7839),
+ (10,'1981-09-08',7844),
+ (11,'2007-07-13',7876),
+ (12,'1981-12-03',7900),
+ (13,'1987-07-13',7980),
+ (14,'2001-11-17',7981),
+ (15,'1951-11-08',7982),
+ (16,'1927-10-13',7983),
+ (17,'1671-12-09',7984),
+ (18,'1981-11-06',7985),
+ (19,'1771-12-06',7986),
+ (20,'1985-10-06',7987),
+ (21,'1771-10-06',7988),
+ (22,'1981-10-05',7989),
+ (23,'2001-12-04',7990),
+ (24,'1999-08-01',7991),
+ (25,'1951-11-08',7992),
+ (26,'1927-10-13',7993),
+ (27,'1971-12-09',7994),
+ (28,'1981-12-09',7995),
+ (29,'2001-11-17',7996),
+ (30,'1981-12-09',7997),
+ (31,'2001-11-17',7998),
+ (32,'2001-11-17',7999);
+ select * from t1 where col3 = 7990;
+ col1 col2 col3
+ 23 2001-12-04 7990
+ select * from t1 where col3 = 7990 or col3 = 7988;
+ col1 col2 col3
+ 23 2001-12-04 7990
+ 21 1771-10-06 7988
+ select * from t1 where col3 in(7990,7698,7988);
+ col1 col2 col3
+ 23 2001-12-04 7990
+ 21 1771-10-06 7988
+ 6 1981-05-01 7698
+ select * from t1 where col3 = 7996 and col1 > 25;
+ col1 col2 col3
+ 29 2001-11-17 7996
+ select * from t1 where col1 = 24 and col3 = 7991 or col3 = 7990 order by col1,col3;
+ col1 col2 col3
+ 23 2001-12-04 7990
+ 24 1999-08-01 7991
+ select * from t1 where col3 > 7992;
+ col1 col2 col3
+ 27 1971-12-09 7994
+ 28 1981-12-09 7995
+ 29 2001-11-17 7996
+ 31 2001-11-17 7998
+ 26 1927-10-13 7993
+ 30 1981-12-09 7997
+ 32 2001-11-17 7999
+ select * from t1 where col3 >= 7992;
+ col1 col2 col3
+ 25 1951-11-08 7992
+ 27 1971-12-09 7994
+ 28 1981-12-09 7995
+ 29 2001-11-17 7996
+ 31 2001-11-17 7998
+ 26 1927-10-13 7993
+ 30 1981-12-09 7997
+ 32 2001-11-17 7999
+ select * from t1 where col1 > 25;
+ col1 col2 col3
+ 27 1971-12-09 7994
+ 28 1981-12-09 7995
+ 29 2001-11-17 7996
+ 31 2001-11-17 7998
+ 26 1927-10-13 7993
+ 30 1981-12-09 7997
+ 32 2001-11-17 7999
+ select * from t1 where col3 != 7782 and col3 != 7980;
+ col1 col2 col3
+ 9 1981-11-17 7839
+ 12 1981-12-03 7900
+ 20 1985-10-06 7987
+ 23 2001-12-04 7990
+ 25 1951-11-08 7992
+ 27 1971-12-09 7994
+ 28 1981-12-09 7995
+ 29 2001-11-17 7996
+ 31 2001-11-17 7998
+ 1 1980-12-17 7369
+ 19 1771-12-06 7986
+ 21 1771-10-06 7988
+ 22 1981-10-05 7989
+ 5 1981-09-28 7654
+ 6 1981-05-01 7698
+ 10 1981-09-08 7844
+ 16 1927-10-13 7983
+ 17 1671-12-09 7984
+ 24 1999-08-01 7991
+ 26 1927-10-13 7993
+ 30 1981-12-09 7997
+ 32 2001-11-17 7999
+ 2 1981-02-20 7499
+ 3 1981-02-22 7521
+ 4 1981-04-02 7566
+ 8 0087-07-13 7788
+ 11 2007-07-13 7876
+ 14 2001-11-17 7981
+ 15 1951-11-08 7982
+ 18 1981-11-06 7985
+ select * from t1 where col3 not in(7990,7698,7983,7980,7988,7995);
+ col1 col2 col3
+ 9 1981-11-17 7839
+ 12 1981-12-03 7900
+ 20 1985-10-06 7987
+ 25 1951-11-08 7992
+ 27 1971-12-09 7994
+ 29 2001-11-17 7996
+ 31 2001-11-17 7998
+ 1 1980-12-17 7369
+ 19 1771-12-06 7986
+ 22 1981-10-05 7989
+ 5 1981-09-28 7654
+ 10 1981-09-08 7844
+ 17 1671-12-09 7984
+ 24 1999-08-01 7991
+ 26 1927-10-13 7993
+ 30 1981-12-09 7997
+ 32 2001-11-17 7999
+ 2 1981-02-20 7499
+ 3 1981-02-22 7521
+ 4 1981-04-02 7566
+ 7 1981-06-09 7782
+ 8 0087-07-13 7788
+ 11 2007-07-13 7876
+ 14 2001-11-17 7981
+ 15 1951-11-08 7982
+ 18 1981-11-06 7985
+ select * from t1 where col3 between 7988 and 7990;
+ col1 col2 col3
+ 23 2001-12-04 7990
+ 21 1771-10-06 7988
+ 22 1981-10-05 7989
+ select * from t1 where col3 = 7996 or col1 > 25;
+ col1 col2 col3
+ 27 1971-12-09 7994
+ 28 1981-12-09 7995
+ 29 2001-11-17 7996
+ 31 2001-11-17 7998
+ 26 1927-10-13 7993
+ 30 1981-12-09 7997
+ 32 2001-11-17 7999
+ drop table if exists t2;
+ CREATE TABLE t2(
+ col1 INT NOT NULL,
+ col2 DATE NOT NULL,
+ col3 INT NOT NULL,
+ PRIMARY KEY(col1,col3)
+ )PARTITION BY KEY(col1,col3)PARTITIONS 4;
+ insert into `t2` values
+ (1,'1980-12-17',7369),
+ (2,'1981-02-20',7499),
+ (3,'1981-02-22',7521),
+ (4,'1981-04-02',7566),
+ (5,'1981-09-28',7654),
+ (6,'1981-05-01',7698),
+ (7,'1981-06-09',7782),
+ (8,'0087-07-13',7788),
+ (9,'1981-11-17',7839),
+ (10,'1981-09-08',7844),
+ (11,'2007-07-13',7876),
+ (12,'1981-12-03',7900),
+ (13,'1987-07-13',7980),
+ (14,'2001-11-17',7981),
+ (15,'1951-11-08',7982),
+ (16,'1927-10-13',7983),
+ (17,'1671-12-09',7984),
+ (18,'1981-11-06',7985),
+ (19,'1771-12-06',7986),
+ (20,'1985-10-06',7987),
+ (21,'1771-10-06',7988),
+ (22,'1981-10-05',7989),
+ (23,'2001-12-04',7990),
+ (24,'1999-08-01',7991),
+ (25,'1951-11-08',7992),
+ (26,'1927-10-13',7993),
+ (27,'1971-12-09',7994),
+ (28,'1981-12-09',7995),
+ (29,'2001-11-17',7996),
+ (30,'1981-12-09',7997),
+ (31,'2001-11-17',7998),
+ (32,'2001-11-17',7999);
+ select * from t2 where((col1 = 1 and col3 = 7369)or(col1 = 27 and col3 = 7994))and((col1 = 1 and col3 = 7369)or(col1 = 29 and col3 = 7996));
+ col1 col2 col3
+ 1 1980-12-17 7369
+ select * from t2 where((col1 = 1 and col3 = 7369)or(col1 = 12 and col3 = 7900))and((col1 = 1 and col3 = 7369)or(col1 = 29 and col3 = 7996));
+ col1 col2 col3
+ 1 1980-12-17 7369
+ select * from t2 where col1 = 23 and col3 = 7990;
+ col1 col2 col3
+ 23 2001-12-04 7990
+ select * from t2 where col1 = 1 and col3 = 7990;
+ col1 col2 col3
+ select * from t2 where col1 = 23 and col3 = 7990 or col1 = 30;
+ col1 col2 col3
+ 23 2001-12-04 7990
+ 30 1981-12-09 7997
+ select * from t2 where col1 in(23,6)and col3 in(7990,7698,7988);
+ col1 col2 col3
+ 6 1981-05-01 7698
+ 23 2001-12-04 7990
+ select * from t2 where col3 = 7996 and col1 > 25;
+ col1 col2 col3
+ 29 2001-11-17 7996
+ select * from t2 where col3 = 7990 or col3 = 7988;
+ col1 col2 col3
+ 21 1771-10-06 7988
+ 23 2001-12-04 7990
+ select * from t2 where(col1 = 1 and col3 = 7369)or(col1 = 27 and col3 = 7994);
+ col1 col2 col3
+ 1 1980-12-17 7369
+ 27 1971-12-09 7994
+ drop table if exists employees;
+ CREATE TABLE employees(
+ id INT NOT NULL,
+ fname VARCHAR(30),
+ lname VARCHAR(30),
+ hired DATE NOT NULL DEFAULT '1970-01-01',
+ separated DATE NOT NULL DEFAULT '9999-12-31',
+ job_code INT,
+ store_id INT
+ )PARTITION BY HASH(store_id)PARTITIONS 4;
+ INSERT INTO employees VALUES
+ (10001,'Georgi','Facello','1953-09-02','1986-06-26',120,1),
+ (10002,'Bezalel','Simmel','1964-06-02','1985-11-21',150,7),
+ (10003,'Parto','Bamford','1959-12-03','1986-08-28',140,3),
+ (10004,'Chirstian','Koblick','1954-05-01','1986-12-01',150,3),
+ (10005,'Kyoichi','Maliniak','1955-01-21','1989-09-12',150,18),
+ (10006,'Anneke','Preusig','1953-04-20','1989-06-02',150,15),
+ (10007,'Tzvetan','Zielinski','1957-05-23','1989-02-10',110,6),
+ (10008,'Saniya','Kalloufi','1958-02-19','1994-09-15',170,10),
+ (10009,'Sumant','Peac','1952-04-19','1985-02-18',110,13),
+ (10010,'Duangkaew','Piveteau','1963-06-01','1989-08-24',160,10),
+ (10011,'Mary','Sluis','1953-11-07','1990-01-22',120,8),
+ (10012,'Patricio','Bridgland','1960-10-04','1992-12-18',120,7),
+ (10013,'Eberhardt','Terkki','1963-06-07','1985-10-20',160,17),
+ (10014,'Berni','Genin','1956-02-12','1987-03-11',120,15),
+ (10015,'Guoxiang','Nooteboom','1959-08-19','1987-07-02',140,8),
+ (10016,'Kazuhito','Cappelletti','1961-05-02','1995-01-27',140,2),
+ (10017,'Cristinel','Bouloucos','1958-07-06','1993-08-03',170,10),
+ (10018,'Kazuhide','Peha','1954-06-19','1987-04-03',170,2),
+ (10019,'Lillian','Haddadi','1953-01-23','1999-04-30',170,13),
+ (10020,'Mayuko','Warwick','1952-12-24','1991-01-26',120,1),
+ (10021,'Ramzi','Erde','1960-02-20','1988-02-10',120,9),
+ (10022,'Shahaf','Famili','1952-07-08','1995-08-22',130,10),
+ (10023,'Bojan','Montemayor','1953-09-29','1989-12-17',120,5),
+ (10024,'Suzette','Pettey','1958-09-05','1997-05-19',130,4),
+ (10025,'Prasadram','Heyers','1958-10-31','1987-08-17',180,8),
+ (10026,'Yongqiao','Berztiss','1953-04-03','1995-03-20',170,4),
+ (10027,'Divier','Reistad','1962-07-10','1989-07-07',180,10),
+ (10028,'Domenick','Tempesti','1963-11-26','1991-10-22',110,11),
+ (10029,'Otmar','Herbst','1956-12-13','1985-11-20',110,12),
+ (10030,'Elvis','Demeyer','1958-07-14','1994-02-17',110,1),
+ (10031,'Karsten','Joslin','1959-01-27','1991-09-01',110,10),
+ (10032,'Jeong','Reistad','1960-08-09','1990-06-20',120,19),
+ (10033,'Arif','Merlo','1956-11-14','1987-03-18',120,14),
+ (10034,'Bader','Swan','1962-12-29','1988-09-21',130,16),
+ (10035,'Alain','Chappelet','1953-02-08','1988-09-05',130,3),
+ (10036,'Adamantios','Portugali','1959-08-10','1992-01-03',130,14),
+ (10037,'Pradeep','Makrucki','1963-07-22','1990-12-05',140,12),
+ (10038,'Huan','Lortz','1960-07-20','1989-09-20',140,7),
+ (10039,'Alejandro','Brender','1959-10-01','1988-01-19',110,20),
+ (10040,'Weiyi','Meriste','1959-09-13','1993-02-14',140,17);
+ select * from employees where store_id = 8;
+ id fname lname hired separated job_code store_id
+ 10011 Mary Sluis 1953-11-07 1990-01-22 120 8
+ 10015 Guoxiang Nooteboom 1959-08-19 1987-07-02 140 8
+ 10025 Prasadram Heyers 1958-10-31 1987-08-17 180 8
+ select * from employees where store_id = 8 or store_id = 10;
+ id fname lname hired separated job_code store_id
+ 10008 Saniya Kalloufi 1958-02-19 1994-09-15 170 10
+ 10010 Duangkaew Piveteau 1963-06-01 1989-08-24 160 10
+ 10017 Cristinel Bouloucos 1958-07-06 1993-08-03 170 10
+ 10022 Shahaf Famili 1952-07-08 1995-08-22 130 10
+ 10027 Divier Reistad 1962-07-10 1989-07-07 180 10
+ 10031 Karsten Joslin 1959-01-27 1991-09-01 110 10
+ 10011 Mary Sluis 1953-11-07 1990-01-22 120 8
+ 10015 Guoxiang Nooteboom 1959-08-19 1987-07-02 140 8
+ 10025 Prasadram Heyers 1958-10-31 1987-08-17 180 8
+ select * from employees where store_id in(1,2,11);
+ id fname lname hired separated job_code store_id
+ 10001 Georgi Facello 1953-09-02 1986-06-26 120 1
+ 10016 Kazuhito Cappelletti 1961-05-02 1995-01-27 140 2
+ 10018 Kazuhide Peha 1954-06-19 1987-04-03 170 2
+ 10020 Mayuko Warwick 1952-12-24 1991-01-26 120 1
+ 10028 Domenick Tempesti 1963-11-26 1991-10-22 110 11
+ 10030 Elvis Demeyer 1958-07-14 1994-02-17 110 1
+ select * from employees where store_id in(1,2,6,7);
+ id fname lname hired separated job_code store_id
+ 10002 Bezalel Simmel 1964-06-02 1985-11-21 150 7
+ 10007 Tzvetan Zielinski 1957-05-23 1989-02-10 110 6
+ 10012 Patricio Bridgland 1960-10-04 1992-12-18 120 7
+ 10038 Huan Lortz 1960-07-20 1989-09-20 140 7
+ 10001 Georgi Facello 1953-09-02 1986-06-26 120 1
+ 10016 Kazuhito Cappelletti 1961-05-02 1995-01-27 140 2
+ 10018 Kazuhide Peha 1954-06-19 1987-04-03 170 2
+ 10020 Mayuko Warwick 1952-12-24 1991-01-26 120 1
+ 10030 Elvis Demeyer 1958-07-14 1994-02-17 110 1
+ select * from employees where store_id in(1,2,11)or store_id in(6,7,18);
+ id fname lname hired separated job_code store_id
+ 10002 Bezalel Simmel 1964-06-02 1985-11-21 150 7
+ 10005 Kyoichi Maliniak 1955-01-21 1989-09-12 150 18
+ 10007 Tzvetan Zielinski 1957-05-23 1989-02-10 110 6
+ 10012 Patricio Bridgland 1960-10-04 1992-12-18 120 7
+ 10038 Huan Lortz 1960-07-20 1989-09-20 140 7
+ 10001 Georgi Facello 1953-09-02 1986-06-26 120 1
+ 10016 Kazuhito Cappelletti 1961-05-02 1995-01-27 140 2
+ 10018 Kazuhide Peha 1954-06-19 1987-04-03 170 2
+ 10020 Mayuko Warwick 1952-12-24 1991-01-26 120 1
+ 10028 Domenick Tempesti 1963-11-26 1991-10-22 110 11
+ 10030 Elvis Demeyer 1958-07-14 1994-02-17 110 1
+ select * from employees where store_id = 3 and id = 10004 or store_id = 10 order by id;
+ id fname lname hired separated job_code store_id
+ 10004 Chirstian Koblick 1954-05-01 1986-12-01 150 3
+ 10008 Saniya Kalloufi 1958-02-19 1994-09-15 170 10
+ 10010 Duangkaew Piveteau 1963-06-01 1989-08-24 160 10
+ 10017 Cristinel Bouloucos 1958-07-06 1993-08-03 170 10
+ 10022 Shahaf Famili 1952-07-08 1995-08-22 130 10
+ 10027 Divier Reistad 1962-07-10 1989-07-07 180 10
+ 10031 Karsten Joslin 1959-01-27 1991-09-01 110 10
+ select * from employees where(store_id = 3 and id = 10004)or(store_id = 10 and id = 10022);
+ id fname lname hired separated job_code store_id
+ 10004 Chirstian Koblick 1954-05-01 1986-12-01 150 3
+ 10022 Shahaf Famili 1952-07-08 1995-08-22 130 10
+ select * from employees where store_id > 15;
+ id fname lname hired separated job_code store_id
+ 10034 Bader Swan 1962-12-29 1988-09-21 130 16
+ 10039 Alejandro Brender 1959-10-01 1988-01-19 110 20
+ 10013 Eberhardt Terkki 1963-06-07 1985-10-20 160 17
+ 10040 Weiyi Meriste 1959-09-13 1993-02-14 140 17
+ 10005 Kyoichi Maliniak 1955-01-21 1989-09-12 150 18
+ 10032 Jeong Reistad 1960-08-09 1990-06-20 120 19
+ select * from employees where store_id = 10 or id = 10004;
+ id fname lname hired separated job_code store_id
+ 10008 Saniya Kalloufi 1958-02-19 1994-09-15 170 10
+ 10010 Duangkaew Piveteau 1963-06-01 1989-08-24 160 10
+ 10017 Cristinel Bouloucos 1958-07-06 1993-08-03 170 10
+ 10022 Shahaf Famili 1952-07-08 1995-08-22 130 10
+ 10027 Divier Reistad 1962-07-10 1989-07-07 180 10
+ 10031 Karsten Joslin 1959-01-27 1991-09-01 110 10
+ 10004 Chirstian Koblick 1954-05-01 1986-12-01 150 3
+ drop table if exists employees;
+ CREATE TABLE employees(
+ id INT NOT NULL,
+ fname VARCHAR(30),
+ lname VARCHAR(30),
+ hired DATE NOT NULL DEFAULT '1970-01-01',
+ separated DATE NOT NULL DEFAULT '9999-12-31',
+ job_code INT,
+ store_id INT
+ )PARTITION BY HASH(store_id)PARTITIONS 4;
+ INSERT INTO employees VALUES
+ (10001,'Georgi','Facello','1953-09-02','1986-06-26',120,1),
+ (10002,'Bezalel','Simmel','1964-06-02','1985-11-21',150,7),
+ (10003,'Parto','Bamford','1959-12-03','1986-08-28',140,3),
+ (10004,'Chirstian','Koblick','1954-05-01','1986-12-01',150,3),
+ (10005,'Kyoichi','Maliniak','1955-01-21','1989-09-12',150,18),
+ (10006,'Anneke','Preusig','1953-04-20','1989-06-02',150,15),
+ (10007,'Tzvetan','Zielinski','1957-05-23','1989-02-10',110,6),
+ (10008,'Saniya','Kalloufi','1958-02-19','1994-09-15',170,10),
+ (10009,'Sumant','Peac','1952-04-19','1985-02-18',110,13),
+ (10010,'Duangkaew','Piveteau','1963-06-01','1989-08-24',160,10),
+ (10011,'Mary','Sluis','1953-11-07','1990-01-22',120,8),
+ (10012,'Patricio','Bridgland','1960-10-04','1992-12-18',120,7),
+ (10013,'Eberhardt','Terkki','1963-06-07','1985-10-20',160,17),
+ (10014,'Berni','Genin','1956-02-12','1987-03-11',120,15),
+ (10015,'Guoxiang','Nooteboom','1959-08-19','1987-07-02',140,8),
+ (10016,'Kazuhito','Cappelletti','1961-05-02','1995-01-27',140,2),
+ (10017,'Cristinel','Bouloucos','1958-07-06','1993-08-03',170,10),
+ (10018,'Kazuhide','Peha','1954-06-19','1987-04-03',170,2),
+ (10019,'Lillian','Haddadi','1953-01-23','1999-04-30',170,13),
+ (10020,'Mayuko','Warwick','1952-12-24','1991-01-26',120,1),
+ (10021,'Ramzi','Erde','1960-02-20','1988-02-10',120,9),
+ (10022,'Shahaf','Famili','1952-07-08','1995-08-22',130,10),
+ (10023,'Bojan','Montemayor','1953-09-29','1989-12-17',120,5),
+ (10024,'Suzette','Pettey','1958-09-05','1997-05-19',130,4),
+ (10025,'Prasadram','Heyers','1958-10-31','1987-08-17',180,8),
+ (10026,'Yongqiao','Berztiss','1953-04-03','1995-03-20',170,4),
+ (10027,'Divier','Reistad','1962-07-10','1989-07-07',180,10),
+ (10028,'Domenick','Tempesti','1963-11-26','1991-10-22',110,11),
+ (10029,'Otmar','Herbst','1956-12-13','1985-11-20',110,12),
+ (10030,'Elvis','Demeyer','1958-07-14','1994-02-17',110,1),
+ (10031,'Karsten','Joslin','1959-01-27','1991-09-01',110,10),
+ (10032,'Jeong','Reistad','1960-08-09','1990-06-20',120,19),
+ (10033,'Arif','Merlo','1956-11-14','1987-03-18',120,14),
+ (10034,'Bader','Swan','1962-12-29','1988-09-21',130,16),
+ (10035,'Alain','Chappelet','1953-02-08','1988-09-05',130,3),
+ (10036,'Adamantios','Portugali','1959-08-10','1992-01-03',130,14),
+ (10037,'Pradeep','Makrucki','1963-07-22','1990-12-05',140,12),
+ (10038,'Huan','Lortz','1960-07-20','1989-09-20',140,7),
+ (10039,'Alejandro','Brender','1959-10-01','1988-01-19',110,20),
+ (10040,'Weiyi','Meriste','1959-09-13','1993-02-14',140,17);
+ insert into employees select * from employees;
+ insert into employees select * from employees;
+ insert into employees select * from employees;
+ insert into employees select * from employees;
+ insert into employees select * from employees;
+ insert into employees select * from employees;
+ insert into employees select * from employees;
+ insert into employees select * from employees;
+ insert into employees select * from employees;
+ insert into employees select * from employees;
+ delete from employees where store_id =(select min(store_id)from employees);
+ drop database db1;
```

---

## 14. rename_to_table.result

**文件路径:** rename_to_table.result

**统计信息:**
- 当前版本行数: 289
- Git版本行数: 290

**内容差异 (3 个):**

### 1. 文件2中独有的内容 (行 118-118)

**Git历史版本内容:**
```
+ a b
```

### 2. 文件1中独有的内容 (行 204-204)

**当前版本内容:**
```
- 90 5983
```

### 3. 文件2中独有的内容 (行 207-207)

**Git历史版本内容:**
```
+ 90 5983
```

---

## 15. secondary_index_alter.result

**文件路径:** secondary_index_alter.result

**统计信息:**
- 当前版本行数: 147
- Git版本行数: 150

**内容差异 (2 个):**

### 1. 文件2中独有的内容 (行 91-91)

**Git历史版本内容:**
```
+ name type column_name
```

### 2. 文件2中独有的内容 (行 149-149)

**Git历史版本内容:**
```
+ name type column_name
```

---

## 16. secondary_index_delete.result

**文件路径:** secondary_index_delete.result

**统计信息:**
- 当前版本行数: 62
- Git版本行数: 64

**内容差异 (2 个):**

### 1. 文件2中独有的内容 (行 12-12)

**Git历史版本内容:**
```
+ name type column_name
```

### 2. 文件2中独有的内容 (行 56-56)

**Git历史版本内容:**
```
+ name type column_name
```

---

## 17. secondary_index_master.result

**文件路径:** secondary_index_master.result

**统计信息:**
- 当前版本行数: 301
- Git版本行数: 304

**内容差异 (5 个):**

### 1. 文件2中独有的内容 (行 69-69)

**Git历史版本内容:**
```
+ a b c
```

### 2. 文件2中独有的内容 (行 72-72)

**Git历史版本内容:**
```
+ a b c
```

### 3. 文件2中独有的内容 (行 75-75)

**Git历史版本内容:**
```
+ t1 0 PRIMARY 1 c A 0 NULL NULL YES c
```

### 4. 文件1中独有的内容 (行 75-75)

**当前版本内容:**
```
- t1 0 PRIMARY 1 c A 0 NULL NULL YES c
```

### 5. 文件2中独有的内容 (行 97-97)

**Git历史版本内容:**
```
+ name type column_name
```

---

## 18. table_kind.result

**文件路径:** table_kind.result

**统计信息:**
- 当前版本行数: 331
- Git版本行数: 334

**内容差异 (3 个):**

### 1. 文件2中独有的内容 (行 67-67)

**Git历史版本内容:**
```
+ a b
```

### 2. 文件2中独有的内容 (行 119-119)

**Git历史版本内容:**
```
+ name type column_name
```

### 3. 文件2中独有的内容 (行 121-121)

**Git历史版本内容:**
```
+ relkind
```

---

## 19. use.result

**文件路径:** use.result

**统计信息:**
- 当前版本行数: 7
- Git版本行数: 6

**内容差异 (2 个):**

### 1. 内容被修改 (文件1行 3-3, 文件2行 3-3)

**当前版本内容:**
```
- use;
```

**Git历史版本内容:**
```
+ use;show tables;
```

### 2. 文件1中独有的内容 (行 5-5)

**当前版本内容:**
```
- show tables;
```

---

