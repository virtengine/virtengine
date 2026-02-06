# Task 29C: lib/portal Shared Component Library

**ID:** 29C  
**Title:** feat(portal): Create lib/portal shared component library  
**Priority:** P0 (Critical Blocker)  
**Wave:** 1 (Parallel with 29A)  
**Estimated LOC:** 3000-5000  
**Dependencies:** None  
**Blocking:** 29D, 29E, 29G, 29H, 29I, 29J  

---

## Problem Statement

The `portal/` Next.js application imports from `../../../lib/portal` which **does not exist**. This is a critical build blocker:

```typescript
// portal/src/app/layout.tsx - BROKEN IMPORT
import { WalletProvider } from '../../../lib/portal';  // ❌ Path doesn't exist
```

Without `lib/portal`:
1. Portal cannot build (`Module not found` error)
2. No shared components for wallet connection
3. No VirtEngine SDK wrapper for TypeScript
4. No React hooks for chain queries

### Current State Analysis

```
lib/
└── (empty or missing portal directory)

portal/
├── package.json           ✅ Exists (Next.js 14.1.3)
├── src/
│   ├── app/              ✅ App router pages
│   └── components/       ✅ Some local components
└── imports from lib/portal ❌ BROKEN
```

---

## Acceptance Criteria

### AC-1: Package Structure
- [ ] Create `lib/portal/` directory with valid package.json
- [ ] Configure TypeScript with strict mode
- [ ] Set up build system (tsup or tsc)
- [ ] Configure exports in package.json
- [ ] Add to pnpm workspace

### AC-2: Wallet Connection Components
- [ ] `WalletProvider` - Context provider for wallet state
- [ ] `WalletConnectButton` - Connect/disconnect button
- [ ] `WalletInfo` - Display address, balance
- [ ] `NetworkSelector` - Switch between networks
- [ ] Support Keplr, Leap, Cosmostation wallets

### AC-3: Deployment Components
- [ ] `DeploymentCard` - Single deployment display
- [ ] `DeploymentList` - List of user's deployments
- [ ] `DeploymentStatus` - Status badge with color coding
- [ ] `DeploymentActions` - Start/stop/close actions
- [ ] `ResourceUsage` - CPU/Memory/Storage bars

### AC-4: Provider Components
- [ ] `ProviderCard` - Provider info display
- [ ] `ProviderList` - Available providers grid
- [ ] `ProviderAttributes` - Capabilities display
- [ ] `ProviderStatus` - Online/offline indicator

### AC-5: Lease Components
- [ ] `LeaseCard` - Active lease display
- [ ] `LeaseList` - User's leases
- [ ] `LeaseStatus` - Lease state badge
- [ ] `LeaseActions` - Close/renew actions

### AC-6: Invoice/Billing Components
- [ ] `InvoiceTable` - Invoice listing
- [ ] `InvoiceRow` - Single invoice row
- [ ] `UsageSummary` - Current period usage
- [ ] `PaymentHistory` - Past payments

### AC-7: React Hooks
- [ ] `useVirtEngine` - Chain connection hook
- [ ] `useWallet` - Wallet state hook
- [ ] `useDeployments` - Query deployments
- [ ] `useProviders` - Query providers
- [ ] `useLeases` - Query user's leases
- [ ] `useBalance` - Query token balance

### AC-8: VirtEngine SDK Wrapper
- [ ] `VirtEngineClient` class for chain operations
- [ ] Sign and broadcast transactions
- [ ] Query blockchain state
- [ ] Parse SDL manifests
- [ ] Type-safe message builders

### AC-9: Portal Build Verification
- [ ] `portal/` builds without errors after lib/portal created
- [ ] All imports resolve correctly
- [ ] TypeScript compilation passes
- [ ] No runtime errors on page load

---

## Technical Requirements

### Package Configuration

```json
// lib/portal/package.json
{
  "name": "@virtengine/portal",
  "version": "0.1.0",
  "description": "Shared components and hooks for VirtEngine portal",
  "main": "./dist/index.js",
  "module": "./dist/index.mjs",
  "types": "./dist/index.d.ts",
  "exports": {
    ".": {
      "import": "./dist/index.mjs",
      "require": "./dist/index.js",
      "types": "./dist/index.d.ts"
    },
    "./components": {
      "import": "./dist/components/index.mjs",
      "require": "./dist/components/index.js",
      "types": "./dist/components/index.d.ts"
    },
    "./hooks": {
      "import": "./dist/hooks/index.mjs",
      "require": "./dist/hooks/index.js",
      "types": "./dist/hooks/index.d.ts"
    },
    "./sdk": {
      "import": "./dist/sdk/index.mjs",
      "require": "./dist/sdk/index.js",
      "types": "./dist/sdk/index.d.ts"
    }
  },
  "scripts": {
    "build": "tsup",
    "dev": "tsup --watch",
    "lint": "eslint src/",
    "typecheck": "tsc --noEmit"
  },
  "peerDependencies": {
    "react": "^18.0.0",
    "react-dom": "^18.0.0"
  },
  "dependencies": {
    "@cosmjs/cosmwasm-stargate": "^0.32.2",
    "@cosmjs/proto-signing": "^0.32.2",
    "@cosmjs/stargate": "^0.32.2",
    "@keplr-wallet/types": "^0.12.0",
    "@tanstack/react-query": "^5.17.0",
    "zustand": "^4.5.0"
  },
  "devDependencies": {
    "@types/react": "^18.2.48",
    "tsup": "^8.0.1",
    "typescript": "^5.3.3"
  }
}
```

### Directory Structure

```
lib/portal/
├── package.json
├── tsconfig.json
├── tsup.config.ts
├── src/
│   ├── index.ts                 # Main exports
│   ├── components/
│   │   ├── index.ts             # Component exports
│   │   ├── wallet/
│   │   │   ├── WalletProvider.tsx
│   │   │   ├── WalletConnectButton.tsx
│   │   │   ├── WalletInfo.tsx
│   │   │   └── NetworkSelector.tsx
│   │   ├── deployment/
│   │   │   ├── DeploymentCard.tsx
│   │   │   ├── DeploymentList.tsx
│   │   │   ├── DeploymentStatus.tsx
│   │   │   └── DeploymentActions.tsx
│   │   ├── provider/
│   │   │   ├── ProviderCard.tsx
│   │   │   ├── ProviderList.tsx
│   │   │   └── ProviderStatus.tsx
│   │   ├── lease/
│   │   │   ├── LeaseCard.tsx
│   │   │   ├── LeaseList.tsx
│   │   │   └── LeaseActions.tsx
│   │   └── billing/
│   │       ├── InvoiceTable.tsx
│   │       ├── UsageSummary.tsx
│   │       └── PaymentHistory.tsx
│   ├── hooks/
│   │   ├── index.ts
│   │   ├── useVirtEngine.ts
│   │   ├── useWallet.ts
│   │   ├── useDeployments.ts
│   │   ├── useProviders.ts
│   │   ├── useLeases.ts
│   │   └── useBalance.ts
│   ├── sdk/
│   │   ├── index.ts
│   │   ├── client.ts            # VirtEngineClient
│   │   ├── messages.ts          # Message builders
│   │   ├── queries.ts           # Query helpers
│   │   └── sdl.ts               # SDL parser
│   ├── types/
│   │   ├── index.ts
│   │   ├── deployment.ts
│   │   ├── provider.ts
│   │   ├── lease.ts
│   │   └── wallet.ts
│   └── utils/
│       ├── format.ts            # Address/amount formatting
│       └── constants.ts         # Chain constants
└── __tests__/
    ├── components/
    ├── hooks/
    └── sdk/
```

### WalletProvider Implementation

```typescript
// lib/portal/src/components/wallet/WalletProvider.tsx
import { createContext, useContext, useState, useCallback, useEffect, ReactNode } from 'react';
import { SigningStargateClient } from '@cosmjs/stargate';

interface WalletState {
  address: string | null;
  client: SigningStargateClient | null;
  isConnected: boolean;
  isConnecting: boolean;
  walletType: 'keplr' | 'leap' | 'cosmostation' | null;
}

interface WalletContextValue extends WalletState {
  connect: (walletType: 'keplr' | 'leap' | 'cosmostation') => Promise<void>;
  disconnect: () => void;
}

const WalletContext = createContext<WalletContextValue | null>(null);

interface WalletProviderProps {
  children: ReactNode;
  chainId: string;
  rpcEndpoint: string;
}

export function WalletProvider({ children, chainId, rpcEndpoint }: WalletProviderProps) {
  const [state, setState] = useState<WalletState>({
    address: null,
    client: null,
    isConnected: false,
    isConnecting: false,
    walletType: null,
  });

  const connect = useCallback(async (walletType: 'keplr' | 'leap' | 'cosmostation') => {
    setState(prev => ({ ...prev, isConnecting: true }));
    
    try {
      let wallet: any;
      
      switch (walletType) {
        case 'keplr':
          if (!window.keplr) throw new Error('Keplr not installed');
          await window.keplr.enable(chainId);
          wallet = window.keplr.getOfflineSigner(chainId);
          break;
        case 'leap':
          if (!window.leap) throw new Error('Leap not installed');
          await window.leap.enable(chainId);
          wallet = window.leap.getOfflineSigner(chainId);
          break;
        case 'cosmostation':
          if (!window.cosmostation) throw new Error('Cosmostation not installed');
          await window.cosmostation.providers.keplr.enable(chainId);
          wallet = window.cosmostation.providers.keplr.getOfflineSigner(chainId);
          break;
      }
      
      const accounts = await wallet.getAccounts();
      const client = await SigningStargateClient.connectWithSigner(rpcEndpoint, wallet);
      
      setState({
        address: accounts[0].address,
        client,
        isConnected: true,
        isConnecting: false,
        walletType,
      });
    } catch (error) {
      setState(prev => ({ ...prev, isConnecting: false }));
      throw error;
    }
  }, [chainId, rpcEndpoint]);

  const disconnect = useCallback(() => {
    setState({
      address: null,
      client: null,
      isConnected: false,
      isConnecting: false,
      walletType: null,
    });
  }, []);

  return (
    <WalletContext.Provider value={{ ...state, connect, disconnect }}>
      {children}
    </WalletContext.Provider>
  );
}

export function useWalletContext() {
  const context = useContext(WalletContext);
  if (!context) {
    throw new Error('useWalletContext must be used within WalletProvider');
  }
  return context;
}
```

### useDeployments Hook

```typescript
// lib/portal/src/hooks/useDeployments.ts
import { useQuery } from '@tanstack/react-query';
import { useWalletContext } from '../components/wallet/WalletProvider';

interface Deployment {
  id: string;
  owner: string;
  state: 'active' | 'closed' | 'pending';
  createdAt: Date;
  resources: {
    cpu: number;
    memory: number;
    storage: number;
  };
}

export function useDeployments() {
  const { address, client } = useWalletContext();
  
  return useQuery({
    queryKey: ['deployments', address],
    queryFn: async (): Promise<Deployment[]> => {
      if (!client || !address) return [];
      
      const response = await client.queryContractSmart(
        process.env.NEXT_PUBLIC_DEPLOYMENT_CONTRACT!,
        { list_deployments: { owner: address } }
      );
      
      return response.deployments.map((d: any) => ({
        id: d.id,
        owner: d.owner,
        state: d.state,
        createdAt: new Date(d.created_at * 1000),
        resources: d.resources,
      }));
    },
    enabled: !!address && !!client,
    staleTime: 30_000, // 30 seconds
  });
}
```

### VirtEngineClient SDK

```typescript
// lib/portal/src/sdk/client.ts
import { SigningStargateClient, StargateClient, DeliverTxResponse } from '@cosmjs/stargate';
import { EncodeObject } from '@cosmjs/proto-signing';

export class VirtEngineClient {
  private signingClient: SigningStargateClient | null = null;
  private queryClient: StargateClient | null = null;
  private address: string | null = null;

  constructor(
    private readonly rpcEndpoint: string,
    private readonly chainId: string,
  ) {}

  async connect(offlineSigner: any): Promise<string> {
    const accounts = await offlineSigner.getAccounts();
    this.address = accounts[0].address;
    
    this.signingClient = await SigningStargateClient.connectWithSigner(
      this.rpcEndpoint,
      offlineSigner,
    );
    
    this.queryClient = await StargateClient.connect(this.rpcEndpoint);
    
    return this.address;
  }

  async createDeployment(sdl: string): Promise<DeliverTxResponse> {
    if (!this.signingClient || !this.address) {
      throw new Error('Client not connected');
    }

    const msg: EncodeObject = {
      typeUrl: '/virtengine.deployment.v1beta1.MsgCreateDeployment',
      value: {
        owner: this.address,
        sdl: sdl,
      },
    };

    return this.signingClient.signAndBroadcast(
      this.address,
      [msg],
      'auto',
      'Create deployment via VirtEngine Portal',
    );
  }

  async closeDeployment(deploymentId: string): Promise<DeliverTxResponse> {
    if (!this.signingClient || !this.address) {
      throw new Error('Client not connected');
    }

    const msg: EncodeObject = {
      typeUrl: '/virtengine.deployment.v1beta1.MsgCloseDeployment',
      value: {
        id: deploymentId,
        owner: this.address,
      },
    };

    return this.signingClient.signAndBroadcast(
      this.address,
      [msg],
      'auto',
      'Close deployment',
    );
  }

  async getBalance(): Promise<{ amount: string; denom: string }> {
    if (!this.queryClient || !this.address) {
      throw new Error('Client not connected');
    }

    const balance = await this.queryClient.getBalance(
      this.address,
      'uvirt',
    );

    return balance;
  }

  // ... more methods for other operations
}
```

---

## Files to Create

| Path | Description | Est. Lines |
|------|-------------|------------|
| `lib/portal/package.json` | Package configuration | 50 |
| `lib/portal/tsconfig.json` | TypeScript config | 30 |
| `lib/portal/tsup.config.ts` | Build config | 20 |
| `lib/portal/src/index.ts` | Main exports | 20 |
| `lib/portal/src/components/index.ts` | Component exports | 30 |
| `lib/portal/src/components/wallet/*.tsx` | Wallet components | 400 |
| `lib/portal/src/components/deployment/*.tsx` | Deployment components | 500 |
| `lib/portal/src/components/provider/*.tsx` | Provider components | 300 |
| `lib/portal/src/components/lease/*.tsx` | Lease components | 300 |
| `lib/portal/src/components/billing/*.tsx` | Billing components | 300 |
| `lib/portal/src/hooks/*.ts` | React hooks | 400 |
| `lib/portal/src/sdk/*.ts` | SDK client | 600 |
| `lib/portal/src/types/*.ts` | TypeScript types | 200 |
| `lib/portal/src/utils/*.ts` | Utility functions | 100 |
| `lib/portal/__tests__/**` | Test files | 500 |

**Total: ~3750 lines**

---

## Implementation Steps

### Step 1: Create Package Structure
```bash
mkdir -p lib/portal/src/{components,hooks,sdk,types,utils}
mkdir -p lib/portal/__tests__
```

### Step 2: Initialize Package
```bash
cd lib/portal
pnpm init
pnpm add @cosmjs/stargate @cosmjs/proto-signing @tanstack/react-query zustand
pnpm add -D typescript tsup @types/react
```

### Step 3: Configure Build
Create tsconfig.json and tsup.config.ts

### Step 4: Implement Core Components
Start with WalletProvider, then expand

### Step 5: Implement Hooks
Create hooks using TanStack Query

### Step 6: Implement SDK
Create VirtEngineClient class

### Step 7: Update pnpm-workspace.yaml
Add lib/portal to workspace

### Step 8: Update Portal Imports
Fix imports in portal/src to use @virtengine/portal

### Step 9: Verify Build
```bash
pnpm -C lib/portal build
pnpm -C portal build
```

---

## Validation Checklist

- [ ] lib/portal package builds without errors
- [ ] portal/ builds without errors
- [ ] All components render correctly
- [ ] Wallet connection works with Keplr
- [ ] TypeScript types are exported correctly
- [ ] Tests pass

---

## Vibe-Kanban Task ID

`2299371b-9ed4-4dee-bc28-417bdf168357`
