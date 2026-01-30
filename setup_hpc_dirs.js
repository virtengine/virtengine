// Create HPC proto directories and copy files
const fs = require('fs');
const path = require('path');

// Create directories
const dirs = [
    'sdk/proto/node/virtengine/hpc/v1',
    'sdk/go/node/hpc/v1',
];

dirs.forEach(dir => {
    fs.mkdirSync(dir, { recursive: true });
    console.log('Created:', dir);
});

// Copy proto files
const protoFiles = {
    'hpc_types.proto.txt': 'sdk/proto/node/virtengine/hpc/v1/types.proto',
    'hpc_tx.proto.txt': 'sdk/proto/node/virtengine/hpc/v1/tx.proto',
    'hpc_query.proto.txt': 'sdk/proto/node/virtengine/hpc/v1/query.proto',
    'hpc_genesis.proto.txt': 'sdk/proto/node/virtengine/hpc/v1/genesis.proto',
};

for (const [src, dst] of Object.entries(protoFiles)) {
    if (fs.existsSync(src)) {
        fs.copyFileSync(src, dst);
        console.log('Copied:', src, '->', dst);
    } else {
        console.log('Warning: Source file not found:', src);
    }
}

console.log('\nHPC proto files created successfully!');
