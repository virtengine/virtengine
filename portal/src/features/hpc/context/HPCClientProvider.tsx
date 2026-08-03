'use client';

import { createContext, useContext, type ReactNode } from 'react';
import { createHPCClient, type HPCClient } from '../lib/hpc-client';

const unavailableHPCClient = createHPCClient();
const HPCClientContext = createContext<HPCClient>(unavailableHPCClient);

export interface HPCClientProviderProps {
  children: ReactNode;
  client?: HPCClient;
}

export function HPCClientProvider({ children, client }: HPCClientProviderProps) {
  return (
    <HPCClientContext.Provider value={client ?? unavailableHPCClient}>
      {children}
    </HPCClientContext.Provider>
  );
}

export function useHPCClient(): HPCClient {
  return useContext(HPCClientContext);
}
