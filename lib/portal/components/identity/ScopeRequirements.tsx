/**
 * Scope Requirements Component
 * VE-701: Shows verification scopes required for an action
 */
import * as React from 'react';
import type { MarketplaceAction, VerificationScope, VerificationScopeType } from '../../types/identity';

export interface ScopeRequirementsProps {
  action: MarketplaceAction;
  completedScopes: VerificationScope[];
  className?: string;
}

const scopeLabels: Record<VerificationScopeType, string> = {
  email: 'Email Verification',
  id_document: 'ID Document',
  selfie: 'Selfie Verification',
  sso: 'SSO Connection',
  domain: 'Domain Ownership',
  biometric: 'Biometric',
};

const actionRequirements: Record<MarketplaceAction, { scopes: VerificationScopeType[]; minScore: number }> = {
  browse_offerings: { scopes: [], minScore: 0 },
  view_offering_details: { scopes: [], minScore: 0 },
  place_order: { scopes: ['email'], minScore: 30 },
  place_high_value_order: { scopes: ['email', 'id_document', 'selfie'], minScore: 60 },
  register_provider: { scopes: ['email', 'id_document', 'selfie', 'domain'], minScore: 70 },
  create_offering: { scopes: ['email', 'id_document', 'selfie', 'domain'], minScore: 70 },
  submit_hpc_job: { scopes: ['email', 'id_document'], minScore: 50 },
  access_outputs: { scopes: ['email'], minScore: 30 },
};

export function ScopeRequirements({ action, completedScopes, className }: ScopeRequirementsProps): JSX.Element {
  const requirements = actionRequirements[action] ?? { scopes: [], minScore: 0 };
  const completedTypes = new Set(completedScopes.filter(s => s.completed).map(s => s.type));

  return (
    <div className={className}>
      <h4 style={{ margin: '0 0 8px', fontSize: '14px', fontWeight: 600 }}>
        Required Verifications
      </h4>
      {requirements.scopes.length === 0 ? (
        <p style={{ margin: 0, color: '#666', fontSize: '14px' }}>
          No verification required for this action.
        </p>
      ) : (
        <ul style={{ margin: 0, padding: 0, listStyle: 'none' }}>
          {requirements.scopes.map((scope) => {
            const isComplete = completedTypes.has(scope);
            return (
              <li
                key={scope}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: '8px',
                  padding: '4px 0',
                  color: isComplete ? '#16a34a' : '#666',
                }}
              >
                <span>{isComplete ? '✓' : '○'}</span>
                <span>{scopeLabels[scope]}</span>
              </li>
            );
          })}
        </ul>
      )}
      {requirements.minScore > 0 && (
        <p style={{ margin: '8px 0 0', fontSize: '12px', color: '#666' }}>
          Minimum identity score: {requirements.minScore}
        </p>
      )}
    </div>
  );
}
