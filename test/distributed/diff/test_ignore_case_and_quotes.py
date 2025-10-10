#!/usr/bin/env python3
"""
测试忽略大小写和引号差异的功能
"""

import sys
import os

# 添加当前目录到 Python 路径
sys.path.insert(0, os.path.dirname(__file__))

from git_diff import SmartDiff

def test_normalize_case_and_quotes():
    """测试大小写和引号标准化"""
    diff_tool = SmartDiff()
    
    test_cases = [
        # 测试大小写忽略
        ("SELECT * FROM table1", "select * from TABLE1", True),
        ("INSERT INTO users", "insert into USERS", True),
        
        # 测试引号差异忽略
        ('name = "John"', "name = 'John'", True),
        ("value = 'test'", 'value = "test"', True),
        ('SELECT "column"', "SELECT 'column'", True),
        
        # 测试中文引号
        ('说明 = "测试"', "说明 = '测试'", True),
        ('"中文字符串"', "'中文字符串'", True),
        
        # 测试组合情况：大小写 + 引号
        ('SELECT "Name" FROM Users', "select 'name' from USERS", True),
        ('WHERE id = "123"', "where ID = '123'", True),
        
        # 真正不同的内容
        ("SELECT * FROM table1", "SELECT * FROM table2", False),
        ("name = 'John'", "name = 'Jane'", False),
    ]
    
    print("=" * 70)
    print("测试大小写和引号差异忽略功能")
    print("=" * 70)
    
    passed = 0
    failed = 0
    
    for i, (text1, text2, should_be_same) in enumerate(test_cases, 1):
        normalized1 = diff_tool.normalize_line(text1)
        normalized2 = diff_tool.normalize_line(text2)
        
        is_same = (normalized1 == normalized2)
        test_passed = (is_same == should_be_same)
        
        status = "✅ 通过" if test_passed else "❌ 失败"
        
        print(f"\n测试 {i}: {status}")
        print(f"  文本1: {text1}")
        print(f"  文本2: {text2}")
        print(f"  标准化1: {normalized1}")
        print(f"  标准化2: {normalized2}")
        print(f"  预期: {'相同' if should_be_same else '不同'}")
        print(f"  实际: {'相同' if is_same else '不同'}")
        
        if test_passed:
            passed += 1
        else:
            failed += 1
    
    print("\n" + "=" * 70)
    print(f"测试结果: {passed} 个通过, {failed} 个失败")
    print("=" * 70)
    
    return failed == 0

if __name__ == '__main__':
    success = test_normalize_case_and_quotes()
    sys.exit(0 if success else 1)


