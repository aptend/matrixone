# 批量Git Diff报告

生成时间: 2025-10-10 10:31:21.972986
比较commit: HEAD
总文件数: 25

## 1. checkpoint.result

**文件路径:** checkpoint/checkpoint.result

**统计信息:**
- 当前版本行数: 124
- Git版本行数: 144

**内容差异 (4 个):**

### 1. 文件2中独有的内容 (行 7-16)

**Git历史版本内容:**
```
+ select mo_ctl('dn','checkpoint','');
+ mo_ctl(dn,checkpoint,)
+ {
+ "method": "Checkpoint",
+ "result": [
+ {
+ "returnStr": "OK"
+ }
+ ]
+ }
```

### 2. 文件2中独有的内容 (行 25-34)

**Git历史版本内容:**
```
+ select mo_ctl('dn','globalcheckpoint','');
+ mo_ctl(dn,GlobalCheckpoint,)
+ {
+ "method": "GlobalCheckpoint",
+ "result": [
+ {
+ "returnStr": "OK"
+ }
+ ]
+ }
```

### 3. 内容被修改 (文件1行 70-70, 文件2行 90-90)

**当前版本内容:**
```
- mo_ctl(dn,globalcheckpoint,)
```

**Git历史版本内容:**
```
+ mo_ctl(dn,GlobalCheckpoint,)
```

### 4. 内容被修改 (文件1行 105-105, 文件2行 125-125)

**当前版本内容:**
```
- mo_ctl(dn,DiskCleaner,stop_gc)
```

**Git历史版本内容:**
```
+ mo_ctl('dn','DiskCleaner','stop_gc')
```

---

## 2. delete.result

**文件路径:** delete/delete.result

**统计信息:**
- 当前版本行数: 438
- Git版本行数: 478

**内容差异 (8 个):**

### 1. 文件2中独有的内容 (行 95-95)

**Git历史版本内容:**
```
+ a b c
```

### 2. 文件2中独有的内容 (行 100-100)

**Git历史版本内容:**
```
+ a b c
```

### 3. 文件2中独有的内容 (行 127-127)

**Git历史版本内容:**
```
+ a
```

### 4. 文件2中独有的内容 (行 195-216)

**Git历史版本内容:**
```
+ drop table if exists t1;
+ create table t1(a int,b int,unique key(a));
+ insert into t1 values(1,1);
+ insert into t1 values(2,2);
+ insert into t1 values(3,3);
+ insert into t1 values(4,4);
+ select * from t1;
+ delete from t1 where a = 1;
+ select * from t1;
+ insert into t1 values(1,2);
+ drop table if exists t1;
+ create table t1(a int,b int,unique key(a,b));
+ insert into t1 values(1,2);
+ insert into t1 values(1,3);
+ insert into t1 values(2,2);
+ insert into t1 values(2,3);
+ select * from t1;
+ delete from t1 where a = 1;
+ select * from t1;
+ insert into t1 values(1,2);
+ insert into t1 values(1,null);
+ delete from t1 where a = 1;
```

### 5. 文件2中独有的内容 (行 231-242)

**Git历史版本内容:**
```
+ insert into t select * from t;
+ begin;
+ insert into t select * from t;
+ delete from t where a = 1;
+ select count(*)from t;
+ rollback;
+ begin;
+ insert into t select * from t;
+ delete from t where a = 1;
+ select count(*)from t;
+ commit;
+ select count(*)from t;
```

### 6. 文件2中独有的内容 (行 327-327)

**Git历史版本内容:**
```
+ a b c
```

### 7. 文件2中独有的内容 (行 347-347)

**Git历史版本内容:**
```
+ a b c
```

### 8. 文件2中独有的内容 (行 371-371)

**Git历史版本内容:**
```
+ id t5_id
```

---

## 3. delete_index.result

**文件路径:** delete/delete_index.result

**统计信息:**
- 当前版本行数: 625
- Git版本行数: 632

**内容差异 (12 个):**

### 1. 文件2中独有的内容 (行 35-35)

**Git历史版本内容:**
```
+ a b c
```

### 2. 文件2中独有的内容 (行 67-67)

**Git历史版本内容:**
```
+ a b c
```

### 3. 文件2中独有的内容 (行 99-99)

**Git历史版本内容:**
```
+ a b c
```

### 4. 文件2中独有的内容 (行 131-131)

**Git历史版本内容:**
```
+ a b c
```

### 5. 文件2中独有的内容 (行 163-163)

**Git历史版本内容:**
```
+ a b c
```

### 6. 内容被修改 (文件1行 239-239, 文件2行 244-244)

**当前版本内容:**
```
- Duplicate entry '5' for key 'b'
```

**Git历史版本内容:**
```
+ Duplicate entry '5' for key '(.*)'
```

### 7. 内容被修改 (文件1行 320-320, 文件2行 325-325)

**当前版本内容:**
```
- Duplicate entry '(5,5)' for key '(b,c)'
```

**Git历史版本内容:**
```
+ Duplicate entry('\(\d\,\d\)'|'\d\w\d{5}\w\d{4}')for key '(.*)'
```

### 8. 内容被修改 (文件1行 401-401, 文件2行 406-406)

**当前版本内容:**
```
- Duplicate entry '5' for key 'b'
```

**Git历史版本内容:**
```
+ Duplicate entry '5' for key '(.*)'
```

### 9. 内容被修改 (文件1行 482-482, 文件2行 487-487)

**当前版本内容:**
```
- Duplicate entry '(5,5)' for key '(b,c)'
```

**Git历史版本内容:**
```
+ Duplicate entry('\(\d\,\d\)'|'\d\w\d{5}\w\d{4}')for key '(.*)'
```

### 10. 内容被修改 (文件1行 523-523, 文件2行 528-528)

**当前版本内容:**
```
- Duplicate entry '5' for key 'c'
```

**Git历史版本内容:**
```
+ Duplicate entry '5' for key '(.*)'
```

### 11. 文件2中独有的内容 (行 560-560)

**Git历史版本内容:**
```
+ a b c
```

### 12. 文件2中独有的内容 (行 592-592)

**Git历史版本内容:**
```
+ a b c
```

---

## 4. delete_index_table.result

**文件路径:** delete/delete_index_table.result

**统计信息:**
- 当前版本行数: 131
- Git版本行数: 134

**内容差异 (3 个):**

### 1. 文件2中独有的内容 (行 125-125)

**Git历史版本内容:**
```
+ empno ename job mgr hiredate sal comm deptno
```

### 2. 文件2中独有的内容 (行 128-128)

**Git历史版本内容:**
```
+ deptno dname loc
```

### 3. 文件2中独有的内容 (行 131-131)

**Git历史版本内容:**
```
+ empno ename job mgr hiredate sal comm deptno
```

---

## 5. delete_multiple_table.result

**文件路径:** delete/delete_multiple_table.result

**统计信息:**
- 当前版本行数: 215
- Git版本行数: 220

**内容差异 (3 个):**

### 1. 文件2中独有的内容 (行 92-92)

**Git历史版本内容:**
```
+ id name
```

### 2. 文件2中独有的内容 (行 144-147)

**Git历史版本内容:**
```
+ DELETE t1,t2,t3 FROM t1 RIGHT JOIN t2 ON t1.t1_id = t2.t1_id LEFT JOIN t3 ON t3.t2_id = t2.t2_id WHERE t2.t2_id = 555;
+ [unknown result because it is related to issue#5219]
+ DELETE t1,t2,t3 FROM t1 LEFT JOIN t2 ON t1.t1_id = t2.t1_id LEFT JOIN t3 ON t3.t2_id = t2.t2_id WHERE t2.t2_id = 555;
+ [unknown result because it is related to issue#5219]
```

### 3. 内容被修改 (文件1行 194-194, 文件2行 199-199)

**当前版本内容:**
```
- 313 1561
```

**Git历史版本内容:**
```
+ 313.0 1561
```

---

## 6. insert_auto_pk.result

**文件路径:** insert/insert_auto_pk.result

**统计信息:**
- 当前版本行数: 267
- Git版本行数: 290

**内容差异 (25 个):**

### 1. 内容被修改 (文件1行 2-3, 文件2行 2-2)

**当前版本内容:**
```
- create database insert_auto_pk;
- use insert_auto_pk;
```

**Git历史版本内容:**
```
+ create database insert_auto_pk;use insert_auto_pk;
```

### 2. 文件2中独有的内容 (行 14-14)

**Git历史版本内容:**
```
+ a count(*)
```

### 3. 文件2中独有的内容 (行 26-26)

**Git历史版本内容:**
```
+ a count(*)
```

### 4. 文件2中独有的内容 (行 38-38)

**Git历史版本内容:**
```
+ a count(*)
```

### 5. 文件2中独有的内容 (行 50-50)

**Git历史版本内容:**
```
+ a count(*)
```

### 6. 文件2中独有的内容 (行 62-62)

**Git历史版本内容:**
```
+ a count(*)
```

### 7. 文件2中独有的内容 (行 74-74)

**Git历史版本内容:**
```
+ a count(*)
```

### 8. 文件2中独有的内容 (行 86-86)

**Git历史版本内容:**
```
+ a count(*)
```

### 9. 文件2中独有的内容 (行 98-98)

**Git历史版本内容:**
```
+ a count(*)
```

### 10. 文件2中独有的内容 (行 110-110)

**Git历史版本内容:**
```
+ a count(*)
```

### 11. 文件2中独有的内容 (行 122-122)

**Git历史版本内容:**
```
+ a count(*)
```

### 12. 文件2中独有的内容 (行 134-134)

**Git历史版本内容:**
```
+ a count(*)
```

### 13. 文件2中独有的内容 (行 146-146)

**Git历史版本内容:**
```
+ a count(*)
```

### 14. 文件2中独有的内容 (行 158-158)

**Git历史版本内容:**
```
+ a count(*)
```

### 15. 文件2中独有的内容 (行 170-170)

**Git历史版本内容:**
```
+ a count(*)
```

### 16. 文件2中独有的内容 (行 182-182)

**Git历史版本内容:**
```
+ a count(*)
```

### 17. 文件2中独有的内容 (行 194-194)

**Git历史版本内容:**
```
+ a count(*)
```

### 18. 文件2中独有的内容 (行 206-206)

**Git历史版本内容:**
```
+ a count(*)
```

### 19. 文件2中独有的内容 (行 218-218)

**Git历史版本内容:**
```
+ a count(*)
```

### 20. 文件2中独有的内容 (行 230-230)

**Git历史版本内容:**
```
+ a count(*)
```

### 21. 文件2中独有的内容 (行 242-242)

**Git历史版本内容:**
```
+ a count(*)
```

### 22. 文件2中独有的内容 (行 254-254)

**Git历史版本内容:**
```
+ a count(*)
```

### 23. 文件2中独有的内容 (行 266-266)

**Git历史版本内容:**
```
+ a count(*)
```

### 24. 文件2中独有的内容 (行 278-278)

**Git历史版本内容:**
```
+ a count(*)
```

### 25. 文件2中独有的内容 (行 290-290)

**Git历史版本内容:**
```
+ a count(*)
```

---

## 7. insert_duplicate.result

**文件路径:** insert/insert_duplicate.result

**统计信息:**
- 当前版本行数: 170
- Git版本行数: 393

**内容差异 (11 个):**

### 1. 内容被修改 (文件1行 13-14, 文件2行 13-14)

**当前版本内容:**
```
- 2 shanghai 002 2 2022-09-23
- 3 guangzhou 003 3 2022-09-23
```

**Git历史版本内容:**
```
+ 3 guangzhou 003 3 2022-09-23
+ 2 shanghai 002 2 2022-09-23
```

### 2. 文件2中独有的内容 (行 17-62)

**Git历史版本内容:**
```
+ select * from indup_00;
+ id act_name spu_id uv update_time
+ 1 beijing 001 1 2021-01-03
+ 5 beijing 010 5 2022-10-23
+ 3 guangzhou 003 3 2022-09-23
+ 2 shanghai 002 2 2022-09-23
+ 4 shenzheng 004 4 2021-05-28
+ insert into indup_00 values(6,'shanghai','002',21,'1999-09-23'),(7,'guangzhou','003',31,'1999-09-23')on duplicate key update `act_name`=VALUES(`act_name`),`spu_id`=VALUES(`spu_id`),`uv`=VALUES(`uv`);
+ select * from indup_00;
+ id act_name spu_id uv update_time
+ 1 beijing 001 1 2021-01-03
+ 5 beijing 010 5 2022-10-23
+ 3 guangzhou 003 31 2022-09-23
+ 2 shanghai 002 21 2022-09-23
+ 4 shenzheng 004 4 2021-05-28
+ insert into indup_00 values(8,'shanghai','002',21,'1999-09-23')on duplicate key update `act_name`=NULL;
+ constraint violation: Column 'act_name' cannot be null
+ select * from indup_00;
+ id act_name spu_id uv update_time
+ 1 beijing 001 1 2021-01-03
+ 5 beijing 010 5 2022-10-23
+ 3 guangzhou 003 31 2022-09-23
+ 2 shanghai 002 21 2022-09-23
+ 4 shenzheng 004 4 2021-05-28
+ insert into indup_00 values(9,'shanxi','005',4,'2022-10-08'),(10,'shandong','006',6,'2022-11-22')on duplicate key update `act_name`='Hongkong';
+ select * from indup_00;
+ id act_name spu_id uv update_time
+ 1 beijing 001 1 2021-01-03
+ 5 beijing 010 5 2022-10-23
+ 3 guangzhou 003 31 2022-09-23
+ 10 shandong 006 6 2022-11-22
+ 2 shanghai 002 21 2022-09-23
+ 9 shanxi 005 4 2022-10-08
+ 4 shenzheng 004 4 2021-05-28
+ insert into indup_00 values(10,'xinjiang','008',7,NULL),(11,'hainan','009',8,NULL)on duplicate key update `act_name`='Hongkong';
+ select * from indup_00;
+ id act_name spu_id uv update_time
+ 1 beijing 001 1 2021-01-03
+ 5 beijing 010 5 2022-10-23
+ 3 guangzhou 003 31 2022-09-23
+ 11 hainan 009 8 null
+ 10 shandong 006 6 2022-11-22
+ 2 shanghai 002 21 2022-09-23
+ 9 shanxi 005 4 2022-10-08
+ 4 shenzheng 004 4 2021-05-28
+ 10 xinjiang 008 7 null
```

### 3. 文件2中独有的内容 (行 80-124)

**Git历史版本内容:**
```
+ select * from indup_01;
+ id act_name spu_id uv update_time
+ 1 beijing 001 1 2021-01-03
+ 2 shanghai 002 2 2022-09-23
+ 3 guangzhou 003 3 2022-09-23
+ 4 shenzheng 004 4 2021-05-28
+ 5 beijing 010 5 2022-10-23
+ insert into indup_01 values(6,'shanghai','002',21,'1999-09-23'),(7,'guangzhou','003',31,'1999-09-23')on duplicate key update `act_name`=VALUES(`act_name`),`spu_id`=VALUES(`spu_id`),`uv`=VALUES(`uv`);
+ select * from indup_01;
+ id act_name spu_id uv update_time
+ 1 beijing 001 1 2021-01-03
+ 2 shanghai 002 21 2022-09-23
+ 3 guangzhou 003 31 2022-09-23
+ 4 shenzheng 004 4 2021-05-28
+ 5 beijing 010 5 2022-10-23
+ insert into indup_01 values(8,'shanghai','002',21,'1999-09-23')on duplicate key update `act_name`=NULL;
+ constraint violation: Column 'act_name' cannot be null
+ select * from indup_01;
+ id act_name spu_id uv update_time
+ 1 beijing 001 1 2021-01-03
+ 2 shanghai 002 21 2022-09-23
+ 3 guangzhou 003 31 2022-09-23
+ 4 shenzheng 004 4 2021-05-28
+ 5 beijing 010 5 2022-10-23
+ insert into indup_01 values(9,'shanxi','005',4,'2022-10-08'),(10,'shandong','006',6,'2022-11-22')on duplicate key update `act_name`='Hongkong';
+ select * from indup_01;
+ id act_name spu_id uv update_time
+ 1 beijing 001 1 2021-01-03
+ 2 shanghai 002 21 2022-09-23
+ 3 guangzhou 003 31 2022-09-23
+ 4 shenzheng 004 4 2021-05-28
+ 5 beijing 010 5 2022-10-23
+ 9 shanxi 005 4 2022-10-08
+ 10 shandong 006 6 2022-11-22
+ insert into indup_01 values(10,'xinjiang','008',7,NULL),(11,'hainan','009',8,NULL)on duplicate key update `act_name`='Hongkong';
+ select * from indup_01;
+ id act_name spu_id uv update_time
+ 1 beijing 001 1 2021-01-03
+ 2 shanghai 002 21 2022-09-23
+ 3 guangzhou 003 31 2022-09-23
+ 4 shenzheng 004 4 2021-05-28
+ 5 beijing 010 5 2022-10-23
+ 9 shanxi 005 4 2022-10-08
+ 10 Hongkong 006 6 2022-11-22
+ 11 hainan 009 8 null
```

### 4. 文件2中独有的内容 (行 139-174)

**Git历史版本内容:**
```
+ select * from indup_02;
+ col1 col2 col3 col4
+ 2 bear right 1000
+ 10 apple left null
+ insert into indup_02(col1,col2,col3)values(2,'wechat','tower'),(3,'paper','up')on duplicate key update col1=col1+20,col3=values(col3);
+ select * from indup_02;
+ col1 col2 col3 col4
+ 3 paper up 30
+ 10 apple left null
+ 22 bear tower 1000
+ insert into indup_02 values(3,'aaa','bbb',30)on duplicate key update col1=col1+7;
+ Duplicate entry '10' for key 'col1'
+ select * from indup_02;
+ col1 col2 col3 col4
+ 3 paper up 30
+ 10 apple left null
+ 22 bear tower 1000
+ insert into indup_02 values(3,'aaa','bbb',30),(30,'abc','abc',10),(11,'a1','b1',300)on duplicate key update col1=col1*10,col4=0;
+ select * from indup_02;
+ col1 col2 col3 col4
+ 10 apple left null
+ 11 a1 b1 300
+ 22 bear tower 1000
+ 300 paper up 0
+ create table indup_tmp(col1 int,col2 varchar(20),col3 varchar(20));
+ insert into indup_tmp values(1,'apple','left'),(2,'bear','right'),(3,'paper','up'),(10,'wine','down'),(300,'box','high');
+ insert into indup_02(col1,col2,col3)select col1,col2,col3 from indup_tmp on duplicate key update indup_02.col3=left(indup_02.col3,2),col2='wow';
+ select * from indup_02;
+ col1 col2 col3 col4
+ 22 bear tower 1000
+ 11 a1 b1 300
+ 1 apple left 30
+ 2 bear right 30
+ 3 paper up 30
+ 10 wow le null
+ 300 wow up 0
```

### 5. 文件2中独有的内容 (行 177-177)

**Git历史版本内容:**
```
+ col1 col2 col3 col4
```

### 6. 文件2中独有的内容 (行 207-226)

**Git历史版本内容:**
```
+ insert into indup_03(col1,col2,col3)values(1,'apple','')on duplicate key update col3='constant';
+ select * from indup_03;
+ col1 col2 col3 col4
+ 1 apple constant null
+ 2 bear right 1000
+ insert into indup_03(col1,col2,col3)values(1,'apple','uuuu')on duplicate key update col3=NULL;
+ select * from indup_03;
+ col1 col2 col3 col4
+ 1 apple null null
+ 2 bear right 1000
+ insert into indup_03(col1,col2,col3)values(1,'apple','uuuu')on duplicate key update col3='';
+ select * from indup_03;
+ col1 col2 col3 col4
+ 1 apple null
+ 2 bear right 1000
+ insert into indup_03(col1,col2,col3)values(1,'apple','uuuu')on duplicate key update col1=2+3;
+ select * from indup_03;
+ col1 col2 col3 col4
+ 2 bear right 1000
+ 5 apple null
```

### 7. 文件2中独有的内容 (行 243-283)

**Git历史版本内容:**
```
+ select * from indup_04;
+ id act_name spu_id uv update_time
+ 2 shanghai 002 2 2022-09-23
+ 3 guangzhou 003 3 2022-09-23
+ 4 shenzheng 004 4 2021-05-28
+ 1 beijing 010 5 2021-01-03
+ insert into indup_04 values(2,'shanghai','002',21,'1999-09-23'),(3,'guangzhou','003',31,'1999-09-23')on duplicate key update `act_name`=VALUES(`act_name`),`spu_id`=VALUES(`spu_id`),`uv`=VALUES(`uv`);
+ select * from indup_04;
+ id act_name spu_id uv update_time
+ 4 shenzheng 004 4 2021-05-28
+ 1 beijing 010 5 2021-01-03
+ 2 shanghai 002 21 2022-09-23
+ 3 guangzhou 003 31 2022-09-23
+ insert into indup_04 values(2,'shanghai','002',21,'1999-09-23')on duplicate key update `act_name`=NULL;
+ constraint violation: Column 'act_name' cannot be null
+ select * from indup_04;
+ id act_name spu_id uv update_time
+ 4 shenzheng 004 4 2021-05-28
+ 1 beijing 010 5 2021-01-03
+ 2 shanghai 002 21 2022-09-23
+ 3 guangzhou 003 31 2022-09-23
+ insert into indup_04 values(5,'shanxi','005',4,'2022-10-08'),(6,'shandong','006',6,'2022-11-22')on duplicate key update `act_name`='Hongkong';
+ select * from indup_04;
+ id act_name spu_id uv update_time
+ 4 shenzheng 004 4 2021-05-28
+ 1 beijing 010 5 2021-01-03
+ 2 shanghai 002 21 2022-09-23
+ 3 guangzhou 003 31 2022-09-23
+ 5 shanxi 005 4 2022-10-08
+ 6 shandong 006 6 2022-11-22
+ insert into indup_04 values(10,'xinjiang','008',7,NULL),(11,'hainan','009',8,NULL)on duplicate key update `act_name`='Hongkong';
+ select * from indup_04;
+ id act_name spu_id uv update_time
+ 4 shenzheng 004 4 2021-05-28
+ 1 beijing 010 5 2021-01-03
+ 2 shanghai 002 21 2022-09-23
+ 3 guangzhou 003 31 2022-09-23
+ 5 shanxi 005 4 2022-10-08
+ 6 shandong 006 6 2022-11-22
+ 10 xinjiang 008 7 null
+ 11 hainan 009 8 null
```

### 8. 文件2中独有的内容 (行 300-300)

**Git历史版本内容:**
```
+ 10 goods 2
```

### 9. 文件1中独有的内容 (行 112-112)

**当前版本内容:**
```
- 10 goods 2
```

### 10. 文件2中独有的内容 (行 324-329)

**Git历史版本内容:**
```
+ truncate table indup_06;
+ insert into indup_06 values(1,1);
+ insert into indup_06 values(1,10),(2,20),(3,30),(4,40),(5,50),(6,60),(7,70),(8,80),(9,90),(10,100),(11,110),(12,120),(13,130),(14,140),(15,150),(16,160),(17,170),(18,180),(19,190),(20,200)on duplicate key update col1=col1+1;
+ insert into indup_06 values(1,10),(2,20),(3,30),(4,40),(5,50),(6,60),(7,70),(8,80),(9,90),(10,100),(11,110),(12,120),(13,130),(14,140),(15,150),(16,160),(17,170),(18,180),(19,190),(20,200)on duplicate key update col1=col1+1,col2=col2*10;
+ insert into indup_06 values(1,10),(2,20),(3,30),(4,40),(5,50),(6,60),(7,70),(8,80),(9,90),(10,100),(11,110),(12,120),(13,130),(14,140),(15,150),(16,160),(17,170),(18,180),(19,190),(20,200)on duplicate key update col1=col1+1,col2=col2/10;
+ constraint violation: Duplicate entry for key 'col1'
```

### 11. 文件2中独有的内容 (行 338-365)

**Git历史版本内容:**
```
+ select * from indup_07;
+ col1 col2 col3 col4
+ 23 22 55 2
+ 24 66 77 1
+ 25 99 88 1
+ 33 11 33 1
+ insert into indup_07 values(24,'1','1',100)on duplicate key update col1=2147483649;
+ Data truncation: data out of range: data type int32,value '2147483649'
+ prepare stmt1 from "insert into indup_07 values(?,'11','33',1)on duplicate key update col1=col1*10";
+ set @a_var = 1;
+ execute stmt1 using @a_var;
+ select * from indup_07;
+ col1 col2 col3 col4
+ 1 11 33 1
+ 23 22 55 2
+ 24 66 77 1
+ 25 99 88 1
+ 33 11 33 1
+ set @a_var = 23;
+ execute stmt1 using @a_var;
+ select * from indup_07;
+ col1 col2 col3 col4
+ 1 11 33 1
+ 24 66 77 1
+ 25 99 88 1
+ 33 11 33 1
+ 230 22 55 2
+ deallocate prepare stmt1;
```

---

## 8. insert_ignore.result

**文件路径:** insert/insert_ignore.result

**统计信息:**
- 当前版本行数: 100
- Git版本行数: 139

**内容差异 (12 个):**

### 1. 文件2中独有的内容 (行 22-22)

**Git历史版本内容:**
```
+ 1 Blue 20
```

### 2. 文件1中独有的内容 (行 23-23)

**当前版本内容:**
```
- 1 Blue 20
```

### 3. 文件2中独有的内容 (行 47-48)

**Git历史版本内容:**
```
+ 6 null
+ 12 null
```

### 4. 文件2中独有的内容 (行 51-51)

**Git历史版本内容:**
```
+ 8 2
```

### 5. 文件2中独有的内容 (行 53-54)

**Git历史版本内容:**
```
+ 9 5
+ 7 7
```

### 6. 文件2中独有的内容 (行 56-56)

**Git历史版本内容:**
```
+ 10 10
```

### 7. 文件1中独有的内容 (行 52-56)

**当前版本内容:**
```
- 6 null
- 7 7
- 8 2
- 9 5
- 10 10
```

### 8. 文件1中独有的内容 (行 58-58)

**当前版本内容:**
```
- 12 null
```

### 9. 文件2中独有的内容 (行 62-86)

**Git历史版本内容:**
```
+ insert ignore into insert_ignore_04(product_name,quantity_in_stock,price)VALUES(NULL,5,1200.00),('board',6,NULL),('phone',NULL,1500.00);
+ select * from insert_ignore_04;
+ product_id product_name quantity_in_stock price
+ 1 Laptop 0 1200.00
+ 2 Monitor 0 150.00
+ 3 Keyboard 0 0.00
+ 4 Mouse 0 15.00
+ 5 5 1200.00
+ 6 board 6 0.00
+ 7 phone null 1500.00
+ create table parent_table(parent_id INT AUTO_INCREMENT PRIMARY KEY,parent_name VARCHAR(255)NOT NULL);
+ create table child_table(child_id INT AUTO_INCREMENT PRIMARY KEY,child_name VARCHAR(255)NOT NULL,parent_id INT,FOREIGN KEY(parent_id)REFERENCES parent_table(parent_id)
+ );
+ insert ignore into parent_table(parent_name)VALUES('Parent 1'),('Parent 2'),('Parent 3');
+ insert ignore into child_table(child_name,parent_id)VALUES('Child 1',1),('Child 2',2),('Child 3',4),('Child 4',1);
+ select * from parent_table;
+ parent_id parent_name
+ 1 Parent 1
+ 2 Parent 2
+ 3 Parent 3
+ select * from child_table;
+ child_id child_name parent_id
+ 1 Child 1 1
+ 2 Child 2 2
+ 3 Child 4 1
```

### 10. 文件2中独有的内容 (行 89-90)

**Git历史版本内容:**
```
+ insert ignore into insert_ignore_02 values("abc",1234.56);
+ insert ignore into insert_ignore_02 select "abc",34.22;
```

### 11. 文件2中独有的内容 (行 93-103)

**Git历史版本内容:**
```
+ create table insert_ignore_05(id TINYINT,created_at DATETIME);
+ insert ignore INTO insert_ignore_05(id,created_at)VALUES(130,'2024-04-03 10:00:00'),(-129,'2024-04-03 11:00:00'),(100,'2024-04-03 12:00:00');
+ insert ignore INTO insert_ignore_05(id,created_at)VALUES(50,'9999-12-31 23:59:59'),(50,'2000-02-29 10:00:00'),(50,'2024-04-03 13:00:00');
+ select * from insert_ignore_05;
+ id created_at
+ 127 2024-04-03 10:00:00
+ -128 2024-04-03 11:00:00
+ 100 2024-04-03 12:00:00
+ 50 9999-12-31 23:59:59
+ 50 2000-02-29 10:00:00
+ 50 2024-04-03 13:00:00
```

### 12. 文件2中独有的内容 (行 108-108)

**Git历史版本内容:**
```
+ sale_id product_id sale_amount sale_date
```

---

## 9. on_duplicate_key.result

**文件路径:** insert/on_duplicate_key.result

**统计信息:**
- 当前版本行数: 210
- Git版本行数: 239

**内容差异 (5 个):**

### 1. 文件2中独有的内容 (行 20-22)

**Git历史版本内容:**
```
+ select * from t1;
+ a b
+ 4 100
```

### 2. 文件2中独有的内容 (行 105-107)

**Git历史版本内容:**
```
+ select * from t1;
+ a b
+ 4 100
```

### 3. 文件2中独有的内容 (行 148-152)

**Git历史版本内容:**
```
+ select * from indup_00 order by id;
+ id act_name spu_id uv update_time
+ 1 beijing 001 1 2021-01-03
+ 2 shanghai 002 21 2022-09-23
+ 3 guangzhou 003 31 2022-09-23
```

### 4. 文件2中独有的内容 (行 162-175)

**Git历史版本内容:**
```
+ select * from indup;
+ col1 col2 col3 col4
+ 33 11 33 1
+ 23 22 55 2
+ 24 66 77 1
+ 25 99 88 1
+ insert into indup values(24,'1','1',100)on duplicate key update col1=2147483649;
+ Data truncation: data out of range: data type int32,value '2147483649'
+ select * from indup;
+ col1 col2 col3 col4
+ 33 11 33 1
+ 23 22 55 2
+ 24 66 77 1
+ 25 99 88 1
```

### 5. 文件2中独有的内容 (行 181-184)

**Git历史版本内容:**
```
+ select * from t1 order by a;
+ a b c
+ 2 2 2
+ 21 1 10
```

---

## 10. limit.result

**文件路径:** select/limit.result

**统计信息:**
- 当前版本行数: 61
- Git版本行数: 68

**内容差异 (7 个):**

### 1. 文件2中独有的内容 (行 13-13)

**Git历史版本内容:**
```
+ a
```

### 2. 文件2中独有的内容 (行 15-15)

**Git历史版本内容:**
```
+ a
```

### 3. 文件2中独有的内容 (行 21-21)

**Git历史版本内容:**
```
+ a
```

### 4. 文件2中独有的内容 (行 45-45)

**Git历史版本内容:**
```
+ a
```

### 5. 文件2中独有的内容 (行 47-47)

**Git历史版本内容:**
```
+ a
```

### 6. 文件2中独有的内容 (行 53-53)

**Git历史版本内容:**
```
+ a
```

### 7. 文件2中独有的内容 (行 68-68)

**Git历史版本内容:**
```
+ a left(b,3)
```

---

## 11. minus.result

**文件路径:** select/minus.result

**统计信息:**
- 当前版本行数: 452
- Git版本行数: 466

**内容差异 (59 个):**

### 1. 文件2中独有的内容 (行 23-23)

**Git历史版本内容:**
```
+ null
```

### 2. 文件1中独有的内容 (行 24-24)

**当前版本内容:**
```
- null
```

### 3. 文件2中独有的内容 (行 27-27)

**Git历史版本内容:**
```
+ null
```

### 4. 文件1中独有的内容 (行 28-28)

**当前版本内容:**
```
- null
```

### 5. 文件2中独有的内容 (行 33-33)

**Git历史版本内容:**
```
+ a
```

### 6. 文件2中独有的内容 (行 36-36)

**Git历史版本内容:**
```
+ bbbb
```

### 7. 文件2中独有的内容 (行 38-40)

**Git历史版本内容:**
```
+ (select b from t1)minus(select b from t1 limit 1);
+ b
+ aaaa
```

### 8. 内容被修改 (文件1行 37-39, 文件2行 42-44)

**当前版本内容:**
```
- (select b from t1)minus(select b from t1 limit 1);
- b
- bbbb
```

**Git历史版本内容:**
```
+ null
+ (select b from t1)minus(select b from t1 limit 2);
+ b
```

### 9. 文件1中独有的内容 (行 42-45)

**当前版本内容:**
```
- (select b from t1)minus(select b from t1 limit 2);
- b
- aaaa
- null
```

### 10. 文件2中独有的内容 (行 51-51)

**Git历史版本内容:**
```
+ b
```

### 11. 文件2中独有的内容 (行 57-57)

**Git历史版本内容:**
```
+ a
```

### 12. 文件2中独有的内容 (行 60-60)

**Git历史版本内容:**
```
+ bbbb
```

### 13. 文件1中独有的内容 (行 58-58)

**当前版本内容:**
```
- bbbb
```

### 14. 文件2中独有的内容 (行 83-83)

**Git历史版本内容:**
```
+ col1
```

### 15. 文件2中独有的内容 (行 85-85)

**Git历史版本内容:**
```
+ col1
```

### 16. 文件1中独有的内容 (行 107-108)

**当前版本内容:**
```
- null
- 2022-01-01 00:00:00
```

### 17. 文件2中独有的内容 (行 113-114)

**Git历史版本内容:**
```
+ null
+ 2022-01-01 00:00:00
```

### 18. 文件1中独有的内容 (行 117-118)

**当前版本内容:**
```
- null
- 2022-01-01 00:00:00
```

### 19. 文件2中独有的内容 (行 123-124)

**Git历史版本内容:**
```
+ null
+ 2022-01-01 00:00:00
```

### 20. 内容被修改 (文件1行 138-141, 文件2行 143-146)

**当前版本内容:**
```
- 20
- 10
- 30
- -10
```

**Git历史版本内容:**
```
+ 30.0
+ 10.0
+ -10.0
+ 20.0
```

### 21. 内容被修改 (文件1行 144-147, 文件2行 149-152)

**当前版本内容:**
```
- 20.00
- 10.00
- 30.00
- -10.00
```

**Git历史版本内容:**
```
+ 30.0
+ 10.0
+ -10.0
+ 20.0
```

### 22. 文件2中独有的内容 (行 155-155)

**Git历史版本内容:**
```
+ 127.0
```

### 23. 文件1中独有的内容 (行 152-153)

**当前版本内容:**
```
- 0
- 127
```

### 24. 文件2中独有的内容 (行 159-159)

**Git历史版本内容:**
```
+ 0.0
```

### 25. 文件2中独有的内容 (行 162-163)

**Git历史版本内容:**
```
+ 127.44
+ 127.0
```

### 26. 内容被修改 (文件1行 158-161, 文件2行 165-166)

**当前版本内容:**
```
- 1.10
- 0.00
- 127.00
- 127.44
```

**Git历史版本内容:**
```
+ 1.1
+ 0.0
```

### 27. 文件2中独有的内容 (行 168-168)

**Git历史版本内容:**
```
+ col1
```

### 28. 文件2中独有的内容 (行 175-175)

**Git历史版本内容:**
```
+ col2
```

### 29. 文件2中独有的内容 (行 178-178)

**Git历史版本内容:**
```
+ 127.44
```

### 30. 文件1中独有的内容 (行 173-173)

**当前版本内容:**
```
- 127.44
```

### 31. 文件2中独有的内容 (行 210-210)

**Git历史版本内容:**
```
+ b
```

### 32. 文件2中独有的内容 (行 213-214)

**Git历史版本内容:**
```
+ dd
+ cc
```

### 33. 文件2中独有的内容 (行 217-219)

**Git历史版本内容:**
```
+ (select b from t5)minus(select col3 from t6);
+ b
+ dd
```

### 34. 文件1中独有的内容 (行 208-212)

**当前版本内容:**
```
- dd
- (select b from t5)minus(select col3 from t6);
- b
- cc
- dd
```

### 35. 文件2中独有的内容 (行 222-222)

**Git历史版本内容:**
```
+ col1
```

### 36. 文件2中独有的内容 (行 225-226)

**Git历史版本内容:**
```
+ 44
+ 22
```

### 37. 文件2中独有的内容 (行 228-230)

**Git历史版本内容:**
```
+ 33
+ (select col3 from t6)minus(select b from t5);
+ col3
```

### 38. 文件1中独有的内容 (行 218-221)

**当前版本内容:**
```
- 33
- 44
- (select col3 from t6)minus(select b from t5);
- col3
```

### 39. 文件1中独有的内容 (行 223-223)

**当前版本内容:**
```
- 22
```

### 40. 文件2中独有的内容 (行 242-242)

**Git历史版本内容:**
```
+ a
```

### 41. 内容被修改 (文件1行 257-258, 文件2行 267-268)

**当前版本内容:**
```
- 2
- 3
```

**Git历史版本内容:**
```
+ 3
+ 2
```

### 42. 文件2中独有的内容 (行 271-271)

**Git历史版本内容:**
```
+ 1
```

### 43. 文件1中独有的内容 (行 264-264)

**当前版本内容:**
```
- 1
```

### 44. 文件2中独有的内容 (行 316-316)

**Git历史版本内容:**
```
+ a
```

### 45. 文件2中独有的内容 (行 326-326)

**Git历史版本内容:**
```
+ b
```

### 46. 内容被修改 (文件1行 362-362, 文件2行 374-374)

**当前版本内容:**
```
- NAME PHONE NAME PHONE
```

**Git历史版本内容:**
```
+ name phone name phone
```

### 47. 内容被修改 (文件1行 370-370, 文件2行 382-383)

**当前版本内容:**
```
- NAME PHONE NAME PHONE
```

**Git历史版本内容:**
```
+ name phone name phone
+ null null g 777
```

### 48. 文件1中独有的内容 (行 372-372)

**当前版本内容:**
```
- null null g 777
```

### 49. 内容被修改 (文件1行 382-382, 文件2行 394-394)

**当前版本内容:**
```
- NAME PHONE NAME PHONE
```

**Git历史版本内容:**
```
+ name phone name phone
```

### 50. 文件2中独有的内容 (行 411-411)

**Git历史版本内容:**
```
+ 20
```

### 51. 文件1中独有的内容 (行 400-400)

**当前版本内容:**
```
- 20
```

### 52. 文件2中独有的内容 (行 414-414)

**Git历史版本内容:**
```
+ a
```

### 53. 文件2中独有的内容 (行 420-420)

**Git历史版本内容:**
```
+ b
```

### 54. 文件1中独有的内容 (行 434-435)

**当前版本内容:**
```
- null
- 2022-01-01 00:00:00
```

### 55. 文件2中独有的内容 (行 449-450)

**Git历史版本内容:**
```
+ null
+ 2022-01-01 00:00:00
```

### 56. 文件1中独有的内容 (行 439-440)

**当前版本内容:**
```
- null
- 2022-01-01 00:00:00
```

### 57. 文件2中独有的内容 (行 454-455)

**Git历史版本内容:**
```
+ null
+ 2022-01-01 00:00:00
```

### 58. 文件1中独有的内容 (行 444-445)

**当前版本内容:**
```
- null
- 2022-01-01 00:00:00
```

### 59. 文件2中独有的内容 (行 459-460)

**Git历史版本内容:**
```
+ null
+ 2022-01-01 00:00:00
```

---

## 12. order_by_with_nulls.result

**文件路径:** select/order_by_with_nulls.result

**统计信息:**
- 当前版本行数: 575
- Git版本行数: 578

**内容差异 (32 个):**

### 1. 文件2中独有的内容 (行 181-181)

**Git历史版本内容:**
```
+ tiny small int_test big
```

### 2. 文件2中独有的内容 (行 183-183)

**Git历史版本内容:**
```
+ tiny small int_test big
```

### 3. 文件2中独有的内容 (行 185-185)

**Git历史版本内容:**
```
+ tiny small int_test big
```

### 4. 内容被修改 (文件1行 192-194, 文件2行 195-197)

**当前版本内容:**
```
- -1 -1.1 -1
- 0.000001 0.000002 0
- 0.000003 0.000001 3
```

**Git历史版本内容:**
```
+ -1.0 -1.1 -1
+ 1.0E-6 2.0E-6 0
+ 3.0E-6 1.0E-6 3
```

### 5. 内容被修改 (文件1行 199-201, 文件2行 202-204)

**当前版本内容:**
```
- 0.000001 0.000002 0
- 0.000003 0.000001 3
- -1 -1.1 -1
```

**Git历史版本内容:**
```
+ 1.0E-6 2.0E-6 0
+ 3.0E-6 1.0E-6 3
+ -1.0 -1.1 -1
```

### 6. 内容被修改 (文件1行 204-204, 文件2行 207-207)

**当前版本内容:**
```
- -1 -1.1 -1
```

**Git历史版本内容:**
```
+ -1.0 -1.1 -1
```

### 7. 内容被修改 (文件1行 206-207, 文件2行 209-210)

**当前版本内容:**
```
- 0.000001 0.000002 0
- 0.000003 0.000001 3
```

**Git历史版本内容:**
```
+ 1.0E-6 2.0E-6 0
+ 3.0E-6 1.0E-6 3
```

### 8. 内容被修改 (文件1行 242-244, 文件2行 245-247)

**当前版本内容:**
```
- 4 modrici 4 Fenland Fenland yisdilne 612094
- 2 kante 2 NA NA bolando 62102
- 1 ronaldo 1 Poutanga Poutanga liseber 520135
```

**Git历史版本内容:**
```
+ 4 modrici 4 Fenland Fenland yisdilne 612094.0
+ 2 kante 2 NA NA bolando 62102.0
+ 1 ronaldo 1 Poutanga Poutanga liseber 520135.0
```

### 9. 内容被修改 (文件1行 247-249, 文件2行 250-252)

**当前版本内容:**
```
- 4 modrici 4 Fenland Fenland yisdilne 612094
- 2 kante 2 NA NA bolando 62102
- 1 ronaldo 1 Poutanga Poutanga liseber 520135
```

**Git历史版本内容:**
```
+ 4 modrici 4 Fenland Fenland yisdilne 612094.0
+ 2 kante 2 NA NA bolando 62102.0
+ 1 ronaldo 1 Poutanga Poutanga liseber 520135.0
```

### 10. 内容被修改 (文件1行 252-254, 文件2行 255-257)

**当前版本内容:**
```
- 1 ronaldo 1 Poutanga Poutanga liseber 520135
- 2 kante 2 NA NA bolando 62102
- 4 modrici 4 Fenland Fenland yisdilne 612094
```

**Git历史版本内容:**
```
+ 1 ronaldo 1 Poutanga Poutanga liseber 520135.0
+ 2 kante 2 NA NA bolando 62102.0
+ 4 modrici 4 Fenland Fenland yisdilne 612094.0
```

### 11. 文件2中独有的内容 (行 276-276)

**Git历史版本内容:**
```
+ 2 kante M
```

### 12. 内容被修改 (文件1行 274-274, 文件2行 278-278)

**当前版本内容:**
```
- 2 kante M
```

**Git历史版本内容:**
```
+ 4 modrici M
```

### 13. 内容被修改 (文件1行 276-276, 文件2行 280-280)

**当前版本内容:**
```
- 4 modrici M
```

**Git历史版本内容:**
```
+ 3 noyer F
```

### 14. 内容被修改 (文件1行 278-278, 文件2行 282-282)

**当前版本内容:**
```
- 3 noyer F
```

**Git历史版本内容:**
```
+ 1 ronaldo F
```

### 15. 文件1中独有的内容 (行 280-280)

**当前版本内容:**
```
- 1 ronaldo F
```

### 16. 文件2中独有的内容 (行 286-286)

**Git历史版本内容:**
```
+ 1 ronaldo F
```

### 17. 内容被修改 (文件1行 284-285, 文件2行 288-288)

**当前版本内容:**
```
- 1 ronaldo F
- 3 noyer Germany
```

**Git历史版本内容:**
```
+ 2 kante M
```

### 18. 文件1中独有的内容 (行 287-287)

**当前版本内容:**
```
- 4 modrici UK
```

### 19. 内容被修改 (文件1行 290-290, 文件2行 292-293)

**当前版本内容:**
```
- 2 kante M
```

**Git历史版本内容:**
```
+ 3 noyer Germany
+ 4 modrici UK
```

### 20. 文件2中独有的内容 (行 352-359)

**Git历史版本内容:**
```
+ 1 ronaldo
+ 2 kante
+ 3 noyer
+ 4 modrici
+ 4 CN
+ 3 RA
+ 2 USA
+ 1 UK
```

### 21. 文件2中独有的内容 (行 364-375)

**Git历史版本内容:**
```
+ ((SELECT * FROM t1 ORDER BY id DESC)UNION(SELECT * FROM t2)UNION ALL(SELECT * FROM t3 ORDER BY area))ORDER BY id;
+ id name
+ 1 ronaldo
+ 1 UK
+ 1 EU
+ 2 kante
+ 2 USA
+ 2 NA
+ 3 noyer
+ 3 RA
+ 3 AU
+ 4 modrici
```

### 22. 文件1中独有的内容 (行 354-371)

**当前版本内容:**
```
- 3 RA
- 2 USA
- 1 UK
- 1 ronaldo
- 2 kante
- 3 noyer
- 4 modrici
- ((SELECT * FROM t1 ORDER BY id DESC)UNION(SELECT * FROM t2)UNION ALL(SELECT * FROM t3 ORDER BY area))ORDER BY id;
- id name
- 1 EU
- 1 UK
- 1 ronaldo
- 2 NA
- 2 USA
- 2 kante
- 3 AU
- 3 RA
- 3 noyer
```

### 23. 文件1中独有的内容 (行 373-374)

**当前版本内容:**
```
- 4 CN
- 4 modrici
```

### 24. 文件1中独有的内容 (行 377-380)

**当前版本内容:**
```
- 4 AS
- 3 AU
- 1 EU
- 2 NA
```

### 25. 文件2中独有的内容 (行 388-391)

**Git历史版本内容:**
```
+ 4 AS
+ 3 AU
+ 1 EU
+ 2 NA
```

### 26. 文件2中独有的内容 (行 396-399)

**Git历史版本内容:**
```
+ 1 ronaldo 1 UK
+ 2 kante 2 USA
+ 3 noyer 3 RA
+ 4 modrici 4 CN
```

### 27. 文件1中独有的内容 (行 397-400)

**当前版本内容:**
```
- 1 ronaldo 1 UK
- 2 kante 2 USA
- 3 noyer 3 RA
- 4 modrici 4 CN
```

### 28. 文件2中独有的内容 (行 408-411)

**Git历史版本内容:**
```
+ 4 modrici 4 AS
+ 3 noyer 3 AU
+ 2 kante 2 NA
+ 1 ronaldo 1 EU
```

### 29. 文件1中独有的内容 (行 409-412)

**当前版本内容:**
```
- 4 modrici 4 AS
- 3 noyer 3 AU
- 2 kante 2 NA
- 1 ronaldo 1 EU
```

### 30. 内容被修改 (文件1行 432-432, 文件2行 435-435)

**当前版本内容:**
```
- DATE(d1)MAX(salary)
```

**Git历史版本内容:**
```
+ date(d1)max(salary)
```

### 31. 内容被修改 (文件1行 437-437, 文件2行 440-440)

**当前版本内容:**
```
- DATE(d1)MAX(salary)
```

**Git历史版本内容:**
```
+ date(d1)max(salary)
```

### 32. 内容被修改 (文件1行 442-442, 文件2行 445-445)

**当前版本内容:**
```
- DATE(d1)MAX(salary)
```

**Git历史版本内容:**
```
+ date(d1)max(salary)
```

---

## 13. select.result

**文件路径:** select/select.result

**统计信息:**
- 当前版本行数: 546
- Git版本行数: 554

**内容差异 (10 个):**

### 1. 文件2中独有的内容 (行 56-56)

**Git历史版本内容:**
```
+ userID count(score)
```

### 2. 文件2中独有的内容 (行 120-120)

**Git历史版本内容:**
```
+ sum
```

### 3. 内容被修改 (文件1行 175-176, 文件2行 177-178)

**当前版本内容:**
```
- 1 1 1 1 1
- 2 2 2 2 2
```

**Git历史版本内容:**
```
+ 1.0 1.0 1.0 1 1
+ 2.0 2.0 2.0 2 2
```

### 4. 内容被修改 (文件1行 322-323, 文件2行 324-325)

**当前版本内容:**
```
- 1 null
- 1 a
```

**Git历史版本内容:**
```
+ 1.0 null
+ 1.0 a
```

### 5. 文件2中独有的内容 (行 379-379)

**Git历史版本内容:**
```
+ a
```

### 6. 文件2中独有的内容 (行 386-386)

**Git历史版本内容:**
```
+ a
```

### 7. 文件2中独有的内容 (行 393-393)

**Git历史版本内容:**
```
+ ;
```

### 8. 文件2中独有的内容 (行 397-397)

**Git历史版本内容:**
```
+ a b
```

### 9. 文件2中独有的内容 (行 537-537)

**Git历史版本内容:**
```
+ a
```

### 10. 文件2中独有的内容 (行 542-542)

**Git历史版本内容:**
```
+ a
```

---

## 14. sp_table.result

**文件路径:** select/sp_table.result

**统计信息:**
- 当前版本行数: 45
- Git版本行数: 45

**内容差异 (4 个):**

### 1. 文件2中独有的内容 (行 7-7)

**Git历史版本内容:**
```
+ mo_iscp_log r
```

### 2. 文件1中独有的内容 (行 17-17)

**当前版本内容:**
```
- mo_iscp_log r
```

### 3. 文件1中独有的内容 (行 23-23)

**当前版本内容:**
```
- mo_pitr r
```

### 4. 文件2中独有的内容 (行 45-45)

**Git历史版本内容:**
```
+ mo_pitr r
```

---

## 15. subquery.result

**文件路径:** select/subquery.result

**统计信息:**
- 当前版本行数: 276
- Git版本行数: 279

**内容差异 (9 个):**

### 1. 内容被修改 (文件1行 25-29, 文件2行 25-29)

**当前版本内容:**
```
- 5 1 2 6 51.26 5126 51 byebye is subquery? 2022-04-28 2022-04-28 22:40:11
- 6 3 2 1 632.1 6321 632 good night maybe subquery 2022-04-28 2022-04-28 22:40:11
- 7 4 4 3 7443.11 744311 7443 yes subquery 2022-04-28 2022-04-28 22:40:11
- 8 7 5 8 8758 875800 8758 nice to meet just subquery 2022-04-28 2022-04-28 22:40:11
- 9 8 4 9 9849.312 9849312 9849 see you subquery 2022-04-28 2022-04-28 22:40:11
```

**Git历史版本内容:**
```
+ 5 1 2 6 51.26 5126.0 51 byebye is subquery? 2022-04-28 2022-04-28 22:40:11
+ 6 3 2 1 632.1 6321.0 632 good night maybe subquery 2022-04-28 2022-04-28 22:40:11
+ 7 4 4 3 7443.11 744311.0 7443 yes subquery 2022-04-28 2022-04-28 22:40:11
+ 8 7 5 8 8758.0 875800.0 8758 nice to meet just subquery 2022-04-28 2022-04-28 22:40:11
+ 9 8 4 9 9849.31 9849312.0 9849 see you subquery 2022-04-28 2022-04-28 22:40:11
```

### 2. 内容被修改 (文件1行 39-40, 文件2行 39-40)

**当前版本内容:**
```
- 7 8758
- 8 9849.312
```

**Git历史版本内容:**
```
+ 7 8758.0
+ 8 9849.31
```

### 3. 内容被修改 (文件1行 50-51, 文件2行 50-51)

**当前版本内容:**
```
- 7 8758
- 8 9849.312
```

**Git历史版本内容:**
```
+ 7 8758.0
+ 8 9849.31
```

### 4. 内容被修改 (文件1行 116-118, 文件2行 116-118)

**当前版本内容:**
```
- 2 2 5 2 2252.05 225205 2252 bye sub query 2022-04-28 2022-04-28 22:40:11
- 3 6 6 3 3663.21 366321 3663 hi subquery 2022-04-28 2022-04-28 22:40:11
- 4 7 1 5 4715.22 471522 4715 good morning my subquery 2022-04-28 2022-04-28 22:40:11
```

**Git历史版本内容:**
```
+ 2 2 5 2 2252.05 225205.0 2252 bye sub query 2022-04-28 2022-04-28 22:40:11
+ 3 6 6 3 3663.21 366321.0 3663 hi subquery 2022-04-28 2022-04-28 22:40:11
+ 4 7 1 5 4715.22 471522.0 4715 good morning my subquery 2022-04-28 2022-04-28 22:40:11
```

### 5. 内容被修改 (文件1行 123-123, 文件2行 123-123)

**当前版本内容:**
```
- max(ti)+ 10 min(si)- 1 avg(fl)
```

**Git历史版本内容:**
```
+ max(ti)+10 min(si)-1 avg(fl)
```

### 6. 内容被修改 (文件1行 140-140, 文件2行 140-140)

**当前版本内容:**
```
- 2 2 5 2 2252.05 225205 2252 bye sub query 2022-04-28 2022-04-28 22:40:11
```

**Git历史版本内容:**
```
+ 2 2 5 2 2252.05 225205.0 2252 bye sub query 2022-04-28 2022-04-28 22:40:11
```

### 7. 文件2中独有的内容 (行 148-148)

**Git历史版本内容:**
```
+ id ti si bi fl dl de ch vch dd dt
```

### 8. 文件2中独有的内容 (行 229-229)

**Git历史版本内容:**
```
+ c1 c2
```

### 9. 文件2中独有的内容 (行 231-231)

**Git历史版本内容:**
```
+ c1 c2
```

---

## 16. union_and_union_all.result

**文件路径:** select/union_and_union_all.result

**统计信息:**
- 当前版本行数: 649
- Git版本行数: 652

**内容差异 (63 个):**

### 1. 内容被修改 (文件1行 46-53, 文件2行 46-53)

**当前版本内容:**
```
- null
- 2022-01-01
- 2022-01-01
- 2022-01-01
- 30
- 20
- 10
- null
```

**Git历史版本内容:**
```
+ 30
+ 20
+ 10
+ null
+ null
+ 2022-01-01
+ 2022-01-01
+ 2022-01-01
```

### 2. 内容被修改 (文件1行 56-63, 文件2行 56-63)

**当前版本内容:**
```
- null
- 2022-01-01
- 2022-01-01
- 2022-01-01
- null
- 10
- 20
- 30
```

**Git历史版本内容:**
```
+ 30
+ 20
+ 10
+ null
+ null
+ 2022-01-01
+ 2022-01-01
+ 2022-01-01
```

### 3. 内容被修改 (文件1行 66-73, 文件2行 66-73)

**当前版本内容:**
```
- null
- 2022-01-01
- 2022-01-01
- 2022-01-01
- 30
- 20
- 10
- null
```

**Git历史版本内容:**
```
+ 30
+ 20
+ 10
+ null
+ null
+ 2022-01-01
+ 2022-01-01
+ 2022-01-01
```

### 4. 内容被修改 (文件1行 76-79, 文件2行 76-79)

**当前版本内容:**
```
- null
- 10
- 20
- 30
```

**Git历史版本内容:**
```
+ 30
+ 20
+ 10
+ null
```

### 5. 文件2中独有的内容 (行 97-99)

**Git历史版本内容:**
```
+ 30
+ 20
+ 10
```

### 6. 文件1中独有的内容 (行 99-101)

**当前版本内容:**
```
- 30
- 20
- 10
```

### 7. 内容被修改 (文件1行 104-107, 文件2行 104-107)

**当前版本内容:**
```
- null
- 10
- 20
- 30
```

**Git历史版本内容:**
```
+ 30
+ 20
+ 10
+ null
```

### 8. 文件2中独有的内容 (行 111-113)

**Git历史版本内容:**
```
+ 30
+ 20
+ 10
```

### 9. 文件1中独有的内容 (行 113-115)

**当前版本内容:**
```
- 30
- 20
- 10
```

### 10. 文件2中独有的内容 (行 118-120)

**Git历史版本内容:**
```
+ 30
+ 20
+ 10
```

### 11. 文件1中独有的内容 (行 120-122)

**当前版本内容:**
```
- 10
- 20
- 30
```

### 12. 文件2中独有的内容 (行 156-160)

**Git历史版本内容:**
```
+ 20
+ 10
+ 30
+ -10
+ 65535
```

### 13. 文件2中独有的内容 (行 162-162)

**Git历史版本内容:**
```
+ 100
```

### 14. 文件1中独有的内容 (行 158-163)

**当前版本内容:**
```
- 100
- 65535
- 20
- 10
- 30
- -10
```

### 15. 内容被修改 (文件1行 176-178, 文件2行 176-182)

**当前版本内容:**
```
- 127
- 1
- 0
```

**Git历史版本内容:**
```
+ 20.0
+ 10.0
+ 30.0
+ -10.0
+ 127.0
+ 1.0
+ 0.0
```

### 16. 文件1中独有的内容 (行 180-183)

**当前版本内容:**
```
- 20
- 10
- 30
- -10
```

### 17. 文件2中独有的内容 (行 348-349)

**Git历史版本内容:**
```
+ 1
+ 2
```

### 18. 文件2中独有的内容 (行 352-354)

**Git历史版本内容:**
```
+ select * from t8 union distinct select * from t9 union distinct select * from t10;
+ a
+ 1
```

### 19. 文件1中独有的内容 (行 351-353)

**当前版本内容:**
```
- 1
- select * from t8 union distinct select * from t9 union distinct select * from t10;
- a
```

### 20. 文件2中独有的内容 (行 357-358)

**Git历史版本内容:**
```
+ select * from(select * from t8 union distinct select * from t9 union all select * from t10)X;
+ a
```

### 21. 文件1中独有的内容 (行 357-358)

**当前版本内容:**
```
- select * from(select * from t8 union distinct select * from t9 union all select * from t10)X;
- a
```

### 22. 文件2中独有的内容 (行 363-368)

**Git历史版本内容:**
```
+ select * from t8 union select * from t9 intersect select * from t10;
+ a
+ 1
+ select * from t8 union select * from t9 minus select * from t10;
+ a
+ 1
```

### 23. 内容被修改 (文件1行 362-367, 文件2行 370-374)

**当前版本内容:**
```
- 1
- select * from t8 union select * from t9 intersect select * from t10;
- a
- 1
- select * from t8 union select * from t9 minus select * from t10;
- a
```

**Git历史版本内容:**
```
+ (select * from t8 union select * from t9)intersect select * from t10;
+ a
+ (select * from t8 union select * from t9)minus select * from t10;
+ a
+ 1
```

### 24. 文件1中独有的内容 (行 369-374)

**当前版本内容:**
```
- 1
- (select * from t8 union select * from t9)intersect select * from t10;
- (select * from t8 union select * from t9)minus select * from t10;
- a
- 2
- 1
```

### 25. 文件2中独有的内容 (行 382-382)

**Git历史版本内容:**
```
+ case+union+test
```

### 26. 文件1中独有的内容 (行 382-382)

**当前版本内容:**
```
- case+union+test
```

### 27. 文件2中独有的内容 (行 387-387)

**Git历史版本内容:**
```
+ case+union+tet
```

### 28. 文件1中独有的内容 (行 387-387)

**当前版本内容:**
```
- case+union+tet
```

### 29. 文件2中独有的内容 (行 393-393)

**Git历史版本内容:**
```
+ a
```

### 30. 文件1中独有的内容 (行 393-393)

**当前版本内容:**
```
- a
```

### 31. 文件2中独有的内容 (行 397-397)

**Git历史版本内容:**
```
+ a
```

### 32. 文件1中独有的内容 (行 397-397)

**当前版本内容:**
```
- a
```

### 33. 文件2中独有的内容 (行 405-405)

**Git历史版本内容:**
```
+ a
```

### 34. 文件1中独有的内容 (行 405-405)

**当前版本内容:**
```
- a
```

### 35. 文件2中独有的内容 (行 409-409)

**Git历史版本内容:**
```
+ a
```

### 36. 文件1中独有的内容 (行 409-409)

**当前版本内容:**
```
- a
```

### 37. 文件2中独有的内容 (行 413-413)

**Git历史版本内容:**
```
+ a
```

### 38. 文件1中独有的内容 (行 413-413)

**当前版本内容:**
```
- a
```

### 39. 文件2中独有的内容 (行 417-417)

**Git历史版本内容:**
```
+ a
```

### 40. 文件1中独有的内容 (行 417-417)

**当前版本内容:**
```
- a
```

### 41. 文件2中独有的内容 (行 425-425)

**Git历史版本内容:**
```
+ a
```

### 42. 文件1中独有的内容 (行 425-425)

**当前版本内容:**
```
- a
```

### 43. 内容被修改 (文件1行 429-429, 文件2行 430-430)

**当前版本内容:**
```
- concat((select x from(select a as x)as t1),(select y from(select b as y)as t2))
```

**Git历史版本内容:**
```
+ concat((select x from(select 'a' as x)as t1),(select y from(select 'b' as y)as t2))
```

### 44. 文件2中独有的内容 (行 438-438)

**Git历史版本内容:**
```
+ 1234562
```

### 45. 文件1中独有的内容 (行 438-438)

**当前版本内容:**
```
- 1234562
```

### 46. 文件2中独有的内容 (行 446-446)

**Git历史版本内容:**
```
+ Mic-5
```

### 47. 文件1中独有的内容 (行 446-446)

**当前版本内容:**
```
- Mic-5
```

### 48. 文件2中独有的内容 (行 454-454)

**Git历史版本内容:**
```
+ Mic-5
```

### 49. 文件1中独有的内容 (行 454-454)

**当前版本内容:**
```
- Mic-5
```

### 50. 文件2中独有的内容 (行 461-461)

**Git历史版本内容:**
```
+ a
```

### 51. 文件2中独有的内容 (行 467-467)

**Git历史版本内容:**
```
+ a
```

### 52. 文件2中独有的内容 (行 493-493)

**Git历史版本内容:**
```
+ dekad
```

### 53. 文件1中独有的内容 (行 491-491)

**当前版本内容:**
```
- dekad
```

### 54. 文件2中独有的内容 (行 497-497)

**Git历史版本内容:**
```
+ joce
```

### 55. 文件1中独有的内容 (行 496-496)

**当前版本内容:**
```
- joce
```

### 56. 文件1中独有的内容 (行 504-504)

**当前版本内容:**
```
- dekad
```

### 57. 文件2中独有的内容 (行 509-509)

**Git历史版本内容:**
```
+ dekad
```

### 58. 内容被修改 (文件1行 534-537, 文件2行 537-540)

**当前版本内容:**
```
- 1 1 foo1 bar1
- 1 2 foo2 bar2
- 1 3 null bar3
- 1 4 foo4 bar4
```

**Git历史版本内容:**
```
+ 1 4 foo4 bar4
+ 1 3 null bar3
+ 1 2 foo2 bar2
+ 1 1 foo1 bar1
```

### 59. 内容被修改 (文件1行 543-546, 文件2行 546-549)

**当前版本内容:**
```
- 1 1 foo1 bar1
- 1 2 foo2 bar2
- 1 3 null bar3
- 1 4 foo4 bar4
```

**Git历史版本内容:**
```
+ 1 4 foo4 bar4
+ 1 3 null bar3
+ 1 2 foo2 bar2
+ 1 1 foo1 bar1
```

### 60. 内容被修改 (文件1行 552-555, 文件2行 555-558)

**当前版本内容:**
```
- 1 1 foo1 bar1
- 1 2 foo2 bar2
- 1 3 null bar3
- 1 4 foo4 bar4
```

**Git历史版本内容:**
```
+ 1 4 foo4 bar4
+ 1 3 null bar3
+ 1 2 foo2 bar2
+ 1 1 foo1 bar1
```

### 61. 内容被修改 (文件1行 579-582, 文件2行 582-585)

**当前版本内容:**
```
- 1 1 foo1 bar1
- 1 2 foo2 bar2
- 1 3 null bar3
- 1 4 foo4 bar4
```

**Git历史版本内容:**
```
+ 1 4 foo4 bar4
+ 1 3 null bar3
+ 1 2 foo2 bar2
+ 1 1 foo1 bar1
```

### 62. 文件2中独有的内容 (行 641-643)

**Git历史版本内容:**
```
+ a 111 null null
+ b 222 null null
+ d 444 d 454
```

### 63. 文件1中独有的内容 (行 641-643)

**当前版本内容:**
```
- a 111 null null
- b 222 null null
- d 444 d 454
```

---

## 17. IndexMetadata.result

**文件路径:** show/IndexMetadata.result

**统计信息:**
- 当前版本行数: 179
- Git版本行数: 180

**内容差异 (2 个):**

### 1. 内容被修改 (文件1行 36-39, 文件2行 36-39)

**当前版本内容:**
```
- 1 Abby 24 zbcvdf
- 2 Bob 25 zbcvdf
- 3 Carol 23 zbcvdf
- 4 Dora 29 zbcvdf
```

**Git历史版本内容:**
```
+ 1 Abby 24.0 zbcvdf
+ 2 Bob 25.0 zbcvdf
+ 3 Carol 23.0 zbcvdf
+ 4 Dora 29.0 zbcvdf
```

### 2. 内容被修改 (文件1行 83-86, 文件2行 83-86)

**当前版本内容:**
```
- 1 Abby 24 zbcvdf
- 2 Bob 25 zbcvdf
- 3 Carol 23 zbcvdf
- 4 Dora 29 zbcvdf
```

**Git历史版本内容:**
```
+ 1 Abby 24.0 zbcvdf
+ 2 Bob 25.0 zbcvdf
+ 3 Carol 23.0 zbcvdf
+ 4 Dora 29.0 zbcvdf
```

---

## 18. cdc.result

**文件路径:** show/cdc.result

**统计信息:**
- 当前版本行数: 13
- Git版本行数: 12

**内容差异 (1 个):**

### 1. 内容被修改 (文件1行 8-9, 文件2行 8-8)

**当前版本内容:**
```
- 0199c83c-07a8-7a87-807f-31751b7716a4 cdc_tpcc 7b225f223a22222c2275736572223a22746573745f6364635f7661725f616363312361646d696e222c226970223a223132372e302e302e31222c22706f7274223a363030312c227265736572766564223a22227d 7b225f223a22222c2275736572223a22746573745f6364635f7661725f616363312361646d696e222c226970223a223132372e302e302e31222c22706f7274223a363030312c227265736572766564223a22227d running {
- } 2025-10-09 17:10:01.611531745 +0800 CST
```

**Git历史版本内容:**
```
+ null cdc_tpcc null null null null null null
```

---

## 19. database_statistics.result

**文件路径:** show/database_statistics.result

**统计信息:**
- 当前版本行数: 532
- Git版本行数: 578

**内容差异 (7 个):**

### 1. 文件2中独有的内容 (行 110-112)

**Git历史版本内容:**
```
+ show column_number from mo_table_partitions;
+ Number of columns in mo_table_partitions
+ 10
```

### 2. 内容被修改 (文件1行 168-168, 文件2行 171-171)

**当前版本内容:**
```
- 30
```

**Git历史版本内容:**
```
+ 28
```

### 3. 文件2中独有的内容 (行 275-312)

**Git历史版本内容:**
```
+ DROP TABLE IF EXISTS partition_table;
+ create table partition_table(
+ empno int unsigned auto_increment,
+ ename varchar(15),
+ job varchar(10),
+ mgr int unsigned,
+ hiredate date,
+ sal decimal(7,2),
+ comm decimal(7,2),
+ deptno int unsigned,
+ primary key(empno,deptno)
+ )
+ PARTITION BY KEY(deptno)
+ PARTITIONS 4;
+ show table_number from test_db;
+ Number of tables in test_db
+ 8
+ show table_values from partition_table;
+ max(empno)min(empno)max(ename)min(ename)max(job)min(job)max(mgr)min(mgr)max(hiredate)min(hiredate)max(sal)min(sal)max(comm)min(comm)max(deptno)min(deptno)
+ null null null null null null null null null null null null null null null null
+ select mo_table_rows("test_db","partition_table"),mo_table_size("test_db","partition_table");
+ mo_table_rows(test_db,partition_table)mo_table_size(test_db,partition_table)
+ 0 0
+ INSERT INTO partition_table VALUES(7369,'SMITH','CLERK',7902,'1980-12-17',800,NULL,20);
+ INSERT INTO partition_table VALUES(7499,'ALLEN','SALESMAN',7698,'1981-02-20',1600,300,30);
+ show table_values from partition_table;
+ max(empno)min(empno)max(ename)min(ename)max(job)min(job)max(mgr)min(mgr)max(hiredate)min(hiredate)max(sal)min(sal)max(comm)min(comm)max(deptno)min(deptno)
+ 7499 7369 SMITH ALLEN SALESMAN CLERK 7902 7698 1981-02-20 1980-12-17 1600.00 800.00 300.00 300.00 30 20
+ INSERT INTO partition_table VALUES(7521,'WARD','SALESMAN',7698,'1981-02-22',1250,500,30);
+ INSERT INTO partition_table VALUES(7566,'JONES','MANAGER',7839,'1981-04-02',2975,NULL,20);
+ show table_values from partition_table;
+ max(empno)min(empno)max(ename)min(ename)max(job)min(job)max(mgr)min(mgr)max(hiredate)min(hiredate)max(sal)min(sal)max(comm)min(comm)max(deptno)min(deptno)
+ 7566 7369 WARD ALLEN SALESMAN CLERK 7902 7698 1981-04-02 1980-12-17 2975.00 800.00 500.00 300.00 30 20
+ set mo_table_stats.use_old_impl = yes;
+ select mo_table_rows("test_db","partition_table"),mo_table_size("test_db","partition_table");
+ mo_table_rows(test_db,partition_table)mo_table_size(test_db,partition_table)
+ 4 512
+ set mo_table_stats.use_old_impl = no;
```

### 4. 文件2中独有的内容 (行 317-317)

**Git历史版本内容:**
```
+ max(col1)min(col1)
```

### 5. 文件2中独有的内容 (行 403-405)

**Git历史版本内容:**
```
+ show column_number from partitions;
+ Number of columns in partitions
+ 25
```

### 6. 内容被修改 (文件1行 408-408, 文件2行 453-453)

**当前版本内容:**
```
- 30
```

**Git历史版本内容:**
```
+ 28
```

### 7. 文件2中独有的内容 (行 546-546)

**Git历史版本内容:**
```
+ max(col1)min(col1)
```

---

## 20. show.result

**文件路径:** show/show.result

**统计信息:**
- 当前版本行数: 493
- Git版本行数: 504

**内容差异 (13 个):**

### 1. 文件2中独有的内容 (行 41-41)

**Git历史版本内容:**
```
+ Charset Description Default collation Maxlen
```

### 2. 文件2中独有的内容 (行 55-56)

**Git历史版本内容:**
```
+ show table status from db11111111111;
+ Name Engine Row_format Rows Avg_row_length Data_length Max_data_length Index_length Data_free Auto_increment Create_time Update_time Check_time Collation Checksum Create_options Comment
```

### 3. 文件2中独有的内容 (行 139-139)

**Git历史版本内容:**
```
+ Grants for test@localhost
```

### 4. 文件2中独有的内容 (行 185-185)

**Git历史版本内容:**
```
+ Variable_name Value
```

### 5. 文件2中独有的内容 (行 245-245)

**Git历史版本内容:**
```
+ Trigger Event Table Statement Timing Created sql_mode Definer character_set_client collation_connection Database Collation
```

### 6. 文件2中独有的内容 (行 247-247)

**Git历史版本内容:**
```
+ Trigger Event Table Statement Timing Created sql_mode Definer character_set_client collation_connection Database Collation
```

### 7. 文件2中独有的内容 (行 252-252)

**Git历史版本内容:**
```
+ mo_iscp_log
```

### 8. 文件1中独有的内容 (行 254-254)

**当前版本内容:**
```
- mo_iscp_log
```

### 9. 文件2中独有的内容 (行 308-308)

**Git历史版本内容:**
```
+ 1
```

### 10. 文件2中独有的内容 (行 310-310)

**Git历史版本内容:**
```
+ 1
```

### 11. 文件2中独有的内容 (行 348-348)

**Git历史版本内容:**
```
+ Tables_in_.quote
```

### 12. 内容被修改 (文件1行 354-354, 文件2行 364-364)

**当前版本内容:**
```
- va create view va as select a from a utf8mb4 utf8mb4_general_ci
```

**Git历史版本内容:**
```
+ va create view va as select a from a;utf8mb4 utf8mb4_general_ci
```

### 13. 内容被修改 (文件1行 357-357, 文件2行 367-367)

**当前版本内容:**
```
- va create view va as select a from a utf8mb4 utf8mb4_general_ci
```

**Git历史版本内容:**
```
+ va create view va as select a from a;utf8mb4 utf8mb4_general_ci
```

---

## 21. show4.result

**文件路径:** show/show4.result

**统计信息:**
- 当前版本行数: 49
- Git版本行数: 49

**内容差异 (1 个):**

### 1. 内容被修改 (文件1行 43-48, 文件2行 43-48)

**当前版本内容:**
```
- error_info null null null null null null null null null 2025-10-09 08:42:58 null null null null null VIEW 0 moadmin
- log_info null null null null null null null null null 2025-10-09 08:42:58 null null null null null VIEW 0 moadmin
- rawlog Tae Dynamic 17666 0 1212737 0 0 NULL 0 2025-10-09 08:42:58 NULL NULL utf8mb4_bin NULL read merge data from log,error,span[mo_no_del_hint] 0 moadmin
- span_info null null null null null null null null null 2025-10-09 08:42:58 null null null null null VIEW 0 moadmin
- sql_statement_hotspot null null null null null null null null null 2025-10-09 08:42:58 null null null null null VIEW 0 moadmin
- statement_info Tae Dynamic 7598 0 1075486 0 0 NULL 0 2025-10-09 08:42:58 NULL NULL utf8mb4_bin NULL record each statement and stats info[mo_no_del_hint] 0 moadmin
```

**Git历史版本内容:**
```
+ error_info null null null null null null null null null 2024-08-29 16:04:58 null null null null null VIEW 0 moadmin
+ log_info null null null null null null null null null 2024-08-29 16:04:58 null null null null null VIEW 0 moadmin
+ rawlog Tae Dynamic 1577 0 335906 0 0 NULL 0 2024-08-29 16:04:58 NULL NULL utf8mb4_bin NULL read merge data from log,error,span[mo_no_del_hint] 0 moadmin
+ span_info null null null null null null null null null 2024-08-29 16:04:58 null null null null null VIEW 0 moadmin
+ sql_statement_hotspot null null null null null null null null null 2024-08-29 16:04:58 null null null null null VIEW 0 moadmin
+ statement_info Tae Dynamic 116 0 102892 0 0 NULL 0 2024-08-29 16:04:58 NULL NULL utf8mb4_bin NULL record each statement and stats info[mo_no_del_hint] 0 moadmin
```

---

## 22. update.result

**文件路径:** update/update.result

**统计信息:**
- 当前版本行数: 299
- Git版本行数: 329

**内容差异 (2 个):**

### 1. 内容被修改 (文件1行 268-268, 文件2行 268-297)

**当前版本内容:**
```
- Duplicate entry '(2,2)' for key '(a,b)'
```

**Git历史版本内容:**
```
+ Duplicate entry('\(\d\,\d\)'|'\d\w\d{5}\w\d{4}')for key '(\(a,b\)|__mo_cpkey_col)'
+ drop table if exists t1;
+ create table t1(a int,b varchar(20),unique key(a));
+ insert into t1 values(1,'1');
+ insert into t1 values(2,'2');
+ insert into t1 values(3,'3');
+ insert into t1 values(4,'4');
+ select * from t1;
+ a b
+ 1 1
+ 2 2
+ 3 3
+ 4 4
+ update t1 set a = 2 where a = 1;
+ tae data: duplicate
+ drop table if exists t1;
+ create table t1(a int,b varchar(20),unique key(a,b));
+ insert into t1 values(1,'2');
+ insert into t1 values(1,'3');
+ insert into t1 values(2,'2');
+ insert into t1 values(2,'3');
+ select * from t1;
+ a b
+ 1 2
+ 1 3
+ 2 2
+ 2 3
+ update t1 set a = 2 where a = 1;
+ tae data: duplicate
+ update t1 set a = null where a = 1;
```

### 2. 文件2中独有的内容 (行 327-327)

**Git历史版本内容:**
```
+ a b
```

---

## 23. update_index.result

**文件路径:** update/update_index.result

**统计信息:**
- 当前版本行数: 241
- Git版本行数: 241

**内容差异 (9 个):**

### 1. 内容被修改 (文件1行 16-16, 文件2行 16-16)

**当前版本内容:**
```
- Duplicate entry '7' for key 'a'
```

**Git历史版本内容:**
```
+ Duplicate entry '7' for key '(.*)'
```

### 2. 内容被修改 (文件1行 41-41, 文件2行 41-41)

**当前版本内容:**
```
- Duplicate entry '(7,7)' for key '(a,b)'
```

**Git历史版本内容:**
```
+ Duplicate entry('\(\d\,\d\)'|'\d\w\d{5}\w\d{4}')for key '(.*)'
```

### 3. 内容被修改 (文件1行 80-80, 文件2行 80-80)

**当前版本内容:**
```
- Duplicate entry '(7,7)' for key '(a,b)'
```

**Git历史版本内容:**
```
+ Duplicate entry('\(\d\,\d\)'|'\d\w\d{5}\w\d{4}')for key '(.*)'
```

### 4. 内容被修改 (文件1行 82-82, 文件2行 82-82)

**当前版本内容:**
```
- Duplicate entry '7' for key 'a'
```

**Git历史版本内容:**
```
+ Duplicate entry '7' for key '(.*)'
```

### 5. 内容被修改 (文件1行 91-91, 文件2行 91-91)

**当前版本内容:**
```
- Duplicate entry '8' for key 'a'
```

**Git历史版本内容:**
```
+ Duplicate entry '8' for key '(.*)'
```

### 6. 内容被修改 (文件1行 126-126, 文件2行 126-126)

**当前版本内容:**
```
- Duplicate entry '7' for key 'a'
```

**Git历史版本内容:**
```
+ Duplicate entry '7' for key '(.*)'
```

### 7. 内容被修改 (文件1行 149-149, 文件2行 149-149)

**当前版本内容:**
```
- Duplicate entry '(7,7)' for key '(a,b)'
```

**Git历史版本内容:**
```
+ Duplicate entry('\(\d\,\d\)'|'\d\w\d{5}\w\d{4}')for key '(.*)'
```

### 8. 内容被修改 (文件1行 179-179, 文件2行 179-179)

**当前版本内容:**
```
- Duplicate entry '7' for key 'a'
```

**Git历史版本内容:**
```
+ Duplicate entry '7' for key '(.*)'
```

### 9. 内容被修改 (文件1行 204-204, 文件2行 204-204)

**当前版本内容:**
```
- Duplicate entry '7' for key 'a'
```

**Git历史版本内容:**
```
+ Duplicate entry '7' for key '(.*)'
```

---

## 24. update_multiple_table.result

**文件路径:** update/update_multiple_table.result

**统计信息:**
- 当前版本行数: 315
- Git版本行数: 315

**内容差异 (47 个):**

### 1. 文件2中独有的内容 (行 23-24)

**Git历史版本内容:**
```
+ 2 3
+ 2 4
```

### 2. 文件2中独有的内容 (行 26-27)

**Git历史版本内容:**
```
+ 3 3
+ 3 4
```

### 3. 文件2中独有的内容 (行 29-30)

**Git历史版本内容:**
```
+ 4 3
+ 4 4
```

### 4. 文件1中独有的内容 (行 26-28)

**当前版本内容:**
```
- 2 3
- 3 3
- 4 3
```

### 5. 文件1中独有的内容 (行 30-32)

**当前版本内容:**
```
- 2 4
- 3 4
- 4 4
```

### 6. 内容被修改 (文件1行 37-40, 文件2行 37-40)

**当前版本内容:**
```
- 3 1002 MySQL教程 80
- 4 1003 Python教程 120
- 5 1004 C语言教程 150
- 2 1001 PHP 100
```

**Git历史版本内容:**
```
+ 3 1002 MySQL教程 80.0
+ 4 1003 Python教程 120.0
+ 5 1004 C语言教程 150.0
+ 2 1001 PHP 100.0
```

### 7. 内容被修改 (文件1行 43-45, 文件2行 43-45)

**当前版本内容:**
```
- 2 1001 100
- 3 1002 100
- 4 1003 100
```

**Git历史版本内容:**
```
+ 2 1001 100.0
+ 3 1002 100.0
+ 4 1003 100.0
```

### 8. 内容被修改 (文件1行 49-60, 文件2行 49-60)

**当前版本内容:**
```
- 3 MySQL教程 64
- 4 Python教程 64
- 5 C语言教程 64
- 2 PHP 64
- 3 MySQL教程 96
- 4 Python教程 96
- 5 C语言教程 96
- 2 PHP 96
- 3 MySQL教程 80
- 4 Python教程 80
- 5 C语言教程 80
- 2 PHP 80
```

**Git历史版本内容:**
```
+ 3 MySQL教程 64.0
+ 3 MySQL教程 96.0
+ 3 MySQL教程 80.0
+ 4 Python教程 64.0
+ 4 Python教程 96.0
+ 4 Python教程 80.0
+ 5 C语言教程 64.0
+ 5 C语言教程 96.0
+ 5 C语言教程 80.0
+ 2 PHP 64.0
+ 2 PHP 96.0
+ 2 PHP 80.0
```

### 9. 内容被修改 (文件1行 64-75, 文件2行 64-75)

**当前版本内容:**
```
- MySQL教程 64
- Python教程 64
- C语言教程 64
- PHP 64
- MySQL教程 96
- Python教程 96
- C语言教程 96
- PHP 96
- MySQL教程 80
- Python教程 80
- C语言教程 80
- PHP 80
```

**Git历史版本内容:**
```
+ MySQL教程 64.0
+ MySQL教程 96.0
+ MySQL教程 80.0
+ Python教程 64.0
+ Python教程 96.0
+ Python教程 80.0
+ C语言教程 64.0
+ C语言教程 96.0
+ C语言教程 80.0
+ PHP 64.0
+ PHP 96.0
+ PHP 80.0
```

### 10. 文件2中独有的内容 (行 96-97)

**Git历史版本内容:**
```
+ Bob null 一班 test00
+ Bob null 二班 test00
```

### 11. 文件2中独有的内容 (行 99-100)

**Git历史版本内容:**
```
+ Ruby null 一班 test00
+ Ruby null 二班 test00
```

### 12. 文件2中独有的内容 (行 102-103)

**Git历史版本内容:**
```
+ 张三 test00 一班 test00
+ 张三 test00 二班 test00
```

### 13. 文件2中独有的内容 (行 105-106)

**Git历史版本内容:**
```
+ 李四 test00 一班 test00
+ 李四 test00 二班 test00
```

### 14. 文件2中独有的内容 (行 108-109)

**Git历史版本内容:**
```
+ 王五 test00 一班 test00
+ 王五 test00 二班 test00
```

### 15. 文件1中独有的内容 (行 101-105)

**当前版本内容:**
```
- Bob null 一班 test00
- Ruby null 一班 test00
- 张三 test00 一班 test00
- 李四 test00 一班 test00
- 王五 test00 一班 test00
```

### 16. 文件1中独有的内容 (行 107-111)

**当前版本内容:**
```
- Bob null 二班 test00
- Ruby null 二班 test00
- 张三 test00 二班 test00
- 李四 test00 二班 test00
- 王五 test00 二班 test00
```

### 17. 文件2中独有的内容 (行 117-118)

**Git历史版本内容:**
```
+ 5 Bob null test11
+ 5 Bob null test11
```

### 18. 文件2中独有的内容 (行 120-121)

**Git历史版本内容:**
```
+ 6 Ruby null test11
+ 6 Ruby null test11
```

### 19. 文件2中独有的内容 (行 123-124)

**Git历史版本内容:**
```
+ 1 张三 test11 test11
+ 1 张三 test11 test11
```

### 20. 文件2中独有的内容 (行 126-127)

**Git历史版本内容:**
```
+ 2 李四 test11 test11
+ 2 李四 test11 test11
```

### 21. 文件2中独有的内容 (行 129-130)

**Git历史版本内容:**
```
+ 3 王五 test11 test11
+ 3 王五 test11 test11
```

### 22. 文件1中独有的内容 (行 122-126)

**当前版本内容:**
```
- 5 Bob null test11
- 6 Ruby null test11
- 1 张三 test11 test11
- 2 李四 test11 test11
- 3 王五 test11 test11
```

### 23. 文件1中独有的内容 (行 128-132)

**当前版本内容:**
```
- 5 Bob null test11
- 6 Ruby null test11
- 1 张三 test11 test11
- 2 李四 test11 test11
- 3 王五 test11 test11
```

### 24. 文件2中独有的内容 (行 139-142)

**Git历史版本内容:**
```
+ Bob test22 test33
+ Bob test22 test33
+ Ruby test22 test33
+ Ruby test22 test33
```

### 25. 文件2中独有的内容 (行 145-148)

**Git历史版本内容:**
```
+ 张三 test33 test33
+ 张三 test33 test33
+ 李四 test33 test33
+ 李四 test33 test33
```

### 26. 内容被修改 (文件1行 143-147, 文件2行 151-151)

**当前版本内容:**
```
- Rose test33 test33
- Bob test22 test33
- Ruby test22 test33
- 张三 test33 test33
- 李四 test33 test33
```

**Git历史版本内容:**
```
+ 王五 test33 test33
```

### 27. 内容被修改 (文件1行 150-154, 文件2行 154-154)

**当前版本内容:**
```
- Bob test22 test33
- Ruby test22 test33
- 张三 test33 test33
- 李四 test33 test33
- 王五 test33 test33
```

**Git历史版本内容:**
```
+ Rose test33 test33
```

### 28. 文件2中独有的内容 (行 160-161)

**Git历史版本内容:**
```
+ Bob test22 张三
+ Bob test22 王五
```

### 29. 文件2中独有的内容 (行 163-164)

**Git历史版本内容:**
```
+ Ruby test22 张三
+ Ruby test22 王五
```

### 30. 文件2中独有的内容 (行 166-167)

**Git历史版本内容:**
```
+ 张三 一班 张三
+ 张三 一班 王五
```

### 31. 文件2中独有的内容 (行 169-170)

**Git历史版本内容:**
```
+ 李四 一班 张三
+ 李四 一班 王五
```

### 32. 文件2中独有的内容 (行 172-173)

**Git历史版本内容:**
```
+ 王五 二班 张三
+ 王五 二班 王五
```

### 33. 文件1中独有的内容 (行 165-169)

**当前版本内容:**
```
- Bob test22 张三
- Ruby test22 张三
- 张三 一班 张三
- 李四 一班 张三
- 王五 二班 张三
```

### 34. 文件1中独有的内容 (行 171-175)

**当前版本内容:**
```
- Bob test22 王五
- Ruby test22 王五
- 张三 一班 王五
- 李四 一班 王五
- 王五 二班 王五
```

### 35. 文件2中独有的内容 (行 212-212)

**Git历史版本内容:**
```
+ 2022-04-12 04:32:46 2022-08-16 20:23:06
```

### 36. 文件1中独有的内容 (行 213-213)

**当前版本内容:**
```
- 2022-04-12 04:32:46 2022-08-16 20:23:06
```

### 37. 文件2中独有的内容 (行 228-228)

**Git历史版本内容:**
```
+ 2 2088-01-01 00:00:00 9999-01-01 00:00:00
```

### 38. 文件1中独有的内容 (行 229-229)

**当前版本内容:**
```
- 2 2088-01-01 00:00:00 9999-01-01 00:00:00
```

### 39. 文件2中独有的内容 (行 254-254)

**Git历史版本内容:**
```
+ this is a test NULL
```

### 40. 文件1中独有的内容 (行 255-255)

**当前版本内容:**
```
- this is a test NULL
```

### 41. 文件2中独有的内容 (行 268-268)

**Git历史版本内容:**
```
+ 0 0
```

### 42. 文件1中独有的内容 (行 269-269)

**当前版本内容:**
```
- 0 0
```

### 43. 内容被修改 (文件1行 287-290, 文件2行 287-287)

**当前版本内容:**
```
- 11 13
- 0 13
- 3 13
- 3 13
```

**Git历史版本内容:**
```
+ 11 13.0
```

### 44. 文件2中独有的内容 (行 289-290)

**Git历史版本内容:**
```
+ 11 0.1411200080598672
+ 0 13.0
```

### 45. 文件2中独有的内容 (行 292-293)

**Git历史版本内容:**
```
+ 0 0.1411200080598672
+ 3 13.0
```

### 46. 内容被修改 (文件1行 295-296, 文件2行 296-296)

**当前版本内容:**
```
- 11 0.1411200080598672
- 0 0.1411200080598672
```

**Git历史版本内容:**
```
+ 3 13.0
```

### 47. 内容被修改 (文件1行 302-313, 文件2行 302-313)

**当前版本内容:**
```
- 11 1
- 0 1
- 3 1
- 0 1
- 11 1
- 0 1
- 3 1
- 0 1
- 11 1
- 0 1
- 3 1
- 0 1
```

**Git历史版本内容:**
```
+ 11 1.0
+ 11 1.0
+ 11 1.0
+ 0 1.0
+ 0 1.0
+ 0 1.0
+ 3 1.0
+ 3 1.0
+ 3 1.0
+ 0 1.0
+ 0 1.0
+ 0 1.0
```

---

## 25. update_workspace.result

**文件路径:** update/update_workspace.result

**统计信息:**
- 当前版本行数: 51
- Git版本行数: 51

**内容差异 (1 个):**

### 1. 内容被修改 (文件1行 48-48, 文件2行 48-48)

**当前版本内容:**
```
- max(a)= min(a)+ 81920 - 1
```

**Git历史版本内容:**
```
+ max(a)=min(a)+81920-1
```

---

