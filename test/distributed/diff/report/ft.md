# 批量Git Diff报告

生成时间: 2025-10-10 14:34:07.665423
比较commit: HEAD
总文件数: 5

## 1. datalink.result

**文件路径:** datalink.result

**统计信息:**
- 当前版本行数: 24
- Git版本行数: 25

**内容差异 (1 个):**

### 1. 内容被修改 (文件1行 5-5, 文件2行 5-6)

**当前版本内容:**
```
- insert into datasrc values(0,'stage://ftstage/mo.pdf'),(1,'file:///$resources/fulltext/chinese.pdf'),(2,'file:///$resources/fulltext/file-sample_100kb.docx');
```

**Git历史版本内容:**
```
+ insert into datasrc values(0,'stage://ftstage/mo.pdf'),(1,'file:///$resources/fulltext/chinese.pdf'),
+ (2,'file:///$resources/fulltext/file-sample_100kb.docx');
```

---

## 2. fulltext.result

**文件路径:** fulltext.result

**统计信息:**
- 当前版本行数: 601
- Git版本行数: 614

**内容差异 (13 个):**

### 1. 文件2中独有的内容 (行 23-23)

**Git历史版本内容:**
```
+ sql parser error: you have an error in your sql syntax;check the manual that corresponds to your matrixone server version for the right syntax to use. syntax error at line 1 column 55 near 'm');';
```

### 2. 文件2中独有的内容 (行 77-77)

**Git历史版本内容:**
```
+ id body title
```

### 3. 文件2中独有的内容 (行 93-93)

**Git历史版本内容:**
```
+ id body title
```

### 4. 文件2中独有的内容 (行 95-95)

**Git历史版本内容:**
```
+ id body title
```

### 5. 文件2中独有的内容 (行 97-97)

**Git历史版本内容:**
```
+ id body title
```

### 6. 文件2中独有的内容 (行 119-119)

**Git历史版本内容:**
```
+ id body title
```

### 7. 文件2中独有的内容 (行 121-121)

**Git历史版本内容:**
```
+ id body title
```

### 8. 文件2中独有的内容 (行 184-184)

**Git历史版本内容:**
```
+ id body title
```

### 9. 文件2中独有的内容 (行 244-244)

**Git历史版本内容:**
```
+ id1 id2 body title
```

### 10. 文件2中独有的内容 (行 383-383)

**Git历史版本内容:**
```
+ id body title
```

### 11. 文件2中独有的内容 (行 416-416)

**Git历史版本内容:**
```
+ id body title
```

### 12. 文件2中独有的内容 (行 419-419)

**Git历史版本内容:**
```
+ id body title
```

### 13. 文件2中独有的内容 (行 565-565)

**Git历史版本内容:**
```
+ n_nationkey n_name n_regionkey n_comment n_dummy
```

---

## 3. fulltext2.result

**文件路径:** fulltext2.result

**统计信息:**
- 当前版本行数: 716
- Git版本行数: 733

**内容差异 (7 个):**

### 1. 文件2中独有的内容 (行 196-196)

**Git历史版本内容:**
```
+ a
```

### 2. 内容被修改 (文件1行 199-199, 文件2行 200-200)

**当前版本内容:**
```
- 时
```

**Git历史版本内容:**
```
+ æ¶
```

### 3. 内容被修改 (文件1行 435-438, 文件2行 436-439)

**当前版本内容:**
```
- true 1 var 2020-09-07 2020-09-07 00:00:00 2020-09-06 16:00:00 18 121.11 ['1',2,null,false,true,{'q': 1}] 1qaz null null
- true 2 var 2020-09-07 2020-09-07 00:00:00 2020-09-06 16:00:00 18 121.11 {'b': ['a','b',{'q': 4}],'c': 1} 1aza null null
- true 3 var 2020-09-07 2020-09-07 00:00:00 2020-09-06 16:00:00 18 121.11 ['1',2,null,false,true,{'q': 1}] 1az null null
- true 4 var 2020-09-07 2020-09-07 00:00:00 2020-09-06 16:00:00 18 121.11 {'b': ['a','b',{'q': 4}],'c': 1} 1qaz null null
```

**Git历史版本内容:**
```
+ true 1 var 2020-09-07 2020-09-07 00:00:00 2020-09-07 00:00:00 18 121.11 ['1',2,null,false,true,{'q': 1}] 1qaz null null
+ true 2 var 2020-09-07 2020-09-07 00:00:00 2020-09-07 00:00:00 18 121.11 {'b': ['a','b',{'q': 4}],'c': 1} 1aza null null
+ true 3 var 2020-09-07 2020-09-07 00:00:00 2020-09-07 00:00:00 18 121.11 ['1',2,null,false,true,{'q': 1}] 1az null null
+ true 4 var 2020-09-07 2020-09-07 00:00:00 2020-09-07 00:00:00 18 121.11 {'b': ['a','b',{'q': 4}],'c': 1} 1qaz null null
```

### 4. 文件2中独有的内容 (行 558-558)

**Git历史版本内容:**
```
+ id content
```

### 5. 文件2中独有的内容 (行 617-617)

**Git历史版本内容:**
```
+ id data
```

### 6. 文件2中独有的内容 (行 656-659)

**Git历史版本内容:**
```
+ select a.title,a.content,au.name
+ from articles a
+ join authors au on a.author_id = au.id
+ where match(a.content)against('mo' in natural language mode);
```

### 7. 文件2中独有的内容 (行 683-692)

**Git历史版本内容:**
```
+ select count(posts.title),count(comments.comment_id)as comment_count
+ from posts
+ left join comments on posts.post_id = comments.post_id
+ where match(posts.content)against('全文索引' in natural language mode)
+ group by posts.post_id;
+ select title,content from articles
+ where match(content)against('全文索引' in natural language mode)
+ union
+ select comment_text as title,comment_text as content from comments
+ where match(comment_text)against('全文索引' in natural language mode);
```

---

## 4. fulltext_bm25.result

**文件路径:** fulltext_bm25.result

**统计信息:**
- 当前版本行数: 576
- Git版本行数: 589

**内容差异 (17 个):**

### 1. 文件2中独有的内容 (行 20-20)

**Git历史版本内容:**
```
+ sql parser error: you have an error in your sql syntax;check the manual that corresponds to your matrixone server version for the right syntax to use. syntax error at line 1 column 55 near 'm');';
```

### 2. 内容被修改 (文件1行 49-49, 文件2行 50-50)

**当前版本内容:**
```
- 4 遠東兒童中文是針對6到9歲的小朋友精心設計的中文學習教材，共三冊，目前已出版一、二冊。 遠東兒童中文 0.8257002
```

**Git历史版本内容:**
```
+ 4 遠東兒童中文是針對6到9歲的小朋友精心設計的中文學習教材，共三冊，目前已出版一、二冊。 遠東兒童中文 0.82570016
```

### 3. 文件2中独有的内容 (行 74-74)

**Git历史版本内容:**
```
+ id body title
```

### 4. 文件2中独有的内容 (行 90-90)

**Git历史版本内容:**
```
+ id body title
```

### 5. 文件2中独有的内容 (行 92-92)

**Git历史版本内容:**
```
+ id body title
```

### 6. 文件2中独有的内容 (行 94-94)

**Git历史版本内容:**
```
+ id body title
```

### 7. 文件2中独有的内容 (行 116-116)

**Git历史版本内容:**
```
+ id body title
```

### 8. 文件2中独有的内容 (行 118-118)

**Git历史版本内容:**
```
+ id body title
```

### 9. 内容被修改 (文件1行 126-126, 文件2行 133-133)

**当前版本内容:**
```
- 0.6888391
```

**Git历史版本内容:**
```
+ 0.688839
```

### 10. 内容被修改 (文件1行 150-150, 文件2行 157-157)

**当前版本内容:**
```
- 4 遠東兒童中文是針對6到9歲的小朋友精心設計的中文學習教材，共三冊，目前已出版一、二冊。 遠東兒童中文 0.8257002
```

**Git历史版本内容:**
```
+ 4 遠東兒童中文是針對6到9歲的小朋友精心設計的中文學習教材，共三冊，目前已出版一、二冊。 遠東兒童中文 0.82570016
```

### 11. 文件2中独有的内容 (行 181-181)

**Git历史版本内容:**
```
+ id body title
```

### 12. 文件2中独有的内容 (行 241-241)

**Git历史版本内容:**
```
+ id1 id2 body title
```

### 13. 内容被修改 (文件1行 350-350, 文件2行 359-359)

**当前版本内容:**
```
- 4 遠東兒童中文是針對6到9歲的小朋友精心設計的中文學習教材，共三冊，目前已出版一、二冊。 遠東兒童中文 0.8257002
```

**Git历史版本内容:**
```
+ 4 遠東兒童中文是針對6到9歲的小朋友精心設計的中文學習教材，共三冊，目前已出版一、二冊。 遠東兒童中文 0.82570016
```

### 14. 文件2中独有的内容 (行 380-380)

**Git历史版本内容:**
```
+ id body title
```

### 15. 文件2中独有的内容 (行 413-413)

**Git历史版本内容:**
```
+ id body title
```

### 16. 文件2中独有的内容 (行 416-416)

**Git历史版本内容:**
```
+ id body title
```

### 17. 文件2中独有的内容 (行 569-569)

**Git历史版本内容:**
```
+ n_nationkey n_name n_regionkey n_comment n_dummy
```

---

## 5. jsonvalue.result

**文件路径:** jsonvalue.result

**统计信息:**
- 当前版本行数: 77
- Git版本行数: 79

**内容差异 (2 个):**

### 1. 文件2中独有的内容 (行 35-35)

**Git历史版本内容:**
```
+ id json1 json2
```

### 2. 文件2中独有的内容 (行 45-45)

**Git历史版本内容:**
```
+ id json1 json2
```

---

