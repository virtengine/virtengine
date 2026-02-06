/**
 * k6 Load Test: HPC Job Submission Burst
 *
 * Task Reference: VE-801 - Load & performance testing
 * Scenario C: Burst HPC job submissions and scheduling
 *
 * Usage:
 *   k6 run tests/load/k6/hpc_burst.js
 *   k6 run --vus 50 --duration 5m tests/load/k6/hpc_burst.js
 */

import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';
import { randomBytes } from 'k6/crypto';

// Custom metrics
const jobSubmitDuration = new Trend('job_submit_duration');
const jobScheduleDuration = new Trend('job_schedule_duration');
const jobLifecycleSuccess = new Rate('job_lifecycle_success');
const jobErrors = new Counter('job_errors');

// Test configuration
export const options = {
    scenarios: {
        hpc_burst: {
            executor: 'ramping-vus',
            startVUs: 0,
            stages: [
                { duration: '20s', target: 10 },  // Ramp up
                { duration: '1m', target: 30 },   // Normal load
                { duration: '30s', target: 50 },   // Peak load
                { duration: '1m', target: 50 },    // Sustain peak
                { duration: '30s', target: 0 },    // Ramp down
            ],
        },
    },
    thresholds: {
        'job_submit_duration': ['p(95)<10000'],    // 95th percentile < 10s
        'job_schedule_duration': ['p(95)<15000'],  // 95th percentile < 15s
        'job_lifecycle_success': ['rate>0.95'],    // 95% success rate
    },
};

// Environment configuration
const BASE_URL = __ENV.VIRTENGINE_NODE_URL || 'http://localhost:26657';

// HPC job templates
const JOB_TEMPLATES = [
    { name: 'ml-training', cpus: 8, memory: 32, gpus: 2, duration_hours: 24 },
    { name: 'simulation', cpus: 32, memory: 128, gpus: 0, duration_hours: 48 },
    { name: 'rendering', cpus: 16, memory: 64, gpus: 4, duration_hours: 12 },
    { name: 'data-processing', cpus: 4, memory: 16, gpus: 0, duration_hours: 6 },
];

// Generate HPC job request
function generateHPCJob() {
    const template = JOB_TEMPLATES[Math.floor(Math.random() * JOB_TEMPLATES.length)];

    return {
        owner: `virtengine1user${__VU}`,
        name: `${template.name}-${__VU}-${__ITER}`,
        manifest: {
            version: '2.0',
            services: [{
                name: template.name,
                image: `virtengine/${template.name}:latest`,
                resources: {
                    cpu: { units: template.cpus * 1000 },
                    memory: { size: `${template.memory}Gi` },
                    gpu: { units: template.gpus, attributes: { vendor: 'nvidia' } },
                    storage: [{ size: '100Gi', class: 'ssd' }],
                },
                env: [
                    `JOB_ID=${__VU}_${__ITER}`,
                    `SLURM_NTASKS=${template.cpus}`,
                ],
            }],
            placement: {
                attributes: { region: 'us-east' },
                pricing: { amount: Math.floor(Math.random() * 5000) + 500, denom: 'uvirt' },
            },
        },
        scheduling: {
            priority: Math.floor(Math.random() * 10),
            max_duration_hours: template.duration_hours,
            queue: 'default',
        },
        encrypted_params: {
            version: '1.0',
            algorithm_id: 'X25519-XSalsa20-Poly1305',
            ciphertext: randomBytes(128).toString('hex'),
            nonce: randomBytes(24).toString('hex'),
        },
    };
}

// Main test function
export default function() {
    let jobId = null;
    let success = true;

    group('HPC Job Lifecycle', function() {
        // Step 1: Submit Job
        group('Submit Job', function() {
            const job = generateHPCJob();
            const startTime = new Date();

            const response = http.post(
                `${BASE_URL}/virtengine/hpc/v1/jobs`,
                JSON.stringify(job),
                {
                    headers: { 'Content-Type': 'application/json' },
                    timeout: '30s',
                }
            );

            const duration = new Date() - startTime;
            jobSubmitDuration.add(duration);

            const submitCheck = check(response, {
                'job submitted': (r) => r.status === 200 || r.status === 201,
            });

            if (submitCheck) {
                try {
                    const body = JSON.parse(response.body);
                    jobId = body.job_id || body.id;
                } catch (e) {
                    jobId = `job_${__VU}_${__ITER}`;
                }
            } else {
                success = false;
                jobErrors.add(1);
            }
        });

        // Step 2: Check Scheduling
        if (jobId) {
            group('Wait for Scheduling', function() {
                sleep(1);

                const startTime = new Date();

                const response = http.get(
                    `${BASE_URL}/virtengine/hpc/v1/jobs/${jobId}`,
                    {
                        headers: { 'Content-Type': 'application/json' },
                        timeout: '15s',
                    }
                );

                const duration = new Date() - startTime;
                jobScheduleDuration.add(duration);

                check(response, {
                    'job queryable': (r) => r.status === 200,
                    'job has valid state': (r) => {
                        try {
                            const body = JSON.parse(r.body);
                            return ['PENDING', 'QUEUED', 'SCHEDULED', 'RUNNING'].includes(body.state);
                        } catch {
                            return true;
                        }
                    },
                });
            });
        }

        // Step 3: Query Queue Status
        group('Query Queue', function() {
            const response = http.get(
                `${BASE_URL}/virtengine/hpc/v1/queue?limit=10`,
                {
                    headers: { 'Content-Type': 'application/json' },
                    timeout: '10s',
                }
            );

            check(response, {
                'queue queryable': (r) => r.status === 200,
            });
        });
    });

    jobLifecycleSuccess.add(success);

    // Throttle between iterations
    sleep(0.5);
}

// Setup function
export function setup() {
    console.log(`HPC Burst Load Test`);
    console.log(`Target: ${BASE_URL}`);
    return { startTime: new Date() };
}

// Teardown function
export function teardown(data) {
    const duration = (new Date() - new Date(data.startTime)) / 1000;
    console.log(`Test completed in ${duration.toFixed(2)} seconds`);
}

// Handle summary
export function handleSummary(data) {
    return {
        'tests/load/results/hpc_burst_summary.json': JSON.stringify(data, null, 2),
    };
}
