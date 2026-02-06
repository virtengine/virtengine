# Task 31C: Push Notification System

**vibe-kanban ID:** `656bf189-e285-4a6f-935e-7cea42af65b4`

## Overview

| Field | Value |
|-------|-------|
| **ID** | 31C |
| **Title** | feat(notifications): Push notification system |
| **Priority** | P1 |
| **Wave** | 3 |
| **Estimated LOC** | 4000 |
| **Duration** | 3-4 weeks |
| **Dependencies** | None |
| **Blocking** | None |

---

## Problem Statement

Users need real-time notifications for important events:
- VEID verification status changes
- Order status updates
- Escrow deposits and settlements
- Provider availability alerts
- Security events (new login, MFA changes)

Currently, only outbound webhooks exist. Users must poll or manually check status.

### Current State Analysis

```
pkg/webhooks/                   ✅ Outbound webhook delivery exists
pkg/notifications/              ❌ Does not exist
portal/src/components/          ❌ No notification UI
lib/portal/notifications/       ❌ No client-side handling
```

---

## Acceptance Criteria

### AC-1: Firebase/APNs Integration
- [ ] Firebase Cloud Messaging (FCM) setup and SDK integration
- [ ] Apple Push Notification Service (APNs) integration
- [ ] Device token registration and storage
- [ ] Topic-based notification routing
- [ ] Silent push for background updates

### AC-2: Notification Preferences Store
- [ ] User notification settings schema
- [ ] Per-channel preferences (push/email/in-app)
- [ ] Per-category toggles (orders, VEID, security, marketing)
- [ ] Quiet hours configuration
- [ ] Frequency controls (immediate/digest)

### AC-3: Email Notification System
- [ ] Email template engine setup (MJML or React Email)
- [ ] Order confirmation email template
- [ ] VEID status change notifications
- [ ] Security alert emails (new login, MFA change)
- [ ] Weekly digest generation
- [ ] Unsubscribe link handling (CAN-SPAM/GDPR compliance)

### AC-4: Portal Notification Center UI
- [ ] Notification bell icon with unread count
- [ ] Notification dropdown/drawer component
- [ ] Mark as read/unread functionality
- [ ] Notification preferences settings page
- [ ] WebSocket real-time updates

---

## Technical Requirements

### Notification Service

```go
// pkg/notifications/service.go

package notifications

import (
    "context"
)

type NotificationType string

const (
    NotificationTypeVEIDStatus    NotificationType = "veid_status"
    NotificationTypeOrderUpdate   NotificationType = "order_update"
    NotificationTypeEscrowDeposit NotificationType = "escrow_deposit"
    NotificationTypeSecurityAlert NotificationType = "security_alert"
    NotificationTypeProviderAlert NotificationType = "provider_alert"
)

type Notification struct {
    ID          string
    UserAddress string
    Type        NotificationType
    Title       string
    Body        string
    Data        map[string]string
    CreatedAt   time.Time
    ReadAt      *time.Time
    Channels    []Channel  // push, email, in_app
}

type Channel string

const (
    ChannelPush  Channel = "push"
    ChannelEmail Channel = "email"
    ChannelInApp Channel = "in_app"
)

type Service interface {
    Send(ctx context.Context, notif Notification) error
    SendBatch(ctx context.Context, notifs []Notification) error
    GetUserNotifications(ctx context.Context, userAddr string, opts ListOptions) ([]Notification, error)
    MarkAsRead(ctx context.Context, userAddr string, notifIDs []string) error
    UpdatePreferences(ctx context.Context, userAddr string, prefs Preferences) error
    GetPreferences(ctx context.Context, userAddr string) (Preferences, error)
}

type Preferences struct {
    UserAddress   string
    Channels      map[NotificationType][]Channel
    QuietHours    *QuietHours
    DigestEnabled bool
    DigestTime    string // "09:00" UTC
}

type QuietHours struct {
    Enabled   bool
    StartHour int  // 0-23
    EndHour   int  // 0-23
    Timezone  string
}
```

### Firebase Integration

```go
// pkg/notifications/firebase/client.go

package firebase

import (
    "context"
    firebase "firebase.google.com/go/v4"
    "firebase.google.com/go/v4/messaging"
)

type Client struct {
    app       *firebase.App
    messaging *messaging.Client
}

func NewClient(ctx context.Context, credentialsPath string) (*Client, error) {
    opt := option.WithCredentialsFile(credentialsPath)
    app, err := firebase.NewApp(ctx, nil, opt)
    if err != nil {
        return nil, fmt.Errorf("firebase app init: %w", err)
    }

    client, err := app.Messaging(ctx)
    if err != nil {
        return nil, fmt.Errorf("firebase messaging: %w", err)
    }

    return &Client{app: app, messaging: client}, nil
}

func (c *Client) SendToDevice(ctx context.Context, token string, notif Notification) error {
    message := &messaging.Message{
        Token: token,
        Notification: &messaging.Notification{
            Title: notif.Title,
            Body:  notif.Body,
        },
        Data: notif.Data,
        Android: &messaging.AndroidConfig{
            Priority: "high",
            Notification: &messaging.AndroidNotification{
                ClickAction: "FLUTTER_NOTIFICATION_CLICK",
            },
        },
        APNS: &messaging.APNSConfig{
            Payload: &messaging.APNSPayload{
                Aps: &messaging.Aps{
                    Sound: "default",
                    Badge: &notif.UnreadCount,
                },
            },
        },
    }

    _, err := c.messaging.Send(ctx, message)
    return err
}

func (c *Client) SendToTopic(ctx context.Context, topic string, notif Notification) error {
    message := &messaging.Message{
        Topic: topic,
        Notification: &messaging.Notification{
            Title: notif.Title,
            Body:  notif.Body,
        },
        Data: notif.Data,
    }

    _, err := c.messaging.Send(ctx, message)
    return err
}
```

### Email Templates

```tsx
// pkg/email/templates/order-confirmation.tsx

import {
  Body,
  Container,
  Head,
  Heading,
  Html,
  Img,
  Preview,
  Section,
  Text,
  Button,
} from '@react-email/components';

interface OrderConfirmationProps {
  orderID: string;
  providerName: string;
  serviceName: string;
  amount: string;
  currency: string;
  dashboardURL: string;
}

export const OrderConfirmation = ({
  orderID,
  providerName,
  serviceName,
  amount,
  currency,
  dashboardURL,
}: OrderConfirmationProps) => (
  <Html>
    <Head />
    <Preview>Your order {orderID} has been confirmed</Preview>
    <Body style={main}>
      <Container style={container}>
        <Img
          src="https://virtengine.com/logo.png"
          width="120"
          height="40"
          alt="VirtEngine"
        />
        <Heading style={h1}>Order Confirmed</Heading>
        <Section>
          <Text style={text}>
            Your order <strong>{orderID}</strong> has been confirmed.
          </Text>
          <Text style={text}>
            <strong>Provider:</strong> {providerName}<br />
            <strong>Service:</strong> {serviceName}<br />
            <strong>Amount:</strong> {amount} {currency}
          </Text>
          <Button style={button} href={dashboardURL}>
            View Order
          </Button>
        </Section>
        <Text style={footer}>
          VirtEngine - Decentralized Cloud Computing
        </Text>
      </Container>
    </Body>
  </Html>
);

const main = { backgroundColor: '#f6f9fc', fontFamily: 'sans-serif' };
const container = { margin: '0 auto', padding: '40px 0' };
const h1 = { color: '#1f2937', fontSize: '24px' };
const text = { color: '#4b5563', fontSize: '16px' };
const button = { backgroundColor: '#4f46e5', color: '#fff', padding: '12px 24px' };
const footer = { color: '#9ca3af', fontSize: '12px', marginTop: '32px' };
```

### Portal Notification Component

```tsx
// portal/src/components/notifications/NotificationCenter.tsx

'use client';

import { useState, useEffect } from 'react';
import { Bell } from 'lucide-react';
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover';
import { ScrollArea } from '@/components/ui/scroll-area';
import { useVirtEngine } from '@virtengine/portal';

interface Notification {
  id: string;
  type: string;
  title: string;
  body: string;
  createdAt: string;
  readAt: string | null;
}

export function NotificationCenter() {
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [unreadCount, setUnreadCount] = useState(0);
  const { wsClient } = useVirtEngine();

  useEffect(() => {
    // Subscribe to real-time notifications
    if (wsClient) {
      wsClient.on('notification', (notif: Notification) => {
        setNotifications(prev => [notif, ...prev]);
        setUnreadCount(prev => prev + 1);
      });
    }
    
    // Load initial notifications
    fetchNotifications();
  }, [wsClient]);

  const fetchNotifications = async () => {
    const res = await fetch('/api/notifications');
    const data = await res.json();
    setNotifications(data.notifications);
    setUnreadCount(data.unreadCount);
  };

  const markAsRead = async (ids: string[]) => {
    await fetch('/api/notifications/read', {
      method: 'POST',
      body: JSON.stringify({ ids }),
    });
    setNotifications(prev =>
      prev.map(n =>
        ids.includes(n.id) ? { ...n, readAt: new Date().toISOString() } : n
      )
    );
    setUnreadCount(prev => Math.max(0, prev - ids.length));
  };

  return (
    <Popover>
      <PopoverTrigger asChild>
        <button className="relative p-2 rounded-full hover:bg-gray-100">
          <Bell className="h-5 w-5" />
          {unreadCount > 0 && (
            <span className="absolute top-0 right-0 h-4 w-4 bg-red-500 rounded-full text-xs text-white flex items-center justify-center">
              {unreadCount > 9 ? '9+' : unreadCount}
            </span>
          )}
        </button>
      </PopoverTrigger>
      <PopoverContent className="w-80 p-0" align="end">
        <div className="p-4 border-b">
          <h3 className="font-semibold">Notifications</h3>
        </div>
        <ScrollArea className="h-[400px]">
          {notifications.length === 0 ? (
            <p className="p-4 text-center text-gray-500">No notifications</p>
          ) : (
            notifications.map(notif => (
              <NotificationItem
                key={notif.id}
                notification={notif}
                onRead={() => markAsRead([notif.id])}
              />
            ))
          )}
        </ScrollArea>
        <div className="p-2 border-t">
          <button
            className="w-full text-sm text-blue-600 hover:underline"
            onClick={() => markAsRead(notifications.filter(n => !n.readAt).map(n => n.id))}
          >
            Mark all as read
          </button>
        </div>
      </PopoverContent>
    </Popover>
  );
}
```

---

## Directory Structure

```
pkg/notifications/
├── service.go            # Main notification service
├── preferences.go        # Preferences management
├── store.go              # Notification storage
├── firebase/
│   └── client.go         # FCM client
├── apns/
│   └── client.go         # APNs client
└── events/
    └── handler.go        # Chain event handler

pkg/email/
├── sender.go             # Email sending service
├── templates/
│   ├── order-confirmation.tsx
│   ├── veid-status.tsx
│   ├── security-alert.tsx
│   └── weekly-digest.tsx
└── mjml/
    └── (alternative MJML templates)

portal/src/components/notifications/
├── NotificationCenter.tsx
├── NotificationItem.tsx
├── NotificationPreferences.tsx
└── hooks/
    └── useNotifications.ts
```

---

## Testing Requirements

### Unit Tests
- Notification routing based on preferences
- Quiet hours filtering
- Email template rendering

### Integration Tests
- FCM delivery (use test tokens)
- Email delivery (use Mailhog/Mailtrap)
- WebSocket notification delivery

### E2E Tests
- Full notification flow from chain event to UI
- Preference changes affecting delivery
- Unsubscribe flow

---

## Security Considerations

1. **Device Token Storage**: Encrypt tokens at rest
2. **Email Deliverability**: Configure SPF/DKIM/DMARC
3. **Unsubscribe Compliance**: One-click unsubscribe (CAN-SPAM/GDPR)
4. **Rate Limiting**: Prevent notification spam
5. **Content Sanitization**: Never include sensitive data in notifications
