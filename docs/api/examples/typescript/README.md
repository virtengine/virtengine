# VirtEngine TypeScript SDK Examples

This directory contains TypeScript examples for interacting with the VirtEngine API.

## Prerequisites

```bash
# Install the VirtEngine SDK
npm install @virtengine/sdk

# Or with yarn
yarn add @virtengine/sdk
```

## Examples

### Basic Query Client

```typescript
import { VirtEngineClient } from '@virtengine/sdk';

async function main() {
  const client = new VirtEngineClient({
    rpcEndpoint: 'https://api.virtengine.com',
  });

  // Query market orders
  const orders = await client.market.orders({
    filters: { state: 'open' },
    pagination: { limit: 10 },
  });

  console.log(`Found ${orders.orders.length} open orders`);
  for (const order of orders.orders) {
    console.log(`  Order: ${order.orderId.owner}/${order.orderId.dseq}`);
  }

  // Query identity
  try {
    const identity = await client.veid.identity({
      accountAddress: 'virtengine1abc...',
    });
    console.log(`Identity score: ${identity.identity?.score?.overall}`);
  } catch (error) {
    console.log('Identity not found');
  }
}

main().catch(console.error);
```

### Transaction Signing

```typescript
import { VirtEngineClient, SigningClient } from '@virtengine/sdk';
import { DirectSecp256k1HdWallet } from '@cosmjs/proto-signing';

async function createBid() {
  // Create wallet from mnemonic
  const wallet = await DirectSecp256k1HdWallet.fromMnemonic(
    'your mnemonic words here...',
    { prefix: 'virtengine' }
  );

  // Create signing client
  const client = await SigningClient.connectWithSigner(
    'https://api.virtengine.com',
    wallet,
    { gasPrice: { denom: 'uvirt', amount: '0.025' } }
  );

  // Create bid message
  const msg = {
    typeUrl: '/virtengine.market.v2beta1.MsgCreateBid',
    value: {
      orderId: {
        owner: 'virtengine1owner...',
        dseq: '12345',
        gseq: 1,
        oseq: 1,
      },
      provider: 'virtengine1provider...',
      price: { denom: 'uvirt', amount: '950' },
      deposit: { denom: 'uvirt', amount: '500000' },
    },
  };

  // Sign and broadcast
  const result = await client.signAndBroadcast(
    'virtengine1provider...',
    [msg],
    'auto',
    'Create bid'
  );

  console.log(`Transaction hash: ${result.transactionHash}`);
}

createBid().catch(console.error);
```

### Query with Pagination

```typescript
import { VirtEngineClient } from '@virtengine/sdk';

async function getAllOrders() {
  const client = new VirtEngineClient({
    rpcEndpoint: 'https://api.virtengine.com',
  });

  const allOrders: any[] = [];
  let nextKey: Uint8Array | undefined;

  do {
    const response = await client.market.orders({
      pagination: {
        limit: 100,
        key: nextKey,
      },
    });

    allOrders.push(...response.orders);
    nextKey = response.pagination?.nextKey;

    console.log(`Fetched ${response.orders.length} orders, total: ${allOrders.length}`);
  } while (nextKey && nextKey.length > 0);

  console.log(`Total orders: ${allOrders.length}`);
  return allOrders;
}

getAllOrders().catch(console.error);
```

### VEID Scope Upload

```typescript
import { VirtEngineClient, encryptPayload } from '@virtengine/sdk';
import * as nacl from 'tweetnacl';

interface IdentityPayload {
  type: string;
  data: string;
  timestamp: number;
}

async function uploadScope() {
  const client = new VirtEngineClient({
    rpcEndpoint: 'https://api.virtengine.com',
  });

  // Prepare identity payload
  const payload: IdentityPayload = {
    type: 'facial_biometric',
    data: 'base64_encoded_biometric_data',
    timestamp: Date.now(),
  };

  // Get validator public key (would be fetched from chain)
  const validatorPubKey = new Uint8Array(32); // ... actual key

  // Encrypt payload
  const envelope = await encryptPayload(
    JSON.stringify(payload),
    validatorPubKey,
    'X25519-XSalsa20-Poly1305'
  );

  // Create and sign message
  const msg = {
    typeUrl: '/virtengine.veid.v1.MsgUploadScope',
    value: {
      sender: 'virtengine1sender...',
      scopeId: `scope_${Date.now()}`,
      scopeType: 'FACIAL_BIOMETRIC',
      encryptedPayload: envelope,
    },
  };

  console.log('Scope upload message prepared:', msg);
  // Sign and broadcast with wallet...
}

// Helper function to encrypt payload
async function encryptPayload(
  plaintext: string,
  recipientPubKey: Uint8Array,
  algorithm: string
): Promise<any> {
  // Generate ephemeral keypair
  const ephemeralKeypair = nacl.box.keyPair();

  // Generate nonce
  const nonce = nacl.randomBytes(24);

  // Encrypt
  const message = new TextEncoder().encode(plaintext);
  const ciphertext = nacl.box(
    message,
    nonce,
    recipientPubKey,
    ephemeralKeypair.secretKey
  );

  return {
    recipientFingerprint: Buffer.from(recipientPubKey.slice(0, 8)).toString('hex'),
    algorithm,
    ciphertext: Buffer.from(ciphertext).toString('base64'),
    nonce: Buffer.from(nonce).toString('base64'),
    ephemeralPublicKey: Buffer.from(ephemeralKeypair.publicKey).toString('base64'),
  };
}

uploadScope().catch(console.error);
```

### MFA Flow

```typescript
import { VirtEngineClient, SigningClient } from '@virtengine/sdk';
import * as readline from 'readline';

async function submitWithMFA() {
  const client = new VirtEngineClient({
    rpcEndpoint: 'https://api.virtengine.com',
  });

  const address = 'virtengine1abc...';
  const txType = 'veid.MsgSubmitScope';

  // Check if MFA is required
  const mfaCheck = await client.mfa.mfaRequired({
    address,
    transactionType: txType,
  });

  if (mfaCheck.required) {
    console.log(`MFA required: ${mfaCheck.factorsNeeded} factor(s) needed`);
    console.log(`Allowed factors: ${mfaCheck.allowedFactors.join(', ')}`);

    // Get enrolled factors
    const enrollments = await client.mfa.factorEnrollments({ address });
    console.log('\nEnrolled factors:');
    for (const factor of enrollments.enrollments) {
      console.log(`  - ${factor.label} (${factor.factorType})`);
    }

    // Create challenge
    const challenge = await createMFAChallenge(client, 'totp', txType);
    console.log(`\nChallenge created: ${challenge.challengeId}`);
    console.log(`Expires at: ${challenge.expiresAt}`);

    // Get code from user
    const code = await promptUser('Enter your 2FA code: ');

    // Verify challenge
    const session = await verifyMFAChallenge(client, challenge.challengeId, code);
    console.log(`\nSession created: ${session.sessionId}`);
    console.log(`Session expires: ${session.expiresAt}`);

    // Now submit transaction with session token
    // Include X-MFA-Session header or session_id in transaction metadata
    return session;
  }

  console.log('MFA not required');
  return null;
}

async function createMFAChallenge(
  client: VirtEngineClient,
  factorType: string,
  transactionType: string
) {
  // This would be a transaction message
  return {
    challengeId: 'chal_abc123',
    expiresAt: new Date(Date.now() + 5 * 60 * 1000).toISOString(),
  };
}

async function verifyMFAChallenge(
  client: VirtEngineClient,
  challengeId: string,
  response: string
) {
  // This would be a transaction message
  return {
    sessionId: 'sess_xyz789',
    expiresAt: new Date(Date.now() + 10 * 60 * 1000).toISOString(),
  };
}

function promptUser(question: string): Promise<string> {
  const rl = readline.createInterface({
    input: process.stdin,
    output: process.stdout,
  });

  return new Promise((resolve) => {
    rl.question(question, (answer) => {
      rl.close();
      resolve(answer);
    });
  });
}

submitWithMFA().catch(console.error);
```

### WebSocket Subscriptions

```typescript
import WebSocket from 'ws';

interface SubscribeRequest {
  jsonrpc: string;
  method: string;
  params: { query: string };
  id: number;
}

async function subscribeToEvents() {
  const ws = new WebSocket('wss://api.virtengine.com/websocket');

  ws.on('open', () => {
    console.log('Connected to WebSocket');

    // Subscribe to new blocks
    const subscribeBlocks: SubscribeRequest = {
      jsonrpc: '2.0',
      method: 'subscribe',
      params: { query: "tm.event='NewBlock'" },
      id: 1,
    };
    ws.send(JSON.stringify(subscribeBlocks));

    // Subscribe to market transactions
    const subscribeMarket: SubscribeRequest = {
      jsonrpc: '2.0',
      method: 'subscribe',
      params: { query: "tm.event='Tx' AND message.module='market'" },
      id: 2,
    };
    ws.send(JSON.stringify(subscribeMarket));
  });

  ws.on('message', (data) => {
    const message = JSON.parse(data.toString());
    
    if (message.result?.data?.value?.block) {
      const height = message.result.data.value.block.header.height;
      console.log(`New block: ${height}`);
    } else if (message.result?.events) {
      console.log('Transaction event:', message.result.events);
    }
  });

  ws.on('error', (error) => {
    console.error('WebSocket error:', error);
  });

  ws.on('close', () => {
    console.log('WebSocket closed');
  });
}

subscribeToEvents();
```

### Error Handling

```typescript
import { VirtEngineClient } from '@virtengine/sdk';

interface VirtEngineError {
  code: string;
  message: string;
  category: string;
  retryable: boolean;
  context?: Record<string, any>;
}

async function handleErrors() {
  const client = new VirtEngineClient({
    rpcEndpoint: 'https://api.virtengine.com',
  });

  try {
    const order = await client.market.order({
      id: {
        owner: 'virtengine1nonexistent...',
        dseq: '99999',
        gseq: 1,
        oseq: 1,
      },
    });
  } catch (error: any) {
    // Parse error response
    const virtError = parseError(error);

    switch (virtError.category) {
      case 'not_found':
        console.log('Order not found - check the order ID');
        break;
      case 'validation':
        console.log('Invalid request:', virtError.message);
        break;
      case 'rate_limit':
        if (virtError.context?.retry_after) {
          console.log(`Rate limited, retry after ${virtError.context.retry_after}s`);
        }
        break;
      case 'unauthorized':
        console.log('Authentication required');
        break;
      default:
        console.log('Error:', virtError.message);
    }

    // Retry if retryable
    if (virtError.retryable) {
      console.log('This error is retryable');
    }
  }
}

function parseError(error: any): VirtEngineError {
  // Try to parse structured error
  if (error.response?.data?.error) {
    return error.response.data.error;
  }

  // Fallback
  return {
    code: 'unknown:0',
    message: error.message || 'Unknown error',
    category: 'internal',
    retryable: false,
  };
}

handleErrors().catch(console.error);
```

### REST API with Fetch

```typescript
const API_BASE = 'https://api.virtengine.com';

interface PaginatedResponse<T> {
  data: T[];
  pagination: {
    nextKey?: string;
    total?: string;
  };
}

async function fetchOrders(
  filters?: { state?: string; owner?: string },
  pagination?: { limit?: number; key?: string }
): Promise<any> {
  const params = new URLSearchParams();

  if (filters?.state) params.append('filters.state', filters.state);
  if (filters?.owner) params.append('filters.owner', filters.owner);
  if (pagination?.limit) params.append('pagination.limit', pagination.limit.toString());
  if (pagination?.key) params.append('pagination.key', pagination.key);

  const response = await fetch(
    `${API_BASE}/virtengine/market/v2beta1/orders/list?${params}`,
    {
      headers: {
        'Accept': 'application/json',
        // 'x-api-key': 'YOUR_API_KEY',  // Optional for higher rate limits
      },
    }
  );

  if (!response.ok) {
    throw new Error(`HTTP error: ${response.status}`);
  }

  // Check rate limit headers
  const remaining = response.headers.get('X-RateLimit-Remaining');
  if (remaining && parseInt(remaining) < 10) {
    console.warn(`Low rate limit: ${remaining} requests remaining`);
  }

  return response.json();
}

async function main() {
  const openOrders = await fetchOrders({ state: 'open' }, { limit: 10 });
  console.log(`Found ${openOrders.orders.length} open orders`);
}

main().catch(console.error);
```

## Running Examples

```bash
# Install dependencies
npm install

# Run with ts-node
npx ts-node example.ts

# Or compile and run
npx tsc example.ts
node example.js
```

## See Also

- [API Reference](../../reference/)
- [Authentication Guide](../../guides/authentication.md)
- [Go Examples](../go/)
- [cURL Examples](../curl.md)
