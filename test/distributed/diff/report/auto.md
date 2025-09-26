# 批量Git Diff报告

生成时间: 2025-09-26 14:19:27.253215
比较commit: HEAD~1
总文件数: 2

## 1. auto_increment.result

**文件路径:** auto_increment.result

**统计信息:**
- 当前版本行数: 434
- Git版本行数: 689

**内容差异 (20 个):**

### 1. 文件2中独有的内容 (行 4-4)

**Git历史版本内容:**
```
+ col1
```

### 2. 内容被修改 (文件1行 24-24, 文件2行 25-25)

**当前版本内容:**
```
- Duplicate entry '10' for key 'col1'
```

**Git历史版本内容:**
```
+ Duplicate entry '10' for key '(.*)'
```

### 3. 内容被修改 (文件1行 85-85, 文件2行 86-86)

**当前版本内容:**
```
- Duplicate entry '10001' for key 'col1'
```

**Git历史版本内容:**
```
+ Duplicate entry '10001' for key '(.*)'
```

### 4. 内容被修改 (文件1行 87-87, 文件2行 88-88)

**当前版本内容:**
```
- Duplicate entry '10002' for key 'col1'
```

**Git历史版本内容:**
```
+ Duplicate entry '10002' for key '(.*)'
```

### 5. 内容被修改 (文件1行 99-99, 文件2行 100-100)

**当前版本内容:**
```
- data out of range: data type int,value 2147483648
```

**Git历史版本内容:**
```
+ Data truncation: data out of range: data type int,value 2147483648
```

### 6. 内容被修改 (文件1行 121-121, 文件2行 122-122)

**当前版本内容:**
```
- data out of range: data type smallint,value 32768
```

**Git历史版本内容:**
```
+ Data truncation: data out of range: data type smallint,value 32768
```

### 7. 内容被修改 (文件1行 135-135, 文件2行 136-136)

**当前版本内容:**
```
- data out of range: data type bigint,value 9223372036854775808
```

**Git历史版本内容:**
```
+ Data truncation: data out of range: data type bigint,value 9223372036854775808
```

### 8. 内容被修改 (文件1行 149-149, 文件2行 150-150)

**当前版本内容:**
```
- data out of range: data type tinyint unsigned,value 256
```

**Git历史版本内容:**
```
+ Data truncation: data out of range: data type tinyint unsigned,value 256
```

### 9. 文件2中独有的内容 (行 159-174)

**Git历史版本内容:**
```
+ Drop table if exists auto_increment10;
+ [unknown result because it is related to issue#10834]
+ Create table auto_increment10(col1 int auto_increment,col2 int,unique index(col1))auto_increment = 254;
+ [unknown result because it is related to issue#10834]
+ Insert into auto_increment10(col2)values(100);
+ [unknown result because it is related to issue#10834]
+ Insert into auto_increment10(col2)values(200);
+ [unknown result because it is related to issue#10834]
+ insert into auto_increment10(col2)values(100);
+ [unknown result because it is related to issue#10834]
+ select last_insert_id();
+ [unknown result because it is related to issue#10834]
+ Select * from auto_increment10;
+ [unknown result because it is related to issue#10834]
+ Drop table auto_increment10;
+ [unknown result because it is related to issue#10834]
```

### 10. 内容被修改 (文件1行 225-225, 文件2行 242-242)

**当前版本内容:**
```
- data out of range: data type int32,value '-2147483649'
```

**Git历史版本内容:**
```
+ Data truncation: data out of range: data type int32,value '-2147483649'
```

### 11. 文件2中独有的内容 (行 343-447)

**Git历史版本内容:**
```
+ drop table if exists auto_increment01;
+ [unknown result because it is related to issue#10903]
+ create temporary table auto_increment01(col1 int auto_increment primary key)auto_increment = 0;
+ [unknown result because it is related to issue#10903]
+ select * from auto_increment01;
+ [unknown result because it is related to issue#10903]
+ Insert into auto_increment01 values();
+ [unknown result because it is related to issue#10903]
+ select last_insert_id();
+ [unknown result because it is related to issue#10903]
+ Select * from auto_increment01;
+ [unknown result because it is related to issue#10903]
+ Insert into auto_increment01 values(1);
+ [unknown result because it is related to issue#10903]
+ Select * from auto_increment01;
+ [unknown result because it is related to issue#10903]
+ drop table auto_increment01;
+ [unknown result because it is related to issue#10903]
+ Drop table if exists auto_increment03;
+ [unknown result because it is related to issue#10903]
+ create temporary table auto_increment03(col1 int auto_increment primary key)auto_increment = 10000;
+ [unknown result because it is related to issue#10903]
+ Insert into auto_increment03 values();
+ [unknown result because it is related to issue#10903]
+ Insert into auto_increment03 values(10000);
+ [unknown result because it is related to issue#10903]
+ Insert into auto_increment03 values(10000);
+ [unknown result because it is related to issue#10903]
+ Insert into auto_increment03 values();
+ [unknown result because it is related to issue#10903]
+ select last_insert_id();
+ [unknown result because it is related to issue#10903]
+ Select * from auto_increment03;
+ [unknown result because it is related to issue#10903]
+ Drop table auto_increment03;
+ [unknown result because it is related to issue#10903]
+ Drop table if exists auto_increment04;
+ [unknown result because it is related to issue#10903]
+ Create temporary table auto_increment04(col1 int primary key auto_increment)auto_increment = 10;
+ [unknown result because it is related to issue#10903]
+ insert into auto_increment04 values();
+ [unknown result because it is related to issue#10903]
+ Select * from auto_increment04;
+ [unknown result because it is related to issue#10903]
+ Insert into auto_increment04 values();
+ [unknown result because it is related to issue#10903]
+ select last_insert_id();
+ [unknown result because it is related to issue#10903]
+ Insert into auto_increment04 values(100);
+ [unknown result because it is related to issue#10903]
+ Insert into auto_increment04 values(200);
+ [unknown result because it is related to issue#10903]
+ Insert into auto_increment04 values(10);
+ [unknown result because it is related to issue#10903]
+ Insert into auto_increment04 values(11);
+ [unknown result because it is related to issue#10903]
+ Select * from auto_increment04;
+ [unknown result because it is related to issue#10903]
+ Drop table auto_increment04;
+ [unknown result because it is related to issue#10903]
+ Drop table if exists auto_increment05;
+ [unknown result because it is related to issue#10834]
+ Create temporary table auto_increment05(col1 int unique key auto_increment)auto_increment = 10000;
+ [unknown result because it is related to issue#10834]
+ Insert into auto_increment05 values();
+ [unknown result because it is related to issue#10834]
+ Insert into auto_increment05 values();
+ [unknown result because it is related to issue#10834]
+ Insert into auto_increment05 values();
+ [unknown result because it is related to issue#10834]
+ select last_insert_id();
+ [unknown result because it is related to issue#10834]
+ Select * from auto_increment05;
+ [unknown result because it is related to issue#10834]
+ Insert into auto_increment05 values(10001);
+ [unknown result because it is related to issue#10834]
+ Insert into auto_increment05 values(10002);
+ [unknown result because it is related to issue#10834]
+ Select * from auto_increment05;
+ [unknown result because it is related to issue#10834]
+ Drop table auto_increment05;
+ [unknown result because it is related to issue#10834]
+ Drop table if exists auto_increment06;
+ Create temporary table auto_increment06(col1 int unsigned auto_increment primary key)auto_increment = 2147483646;
+ Insert into auto_increment06 values();
+ Insert into auto_increment06 values();
+ Insert into auto_increment06 values();
+ select last_insert_id();
+ last_insert_id()
+ 2147483648
+ Select * from auto_increment06;
+ col1
+ 2147483646
+ 2147483647
+ 2147483648
+ Insert into auto_increment06 values(10001);
+ Insert into auto_increment06 values(10002);
+ Select * from auto_increment06;
+ col1
+ 10001
+ 10002
+ 2147483646
+ 2147483647
+ 2147483648
+ Drop table auto_increment06;
```

### 12. 内容被修改 (文件1行 331-331, 文件2行 453-453)

**当前版本内容:**
```
- data out of range: data type smallint unsigned,value 65536
```

**Git历史版本内容:**
```
+ Data truncation: data out of range: data type smallint unsigned,value 65536
```

### 13. 内容被修改 (文件1行 333-333, 文件2行 455-455)

**当前版本内容:**
```
- data out of range: data type smallint unsigned,value 65537
```

**Git历史版本内容:**
```
+ Data truncation: data out of range: data type smallint unsigned,value 65537
```

### 14. 内容被修改 (文件1行 359-359, 文件2行 481-481)

**当前版本内容:**
```
- data out of range: data type tinyint,value 254
```

**Git历史版本内容:**
```
+ Data truncation: data out of range: data type tinyint,value 254
```

### 15. 内容被修改 (文件1行 361-361, 文件2行 483-483)

**当前版本内容:**
```
- data out of range: data type tinyint,value 255
```

**Git历史版本内容:**
```
+ Data truncation: data out of range: data type tinyint,value 255
```

### 16. 内容被修改 (文件1行 363-363, 文件2行 485-485)

**当前版本内容:**
```
- data out of range: data type tinyint,value 256
```

**Git历史版本内容:**
```
+ Data truncation: data out of range: data type tinyint,value 256
```

### 17. 文件2中独有的内容 (行 490-490)

**Git历史版本内容:**
```
+ col1
```

### 18. 文件2中独有的内容 (行 492-549)

**Git历史版本内容:**
```
+ Drop table if exists auto_increment10;
+ [unknown result because it is related to issue#10834]
+ Create temporary table auto_increment10(col1 int auto_increment,col2 int,unique index(col1))auto_increment = 3267183;
+ [unknown result because it is related to issue#10834]
+ Insert into auto_increment10(col2)values(100);
+ [unknown result because it is related to issue#10834]
+ Insert into auto_increment10(col2)values(200);
+ [unknown result because it is related to issue#10834]
+ insert into auto_increment10(col2)values(100);
+ [unknown result because it is related to issue#10834]
+ select last_insert_id();
+ [unknown result because it is related to issue#10834]
+ Select * from auto_increment10;
+ [unknown result because it is related to issue#10834]
+ Drop table auto_increment10;
+ [unknown result because it is related to issue#10834]
+ Drop table if exists auto_increment11;
+ [unknown result because it is related to issue#10903]
+ Create temporary table auto_increment11(col1 int auto_increment primary key)auto_increment = 100;
+ [unknown result because it is related to issue#10903]
+ insert into auto_increment11 values();
+ [unknown result because it is related to issue#10903]
+ Insert into auto_increment11 values();
+ [unknown result because it is related to issue#10903]
+ Insert into auto_increment11 values();
+ [unknown result because it is related to issue#10903]
+ select last_insert_id();
+ [unknown result because it is related to issue#10903]
+ Select * from auto_increment11;
+ [unknown result because it is related to issue#10903]
+ Delete from auto_increment11 where col1 = 100;
+ [unknown result because it is related to issue#10903]
+ Update auto_increment11 set col1 = 200 where col1 = 101;
+ [unknown result because it is related to issue#10834]
+ Select * from auto_increment11;
+ [unknown result because it is related to issue#10834]
+ Drop table auto_increment11;
+ [unknown result because it is related to issue#10834]
+ Drop table if exists auto_increment12;
+ create temporary table auto_increment12(col1 int auto_increment primary key)auto_increment = 10;
+ Insert into auto_increment12 values();
+ Insert into auto_increment12 values();
+ Select * from auto_increment12;
+ col1
+ 10
+ 11
+ Insert into auto_increment12 values(16.898291);
+ insert into auto_increment12 values();
+ select last_insert_id();
+ last_insert_id()
+ 17
+ Select * from auto_increment12;
+ col1
+ 10
+ 11
+ 16
+ 17
+ Drop table auto_increment12;
```

### 19. 内容被修改 (文件1行 394-394, 文件2行 575-575)

**当前版本内容:**
```
- data out of range: data type int32,value '-2147483649'
```

**Git历史版本内容:**
```
+ Data truncation: data out of range: data type int32,value '-2147483649'
```

### 20. 文件2中独有的内容 (行 595-668)

**Git历史版本内容:**
```
+ drop table if exists auto_increment15;
+ create temporary table auto_increment15(
+ a int primary key auto_increment,
+ b bigint auto_increment,
+ c int auto_increment,
+ d int auto_increment,
+ e bigint auto_increment
+ )auto_increment = 100;
+ show create table auto_increment15;
+ Table Create Table
+ auto_increment15 CREATE TABLE `auto_increment15`(
+ `a` INT NOT NULL AUTO_INCREMENT,
+ `b` BIGINT NOT NULL AUTO_INCREMENT,
+ `c` INT NOT NULL AUTO_INCREMENT,
+ `d` INT NOT NULL AUTO_INCREMENT,
+ `e` BIGINT NOT NULL AUTO_INCREMENT,
+ PRIMARY KEY(`a`)
+ )
+ insert into auto_increment15 values(),(),(),();
+ select * from auto_increment15 order by a;
+ a b c d e
+ 100 100 100 100 100
+ 101 101 101 101 101
+ 102 102 102 102 102
+ 103 103 103 103 103
+ insert into auto_increment15 values(NULL,NULL,NULL,NULL,NULL);
+ select * from auto_increment15 order by a;
+ a b c d e
+ 100 100 100 100 100
+ 101 101 101 101 101
+ 102 102 102 102 102
+ 103 103 103 103 103
+ 104 104 104 104 104
+ insert into auto_increment15(b,c,d)values(NULL,NULL,NULL);
+ select * from auto_increment15 order by a;
+ a b c d e
+ 100 100 100 100 100
+ 101 101 101 101 101
+ 102 102 102 102 102
+ 103 103 103 103 103
+ 104 104 104 104 104
+ 105 105 105 105 105
+ insert into auto_increment15(a,b)values(100,400);
+ tae data: duplicate
+ select * from auto_increment15 order by a;
+ a b c d e
+ 100 100 100 100 100
+ 101 101 101 101 101
+ 102 102 102 102 102
+ 103 103 103 103 103
+ 104 104 104 104 104
+ 105 105 105 105 105
+ insert into auto_increment15(c,d,e)values(200,200,200);
+ select * from auto_increment15;
+ a b c d e
+ 100 100 100 100 100
+ 101 101 101 101 101
+ 102 102 102 102 102
+ 103 103 103 103 103
+ 104 104 104 104 104
+ 105 105 105 105 105
+ 106 401 200 200 200
+ insert into auto_increment15(c,d,e)values(200,400,600);
+ select * from auto_increment15;
+ a b c d e
+ 100 100 100 100 100
+ 101 101 101 101 101
+ 102 102 102 102 102
+ 103 103 103 103 103
+ 104 104 104 104 104
+ 105 105 105 105 105
+ 106 401 200 200 200
+ 107 402 200 400 600
+ Drop table auto_increment15;
```

---

## 2. auto_increment_columns.result

**文件路径:** auto_increment_columns.result

**统计信息:**
- 当前版本行数: 835
- Git版本行数: 838

**内容差异 (24 个):**

### 1. 内容被修改 (文件1行 121-121, 文件2行 121-121)

**当前版本内容:**
```
- data out of range: data type int64,value '-9223372036854775809'
```

**Git历史版本内容:**
```
+ Data truncation: data out of range: data type int64,value '-9223372036854775809'
```

### 2. 内容被修改 (文件1行 124-124, 文件2行 124-124)

**当前版本内容:**
```
- data out of range: data type int64,value '9223372036854775808'
```

**Git历史版本内容:**
```
+ Data truncation: data out of range: data type int64,value '9223372036854775808'
```

### 3. 内容被修改 (文件1行 126-126, 文件2行 126-126)

**当前版本内容:**
```
- data out of range: data type bigint,value 9223372036854775808
```

**Git历史版本内容:**
```
+ Data truncation: data out of range: data type bigint,value 9223372036854775808
```

### 4. 内容被修改 (文件1行 138-138, 文件2行 138-138)

**当前版本内容:**
```
- data out of range: data type int32,value '-2147483649'
```

**Git历史版本内容:**
```
+ Data truncation: data out of range: data type int32,value '-2147483649'
```

### 5. 内容被修改 (文件1行 146-146, 文件2行 146-146)

**当前版本内容:**
```
- data out of range: data type int32,value '2147483648'
```

**Git历史版本内容:**
```
+ Data truncation: data out of range: data type int32,value '2147483648'
```

### 6. 内容被修改 (文件1行 148-148, 文件2行 148-148)

**当前版本内容:**
```
- data out of range: data type int,value 2147483648
```

**Git历史版本内容:**
```
+ Data truncation: data out of range: data type int,value 2147483648
```

### 7. 内容被修改 (文件1行 314-314, 文件2行 314-314)

**当前版本内容:**
```
- data out of range: data type int64,value '-9223372036854775809'
```

**Git历史版本内容:**
```
+ Data truncation: data out of range: data type int64,value '-9223372036854775809'
```

### 8. 内容被修改 (文件1行 317-317, 文件2行 317-317)

**当前版本内容:**
```
- data out of range: data type int64,value '9223372036854775808'
```

**Git历史版本内容:**
```
+ Data truncation: data out of range: data type int64,value '9223372036854775808'
```

### 9. 内容被修改 (文件1行 319-319, 文件2行 319-319)

**当前版本内容:**
```
- data out of range: data type bigint,value 9223372036854775808
```

**Git历史版本内容:**
```
+ Data truncation: data out of range: data type bigint,value 9223372036854775808
```

### 10. 内容被修改 (文件1行 331-331, 文件2行 331-331)

**当前版本内容:**
```
- data out of range: data type int32,value '-2147483649'
```

**Git历史版本内容:**
```
+ Data truncation: data out of range: data type int32,value '-2147483649'
```

### 11. 内容被修改 (文件1行 339-339, 文件2行 339-339)

**当前版本内容:**
```
- data out of range: data type int32,value '2147483648'
```

**Git历史版本内容:**
```
+ Data truncation: data out of range: data type int32,value '2147483648'
```

### 12. 内容被修改 (文件1行 346-346, 文件2行 346-346)

**当前版本内容:**
```
- data out of range: data type int,value 2147483648
```

**Git历史版本内容:**
```
+ Data truncation: data out of range: data type int,value 2147483648
```

### 13. 文件2中独有的内容 (行 599-599)

**Git历史版本内容:**
```
+ a b c
```

### 14. 文件2中独有的内容 (行 610-610)

**Git历史版本内容:**
```
+ a b c
```

### 15. 内容被修改 (文件1行 653-653, 文件2行 655-655)

**当前版本内容:**
```
- data out of range: data type tinyint,value 128
```

**Git历史版本内容:**
```
+ Data truncation: data out of range: data type tinyint,value 128
```

### 16. 内容被修改 (文件1行 666-666, 文件2行 668-668)

**当前版本内容:**
```
- data out of range: data type smallint,value 32768
```

**Git历史版本内容:**
```
+ Data truncation: data out of range: data type smallint,value 32768
```

### 17. 内容被修改 (文件1行 679-679, 文件2行 681-681)

**当前版本内容:**
```
- data out of range: data type int,value 2147483648
```

**Git历史版本内容:**
```
+ Data truncation: data out of range: data type int,value 2147483648
```

### 18. 内容被修改 (文件1行 692-692, 文件2行 694-694)

**当前版本内容:**
```
- data out of range: data type bigint,value 9223372036854775808
```

**Git历史版本内容:**
```
+ Data truncation: data out of range: data type bigint,value 9223372036854775808
```

### 19. 内容被修改 (文件1行 705-705, 文件2行 707-707)

**当前版本内容:**
```
- data out of range: data type tinyint unsigned,value 256
```

**Git历史版本内容:**
```
+ Data truncation: data out of range: data type tinyint unsigned,value 256
```

### 20. 内容被修改 (文件1行 718-718, 文件2行 720-720)

**当前版本内容:**
```
- data out of range: data type smallint unsigned,value 65536
```

**Git历史版本内容:**
```
+ Data truncation: data out of range: data type smallint unsigned,value 65536
```

### 21. 内容被修改 (文件1行 731-731, 文件2行 733-733)

**当前版本内容:**
```
- data out of range: data type int unsigned,value 4294967296
```

**Git历史版本内容:**
```
+ Data truncation: data out of range: data type int unsigned,value 4294967296
```

### 22. 内容被修改 (文件1行 744-744, 文件2行 746-746)

**当前版本内容:**
```
- data out of range: data type bigint unsigned,auto_incrment column constant value overflows bigint unsigned
```

**Git历史版本内容:**
```
+ Data truncation: data out of range: data type bigint unsigned,auto_incrment column constant value overflows bigint unsigned
```

### 23. 文件2中独有的内容 (行 819-819)

**Git历史版本内容:**
```
+ col1
```

### 24. 内容被修改 (文件1行 828-828, 文件2行 831-831)

**当前版本内容:**
```
- Duplicate entry '10' for key 'col1'
```

**Git历史版本内容:**
```
+ Duplicate entry '10' for key '(.*)'
```

---

