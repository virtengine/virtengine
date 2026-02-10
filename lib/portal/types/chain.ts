/**
 * Chain Types
 * VE-700: Chain RPC/WebSocket integration
 *
 * @packageDocumentation
 */

/**
 * Chain connection state
 */
export interface ChainState {
  /**
   * Whether connected to chain
   */
  isConnected: boolean;

  /**
   * Whether connecting
   */
  isConnecting: boolean;

  /**
   * Current block height
   */
  blockHeight: number;

  /**
   * Chain ID
   */
  chainId: string;

  /**
   * Network name
   */
  networkName: string;

  /**
   * Connection latency in ms
   */
  latencyMs: number | null;

  /**
   * Active subscriptions
   */
  subscriptions: EventSubscription[];

  /**
   * Connection error
   */
  error: ChainError | null;
}

/**
 * Chain configuration
 */
export interface ChainConfig {
  /**
   * WebSocket endpoint
   */
  wsEndpoint: string;

  /**
   * REST endpoint
   */
  restEndpoint: string;

  /**
   * Chain ID
   */
  chainId: string;

  /**
   * Reconnect on disconnect
   */
  autoReconnect: boolean;

  /**
   * Reconnect delay in ms
   */
  reconnectDelayMs: number;

  /**
   * Maximum reconnect attempts
   */
  maxReconnectAttempts: number;

  /**
   * Heartbeat interval in ms
   */
  heartbeatIntervalMs: number;

  /**
   * Request timeout in ms
   */
  requestTimeoutMs: number;
}

/**
 * Event subscription
 */
export interface EventSubscription {
  /**
   * Subscription ID
   */
  id: string;

  /**
   * Event query (Tendermint event query format)
   */
  query: string;

  /**
   * Subscription status
   */
  status: SubscriptionStatus;

  /**
   * When subscription was created
   */
  createdAt: number;

  /**
   * Events received count
   */
  eventsReceived: number;
}

/**
 * Subscription status
 */
export type SubscriptionStatus =
  | 'pending'
  | 'active'
  | 'paused'
  | 'error'
  | 'closed';

/**
 * Query client interface
 */
export interface QueryClient {
  /**
   * Query account info
   */
  queryAccount(address: string): Promise<AccountInfo>;

  /**
   * Query balance
   */
  queryBalance(address: string, denom: string): Promise<Balance>;

  /**
   * Query identity
   */
  queryIdentity(address: string): Promise<IdentityInfo>;

  /**
   * Query offering
   */
  queryOffering(id: string): Promise<OfferingInfo>;

  /**
   * Query order
   */
  queryOrder(id: string): Promise<OrderInfo>;

  /**
   * Query job
   */
  queryJob(id: string): Promise<JobInfo>;

  /**
   * Query provider
   */
  queryProvider(address: string): Promise<ProviderInfo>;

  /**
   * Generic query
   */
  query<T>(path: string, params?: Record<string, string>): Promise<T>;
}

/**
 * Account info
 */
export interface AccountInfo {
  address: string;
  publicKey: string | null;
  accountNumber: number;
  sequence: number;
}

/**
 * Balance info
 */
export interface Balance {
  denom: string;
  amount: string;
}

/**
 * Identity info (from chain)
 */
export interface IdentityInfo {
  address: string;
  status: string;
  score: number;
  modelVersion: string;
  updatedAt: number;
  blockHeight: number;
}

/**
 * Offering info (from chain)
 */
export interface OfferingInfo {
  id: string;
  providerAddress: string;
  status: string;
  metadata: Record<string, unknown>;
  createdAt: number;
}

/**
 * Order info (from chain)
 */
export interface OrderInfo {
  id: string;
  offeringId: string;
  customerAddress: string;
  providerAddress: string;
  state: string;
  createdAt: number;
}

/**
 * Job info (from chain)
 */
export interface JobInfo {
  id: string;
  customerAddress: string;
  providerAddress: string;
  status: string;
  createdAt: number;
}

/**
 * Provider info (from chain)
 */
export interface ProviderInfo {
  address: string;
  status: string;
  reliabilityScore: number;
  registeredAt: number;
}

/**
 * Transaction result
 */
export interface TransactionResult {
  /**
   * Transaction hash
   */
  txHash: string;

  /**
   * Block height (after confirmation)
   */
  blockHeight: number | null;

  /**
   * Transaction code (0 = success)
   */
  code: number;

  /**
   * Raw log
   */
  rawLog: string;

  /**
   * Parsed events
   */
  events: TransactionEvent[];

  /**
   * Gas used
   */
  gasUsed: number;

  /**
   * Gas wanted
   */
  gasWanted: number;
}

/**
 * Transaction event
 */
export interface TransactionEvent {
  /**
   * Event type
   */
  type: string;

  /**
   * Event attributes
   */
  attributes: { key: string; value: string }[];
}

/**
 * Chain event (from subscription)
 */
export interface ChainEvent {
  /**
   * Event query that matched
   */
  query: string;

  /**
   * Event type
   */
  type: string;

  /**
   * Event attributes
   */
  attributes: Record<string, string>;

  /**
   * Block height
   */
  blockHeight: number;

  /**
   * Transaction hash (if from tx)
   */
  txHash?: string;

  /**
   * Timestamp
   */
  timestamp: number;
}

/**
 * Chain error
 */
export interface ChainError {
  /**
   * Error code
   */
  code: ChainErrorCode;

  /**
   * Human-readable message
   */
  message: string;

  /**
   * Original error
   */
  originalError?: Error;
}

/**
 * Chain error codes
 */
export type ChainErrorCode =
  | 'connection_failed'
  | 'connection_lost'
  | 'timeout'
  | 'invalid_response'
  | 'subscription_failed'
  | 'query_failed'
  | 'broadcast_failed'
  | 'unknown';

/**
 * Default chain configuration
 */
export const defaultChainConfig: ChainConfig = {
  wsEndpoint: 'wss://rpc.virtengine.com/websocket',
  restEndpoint: 'https://api.virtengine.com',
  chainId: 'virtengine-1',
  autoReconnect: true,
  reconnectDelayMs: 1000,
  maxReconnectAttempts: 10,
  heartbeatIntervalMs: 30000,
  requestTimeoutMs: 30000,
};

/**
 * Initial chain state
 */
export const initialChainState: ChainState = {
  isConnected: false,
  isConnecting: false,
  blockHeight: 0,
  chainId: '',
  networkName: '',
  latencyMs: null,
  subscriptions: [],
  error: null,
};
