#!/usr/bin/env python3
"""
Git版本diff工具 - 比较文件的当前版本和git历史版本
专门用于比较SQL测试结果文件的格式变化
"""

import os
import sys
import argparse
import subprocess
import tempfile
import re
from pathlib import Path
from typing import List, Tuple, Dict, Set
from difflib import unified_diff, SequenceMatcher

class SmartDiff:
    def __init__(self):
        # 匹配SQL元数据header的正则表达式
        self.metadata_pattern = re.compile(r'^#SQL\[@\d+,N\d+\]Result\[\]|^#SQL\[@\d+,N\d+\]Error\[\d+\]')
        # 匹配分隔符模式
        self.separator_pattern = re.compile(r'\s*¦\s*')
        
    def normalize_line(self, line: str) -> str:
        """标准化单行内容，忽略格式差异"""
        # 移除SQL元数据header
        if (re.match(r'^#SQL\[@\d+,N\d+\]', line.strip()) or
            re.match(r'^Result\[\d+.*\]$', line.strip()) or
            re.match(r'^Error\[\d+\]$', line.strip())):
            return ""
        
        # 处理转义字符和换行
        line = line.replace('\\n', '\n').replace('\\"', '"')
        
        # 标准化分隔符和空格
        normalized = self.separator_pattern.sub(' ', line)
        normalized = re.sub(r'\s+', ' ', normalized)
        normalized = normalized.replace('\n', ' ').strip()
        
        # 标准化标点符号
        normalized = re.sub(r'\s*([(),;])\s*', r'\1', normalized)
        
        return normalized
    
    def parse_file(self, filepath: str) -> List[str]:
        """解析文件，返回标准化后的内容行"""
        try:
            with open(filepath, 'r', encoding='utf-8') as f:
                content = f.read()
        except FileNotFoundError:
            print(f"错误：文件 {filepath} 不存在")
            return []
        except Exception as e:
            print(f"错误：读取文件 {filepath} 失败: {e}")
            return []
        
        # 处理转义的换行符
        content = content.replace('\\n', '\n')
        
        # 按行分割
        lines = content.split('\n')
        
        normalized_lines = []
        for line in lines:
            normalized = self.normalize_line(line)
            if normalized:  # 只保留非空行
                normalized_lines.append(normalized)
        
        return normalized_lines
    
    
    def compare_files(self, file1: str, file2: str, show_context: bool = True) -> Dict:
        """比较两个文件，返回差异分析"""
        # 标准化内容
        lines1 = self.parse_file(file1)
        lines2 = self.parse_file(file2)
        
        result = {
            'file1': file1,
            'file2': file2,
            'lines1_count': len(lines1),
            'lines2_count': len(lines2),
            'differences': []
        }
        
        # 比较标准化后的行
        matcher = SequenceMatcher(None, lines1, lines2)
        
        # 分析差异
        for tag, i1, i2, j1, j2 in matcher.get_opcodes():
            if tag == 'equal':
                continue
            elif tag == 'delete':
                result['differences'].append({
                    'type': 'deleted',
                    'file1_lines': lines1[i1:i2],
                    'file2_lines': [],
                    'description': f"文件1中独有的内容 (行 {i1+1}-{i2})"
                })
            elif tag == 'insert':
                result['differences'].append({
                    'type': 'added',
                    'file1_lines': [],
                    'file2_lines': lines2[j1:j2],
                    'description': f"文件2中独有的内容 (行 {j1+1}-{j2})"
                })
            elif tag == 'replace':
                result['differences'].append({
                    'type': 'modified',
                    'file1_lines': lines1[i1:i2],
                    'file2_lines': lines2[j1:j2],
                    'description': f"内容被修改 (文件1行 {i1+1}-{i2}, 文件2行 {j1+1}-{j2})"
                })
        
        return result
    
    def print_differences(self, result: Dict, show_context: bool = True):
        """打印差异结果"""
        if not result['differences'] and not result.get('sql_differences'):
            print("✅ 没有发现实际的数据差异（忽略格式变化）")
            return
        
        # 只显示真正的内容差异，过滤掉格式差异
        real_differences = []
        for diff in result['differences']:
            # 检查是否是真正的数据差异
            if self.is_real_difference(diff):
                real_differences.append(diff)
        
        
        if not real_differences:
            print("✅ 没有发现实际的数据差异（忽略格式变化）")
            return
        
        print(f"\n发现 {len(real_differences)} 个实际数据差异:")
        print("-" * 40)
        
        for i, diff in enumerate(real_differences, 1):
            print(f"\n{i}. {diff['description']}")
            print("   " + "-" * 30)
            
            if diff['file1_lines']:
                print("   当前版本:")
                for line in diff['file1_lines']:
                    print(f"   - {line}")
            
            if diff['file2_lines']:
                print("   Git历史版本:")
                for line in diff['file2_lines']:
                    print(f"   + {line}")
    
    def is_real_difference(self, diff: Dict) -> bool:
        """判断是否是真正的数据差异，而不是格式差异"""
        if not diff['file1_lines'] or not diff['file2_lines']:
            # 检查是否是表头行
            content = ' '.join(diff['file1_lines'] + diff['file2_lines']).strip()
            if 'Table Non_unique Key_name' in content:
                return False
            return True  # 新增或删除的内容
        
        # 比较标准化后的内容
        content1 = ' '.join(diff['file1_lines']).strip()
        content2 = ' '.join(diff['file2_lines']).strip()
        
        # 如果标准化后内容相同，则是格式差异
        if content1 == content2:
            return False
        
        # 检查是否是注释差异
        # 移除注释部分后比较
        content1_no_comment = content1.split('--')[0].strip()
        content2_no_comment = content2.split('--')[0].strip()
        if content1_no_comment == content2_no_comment:
            return False
        
        # 检查是否是引号转义差异
        if (content1.replace('\\"', '"') == content2.replace('\\"', '"')):
            return False
        
        # 检查是否是标点符号差异（分号、引号等）
        content1_normalized = content1.replace(';', '').replace('\\"', '"').strip()
        content2_normalized = content2.replace(';', '').replace('\\"', '"').strip()
        if content1_normalized == content2_normalized:
            return False
        
        # 检查是否是表头行
        if 'Table Non_unique Key_name' in content1 or 'Table Non_unique Key_name' in content2:
            return False
        
        return True

class GitDiff:
    def __init__(self):
        self.diff_tool = SmartDiff()
    
    def get_git_file_content(self, filepath: str, commit: str = "HEAD~1") -> str:
        """获取git历史版本的文件内容"""
        try:
            # 获取相对于git根目录的文件路径
            git_root = subprocess.run(
                ['git', 'rev-parse', '--show-toplevel'],
                capture_output=True, text=True, cwd=os.path.dirname(filepath) or '.'
            ).stdout.strip()
            
            # 计算相对路径
            abs_filepath = os.path.abspath(filepath)
            rel_path = os.path.relpath(abs_filepath, git_root)
            
            # 检查文件是否在git中
            result = subprocess.run(
                ['git', 'log', '--oneline', '-1', '--', rel_path],
                capture_output=True, text=True, cwd=git_root
            )
            
            if result.returncode != 0 or not result.stdout.strip():
                print(f"警告: 文件 {rel_path} 不在git历史中")
                return None
            
            # 获取指定commit的文件内容
            result = subprocess.run(
                ['git', 'show', f'{commit}:{rel_path}'],
                capture_output=True, text=True, cwd=git_root
            )
            
            if result.returncode != 0:
                print(f"错误: 无法获取文件 {rel_path} 在commit {commit} 的内容")
                return None
            
            return result.stdout
            
        except Exception as e:
            print(f"错误: 获取git文件内容失败: {e}")
            return None
    
    def get_current_file_content(self, filepath: str) -> str:
        """获取当前工作区文件内容"""
        try:
            with open(filepath, 'r', encoding='utf-8') as f:
                return f.read()
        except Exception as e:
            print(f"错误: 读取当前文件失败: {e}")
            return None
    
    def compare_with_git(self, filepath: str, commit: str = "HEAD~1", show_context: bool = True) -> dict:
        """比较文件与git历史版本"""
        print(f"比较文件: {filepath}")
        print(f"当前版本 vs Git版本 ({commit})")
        print("=" * 60)
        
        # 获取当前版本内容
        current_content = self.get_current_file_content(filepath)
        if current_content is None:
            return None
        
        # 获取git历史版本内容
        git_content = self.get_git_file_content(filepath, commit)
        if git_content is None:
            return None
        
        # 创建临时文件进行比较
        with tempfile.NamedTemporaryFile(mode='w', suffix='_current', delete=False, encoding='utf-8') as f:
            f.write(current_content)
            current_file = f.name
        
        with tempfile.NamedTemporaryFile(mode='w', suffix='_git', delete=False, encoding='utf-8') as f:
            f.write(git_content)
            git_file = f.name
        
        try:
            # 使用smart_diff进行比较
            result = self.diff_tool.compare_files(current_file, git_file, show_context)
            result['current_file'] = filepath
            result['git_commit'] = commit
            return result
        finally:
            # 清理临时文件
            try:
                os.unlink(current_file)
                os.unlink(git_file)
            except:
                pass
    
    def get_git_commits(self, filepath: str, limit: int = 10) -> list:
        """获取文件的git提交历史"""
        try:
            result = subprocess.run(
                ['git', 'log', '--oneline', f'-{limit}', '--', filepath],
                capture_output=True, text=True, cwd=os.path.dirname(filepath) or '.'
            )
            
            if result.returncode != 0:
                return []
            
            commits = []
            for line in result.stdout.strip().split('\n'):
                if line.strip():
                    parts = line.split(' ', 1)
                    if len(parts) >= 2:
                        commits.append({
                            'hash': parts[0],
                            'message': parts[1]
                        })
            
            return commits
            
        except Exception as e:
            print(f"错误: 获取git历史失败: {e}")
            return []
    
    def print_differences(self, result: dict, show_context: bool = True):
        """打印差异结果"""
        if not result:
            return
        
        if not result['differences'] and not result.get('sql_differences'):
            print("✅ 没有发现实际的数据差异（忽略格式变化）")
            return
        
        print(f"\n发现 {len(result['differences'])} 个内容差异:")
        print("-" * 40)
        
        for i, diff in enumerate(result['differences'], 1):
            print(f"\n{i}. {diff['description']}")
            print("   " + "-" * 30)
            
            if diff['file1_lines']:
                print("   当前版本内容:")
                for line in diff['file1_lines']:
                    print(f"   - {line}")
            
            if diff['file2_lines']:
                print("   Git历史版本内容:")
                for line in diff['file2_lines']:
                    print(f"   + {line}")
        
        if result.get('sql_differences'):
            print(f"\n发现 {len(result['sql_differences'])} 个SQL语句差异:")
            print("-" * 40)
            
            for i, diff in enumerate(result['sql_differences'], 1):
                print(f"\n{i}. {diff['description']}")
                print("   " + "-" * 30)
                
                if 'content' in diff:
                    for sql in diff['content']:
                        print(f"   {sql}")
                elif 'content1' in diff:
                    print("   当前版本 SQL:")
                    for sql in diff['content1']:
                        print(f"   - {sql}")
                    print("   Git历史版本 SQL:")
                    for sql in diff['content2']:
                        print(f"   + {sql}")

def main():
    parser = argparse.ArgumentParser(description='Git版本diff工具 - 比较文件与git历史版本')
    parser.add_argument('file', help='要比较的文件路径')
    parser.add_argument('--commit', '-c', default='HEAD~1', help='要比较的git commit (默认: HEAD~1)')
    parser.add_argument('--context', action='store_true', help='显示上下文信息')
    parser.add_argument('--verbose', '-v', action='store_true', help='详细输出')
    parser.add_argument('--history', action='store_true', help='显示文件的git提交历史')
    
    args = parser.parse_args()
    
    if not os.path.exists(args.file):
        print(f"错误：文件 {args.file} 不存在")
        sys.exit(1)
    
    git_diff = GitDiff()
    
    # 显示git历史
    if args.history:
        print(f"文件 {args.file} 的git提交历史:")
        print("-" * 40)
        commits = git_diff.get_git_commits(args.file)
        for i, commit in enumerate(commits, 1):
            print(f"{i}. {commit['hash']} - {commit['message']}")
        print()
    
    # 比较文件
    result = git_diff.compare_with_git(args.file, args.commit, args.context)
    
    if result:
        git_diff.diff_tool.print_differences(result, args.context)
        
        if args.verbose:
            print(f"\n统计信息:")
            print(f"  当前版本: {result['lines1_count']} 行")
            print(f"  Git版本: {result['lines2_count']} 行")
            print(f"  差异数: {len(result['differences'])}")
            print(f"  比较的commit: {result['git_commit']}")

if __name__ == '__main__':
    main()
