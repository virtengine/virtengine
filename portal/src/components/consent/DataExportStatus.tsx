'use client';

import { useCallback, useEffect, useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import type { DataExportRequest, DeletionRequest } from '@/types/consent';

type RequestResponse = {
  exports: DataExportRequest[];
  deletions: DeletionRequest[];
};

export function DataExportStatus({ dataSubject }: { dataSubject: string }) {
  const [requests, setRequests] = useState<RequestResponse>({ exports: [], deletions: [] });
  const [loading, setLoading] = useState(true);

  const loadRequests = useCallback(async () => {
    setLoading(true);
    try {
      const res = await fetch('/api/consent/requests', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ dataSubject }),
      });
      const data = (await res.json()) as RequestResponse;
      setRequests(data);
    } finally {
      setLoading(false);
    }
  }, [dataSubject]);

  useEffect(() => {
    void loadRequests();
  }, [loadRequests]);

  const handleExport = async (format: 'json' | 'csv') => {
    await fetch('/api/consent/export', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ dataSubject, format }),
    });
    await loadRequests();
  };

  const handleDeletion = async () => {
    await fetch('/api/consent/deletion', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ dataSubject }),
    });
    await loadRequests();
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>Your data rights</CardTitle>
        <CardDescription>Request exports or deletions under GDPR Articles 15–20.</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="flex flex-wrap gap-3">
          <Button variant="secondary" onClick={() => handleExport('json')}>
            Request JSON export
          </Button>
          <Button variant="secondary" onClick={() => handleExport('csv')}>
            Request CSV export
          </Button>
          <Button variant="destructive" onClick={handleDeletion}>
            Request deletion
          </Button>
        </div>

        {loading ? (
          <p className="text-sm text-muted-foreground">Loading export status…</p>
        ) : (
          <div className="grid gap-4 lg:grid-cols-2">
            <div className="rounded-lg border border-border bg-muted/20 p-4">
              <p className="text-sm font-semibold">Export requests</p>
              <div className="mt-3 space-y-2 text-sm">
                {requests.exports.length === 0 ? (
                  <p className="text-xs text-muted-foreground">No export requests yet.</p>
                ) : (
                  requests.exports.map((item) => (
                    <div key={item.id} className="flex items-center justify-between">
                      <div>
                        <p className="text-xs text-muted-foreground">{item.id}</p>
                        <p className="text-xs text-muted-foreground">
                          {new Date(item.requestedAt).toLocaleString()}
                        </p>
                      </div>
                      <Badge variant="outline">{item.status}</Badge>
                    </div>
                  ))
                )}
              </div>
            </div>

            <div className="rounded-lg border border-border bg-muted/20 p-4">
              <p className="text-sm font-semibold">Deletion requests</p>
              <div className="mt-3 space-y-2 text-sm">
                {requests.deletions.length === 0 ? (
                  <p className="text-xs text-muted-foreground">No deletion requests yet.</p>
                ) : (
                  requests.deletions.map((item) => (
                    <div key={item.id} className="flex items-center justify-between">
                      <div>
                        <p className="text-xs text-muted-foreground">{item.id}</p>
                        <p className="text-xs text-muted-foreground">
                          {new Date(item.requestedAt).toLocaleString()}
                        </p>
                      </div>
                      <Badge variant="outline">{item.status}</Badge>
                    </div>
                  ))
                )}
              </div>
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
