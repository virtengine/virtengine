/**
 * k6 Load Test: Identity Scope Upload Burst
 * 
 * Task Reference: VE-801 - Load & performance testing
 * Scenario A: Burst identity scope uploads
 * 
 * Usage:
 *   k6 run tests/load/k6/identity_burst.js
 *   k6 run --vus 100 --duration 5m tests/load/k6/identity_burst.js
 */

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';
import { randomBytes } from 'k6/crypto';

// Custom metrics
const identityUploadDuration = new Trend('identity_upload_duration');
const identityUploadSuccess = new Rate('identity_upload_success');
const identityUploadErrors = new Counter('identity_upload_errors');

// Test configuration
export const options = {
    scenarios: {
        identity_burst: {
            executor: 'ramping-vus',
            startVUs: 0,
            stages: [
                { duration: '30s', target: 20 },  // Ramp up to 20 VUs
                { duration: '1m', target: 50 },   // Ramp up to 50 VUs
                { duration: '2m', target: 50 },   // Stay at 50 VUs
                { duration: '30s', target: 100 }, // Spike to 100 VUs
                { duration: '1m', target: 100 },  // Stay at 100 VUs
                { duration: '30s', target: 0 },   // Ramp down
            ],
        },
    },
    thresholds: {
        'identity_upload_duration': ['p(95)<5000'], // 95th percentile < 5s
        'identity_upload_success': ['rate>0.99'],   // 99% success rate
        'http_req_failed': ['rate<0.01'],           // Less than 1% HTTP errors
    },
};

// Environment configuration
const BASE_URL = __ENV.VIRTENGINE_NODE_URL || 'http://localhost:26657';
const GRPC_URL = __ENV.VIRTENGINE_GRPC_URL || 'localhost:9090';

// Generate random identity payload
function generateIdentityPayload() {
    const scopes = randomBytes(1024);
    const salt = randomBytes(32);
    const timestamp = new Date().toISOString();
    
    return {
        scopes: scopes,
        salt: salt,
        timestamp: timestamp,
        client_signature: randomBytes(64),
        user_signature: randomBytes(64),
    };
}

// Generate encrypted envelope
function generateEncryptedEnvelope(payload) {
    return {
        version: '1.0',
        algorithm_id: 'X25519-XSalsa20-Poly1305',
        recipient_key_ids: ['validator_key_001'],
        nonce: randomBytes(24),
        ciphertext: randomBytes(payload.scopes.length + 16),
        sender_pub_key: randomBytes(32),
    };
}

// Main test function
export default function() {
    // Generate identity upload request
    const payload = generateIdentityPayload();
    const envelope = generateEncryptedEnvelope(payload);
    
    const requestBody = JSON.stringify({
        owner: `virtengine1user${__VU}${__ITER}`,
        scopes: [
            {
                scope_type: 'DOCUMENT',
                encrypted_data: envelope,
            },
            {
                scope_type: 'SELFIE',
                encrypted_data: generateEncryptedEnvelope({scopes: randomBytes(512)}),
            },
        ],
        salt: payload.salt.toString('hex'),
        client_signature: payload.client_signature.toString('hex'),
        user_signature: payload.user_signature.toString('hex'),
    });
    
    const params = {
        headers: {
            'Content-Type': 'application/json',
        },
        timeout: '30s',
    };
    
    const startTime = new Date();
    
    // Submit identity upload transaction
    // Note: In real test, this would call the chain's broadcast_tx endpoint
    const response = http.post(
        `${BASE_URL}/cosmos/tx/v1beta1/txs`,
        requestBody,
        params
    );
    
    const duration = new Date() - startTime;
    identityUploadDuration.add(duration);
    
    // Check response
    const success = check(response, {
        'status is 200': (r) => r.status === 200,
        'response has tx_hash': (r) => {
            try {
                const body = JSON.parse(r.body);
                return body.tx_response && body.tx_response.txhash;
            } catch {
                return false;
            }
        },
        'no error in response': (r) => {
            try {
                const body = JSON.parse(r.body);
                return !body.tx_response.code || body.tx_response.code === 0;
            } catch {
                return true; // If we can't parse, don't count as error
            }
        },
    });
    
    identityUploadSuccess.add(success);
    if (!success) {
        identityUploadErrors.add(1);
    }
    
    // Small delay between requests
    sleep(0.1);
}

// Setup function - runs once before the test
export function setup() {
    console.log(`Identity Burst Load Test`);
    console.log(`Target: ${BASE_URL}`);
    console.log(`Test starting...`);
    
    // Verify chain is accessible
    const healthCheck = http.get(`${BASE_URL}/health`);
    if (healthCheck.status !== 200) {
        console.warn(`Health check failed: ${healthCheck.status}`);
    }
    
    return { startTime: new Date() };
}

// Teardown function - runs once after the test
export function teardown(data) {
    const duration = (new Date() - new Date(data.startTime)) / 1000;
    console.log(`Test completed in ${duration.toFixed(2)} seconds`);
}

// Handle summary output
export function handleSummary(data) {
    return {
        'stdout': textSummary(data, { indent: ' ', enableColors: true }),
        'tests/load/results/identity_burst_summary.json': JSON.stringify(data, null, 2),
    };
}

function textSummary(data, options) {
    const metrics = data.metrics;
    let output = '\n=== Identity Burst Load Test Results ===\n\n';
    
    if (metrics.identity_upload_duration) {
        const dur = metrics.identity_upload_duration.values;
        output += `Upload Duration:\n`;
        output += `  avg: ${dur.avg.toFixed(2)}ms\n`;
        output += `  p50: ${dur.med.toFixed(2)}ms\n`;
        output += `  p95: ${dur['p(95)'].toFixed(2)}ms\n`;
        output += `  p99: ${dur['p(99)'].toFixed(2)}ms\n\n`;
    }
    
    if (metrics.identity_upload_success) {
        output += `Success Rate: ${(metrics.identity_upload_success.values.rate * 100).toFixed(2)}%\n`;
    }
    
    if (metrics.http_reqs) {
        output += `Total Requests: ${metrics.http_reqs.values.count}\n`;
        output += `Throughput: ${metrics.http_reqs.values.rate.toFixed(2)} req/s\n`;
    }
    
    return output;
}
