# 批量Git Diff报告

生成时间: 2025-09-26 14:13:15.576279
比较commit: HEAD~1
总文件数: 3

## 1. array.result

**文件路径:** array.result

**统计信息:**
- 当前版本行数: 568
- Git版本行数: 571

**内容差异 (18 个):**

### 1. 内容被修改 (文件1行 16-16, 文件2行 16-16)

**当前版本内容:**
```
- select * from vec_table /* save_result */;
```

**Git历史版本内容:**
```
+ /* save_result */ select * from vec_table;
```

### 2. 文件2中独有的内容 (行 36-36)

**Git历史版本内容:**
```
+ a b c
```

### 3. 文件2中独有的内容 (行 38-38)

**Git历史版本内容:**
```
+ a b c
```

### 4. 文件2中独有的内容 (行 46-46)

**Git历史版本内容:**
```
+ a b c
```

### 5. 内容被修改 (文件1行 80-80, 文件2行 83-83)

**当前版本内容:**
```
- 6
```

**Git历史版本内容:**
```
+ 6.0
```

### 6. 内容被修改 (文件1行 83-83, 文件2行 86-86)

**当前版本内容:**
```
- 6
```

**Git历史版本内容:**
```
+ 6.0
```

### 7. 内容被修改 (文件1行 92-92, 文件2行 95-95)

**当前版本内容:**
```
- -14
```

**Git历史版本内容:**
```
+ -14.0
```

### 8. 内容被修改 (文件1行 95-95, 文件2行 98-98)

**当前版本内容:**
```
- 1
```

**Git历史版本内容:**
```
+ 1.0
```

### 9. 内容被修改 (文件1行 98-98, 文件2行 101-101)

**当前版本内容:**
```
- 0
```

**Git历史版本内容:**
```
+ 0.0
```

### 10. 内容被修改 (文件1行 101-101, 文件2行 104-104)

**当前版本内容:**
```
- 0
```

**Git历史版本内容:**
```
+ 0.0
```

### 11. 内容被修改 (文件1行 123-123, 文件2行 126-126)

**当前版本内容:**
```
- data out of range: data type vecf32,typeLen is over the MaxVectorLen : 65535
```

**Git历史版本内容:**
```
+ Data truncation: data out of range: data type vecf32,typeLen is over the MaxVectorLen : 65535
```

### 12. 内容被修改 (文件1行 127-127, 文件2行 130-130)

**当前版本内容:**
```
- division by zero
```

**Git历史版本内容:**
```
+ Data truncation: division by zero
```

### 13. 内容被修改 (文件1行 201-201, 文件2行 204-204)

**当前版本内容:**
```
- select * from t7 /* save_result */;
```

**Git历史版本内容:**
```
+ /* save_result */ select * from t7;
```

### 14. 内容被修改 (文件1行 320-320, 文件2行 323-323)

**当前版本内容:**
```
- 1 null
```

**Git历史版本内容:**
```
+ 1.0 null
```

### 15. 内容被修改 (文件1行 322-322, 文件2行 325-325)

**当前版本内容:**
```
- 1 1
```

**Git历史版本内容:**
```
+ 1.0 1.0
```

### 16. 内容被修改 (文件1行 331-333, 文件2行 334-336)

**当前版本内容:**
```
- 1 null
- 1 1
- 1 1
```

**Git历史版本内容:**
```
+ 1.0 null
+ 1.0 1.0
+ 1.0 1.0
```

### 17. 内容被修改 (文件1行 527-527, 文件2行 530-530)

**当前版本内容:**
```
- division by zero
```

**Git历史版本内容:**
```
+ Data truncation: division by zero
```

### 18. 内容被修改 (文件1行 560-560, 文件2行 563-563)

**当前版本内容:**
```
- 5 25
```

**Git历史版本内容:**
```
+ 5.0 25.0
```

---

## 2. array_index.result

**文件路径:** array_index.result

**统计信息:**
- 当前版本行数: 717
- Git版本行数: 721

**内容差异 (14 个):**

### 1. 文件2中独有的内容 (行 187-187)

**Git历史版本内容:**
```
+ name type column_name algo algo_table_type algo_params
```

### 2. 文件2中独有的内容 (行 221-221)

**Git历史版本内容:**
```
+ name type column_name algo algo_table_type algo_params
```

### 3. 内容被修改 (文件1行 243-244, 文件2行 245-246)

**当前版本内容:**
```
- insert into tbl values(15,"[1,3,5]");
- insert into tbl values(18,"[130,40,90]");
```

**Git历史版本内容:**
```
+ insert into tbl values(15,"[1,3,5]");-- inserted to centroid 1 of version 0
+ insert into tbl values(18,"[130,40,90]");-- inserted to centroid 2 of version 0
```

### 4. 内容被修改 (文件1行 246-247, 文件2行 248-249)

**当前版本内容:**
```
- insert into tbl values(25,"[2,4,5]");
- insert into tbl values(28,"[131,41,91]");
```

**Git历史版本内容:**
```
+ insert into tbl values(25,"[2,4,5]");-- inserted to cluster 1 of version 1
+ insert into tbl values(28,"[131,41,91]");-- inserted to cluster 2 of version 1
```

### 5. 内容被修改 (文件1行 273-274, 文件2行 275-276)

**当前版本内容:**
```
- insert into tbl values(15,90,"[1,3,5]");
- insert into tbl values(18,100,"[130,40,90]");
```

**Git历史版本内容:**
```
+ insert into tbl values(15,90,"[1,3,5]");-- inserted to centroid 1 of version 0
+ insert into tbl values(18,100,"[130,40,90]");-- inserted to centroid 2 of version 0
```

### 6. 内容被修改 (文件1行 276-277, 文件2行 278-279)

**当前版本内容:**
```
- insert into tbl values(25,110,"[2,4,5]");
- insert into tbl values(28,120,"[131,41,91]");
```

**Git历史版本内容:**
```
+ insert into tbl values(25,110,"[2,4,5]");-- inserted to cluster 1 of version 1
+ insert into tbl values(28,120,"[131,41,91]");-- inserted to cluster 2 of version 1
```

### 7. 内容被修改 (文件1行 299-300, 文件2行 301-302)

**当前版本内容:**
```
- insert into tbl values(15,"[1,3,5]");
- insert into tbl values(18,"[130,40,90]");
```

**Git历史版本内容:**
```
+ insert into tbl values(15,"[1,3,5]");-- inserted to centroid 1 of version 0
+ insert into tbl values(18,"[130,40,90]");-- inserted to centroid 2 of version 0
```

### 8. 内容被修改 (文件1行 302-303, 文件2行 304-305)

**当前版本内容:**
```
- insert into tbl values(25,"[2,4,5]");
- insert into tbl values(28,"[131,41,91]");
```

**Git历史版本内容:**
```
+ insert into tbl values(25,"[2,4,5]");-- inserted to cluster 1 of version 1
+ insert into tbl values(28,"[131,41,91]");-- inserted to cluster 2 of version 1
```

### 9. 内容被修改 (文件1行 316-317, 文件2行 318-319)

**当前版本内容:**
```
- delete from tbl where id=9;
- delete from tbl where embedding="[130,40,90]";
```

**Git历史版本内容:**
```
+ delete from tbl where id=9;-- delete 9
+ delete from tbl where embedding="[130,40,90]";-- delete 8
```

### 10. 内容被修改 (文件1行 320-322, 文件2行 322-324)

**当前版本内容:**
```
- delete from tbl where id=6;
- delete from tbl where embedding="[1,3,5]";
- delete from tbl where id=10;
```

**Git历史版本内容:**
```
+ delete from tbl where id=6;-- removes both(0,6)and(1,6)entries
+ delete from tbl where embedding="[1,3,5]";-- removes both(0,5)and(1,5)entries
+ delete from tbl where id=10;-- removes(1,10)
```

### 11. 内容被修改 (文件1行 358-359, 文件2行 360-361)

**当前版本内容:**
```
- update tbl set embedding="[1,2,3]" where id=8;
- update tbl set id=9 where id=8;
```

**Git历史版本内容:**
```
+ update tbl set embedding="[1,2,3]" where id=8;-- update 8 to cluster 1 from cluster 2
+ update tbl set id=9 where id=8;-- update 8 to 9
```

### 12. 内容被修改 (文件1行 361-362, 文件2行 363-364)

**当前版本内容:**
```
- update tbl set embedding="[1,2,3]" where id=7;
- update tbl set id=10 where id=7;
```

**Git历史版本内容:**
```
+ update tbl set embedding="[1,2,3]" where id=7;-- update 7 to cluster 1 from cluster 2 for the latest versions
+ update tbl set id=10 where id=7;-- update 7 to 10
```

### 13. 文件2中独有的内容 (行 672-672)

**Git历史版本内容:**
```
+ a b c
```

### 14. 文件2中独有的内容 (行 675-675)

**Git历史版本内容:**
```
+ a b c
```

---

## 3. array_index_1.result

**文件路径:** array_index_1.result

**统计信息:**
- 当前版本行数: 717
- Git版本行数: 721

**内容差异 (14 个):**

### 1. 文件2中独有的内容 (行 187-187)

**Git历史版本内容:**
```
+ name type column_name algo algo_table_type algo_params
```

### 2. 文件2中独有的内容 (行 221-221)

**Git历史版本内容:**
```
+ name type column_name algo algo_table_type algo_params
```

### 3. 内容被修改 (文件1行 243-244, 文件2行 245-246)

**当前版本内容:**
```
- insert into tbl values(15,"[1,3,5]");
- insert into tbl values(18,"[130,40,90]");
```

**Git历史版本内容:**
```
+ insert into tbl values(15,"[1,3,5]");-- inserted to centroid 1 of version 0
+ insert into tbl values(18,"[130,40,90]");-- inserted to centroid 2 of version 0
```

### 4. 内容被修改 (文件1行 246-247, 文件2行 248-249)

**当前版本内容:**
```
- insert into tbl values(25,"[2,4,5]");
- insert into tbl values(28,"[131,41,91]");
```

**Git历史版本内容:**
```
+ insert into tbl values(25,"[2,4,5]");-- inserted to cluster 1 of version 1
+ insert into tbl values(28,"[131,41,91]");-- inserted to cluster 2 of version 1
```

### 5. 内容被修改 (文件1行 273-274, 文件2行 275-276)

**当前版本内容:**
```
- insert into tbl values(15,90,"[1,3,5]");
- insert into tbl values(18,100,"[130,40,90]");
```

**Git历史版本内容:**
```
+ insert into tbl values(15,90,"[1,3,5]");-- inserted to centroid 1 of version 0
+ insert into tbl values(18,100,"[130,40,90]");-- inserted to centroid 2 of version 0
```

### 6. 内容被修改 (文件1行 276-277, 文件2行 278-279)

**当前版本内容:**
```
- insert into tbl values(25,110,"[2,4,5]");
- insert into tbl values(28,120,"[131,41,91]");
```

**Git历史版本内容:**
```
+ insert into tbl values(25,110,"[2,4,5]");-- inserted to cluster 1 of version 1
+ insert into tbl values(28,120,"[131,41,91]");-- inserted to cluster 2 of version 1
```

### 7. 内容被修改 (文件1行 299-300, 文件2行 301-302)

**当前版本内容:**
```
- insert into tbl values(15,"[1,3,5]");
- insert into tbl values(18,"[130,40,90]");
```

**Git历史版本内容:**
```
+ insert into tbl values(15,"[1,3,5]");-- inserted to centroid 1 of version 0
+ insert into tbl values(18,"[130,40,90]");-- inserted to centroid 2 of version 0
```

### 8. 内容被修改 (文件1行 302-303, 文件2行 304-305)

**当前版本内容:**
```
- insert into tbl values(25,"[2,4,5]");
- insert into tbl values(28,"[131,41,91]");
```

**Git历史版本内容:**
```
+ insert into tbl values(25,"[2,4,5]");-- inserted to cluster 1 of version 1
+ insert into tbl values(28,"[131,41,91]");-- inserted to cluster 2 of version 1
```

### 9. 内容被修改 (文件1行 316-317, 文件2行 318-319)

**当前版本内容:**
```
- delete from tbl where id=9;
- delete from tbl where embedding="[130,40,90]";
```

**Git历史版本内容:**
```
+ delete from tbl where id=9;-- delete 9
+ delete from tbl where embedding="[130,40,90]";-- delete 8
```

### 10. 内容被修改 (文件1行 320-322, 文件2行 322-324)

**当前版本内容:**
```
- delete from tbl where id=6;
- delete from tbl where embedding="[1,3,5]";
- delete from tbl where id=10;
```

**Git历史版本内容:**
```
+ delete from tbl where id=6;-- removes both(0,6)and(1,6)entries
+ delete from tbl where embedding="[1,3,5]";-- removes both(0,5)and(1,5)entries
+ delete from tbl where id=10;-- removes(1,10)
```

### 11. 内容被修改 (文件1行 358-359, 文件2行 360-361)

**当前版本内容:**
```
- update tbl set embedding="[1,2,3]" where id=8;
- update tbl set id=9 where id=8;
```

**Git历史版本内容:**
```
+ update tbl set embedding="[1,2,3]" where id=8;-- update 8 to cluster 1 from cluster 2
+ update tbl set id=9 where id=8;-- update 8 to 9
```

### 12. 内容被修改 (文件1行 361-362, 文件2行 363-364)

**当前版本内容:**
```
- update tbl set embedding="[1,2,3]" where id=7;
- update tbl set id=10 where id=7;
```

**Git历史版本内容:**
```
+ update tbl set embedding="[1,2,3]" where id=7;-- update 7 to cluster 1 from cluster 2 for the latest versions
+ update tbl set id=10 where id=7;-- update 7 to 10
```

### 13. 文件2中独有的内容 (行 672-672)

**Git历史版本内容:**
```
+ a b c
```

### 14. 文件2中独有的内容 (行 675-675)

**Git历史版本内容:**
```
+ a b c
```

---

