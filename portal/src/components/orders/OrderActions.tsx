/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * Order action components: extend, cancel, and support ticket modals.
 */

'use client';

import { useState, useCallback } from 'react';
import { Button } from '@/components/ui/Button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card';
import { Input } from '@/components/ui/Input';
import { Label } from '@/components/ui/Label';
import { Textarea } from '@/components/ui/Textarea';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/Select';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/dialog';
import { useToast } from '@/hooks/use-toast';
import type { OrderStatus } from '@/stores/orderStore';
import type {
  ExtendOrderRequest,
  CancelOrderRequest,
  SupportTicketRequest,
  OrderActionResult,
} from '@/features/orders/tracking-types';
import { isOrderActive, isOrderTerminal } from '@/features/orders/tracking-types';

// =============================================================================
// Main Actions Panel
// =============================================================================

interface OrderActionsProps {
  orderId: string;
  status: OrderStatus;
  providerName: string;
  onExtend?: (req: ExtendOrderRequest) => Promise<OrderActionResult>;
  onCancel?: (req: CancelOrderRequest) => Promise<OrderActionResult>;
  onSupport?: (req: SupportTicketRequest) => Promise<OrderActionResult>;
}

export function OrderActions({
  orderId,
  status,
  providerName,
  onExtend,
  onCancel,
  onSupport,
}: OrderActionsProps) {
  const [extendOpen, setExtendOpen] = useState(false);
  const [cancelOpen, setCancelOpen] = useState(false);
  const [supportOpen, setSupportOpen] = useState(false);

  const active = isOrderActive(status);
  const terminal = isOrderTerminal(status);

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-lg">Actions</CardTitle>
      </CardHeader>
      <CardContent className="space-y-3">
        {active && (
          <Button
            variant="outline"
            className="w-full justify-start"
            onClick={() => setExtendOpen(true)}
          >
            ⏱ Extend Duration
          </Button>
        )}

        <Button
          variant="outline"
          className="w-full justify-start"
          onClick={() => setSupportOpen(true)}
        >
          💬 Request Support
        </Button>

        {active && (
          <Button
            variant="outline"
            className="w-full justify-start text-destructive hover:bg-destructive/10 hover:text-destructive"
            onClick={() => setCancelOpen(true)}
          >
            ✕ Cancel Order
          </Button>
        )}

        {terminal && (
          <p className="text-center text-xs text-muted-foreground">
            This order has been {status}. No further actions available.
          </p>
        )}

        {/* Modals */}
        <ExtendOrderDialog
          orderId={orderId}
          open={extendOpen}
          onOpenChange={setExtendOpen}
          onSubmit={onExtend}
        />
        <CancelOrderDialog
          orderId={orderId}
          providerName={providerName}
          open={cancelOpen}
          onOpenChange={setCancelOpen}
          onSubmit={onCancel}
        />
        <SupportTicketDialog
          orderId={orderId}
          open={supportOpen}
          onOpenChange={setSupportOpen}
          onSubmit={onSupport}
        />
      </CardContent>
    </Card>
  );
}

// =============================================================================
// Extend Order Dialog
// =============================================================================

interface ExtendOrderDialogProps {
  orderId: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSubmit?: (req: ExtendOrderRequest) => Promise<OrderActionResult>;
}

function ExtendOrderDialog({ orderId, open, onOpenChange, onSubmit }: ExtendOrderDialogProps) {
  const [duration, setDuration] = useState('24');
  const [unit, setUnit] = useState<'hours' | 'days' | 'months'>('hours');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const { toast } = useToast();

  const handleSubmit = useCallback(async () => {
    if (!onSubmit) return;
    setIsSubmitting(true);
    try {
      const result = await onSubmit({
        orderId,
        additionalDuration: parseInt(duration, 10),
        durationUnit: unit,
      });
      if (result.success) {
        toast({ title: 'Order Extended', description: result.message });
        onOpenChange(false);
      } else {
        toast({ title: 'Failed', description: result.message, variant: 'destructive' });
      }
    } catch (err) {
      toast({
        title: 'Error',
        description: err instanceof Error ? err.message : 'Failed to extend order',
        variant: 'destructive',
      });
    } finally {
      setIsSubmitting(false);
    }
  }, [orderId, duration, unit, onSubmit, onOpenChange, toast]);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Extend Order Duration</DialogTitle>
          <DialogDescription>
            Add more time to your current order. Additional escrow deposit may be required.
          </DialogDescription>
        </DialogHeader>

        <div className="grid gap-4 py-4">
          <div className="grid gap-2">
            <Label htmlFor="extend-duration">Additional Duration</Label>
            <div className="flex gap-2">
              <Input
                id="extend-duration"
                type="number"
                min="1"
                value={duration}
                onChange={(e) => setDuration(e.target.value)}
                className="flex-1"
              />
              <Select value={unit} onValueChange={(v) => setUnit(v as typeof unit)}>
                <SelectTrigger className="w-32">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="hours">Hours</SelectItem>
                  <SelectItem value="days">Days</SelectItem>
                  <SelectItem value="months">Months</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button onClick={handleSubmit} loading={isSubmitting}>
            Extend Order
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

// =============================================================================
// Cancel Order Dialog
// =============================================================================

interface CancelOrderDialogProps {
  orderId: string;
  providerName: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSubmit?: (req: CancelOrderRequest) => Promise<OrderActionResult>;
}

function CancelOrderDialog({
  orderId,
  providerName,
  open,
  onOpenChange,
  onSubmit,
}: CancelOrderDialogProps) {
  const [reason, setReason] = useState('');
  const [immediate, setImmediate] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const { toast } = useToast();

  const handleSubmit = useCallback(async () => {
    if (!onSubmit) return;
    setIsSubmitting(true);
    try {
      const result = await onSubmit({ orderId, reason, immediate });
      if (result.success) {
        toast({ title: 'Order Cancelled', description: result.message });
        onOpenChange(false);
      } else {
        toast({ title: 'Failed', description: result.message, variant: 'destructive' });
      }
    } catch (err) {
      toast({
        title: 'Error',
        description: err instanceof Error ? err.message : 'Failed to cancel order',
        variant: 'destructive',
      });
    } finally {
      setIsSubmitting(false);
    }
  }, [orderId, reason, immediate, onSubmit, onOpenChange, toast]);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Cancel Order</DialogTitle>
          <DialogDescription>
            This will stop your deployment on {providerName}. Remaining escrow funds will be
            returned after settlement.
          </DialogDescription>
        </DialogHeader>

        <div className="grid gap-4 py-4">
          <div className="grid gap-2">
            <Label htmlFor="cancel-reason">Reason (optional)</Label>
            <Textarea
              id="cancel-reason"
              placeholder="Why are you cancelling this order?"
              value={reason}
              onChange={(e) => setReason(e.target.value)}
              rows={3}
            />
          </div>

          <label className="flex items-center gap-2 text-sm">
            <input
              type="checkbox"
              checked={immediate}
              onChange={(e) => setImmediate(e.target.checked)}
              className="rounded border-border"
            />
            Cancel immediately (skip graceful shutdown)
          </label>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Keep Order
          </Button>
          <Button variant="destructive" onClick={handleSubmit} loading={isSubmitting}>
            Cancel Order
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

// =============================================================================
// Support Ticket Dialog
// =============================================================================

interface SupportTicketDialogProps {
  orderId: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSubmit?: (req: SupportTicketRequest) => Promise<OrderActionResult>;
}

function SupportTicketDialog({ orderId, open, onOpenChange, onSubmit }: SupportTicketDialogProps) {
  const [subject, setSubject] = useState('');
  const [description, setDescription] = useState('');
  const [priority, setPriority] = useState<'low' | 'medium' | 'high' | 'critical'>('medium');
  const [category, setCategory] = useState<
    'billing' | 'technical' | 'access' | 'performance' | 'other'
  >('technical');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const { toast } = useToast();

  const handleSubmit = useCallback(async () => {
    if (!subject.trim() || !description.trim()) {
      toast({
        title: 'Validation Error',
        description: 'Please fill in subject and description.',
        variant: 'destructive',
      });
      return;
    }

    if (!onSubmit) return;
    setIsSubmitting(true);
    try {
      const result = await onSubmit({ orderId, subject, description, priority, category });
      if (result.success) {
        toast({
          title: 'Ticket Created',
          description: result.ticketId ? `Ticket #${result.ticketId} created.` : result.message,
        });
        onOpenChange(false);
        setSubject('');
        setDescription('');
      } else {
        toast({ title: 'Failed', description: result.message, variant: 'destructive' });
      }
    } catch (err) {
      toast({
        title: 'Error',
        description: err instanceof Error ? err.message : 'Failed to create ticket',
        variant: 'destructive',
      });
    } finally {
      setIsSubmitting(false);
    }
  }, [orderId, subject, description, priority, category, onSubmit, onOpenChange, toast]);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>Request Support</DialogTitle>
          <DialogDescription>
            Create a support ticket for order #{orderId}. Our team will respond within 24 hours.
          </DialogDescription>
        </DialogHeader>

        <div className="grid gap-4 py-4">
          <div className="grid gap-2">
            <Label htmlFor="support-subject">Subject</Label>
            <Input
              id="support-subject"
              placeholder="Brief description of the issue"
              value={subject}
              onChange={(e) => setSubject(e.target.value)}
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="grid gap-2">
              <Label>Category</Label>
              <Select value={category} onValueChange={(v) => setCategory(v as typeof category)}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="technical">Technical</SelectItem>
                  <SelectItem value="billing">Billing</SelectItem>
                  <SelectItem value="access">Access</SelectItem>
                  <SelectItem value="performance">Performance</SelectItem>
                  <SelectItem value="other">Other</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="grid gap-2">
              <Label>Priority</Label>
              <Select value={priority} onValueChange={(v) => setPriority(v as typeof priority)}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="low">Low</SelectItem>
                  <SelectItem value="medium">Medium</SelectItem>
                  <SelectItem value="high">High</SelectItem>
                  <SelectItem value="critical">Critical</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

          <div className="grid gap-2">
            <Label htmlFor="support-description">Description</Label>
            <Textarea
              id="support-description"
              placeholder="Detailed description of the issue, steps to reproduce, etc."
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              rows={5}
            />
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button onClick={handleSubmit} loading={isSubmitting}>
            Submit Ticket
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
