#!/usr/bin/env python3
"""
脚本：将指定目录中的 .test 文件重命名为 .sql 文件

使用方法:
    python3 rename.py <目录路径>

示例:
    python3 rename.py /path/to/directory
    python3 rename.py .  # 当前目录
"""

import os
import sys
import argparse
from pathlib import Path


def rename_test_files(directory_path):
    """
    将指定目录中的 .test 文件重命名为 .sql 文件
    
    Args:
        directory_path (str): 目标目录路径
        
    Returns:
        tuple: (成功重命名的文件数量, 失败的文件列表)
    """
    directory = Path(directory_path)
    
    if not directory.exists():
        print(f"错误: 目录 '{directory_path}' 不存在")
        return 0, []
    
    if not directory.is_dir():
        print(f"错误: '{directory_path}' 不是一个目录")
        return 0, []
    
    # 查找所有 .test 文件
    test_files = list(directory.glob("**/*.test"))
    
    if not test_files:
        print(f"在目录 '{directory_path}' 中没有找到 .test 文件")
        return 0, []
    
    print(f"在目录 '{directory_path}' 中找到 {len(test_files)} 个 .test 文件")
    
    success_count = 0
    failed_files = []
    
    for test_file in test_files:
        try:
            # 生成新的 .sql 文件名
            sql_file = test_file.with_suffix('.sql')
            
            # 检查目标文件是否已存在
            if sql_file.exists():
                print(f"警告: 目标文件 '{sql_file.name}' 已存在，跳过 '{test_file.name}'")
                failed_files.append(str(test_file))
                continue
            
            # 重命名文件
            test_file.rename(sql_file)
            print(f"✓ 重命名: {test_file.name} -> {sql_file.name}")
            success_count += 1
            
        except Exception as e:
            print(f"✗ 重命名失败: {test_file.name} - {str(e)}")
            failed_files.append(str(test_file))
    
    return success_count, failed_files


def main():
    """主函数"""
    parser = argparse.ArgumentParser(
        description="将指定目录中的 .test 文件重命名为 .sql 文件",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
示例:
  %(prog)s /path/to/directory    # 重命名指定目录中的文件
  %(prog)s .                     # 重命名当前目录中的文件
  %(prog)s --dry-run /path/to    # 预览模式，不实际重命名
        """
    )
    
    parser.add_argument(
        'directory',
        help='目标目录路径'
    )
    
    parser.add_argument(
        '--dry-run',
        action='store_true',
        help='预览模式，显示将要重命名的文件但不实际执行'
    )
    
    parser.add_argument(
        '--verbose', '-v',
        action='store_true',
        help='显示详细信息'
    )
    
    args = parser.parse_args()
    
    if args.dry_run:
        print("预览模式 - 不会实际重命名文件")
        directory = Path(args.directory)
        
        if not directory.exists():
            print(f"错误: 目录 '{args.directory}' 不存在")
            sys.exit(1)
        
        test_files = list(directory.glob("*.test"))
        if not test_files:
            print(f"在目录 '{args.directory}' 中没有找到 .test 文件")
            sys.exit(0)
        
        print(f"将重命名以下 {len(test_files)} 个文件:")
        for test_file in test_files:
            sql_file = test_file.with_suffix('.sql')
            status = "✓" if not sql_file.exists() else "⚠ (目标文件已存在)"
            print(f"  {status} {test_file.name} -> {sql_file.name}")
        
        sys.exit(0)
    
    # 执行重命名
    success_count, failed_files = rename_test_files(args.directory)
    
    # 输出结果摘要
    print(f"\n重命名完成!")
    print(f"成功: {success_count} 个文件")
    if failed_files:
        print(f"失败: {len(failed_files)} 个文件")
        if args.verbose:
            print("失败的文件:")
            for failed_file in failed_files:
                print(f"  - {failed_file}")


if __name__ == "__main__":
    main()
