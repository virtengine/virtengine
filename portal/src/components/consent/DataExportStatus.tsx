'use client';

import { useCallback, useEffect, useRef, useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import type { DataExportRequest, DeletionRequest } from '@/types/consent';
import { useTranslation } from 'react-i18next';

type RequestResponse = {
  exports: DataExportRequest[];
  deletions: DeletionRequest[];
};

export function DataExportStatus({ dataSubject }: { dataSubject: string }) {
  const { t } = useTranslation();
  const [requests, setRequests] = useState<RequestResponse>({ exports: [], deletions: [] });
  const [loading, setLoading] = useState(true);
  const [deletionError, setDeletionError] = useState<string | null>(null);
  const [deletionSubmitting, setDeletionSubmitting] = useState(false);
  const requestGeneration = useRef(0);
  const loadSequence = useRef(0);
  const deletionInFlight = useRef(false);
  const acceptedDeletions = useRef(new Map<string, DeletionRequest>());
  const activeControllers = useRef(new Set<AbortController>());

  const loadRequests = useCallback(async () => {
    const generation = requestGeneration.current;
    const sequence = ++loadSequence.current;
    const controller = new AbortController();
    activeControllers.current.add(controller);
    setLoading(true);
    try {
      const res = await fetch('/api/consent/requests', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ dataSubject }),
        signal: controller.signal,
      });
      if (!res.ok) throw new Error(t('Deletion workflow is unavailable.'));
      const data = (await res.json()) as RequestResponse;
      if (generation !== requestGeneration.current || sequence !== loadSequence.current) return;
      const serverIds = new Set(data.deletions.map((item) => item.id));
      const accepted = Array.from(acceptedDeletions.current.values()).filter(
        (item) => !serverIds.has(item.id)
      );
      for (const item of data.deletions) acceptedDeletions.current.delete(item.id);
      setRequests({
        exports: data.exports,
        deletions: [
          ...accepted,
          ...data.deletions.filter((item) => !acceptedDeletions.current.has(item.id)),
        ],
      });
    } catch (error) {
      if (
        generation === requestGeneration.current &&
        sequence === loadSequence.current &&
        !controller.signal.aborted
      ) {
        setDeletionError(
          error instanceof Error ? error.message : t('Deletion workflow is unavailable.')
        );
      }
    } finally {
      activeControllers.current.delete(controller);
      if (generation === requestGeneration.current && sequence === loadSequence.current) {
        setLoading(false);
      }
    }
  }, [dataSubject, t]);

  useEffect(() => {
    const controllers = activeControllers.current;
    requestGeneration.current += 1;
    loadSequence.current += 1;
    acceptedDeletions.current.clear();
    setRequests({ exports: [], deletions: [] });
    setDeletionError(null);
    setDeletionSubmitting(false);
    void loadRequests();
    return () => {
      requestGeneration.current += 1;
      loadSequence.current += 1;
      deletionInFlight.current = false;
      for (const controller of controllers) controller.abort();
      controllers.clear();
    };
  }, [dataSubject, loadRequests]);

  const handleExport = async (format: 'json' | 'csv') => {
    const generation = requestGeneration.current;
    const subject = dataSubject;
    const controller = new AbortController();
    activeControllers.current.add(controller);
    try {
      const response = await fetch('/api/consent/export', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ dataSubject: subject, format }),
        signal: controller.signal,
      });
      if (!response.ok) throw new Error(t('Export workflow is unavailable.'));
      if (generation !== requestGeneration.current || controller.signal.aborted) return;
      await loadRequests();
    } catch (error) {
      if (generation !== requestGeneration.current || controller.signal.aborted) return;
      setDeletionError(
        error instanceof Error ? error.message : t('Export workflow is unavailable.')
      );
    } finally {
      activeControllers.current.delete(controller);
    }
  };

  const handleDeletion = async () => {
    if (deletionInFlight.current) return;
    deletionInFlight.current = true;
    const generation = requestGeneration.current;
    const subject = dataSubject;
    const controller = new AbortController();
    activeControllers.current.add(controller);
    setDeletionSubmitting(true);
    setDeletionError(null);
    try {
      const response = await fetch('/api/consent/deletion', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ dataSubject: subject }),
        signal: controller.signal,
      });
      if (!response.ok) throw new Error(t('Deletion workflow is unavailable.'));
      const record = (await response.json()) as DeletionRequest;
      if (generation !== requestGeneration.current || record.dataSubject !== subject) return;
      acceptedDeletions.current.set(record.id, record);
      setRequests((current) => ({
        ...current,
        deletions: [record, ...current.deletions.filter((item) => item.id !== record.id)],
      }));
    } catch (error) {
      if (generation !== requestGeneration.current || controller.signal.aborted) return;
      setDeletionError(
        error instanceof Error ? error.message : t('Deletion workflow is unavailable.')
      );
    } finally {
      activeControllers.current.delete(controller);
      if (generation === requestGeneration.current) {
        deletionInFlight.current = false;
        setDeletionSubmitting(false);
      }
    }
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t('Your data rights')}</CardTitle>
        <CardDescription>
          {t('Request exports or deletions under GDPR Articles 15–20.')}
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="flex flex-wrap gap-3">
          <Button variant="secondary" onClick={() => handleExport('json')}>
            {t('Request JSON export')}
          </Button>
          <Button variant="secondary" onClick={() => handleExport('csv')}>
            {t('Request CSV export')}
          </Button>
          <Button variant="destructive" onClick={handleDeletion} disabled={deletionSubmitting}>
            {deletionSubmitting ? t('Requesting deletion…') : t('Request deletion')}
          </Button>
        </div>
        {deletionError && (
          <p role="alert" className="text-sm text-destructive">
            {deletionError}
          </p>
        )}

        {loading ? (
          <p className="text-sm text-muted-foreground">{t('Loading export status…')}</p>
        ) : (
          <div className="grid gap-4 lg:grid-cols-2">
            <div className="rounded-lg border border-border bg-muted/20 p-4">
              <p className="text-sm font-semibold">{t('Export requests')}</p>
              <div className="mt-3 space-y-2 text-sm">
                {requests.exports.length === 0 ? (
                  <p className="text-xs text-muted-foreground">{t('No export requests yet.')}</p>
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
              <p className="text-sm font-semibold">{t('Deletion requests')}</p>
              <div className="mt-3 space-y-2 text-sm">
                {requests.deletions.length === 0 ? (
                  <p className="text-xs text-muted-foreground">{t('No deletion requests yet.')}</p>
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
