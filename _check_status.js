const { execSync } = require('child_process');
const path = require('path');
const fs = require('fs');

const workdir = 'C:\\Users\\jonathan\\AppData\\Local\\Temp\\vibe-kanban\\worktrees\\42bf-test-hpc-impleme\\virtengine';
process.chdir(workdir);

console.log('============================================================');
console.log('1. CURRENT BRANCH:');
console.log('============================================================');
try {
    const branch = execSync('git branch --show-current', { encoding: 'utf8', timeout: 30000 }).trim();
    console.log(branch || '(empty output)');
} catch (e) {
    console.log('ERROR:', e.message);
}

console.log('\n============================================================');
console.log('2. GIT STATUS (modified/new files):');
console.log('============================================================');
try {
    const status = execSync('git status --short', { encoding: 'utf8', timeout: 30000 }).trim();
    console.log(status || '(no changes)');
} catch (e) {
    console.log('ERROR:', e.message);
}

console.log('\n============================================================');
console.log('3. GO BUILD ERRORS:');
console.log('============================================================');
try {
    const result = execSync('go build -tags=e2e.integration ./tests/e2e/...', { 
        encoding: 'utf8', 
        timeout: 300000,
        stdio: ['pipe', 'pipe', 'pipe']
    });
    console.log('BUILD SUCCESSFUL - No compilation errors');
} catch (e) {
    console.log('BUILD FAILED:');
    if (e.stderr) console.log(e.stderr);
    if (e.stdout) console.log(e.stdout);
    if (e.message) console.log(e.message);
}
