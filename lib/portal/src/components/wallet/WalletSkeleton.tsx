'use client';

export interface WalletSkeletonProps {
  variant?: 'button' | 'account' | 'modal';
}

const pulseAnimation = `
@keyframes wallet-skeleton-pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}
`;

const baseSkeletonStyle: React.CSSProperties = {
  backgroundColor: '#e5e7eb',
  borderRadius: 4,
  animation: 'wallet-skeleton-pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite',
};

function SkeletonBox({ width, height, style }: { width: number | string; height: number; style?: React.CSSProperties }) {
  return (
    <div
      style={{
        ...baseSkeletonStyle,
        width,
        height,
        ...style,
      }}
      aria-hidden="true"
    />
  );
}

function ButtonSkeleton() {
  return (
    <div
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        gap: 8,
        padding: '10px 20px',
        borderRadius: 8,
        backgroundColor: '#f3f4f6',
      }}
    >
      <SkeletonBox width={20} height={20} style={{ borderRadius: '50%' }} />
      <SkeletonBox width={100} height={16} />
    </div>
  );
}

function AccountSkeleton() {
  return (
    <div
      style={{
        display: 'flex',
        alignItems: 'center',
        gap: 12,
        padding: 12,
        borderRadius: 8,
        backgroundColor: '#f9fafb',
        border: '1px solid #e5e7eb',
      }}
    >
      <SkeletonBox width={40} height={40} style={{ borderRadius: '50%' }} />
      <div style={{ flex: 1, display: 'flex', flexDirection: 'column', gap: 6 }}>
        <SkeletonBox width={120} height={14} />
        <SkeletonBox width={180} height={12} />
      </div>
      <SkeletonBox width={24} height={24} />
    </div>
  );
}

function ModalSkeleton() {
  return (
    <div
      style={{
        width: '100%',
        maxWidth: 400,
        padding: 24,
        borderRadius: 12,
        backgroundColor: '#fff',
        boxShadow: '0 10px 25px rgba(0,0,0,0.1)',
      }}
    >
      {/* Header */}
      <div style={{ marginBottom: 20 }}>
        <SkeletonBox width={160} height={20} style={{ marginBottom: 8 }} />
        <SkeletonBox width={220} height={14} />
      </div>
      
      {/* Wallet options */}
      <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
        {[1, 2, 3, 4].map((i) => (
          <div
            key={i}
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: 12,
              padding: 14,
              borderRadius: 10,
              backgroundColor: '#f3f4f6',
            }}
          >
            <SkeletonBox width={36} height={36} style={{ borderRadius: 8 }} />
            <div style={{ flex: 1 }}>
              <SkeletonBox width={80} height={14} style={{ marginBottom: 6 }} />
              <SkeletonBox width={140} height={11} />
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

export function WalletSkeleton({ variant = 'button' }: WalletSkeletonProps) {
  return (
    <>
      <style dangerouslySetInnerHTML={{ __html: pulseAnimation }} />
      {variant === 'button' && <ButtonSkeleton />}
      {variant === 'account' && <AccountSkeleton />}
      {variant === 'modal' && <ModalSkeleton />}
    </>
  );
}
