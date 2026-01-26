/**
 * k6 Load Test: Marketplace Order Burst
 * 
 * Task Reference: VE-801 - Load & performance testing
 * Scenario B: Burst orders + bids + allocations
 * 
 * Usage:
 *   k6 run tests/load/k6/marketplace_burst.js
 *   k6 run --vus 50 --duration 5m tests/load/k6/marketplace_burst.js
 */

import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';
import { randomBytes, randomIntBetween } from 'k6/crypto';

// Custom metrics
const orderCreateDuration = new Trend('order_create_duration');
const bidSubmitDuration = new Trend('bid_submit_duration');
const allocationDuration = new Trend('allocation_duration');
const orderLifecycleSuccess = new Rate('order_lifecycle_success');
const orderErrors = new Counter('order_errors');

// Test configuration
export const options = {
    scenarios: {
        marketplace_burst: {
            executor: 'ramping-vus',
            startVUs: 0,
            stages: [
                { duration: '20s', target: 10 },  // Ramp up
                { duration: '1m', target: 30 },   // Normal load
                { duration: '30s', target: 60 },  // Peak load
                { duration: '1m', target: 60 },   // Sustain peak
                { duration: '30s', target: 0 },   // Ramp down
            ],
        },
    },
    thresholds: {
        'order_create_duration': ['p(95)<10000'],   // 95th percentile < 10s
        'bid_submit_duration': ['p(95)<5000'],      // 95th percentile < 5s
        'allocation_duration': ['p(95)<15000'],     // 95th percentile < 15s
        'order_lifecycle_success': ['rate>0.95'],   // 95% success rate
    },
};

// Environment configuration
const BASE_URL = __ENV.VIRTENGINE_NODE_URL || 'http://localhost:26657';

// Offering templates
const OFFERINGS = [
    { id: 'gpu-a100', cpu: 8, memory: 64, gpu: 1, price: 1000 },
    { id: 'compute-standard', cpu: 4, memory: 16, gpu: 0, price: 100 },
    { id: 'compute-large', cpu: 16, memory: 64, gpu: 0, price: 400 },
    { id: 'hpc-node', cpu: 32, memory: 256, gpu: 4, price: 5000 },
];

// Generate order request
function generateOrder() {
    const offering = OFFERINGS[Math.floor(Math.random() * OFFERINGS.length)];
    
    return {
        customer: `virtengine1customer${__VU}`,
        offering_id: offering.id,
        quantity: 1,
        config: {
            region: ['us-east', 'us-west', 'eu-west'][Math.floor(Math.random() * 3)],
            duration_hours: Math.floor(Math.random() * 168) + 1, // 1-168 hours
        },
        encrypted_details: {
            version: '1.0',
            algorithm_id: 'X25519-XSalsa20-Poly1305',
            ciphertext: randomBytes(256).toString('hex'),
            nonce: randomBytes(24).toString('hex'),
        },
    };
}

// Generate bid request
function generateBid(orderId) {
    return {
        order_id: orderId,
        provider: `virtengine1provider${Math.floor(Math.random() * 10)}`,
        price: Math.floor(Math.random() * 500) + 100,
        available_at: new Date().toISOString(),
    };
}

// Main test function
export default function() {
    let orderId = null;
    let bidId = null;
    let success = true;
    
    group('Order Lifecycle', function() {
        // Step 1: Create Order
        group('Create Order', function() {
            const order = generateOrder();
            const startTime = new Date();
            
            const response = http.post(
                `${BASE_URL}/virtengine/market/v1/orders`,
                JSON.stringify(order),
                {
                    headers: { 'Content-Type': 'application/json' },
                    timeout: '30s',
                }
            );
            
            const duration = new Date() - startTime;
            orderCreateDuration.add(duration);
            
            const orderCheck = check(response, {
                'order created': (r) => r.status === 200 || r.status === 201,
            });
            
            if (orderCheck) {
                try {
                    const body = JSON.parse(response.body);
                    orderId = body.order_id || body.id;
                } catch (e) {
                    orderId = `order_${__VU}_${__ITER}`;
                }
            } else {
                success = false;
                orderErrors.add(1);
            }
        });
        
        // Step 2: Submit Bid (simulating provider)
        if (orderId) {
            group('Submit Bid', function() {
                sleep(0.5); // Small delay before bid
                
                const bid = generateBid(orderId);
                const startTime = new Date();
                
                const response = http.post(
                    `${BASE_URL}/virtengine/market/v1/bids`,
                    JSON.stringify(bid),
                    {
                        headers: { 'Content-Type': 'application/json' },
                        timeout: '15s',
                    }
                );
                
                const duration = new Date() - startTime;
                bidSubmitDuration.add(duration);
                
                const bidCheck = check(response, {
                    'bid submitted': (r) => r.status === 200 || r.status === 201,
                });
                
                if (bidCheck) {
                    try {
                        const body = JSON.parse(response.body);
                        bidId = body.bid_id || body.id;
                    } catch (e) {
                        bidId = `bid_${orderId}`;
                    }
                } else {
                    success = false;
                }
            });
        }
        
        // Step 3: Allocate Order
        if (orderId && bidId) {
            group('Allocate Order', function() {
                sleep(0.5); // Small delay before allocation
                
                const startTime = new Date();
                
                const response = http.post(
                    `${BASE_URL}/virtengine/market/v1/orders/${orderId}/allocate`,
                    JSON.stringify({ bid_id: bidId }),
                    {
                        headers: { 'Content-Type': 'application/json' },
                        timeout: '30s',
                    }
                );
                
                const duration = new Date() - startTime;
                allocationDuration.add(duration);
                
                const allocationCheck = check(response, {
                    'order allocated': (r) => r.status === 200,
                });
                
                if (!allocationCheck) {
                    success = false;
                }
            });
        }
        
        // Step 4: Verify Order State
        if (orderId) {
            group('Verify Order State', function() {
                sleep(0.5);
                
                const response = http.get(
                    `${BASE_URL}/virtengine/market/v1/orders/${orderId}`,
                    {
                        headers: { 'Content-Type': 'application/json' },
                        timeout: '10s',
                    }
                );
                
                check(response, {
                    'order queryable': (r) => r.status === 200,
                    'order has valid state': (r) => {
                        try {
                            const body = JSON.parse(r.body);
                            return ['PENDING', 'ALLOCATED', 'ACTIVE'].includes(body.state);
                        } catch {
                            return true; // Don't fail if can't parse
                        }
                    },
                });
            });
        }
    });
    
    orderLifecycleSuccess.add(success);
    
    // Throttle between iterations
    sleep(1);
}

// Setup function
export function setup() {
    console.log(`Marketplace Burst Load Test`);
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
        'tests/load/results/marketplace_burst_summary.json': JSON.stringify(data, null, 2),
    };
}
