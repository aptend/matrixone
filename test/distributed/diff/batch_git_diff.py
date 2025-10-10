#!/usr/bin/env python3
"""
批量Git diff工具 - 比较所有修改过的.result文件与git历史版本
"""

import os
import sys
import argparse
import subprocess
from pathlib import Path
from git_diff import GitDiff

def get_result_files(directory=None):
    """获取指定目录下的所有.result文件"""
    if directory is None:
        # 如果没有指定目录，获取git状态中修改的文件
        try:
            result = subprocess.run(
                ['git', 'status', '--porcelain'],
                capture_output=True, text=True
            )
            
            if result.returncode != 0:
                print("错误: 无法获取git状态")
                return []
            
            modified_files = []
            for line in result.stdout.strip().split('\n'):
                if line.strip():
                    # 解析git状态行: M  filename 或 MM filename
                    parts = line.split()
                    if len(parts) >= 2 and parts[0] in ['M', 'MM'] and parts[1].endswith('.result'):
                        # 只处理当前目录下的文件
                        if 'test/distributed/cases/ddl/' in parts[1]:
                            # 提取文件名
                            filename = os.path.basename(parts[1])
                            modified_files.append(filename)
            
            return modified_files
            
        except Exception as e:
            print(f"错误: 获取修改文件失败: {e}")
            return []
    else:
        # 获取指定目录下的所有.result文件（递归搜索所有子目录）
        try:
            if not os.path.exists(directory):
                print(f"错误: 目录 {directory} 不存在")
                return []
            
            result_files = []
            # 使用os.walk递归遍历所有子目录
            for root, dirs, files in os.walk(directory):
                for filename in files:
                    if filename.endswith('.result'):
                        # 获取相对于指定目录的路径
                        rel_path = os.path.relpath(os.path.join(root, filename), directory)
                        result_files.append(rel_path)
            
            return sorted(result_files)
            
        except Exception as e:
            print(f"错误: 读取目录 {directory} 失败: {e}")
            return []

def batch_compare_result_files(directory=None, commit="HEAD~1", output_file=None):
    """批量比较指定目录下的所有.result文件"""
    result_files = get_result_files(directory)
    
    if not result_files:
        if directory:
            print(f"目录 {directory} 中没有找到.result文件")
        else:
            print("没有找到修改过的.result文件")
        return
    
    if directory:
        print(f"找到 {len(result_files)} 个.result文件在目录 {directory}")
    else:
        print(f"找到 {len(result_files)} 个修改过的.result文件")
    print("=" * 60)
    
    git_diff = GitDiff()
    all_differences = []
    
    for result_file in result_files:
        # 构建完整路径
        if directory:
            full_path = os.path.join(directory, result_file)
        else:
            full_path = os.path.join("../cases/ddl", result_file)
        if not os.path.exists(full_path):
            print(f"⚠️  文件不存在: {full_path}")
            continue
            
        print(f"\n比较: {result_file}")
        print("-" * 40)
        
        try:
            result = git_diff.compare_with_git(full_path, commit, False)
            
            if result and result['differences']:
                # 使用SmartDiff的格式过滤逻辑
                real_differences = []
                for diff in result['differences']:
                    if git_diff.diff_tool.is_real_difference(diff):
                        real_differences.append(diff)
                
                if real_differences:
                    all_differences.append({
                        'file': result_file,
                        'differences': real_differences,
                        'stats': {
                            'lines1': result['lines1_count'],
                            'lines2': result['lines2_count']
                        }
                    })
                    print(f"❌ 发现差异: {len(real_differences)} 个实际数据差异")
                else:
                    print("✅ 无实际数据差异（忽略格式变化）")
            else:
                print("✅ 无差异")
                
        except Exception as e:
            print(f"❌ 比较失败: {e}")
            all_differences.append({
                'file': result_file,
                'error': str(e)
            })
    
    # 生成报告
    if output_file:
        generate_git_diff_report(all_differences, output_file, commit)
    
    # 打印总结
    print("\n" + "=" * 60)
    print("批量Git比较总结:")
    print(f"总文件数: {len(result_files)}")
    print(f"有差异的文件数: {len(all_differences)}")
    
    if all_differences:
        print("\n有差异的文件:")
        for diff in all_differences:
            if 'error' in diff:
                print(f"  ❌ {os.path.basename(diff['file'])} - 错误: {diff['error']}")
            else:
                content_diffs = len(diff['differences'])
                print(f"  ⚠️  {os.path.basename(diff['file'])} - {content_diffs} 内容差异")

def generate_git_diff_report(all_differences, output_file, commit):
    """生成详细的Git差异报告"""
    with open(output_file, 'w', encoding='utf-8') as f:
        f.write("# 批量Git Diff报告\n\n")
        f.write(f"生成时间: {__import__('datetime').datetime.now()}\n")
        f.write(f"比较commit: {commit}\n")
        f.write(f"总文件数: {len(all_differences)}\n\n")
        
        for i, diff in enumerate(all_differences, 1):
            f.write(f"## {i}. {os.path.basename(diff['file'])}\n\n")
            f.write(f"**文件路径:** {diff['file']}\n\n")
            
            if 'error' in diff:
                f.write(f"**错误:** {diff['error']}\n\n")
                continue
            
            f.write(f"**统计信息:**\n")
            f.write(f"- 当前版本行数: {diff['stats']['lines1']}\n")
            f.write(f"- Git版本行数: {diff['stats']['lines2']}\n\n")
            
            if diff['differences']:
                f.write(f"**内容差异 ({len(diff['differences'])} 个):**\n\n")
                for j, content_diff in enumerate(diff['differences'], 1):
                    f.write(f"### {j}. {content_diff['description']}\n\n")
                    if content_diff['file1_lines']:
                        f.write("**当前版本内容:**\n```\n")
                        for line in content_diff['file1_lines']:
                            f.write(f"- {line}\n")
                        f.write("```\n\n")
                    if content_diff['file2_lines']:
                        f.write("**Git历史版本内容:**\n```\n")
                        for line in content_diff['file2_lines']:
                            f.write(f"+ {line}\n")
                        f.write("```\n\n")
            
            
            f.write("---\n\n")

def main():
    parser = argparse.ArgumentParser(description='批量Git diff工具 - 比较指定目录下的.result文件')
    parser.add_argument('--directory', '-d', help='指定目录路径，如果不指定则比较git状态中修改的文件')
    parser.add_argument('--commit', '-c', default='HEAD', help='要比较的git commit (默认: HEAD)')
    parser.add_argument('--output', '-o', help='输出报告文件路径')
    parser.add_argument('--verbose', '-v', action='store_true', help='详细输出')
    
    args = parser.parse_args()
    
    batch_compare_result_files(args.directory, args.commit, args.output)

if __name__ == '__main__':
    main()
