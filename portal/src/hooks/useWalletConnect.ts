/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

'use client';

import { useState, useCallback, useRef, useEffect } from 'react';
import { WALLET_CONNECT_PROJECT_ID } from '@/config';

export type WalletConnectStatus = 'idle' | 'connecting' | 'connected' | 'error';

export interface WalletConnectSession {
  topic: string;
  peerName: string;
  peerIcon?: string;
  accounts: string[];
  chainId: string;
}

export interface UseWalletConnectResult {
  status: WalletConnectStatus;
  uri: string | null;
  session: WalletConnectSession | null;
  error: string | null;
  connect: () => Promise<void>;
  disconnect: () => Promise<void>;
  isSupported: boolean;
}

/**
 * Hook wrapping WalletConnect v2 sign client for mobile wallet pairing.
 *
 * Generates a WalletConnect URI for QR code display. When a mobile wallet
 * scans the code and approves, the session is stored and accounts are
 * available for signing.
 *
 * Requires NEXT_PUBLIC_WALLET_CONNECT_PROJECT_ID in env.
 */
export function useWalletConnect(): UseWalletConnectResult {
  const [status, setStatus] = useState<WalletConnectStatus>('idle');
  const [uri, setUri] = useState<string | null>(null);
  const [session, setSession] = useState<WalletConnectSession | null>(null);
  const [error, setError] = useState<string | null>(null);
  const clientRef = useRef<unknown>(null);

  const isSupported = !!WALLET_CONNECT_PROJECT_ID;

  // Lazy-load sign client to avoid bundling when not configured
  const getClient = useCallback(async () => {
    if (clientRef.current) return clientRef.current;
    if (!WALLET_CONNECT_PROJECT_ID) {
      throw new Error('WalletConnect project ID not configured');
    }

    const { SignClient } = await import('@walletconnect/sign-client');
    const client = await SignClient.init({
      projectId: WALLET_CONNECT_PROJECT_ID,
      metadata: {
        name: 'VirtEngine Portal',
        description: 'Decentralized cloud computing marketplace',
        url:
          typeof window !== 'undefined' ? window.location.origin : 'https://portal.virtengine.io',
        icons: ['/apple-touch-icon.png'],
      },
    });
    clientRef.current = client;
    return client;
  }, []);

  const connect = useCallback(async () => {
    try {
      setStatus('connecting');
      setError(null);

      const client = (await getClient()) as {
        connect: (params: {
          requiredNamespaces: Record<
            string,
            { methods: string[]; chains: string[]; events: string[] }
          >;
        }) => Promise<{
          uri?: string;
          approval: () => Promise<{
            topic: string;
            peer: { metadata: { name: string; icons?: string[] } };
            namespaces: Record<string, { accounts: string[] }>;
          }>;
        }>;
      };

      const { uri: wcUri, approval } = await client.connect({
        requiredNamespaces: {
          cosmos: {
            methods: ['cosmos_signDirect', 'cosmos_signAmino'],
            chains: ['cosmos:virtengine-1'],
            events: ['chainChanged', 'accountsChanged'],
          },
        },
      });

      if (wcUri) {
        setUri(wcUri);
      }

      const approved = await approval();

      const accounts =
        approved.namespaces?.cosmos?.accounts?.map((a: string) => a.split(':').pop() ?? '') ?? [];

      setSession({
        topic: approved.topic,
        peerName: approved.peer.metadata.name,
        peerIcon: approved.peer.metadata.icons?.[0],
        accounts,
        chainId: 'virtengine-1',
      });
      setStatus('connected');
      setUri(null);
    } catch (err) {
      setStatus('error');
      setError(err instanceof Error ? err.message : 'WalletConnect connection failed');
    }
  }, [getClient]);

  const disconnect = useCallback(async () => {
    if (session && clientRef.current) {
      try {
        const client = clientRef.current as {
          disconnect: (params: {
            topic: string;
            reason: { code: number; message: string };
          }) => Promise<void>;
        };
        await client.disconnect({
          topic: session.topic,
          reason: { code: 6000, message: 'User disconnected' },
        });
      } catch {
        // Best-effort disconnect
      }
    }
    setSession(null);
    setStatus('idle');
    setUri(null);
    setError(null);
  }, [session]);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      if (session && clientRef.current) {
        const client = clientRef.current as {
          disconnect: (params: {
            topic: string;
            reason: { code: number; message: string };
          }) => Promise<void>;
        };
        void client.disconnect({
          topic: session.topic,
          reason: { code: 6000, message: 'Component unmounted' },
        });
      }
    };
    // Only run cleanup on unmount
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return {
    status,
    uri,
    session,
    error,
    connect,
    disconnect,
    isSupported,
  };
}
