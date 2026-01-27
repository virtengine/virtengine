#!/usr/bin/env python3
import re

# Fix veid module - resequence ALL codes starting from 100
file_path = r'x\veid\types\errors.go'

with open(file_path, 'r', encoding='utf-8') as f:
    content = f.read()

code = [100]
def replacer(match):
    result = f'{match.group(1)}{code[0]},'
    code[0] += 1
    return result

# Match Register(ModuleName, NUMBER, pattern
new_content = re.sub(r'(Register\(ModuleName,\s*)\d+(,)', replacer, content)

with open(file_path, 'w', encoding='utf-8', newline='') as f:
    f.write(new_content)

print(f'Fixed {file_path} - {code[0]-100} error codes renumbered from 100')
