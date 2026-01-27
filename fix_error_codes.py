#!/usr/bin/env python3
import re

modules = {
    r'x\mfa\types\errors.go': 200,
    r'x\encryption\types\errors.go': 300,
    r'x\roles\types\errors.go': 400,
    r'x\settlement\types\errors.go': 500,
    r'x\config\types\errors.go': 600,
    r'x\delegation\types\errors.go': 800,
    r'x\enclave\types\errors.go': 900,
    r'x\fraud\types\errors.go': 1000,
    r'x\hpc\types\errors.go': 1100,
    r'x\market\types\marketplace\errors.go': 1200,
    r'x\review\types\errors.go': 1300,
    r'x\staking\types\errors.go': 1400,
    r'pkg\artifact_store\errors.go': 1500,
    r'sdk\go\node\bme\v1\errors.go': 1600,
}

for file_path, start_code in modules.items():
    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            content = f.read()
        
        # Use list to allow mutation in nested function
        code = [start_code]
        def replacer(match):
            result = f'{match.group(1)}{code[0]},'
            code[0] += 1
            return result
        
        # Match Register(..., NUMBER, pattern
        new_content = re.sub(r'(Register\([^,]+,\s*)\d+(,)', replacer, content)
        
        with open(file_path, 'w', encoding='utf-8', newline='') as f:
            f.write(new_content)
        
        print(f'Fixed {file_path} starting at {start_code}')
    except Exception as e:
        print(f'Error fixing {file_path}: {e}')
