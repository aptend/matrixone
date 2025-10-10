#!/usr/bin/env python3
"""
SQL Test Format Converter

Converts SQL test files from old format to new format.
Supports the following builtin commands:
- @pattern -> @regex(pattern=r"")
- @session:id=X&user=Y&password=Z -> @session(id=X, user="abc:Y", password="Z") { ... }
- @bvt:issue#N -> @issue(no=N) { ... }
- @ignore:X,Y -> | @ignore(X,Y);
"""

import re
import sys
import argparse
import os
import glob
from typing import List, Tuple, Optional


class FormatConverter:
    def __init__(self):
        # Patterns for old format commands
        self.patterns = {
            'pattern': re.compile(r'^\s*--\s*@pattern\s*$', re.MULTILINE),
            'session_start': re.compile(r'^\s*--\s*@session:id=(\d+)(?:&user=([^&]+))?(?:&password=([^\s]+))?\s*{?\s*$', re.MULTILINE),
            'session_end': re.compile(r'^\s*--\s*@session\s*}?\s*$', re.MULTILINE),
            'bvt_issue_start': re.compile(r'^\s*--\s*@bvt:issue#([0-9a-zA-Z"]+)\s*$', re.MULTILINE),
            'bvt_issue_end': re.compile(r'^\s*--\s*@bvt:issue\s*$', re.MULTILINE),
            'ignore': re.compile(r'^\s*--\s*@ignore:([0-9,]+)\s*$', re.MULTILINE),
        }
    
    def _process_save_result_sql(self, lines: List[str], start_idx: int) -> Tuple[List[str], int]:
        """Process a multi-line SQL statement with /* save_result */ comment"""
        # Check if this is a single line SQL statement
        first_line = lines[start_idx].strip()
        sql_part = first_line[17:].strip()  # Remove '/* save_result */' and whitespace
        
        if sql_part and sql_part.endswith(';'):
            # Single line SQL statement
            sql_without_semicolon = sql_part[:-1].rstrip()
            new_line = f"{sql_without_semicolon} /* save_result */;"
            return [new_line], start_idx + 1
        
        # Multi-line SQL statement - find the end
        sql_end_idx = self._find_sql_end(lines, start_idx)
        if sql_end_idx is None:
            # No semicolon found, treat as single line without semicolon
            if sql_part:
                new_line = f"{sql_part} /* save_result */"
                return [new_line], start_idx + 1
            else:
                return [lines[start_idx]], start_idx + 1
        
        # Process multi-line SQL statement
        result_lines = []
        
        # First line: remove /* save_result */ comment
        if sql_part:
            result_lines.append(sql_part)
        else:
            # If no SQL on first line, skip it
            pass
        
        # Add all middle lines (remove trailing spaces)
        for j in range(start_idx + 1, sql_end_idx):
            result_lines.append(lines[j].rstrip())
        
        # Last line: add /* save_result */ comment before semicolon
        last_line = lines[sql_end_idx]
        if last_line.strip().endswith(';'):
            sql_without_semicolon = last_line.rstrip()[:-1].rstrip()
            new_last_line = f"{sql_without_semicolon} /* save_result */;"
        else:
            new_last_line = f"{last_line.rstrip()} /* save_result */"
        result_lines.append(new_last_line)
        
        return result_lines, sql_end_idx + 1

    def _process_sql_with_command(self, lines: List[str], start_idx: int, command_type: str, command_params: str = "", indent: str = '') -> Tuple[List[str], int]:
        """Process a SQL statement with its associated command"""
        sql_start_idx = self._find_next_sql_line(lines, start_idx)
        if sql_start_idx is None:
            return [], start_idx + 1

        sql_end_idx = self._find_sql_end(lines, sql_start_idx)
        if sql_end_idx is None:
            # Fallback to single line (remove trailing spaces)
            result = [indent + lines[sql_start_idx].rstrip()]
            if command_type == "pattern":
                result.append(f'{indent}| @regex(pattern=r"");')
            elif command_type == "ignore":
                result.append(f'{indent}| @ignore({command_params});')
            return result, sql_start_idx + 1
        
        # Add all lines of the SQL statement (remove trailing spaces)
        result = []
        for j in range(sql_start_idx, sql_end_idx + 1):
            result.append(indent + lines[j].rstrip())
        
        # Add the command
        if command_type == "pattern":
            result.append(f'{indent}| @regex(pattern=r"");')
        elif command_type == "ignore":
            result.append(f'{indent}| @ignore({command_params});')
        
        return result, sql_end_idx + 1

    def find_sql_files(self, path: str) -> List[str]:
        """Find all .sql and .test files in the given path"""
        sql_files = []
        
        if os.path.isfile(path):
            # Single file
            if path.endswith(('.sql', '.test')):
                sql_files.append(path)
        elif os.path.isdir(path):
            # Directory - find all .sql and .test files recursively
            for pattern in ['**/*.sql', '**/*.test']:
                sql_files.extend(glob.glob(os.path.join(path, pattern), recursive=True))
        else:
            raise FileNotFoundError(f"Path '{path}' does not exist")
        
        return sorted(sql_files)


    def convert(self, content: str) -> str:
        """Convert old format to new format"""
        lines = content.split('\n')
        
        # Check if any line contains @skip:issue#xxxx pattern (skip empty lines and non-skip comments)
        non_empty_idx = 0
        for i, line in enumerate(lines):
            if len(line.strip()) > 0:
                non_empty_idx = i
                break
        
        skip_issue_match = None
        first_line = lines[non_empty_idx]
        if '@skip:issue#' in first_line:
            skip_pattern = re.compile(r'@skip:issue#([0-9a-zA-Z]+)')
            skip_issue_match = skip_pattern.search(first_line)
        
        # If @skip:issue#xxxx is detected in first line, wrap entire content
        if skip_issue_match:
            issue_num = skip_issue_match.group(1)
            
            # First, convert the content (excluding the first line with @skip:issue#xxxx)
            content_to_convert = '\n'.join(lines[non_empty_idx:])
            converted_content = self._convert_without_skip_check(content_to_convert)
            
            # Then wrap the converted content
            result_lines = [f'@issue(no={issue_num}) {{']
            
            # Add the converted content with proper indentation
            for line in converted_content.split('\n'):
                if line.strip():  # Only indent non-empty lines
                    result_lines.append('    ' + line)
                else:
                    result_lines.append(line)
            
            result_lines.append('}')
            return '\n'.join(result_lines)
        
        # Normal conversion without @skip:issue wrapper
        return self._convert_without_skip_check(content)
    
    def _convert_without_skip_check(self, content: str) -> str:
        """Convert content without checking for @skip:issue in first line (original logic)"""
        lines = content.split('\n')
        result_lines = []
        i = 0
        in_session_block = False
        in_issue_block = False
        
        while i < len(lines):
            line = lines[i]
            
            # Check for /* save_result */ comment at the beginning of SQL line
            if line.strip().startswith('/* save_result */'):
                # Handle multi-line SQL statement with /* save_result */ comment
                sql_lines, new_i = self._process_save_result_sql(lines, i)
                result_lines.extend(sql_lines)
                i = new_i
                continue
            
            # Check for @pattern
            if self.patterns['pattern'].match(line):
                if in_session_block or in_issue_block:
                    indent = '    '
                else:
                    indent = ''
                sql_lines, new_i = self._process_sql_with_command(lines, i + 1, "pattern", indent=indent)
                result_lines.extend(sql_lines)
                i = new_i
                continue
            
            # Check for @session start
            session_match = self.patterns['session_start'].match(line)
            if session_match:
                session_id = session_match.group(1)
                user = session_match.group(2) or ""
                password = session_match.group(3) or ""
                
                # Add comment before session block
                
                # Create new session block
                session_line = f'@session(id={session_id}'
                if user:
                    session_line += f', user="{user}"'
                if password:
                    session_line += f', password="{password}"'
                session_line += ') {'
                
                result_lines.append(session_line)
                in_session_block = True
                i += 1
                continue
            
            # Check for @session end
            if self.patterns['session_end'].match(line):
                if not in_session_block:
                    raise ValueError("Error: Found @session end marker without corresponding @session start marker.")
                result_lines.append('}')
                in_session_block = False
                i += 1
                continue
            
            # Check for @bvt:issue start
            bvt_match = self.patterns['bvt_issue_start'].match(line)
            if bvt_match:
                issue_num = bvt_match.group(1)
                result_lines.append(f'@issue(no={issue_num}) {{')
                in_issue_block = True
                i += 1
                continue
            
            # Check for @bvt:issue end
            if self.patterns['bvt_issue_end'].match(line):
                if not in_issue_block:
                    raise ValueError("Error: Found @bvt:issue end marker without corresponding @bvt:issue start marker.")
                result_lines.append('}')
                in_issue_block = False
                i += 1
                continue
            
            # Check for @ignore
            ignore_match = self.patterns['ignore'].match(line)
            if ignore_match:
                indices = ignore_match.group(1)
                if in_session_block or in_issue_block:
                    indent = '    '
                else:
                    indent = ''
                sql_lines, new_i = self._process_sql_with_command(lines, i + 1, "ignore", indices, indent=indent)
                result_lines.extend(sql_lines)
                i = new_i
                continue
            
            # Regular line - add proper indentation if in block
            if in_session_block or in_issue_block:
                # Add 4-space indentation for content inside blocks
                if line.strip():  # Only indent non-empty lines
                    # Remove trailing spaces from the line before adding indentation
                    cleaned_line = line.rstrip()
                    result_lines.append('    ' + cleaned_line)
                else:
                    result_lines.append(line)
            else:
                # Skip extra blank lines before session blocks
                if (line.strip() == '' and 
                    i + 1 < len(lines) and 
                    self.patterns['session_start'].match(lines[i + 1])):
                    i += 1
                    continue
                # Remove trailing spaces from regular lines
                cleaned_line = line.rstrip()
                result_lines.append(cleaned_line)
            i += 1
        
        # Check for unclosed blocks
        if in_session_block:
            raise ValueError("Error: Found unclosed @session block. Missing @session end marker.")
        if in_issue_block:
            raise ValueError("Error: Found unclosed @bvt:issue block. Missing @bvt:issue end marker.")
        
        return '\n'.join(result_lines)
    
    def _find_previous_sql_line(self, lines: List[str]) -> Optional[int]:
        """Find the most recent non-empty, non-comment SQL line"""
        for i in range(len(lines) - 1, -1, -1):
            line = lines[i].strip()
            if line and not line.startswith('--') and not line.startswith('@'):
                return i
        return None
    
    def _find_next_sql_line(self, lines: List[str], start_idx: int) -> Optional[int]:
        """Find the next non-empty, non-comment SQL line"""
        for i in range(start_idx, len(lines)):
            line = lines[i].strip()
            if line:
                # Handle lines that start with SQL but have trailing comments
                if '--' in line:
                    # Check if this line has SQL content before the comment
                    sql_part = line.split('--')[0].strip()
                    if sql_part and not sql_part.startswith('@'):
                        return i
                # Handle lines that don't start with -- or @
                elif not line.startswith('--') and not line.startswith('@'):
                    return i
        return None
    
    def _find_sql_end(self, lines: List[str], start_idx: int) -> Optional[int]:
        """Find the end of a multi-line SQL statement (ending with semicolon)"""
        for i in range(start_idx, len(lines)):
            line = lines[i].strip()
            if line:
                # Handle lines that start with SQL but have trailing comments
                if '--' in line:
                    # Check if this line has SQL content before the comment
                    sql_part = line.split('--')[0].strip()
                    if sql_part and not sql_part.startswith('@'):
                        # Check if this line ends with semicolon
                        if sql_part.endswith(';'):
                            return i
                # Handle lines that don't start with -- or @
                elif not line.startswith('--') and not line.startswith('@'):
                    # Check if this line ends with semicolon
                    if line.endswith(';'):
                        return i
        return None


def process_single_file(converter: FormatConverter, input_file: str, output_file: str = None, 
                       backup: bool = False, no_overwrite: bool = False) -> bool:
    """Process a single file and return success status"""
    try:
        # Read input file
        with open(input_file, 'r', encoding='utf-8') as f:
            content = f.read()
        
        # Convert format
        converted_content = converter.convert(content)
        
        # Create backup if requested
        if backup:
            backup_file = input_file + '.backup'
            with open(backup_file, 'w', encoding='utf-8') as f:
                f.write(content)
            print(f"Backup created: {backup_file}")
        
        # Write output
        if output_file == '__stdout__':
            # Output to stdout
            print(f"=== {input_file} ===")
            print(converted_content)
            print()
        elif output_file:
            # Check if output file exists and handle overwrite
            if os.path.exists(output_file) and no_overwrite:
                print(f"Error: File '{output_file}' already exists. Skipping.", file=sys.stderr)
                return False
            
            with open(output_file, 'w', encoding='utf-8') as f:
                f.write(converted_content)
            print(f"Converted: {input_file} -> {output_file}")
        else:
            # Default: overwrite original file
            with open(input_file, 'w', encoding='utf-8') as f:
                f.write(converted_content)
            print(f"Converted: {input_file} (overwritten)")
        
        return True
        
    except Exception as e:
        print(f"Error processing {input_file}: {e}", file=sys.stderr)
        return False

def main():
    parser = argparse.ArgumentParser(description='Convert SQL test files from old format to new format')
    parser.add_argument('input_path', help='Input SQL file or directory (old format)')
    parser.add_argument('-o', '--output', help='Output file, directory, or __stdout__ (new format). If not specified, overwrites original files')
    parser.add_argument('--backup', action='store_true', help='Create backup of original files')
    parser.add_argument('--no-overwrite', action='store_true', help='Do not overwrite output files if they exist (default: overwrite)')
    parser.add_argument('--suffix', default='', help='Suffix for output files when processing directory (default: )')
    
    args = parser.parse_args()
    
    try:
        converter = FormatConverter()
        input_files = converter.find_sql_files(args.input_path)
        
        if not input_files:
            print(f"No .sql or .test files found in '{args.input_path}'", file=sys.stderr)
            sys.exit(1)
        
        print(f"Found {len(input_files)} file(s) to process:")
        for file in input_files:
            print(f"  - {file}")
        print()
        
        success_count = 0
        total_count = len(input_files)
        
        for input_file in input_files:
            if args.output:
                if args.output == '__stdout__':
                    # Output to stdout
                    output_file = '__stdout__'
                elif os.path.isdir(args.output):
                    # Output is a directory
                    rel_path = os.path.relpath(input_file, args.input_path)
                    output_file = os.path.join(args.output, rel_path)
                    # Add suffix before extension
                    name, ext = os.path.splitext(output_file)
                    output_file = name + args.suffix + ext
                else:
                    # Output is a single file (only valid for single input file)
                    if len(input_files) > 1:
                        print(f"Error: Cannot specify single output file when processing multiple input files", file=sys.stderr)
                        sys.exit(1)
                    output_file = args.output
            else:
                # Default: overwrite original file
                output_file = None
            
            if process_single_file(converter, input_file, output_file, args.backup, args.no_overwrite):
                success_count += 1
        
        print(f"\nProcessed {success_count}/{total_count} files successfully")
        if success_count < total_count:
            sys.exit(1)
            
    except FileNotFoundError as e:
        print(f"Error: {e}", file=sys.stderr)
        sys.exit(1)
    except Exception as e:
        print(f"Error: {e}", file=sys.stderr)
        sys.exit(1)


if __name__ == '__main__':
    main()
