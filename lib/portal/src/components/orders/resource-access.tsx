import * as React from "react";
import type { OrderResourceAccess } from "../../hooks/useOrderTracking";

export interface ResourceAccessProps {
  access?: OrderResourceAccess;
  className?: string;
}

const maskSecret = (value: string, keep: number = 4) => {
  if (!value) return "";
  if (value.length <= keep) return "*".repeat(value.length);
  return `${"*".repeat(Math.max(4, value.length - keep))}${value.slice(-keep)}`;
};

export function ResourceAccess({
  access,
  className = "",
}: ResourceAccessProps): JSX.Element {
  const [revealed, setRevealed] = React.useState<Record<string, boolean>>({});
  const [copiedId, setCopiedId] = React.useState<string | null>(null);

  const handleCopy = async (id: string, value: string) => {
    try {
      await navigator.clipboard.writeText(value);
      setCopiedId(id);
      setTimeout(() => setCopiedId(null), 2000);
    } catch {
      setCopiedId(null);
    }
  };

  if (!access) {
    return (
      <section className={`ve-resource-access ${className}`}>
        <h3>Resource access</h3>
        <p>
          Access details will appear once the provider provisions your
          resources.
        </p>
        <style>{resourceAccessStyles}</style>
      </section>
    );
  }

  return (
    <section className={`ve-resource-access ${className}`}>
      <div className="ve-resource-access__header">
        <div>
          <h3>Resource access</h3>
          <p>Secure connection details for your deployment.</p>
        </div>
        {access.consoleUrl && (
          <a
            className="ve-resource-access__console"
            href={access.consoleUrl}
            target="_blank"
            rel="noreferrer"
          >
            Open console
          </a>
        )}
      </div>

      <div className="ve-resource-access__section">
        <h4>Connections</h4>
        <div className="ve-resource-access__cards">
          {access.connections.map((connection) => (
            <div key={connection.label} className="ve-resource-access__card">
              <div>
                <strong>{connection.label}</strong>
                <p>{connection.protocol.toUpperCase()}</p>
              </div>
              <div>
                <span>{connection.host}</span>
                <span>Port {connection.port}</span>
                {connection.username && <span>User {connection.username}</span>}
              </div>
            </div>
          ))}
        </div>
      </div>

      <div className="ve-resource-access__section">
        <h4>Credentials</h4>
        <div className="ve-resource-access__credentials">
          {access.credentials.map((credential) => {
            const isRevealed = revealed[credential.id];
            return (
              <div
                key={credential.id}
                className="ve-resource-access__credential"
              >
                <div>
                  <strong>{credential.label}</strong>
                  <small>{credential.type.replace(/_/g, " ")}</small>
                </div>
                <code>
                  {isRevealed ? credential.value : maskSecret(credential.value)}
                </code>
                <div>
                  <button
                    type="button"
                    onClick={() =>
                      setRevealed((prev) => ({
                        ...prev,
                        [credential.id]: !prev[credential.id],
                      }))
                    }
                  >
                    {isRevealed ? "Hide" : "Reveal"}
                  </button>
                  <button
                    type="button"
                    onClick={() =>
                      void handleCopy(credential.id, credential.value)
                    }
                  >
                    {copiedId === credential.id ? "Copied" : "Copy"}
                  </button>
                </div>
              </div>
            );
          })}
        </div>

        {access.sshPublicKey && (
          <div className="ve-resource-access__ssh">
            <strong>SSH public key</strong>
            <code>{access.sshPublicKey}</code>
            {access.sshFingerprint && (
              <small>Fingerprint: {access.sshFingerprint}</small>
            )}
          </div>
        )}
      </div>

      <div className="ve-resource-access__section">
        <h4>API endpoints</h4>
        <div className="ve-resource-access__api">
          {access.apiEndpoints.map((endpoint) => (
            <div key={endpoint.id} className="ve-resource-access__api-row">
              <div>
                <strong>{endpoint.label}</strong>
                {endpoint.description && <span>{endpoint.description}</span>}
              </div>
              <div>
                <code>{endpoint.method.toUpperCase()}</code>
                <a href={endpoint.url} target="_blank" rel="noreferrer">
                  {endpoint.url}
                </a>
              </div>
            </div>
          ))}
        </div>
      </div>

      {access.notes && (
        <p className="ve-resource-access__notes">{access.notes}</p>
      )}

      <style>{resourceAccessStyles}</style>
    </section>
  );
}

const resourceAccessStyles = `
  .ve-resource-access {
    background: #fff;
    border-radius: 22px;
    padding: 20px;
    border: 1px solid #e2e8f0;
    display: grid;
    gap: 18px;
  }

  .ve-resource-access__header {
    display: flex;
    justify-content: space-between;
    gap: 12px;
    align-items: center;
  }

  .ve-resource-access h3 {
    margin: 0;
    font-size: 1.2rem;
    font-family: 'Space Grotesk', 'DM Sans', sans-serif;
  }

  .ve-resource-access__header p {
    margin: 4px 0 0;
    color: #64748b;
  }

  .ve-resource-access__console {
    padding: 8px 14px;
    border-radius: 999px;
    background: #0f172a;
    color: #fff;
    font-size: 0.85rem;
    font-weight: 600;
    text-decoration: none;
  }

  .ve-resource-access__section h4 {
    margin: 0 0 8px;
    font-size: 0.95rem;
    color: #0f172a;
  }

  .ve-resource-access__cards {
    display: grid;
    gap: 10px;
  }

  .ve-resource-access__card {
    border-radius: 16px;
    border: 1px solid #e2e8f0;
    padding: 12px 14px;
    display: grid;
    gap: 6px;
    background: #f8fafc;
  }

  .ve-resource-access__card strong {
    color: #0f172a;
  }

  .ve-resource-access__card span {
    display: block;
    color: #475569;
    font-size: 0.85rem;
  }

  .ve-resource-access__credentials {
    display: grid;
    gap: 12px;
  }

  .ve-resource-access__credential {
    border-radius: 16px;
    border: 1px solid #e2e8f0;
    padding: 12px;
    display: grid;
    gap: 8px;
    background: #0f172a;
    color: #e2e8f0;
  }

  .ve-resource-access__credential code {
    font-family: 'JetBrains Mono', 'Fira Code', monospace;
    font-size: 0.8rem;
    color: #bae6fd;
    word-break: break-all;
  }

  .ve-resource-access__credential button {
    border: 1px solid rgba(255, 255, 255, 0.2);
    background: transparent;
    color: #fff;
    font-size: 0.75rem;
    padding: 6px 10px;
    border-radius: 999px;
    cursor: pointer;
    margin-right: 6px;
  }

  .ve-resource-access__ssh {
    margin-top: 12px;
    padding: 10px 12px;
    border-radius: 12px;
    border: 1px dashed #e2e8f0;
    background: #f8fafc;
    display: grid;
    gap: 6px;
  }

  .ve-resource-access__ssh code {
    font-size: 0.75rem;
    word-break: break-all;
  }

  .ve-resource-access__api {
    display: grid;
    gap: 10px;
  }

  .ve-resource-access__api-row {
    display: flex;
    justify-content: space-between;
    gap: 12px;
    border: 1px solid #e2e8f0;
    border-radius: 14px;
    padding: 12px;
    background: #f8fafc;
  }

  .ve-resource-access__api-row code {
    padding: 4px 8px;
    background: #0f172a;
    color: #fff;
    border-radius: 999px;
    font-size: 0.7rem;
  }

  .ve-resource-access__api-row a {
    margin-left: 10px;
    color: #0f172a;
    font-weight: 600;
    text-decoration: none;
    font-size: 0.85rem;
  }

  .ve-resource-access__notes {
    margin: 0;
    color: #64748b;
    font-size: 0.85rem;
  }

  @media (max-width: 960px) {
    .ve-resource-access__header {
      flex-direction: column;
      align-items: flex-start;
    }

    .ve-resource-access__api-row {
      flex-direction: column;
      align-items: flex-start;
    }
  }
`;
