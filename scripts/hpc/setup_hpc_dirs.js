// Create HPC proto directories and copy files
const fs = require('fs');
const path = require('path');

const repoRoot = path.resolve(__dirname, '..', '..');
const protoDir = path.join(__dirname, 'proto');

// Create directories
const dirs = [
    path.join(repoRoot, 'sdk/proto/node/virtengine/hpc/v1'),
    path.join(repoRoot, 'sdk/go/node/hpc/v1'),
];

dirs.forEach(dir => {
    fs.mkdirSync(dir, { recursive: true });
    console.log('Created:', dir);
});

// Copy proto files
const protoFiles = {
    'hpc_types.proto.txt': path.join(repoRoot, 'sdk/proto/node/virtengine/hpc/v1/types.proto'),
    'hpc_tx.proto.txt': path.join(repoRoot, 'sdk/proto/node/virtengine/hpc/v1/tx.proto'),
    'hpc_query.proto.txt': path.join(repoRoot, 'sdk/proto/node/virtengine/hpc/v1/query.proto'),
    'hpc_genesis.proto.txt': path.join(repoRoot, 'sdk/proto/node/virtengine/hpc/v1/genesis.proto'),
};

for (const [src, dst] of Object.entries(protoFiles)) {
    const srcPath = path.join(protoDir, src);
    if (fs.existsSync(srcPath)) {
        fs.copyFileSync(srcPath, dst);
        console.log('Copied:', src, '->', dst);
    } else {
        console.log('Warning: Source file not found:', srcPath);
    }
}

console.log('\nHPC proto files created successfully!');
