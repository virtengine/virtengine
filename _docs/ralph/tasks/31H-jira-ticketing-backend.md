# Task 31H: Jira Ticketing Backend Integration

**vibe-kanban ID:** `1b15e422-7f12-4c71-85ce-70a4f361f2f6`

## Overview

| Field | Value |
|-------|-------|
| **ID** | 31H |
| **Title** | feat(support): Jira ticketing backend |
| **Priority** | P2 |
| **Wave** | 4 |
| **Estimated LOC** | 2000 |
| **Duration** | 2 weeks |
| **Dependencies** | None |
| **Blocking** | None |

---

## Problem Statement

Enterprise customers require integration with their existing ticketing systems. Customer support workflows need:

- Automatic ticket creation from portal support requests
- Bi-directional synchronization of ticket status
- SLA tracking and escalation
- Internal notes vs customer-facing comments
- Attachment handling

### Current State Analysis

```
pkg/support/                    ❌ Does not exist
portal/src/pages/support/       ⚠️  Basic contact form only
integration connectors:         ❌ None exist
```

---

## Acceptance Criteria

### AC-1: Jira Integration
- [ ] OAuth 2.0 authentication with Jira Cloud
- [ ] On-premise Jira Server support (app links)
- [ ] Issue creation with custom field mapping
- [ ] Issue status synchronization
- [ ] Comment synchronization (bi-directional)
- [ ] Attachment upload/download

### AC-2: Ticket Management
- [ ] Create ticket from portal
- [ ] View ticket status and history
- [ ] Add comments/replies
- [ ] Attach files (screenshots, logs)
- [ ] Close/reopen tickets

### AC-3: SLA Tracking
- [ ] Define SLA rules by ticket priority
- [ ] Track response time SLA
- [ ] Track resolution time SLA
- [ ] Escalation workflows
- [ ] SLA breach notifications

### AC-4: Internal Operations
- [ ] Admin ticket dashboard
- [ ] Ticket assignment and routing
- [ ] Internal notes (not visible to customer)
- [ ] Ticket merge capability
- [ ] Bulk operations

---

## Technical Requirements

### Jira Client

```go
// pkg/support/jira/client.go

package jira

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "net/url"
    
    "golang.org/x/oauth2"
)

type Config struct {
    SiteURL       string  // https://yoursite.atlassian.net
    ProjectKey    string  // VE
    IssueType     string  // Support Request
    
    // OAuth 2.0 (Jira Cloud)
    ClientID      string
    ClientSecret  string
    RedirectURI   string
    
    // Basic Auth (Jira Server)
    Username      string
    APIToken      string
    
    // Field mappings
    CustomFields  map[string]string  // "account_address" -> "customfield_10001"
}

type Client struct {
    config     Config
    httpClient *http.Client
    baseURL    string
}

func NewClient(cfg Config) (*Client, error) {
    baseURL := cfg.SiteURL + "/rest/api/3"
    
    // Create authenticated HTTP client
    httpClient := &http.Client{}
    
    if cfg.APIToken != "" {
        // Basic auth for Jira Server
        httpClient.Transport = &basicAuthTransport{
            username: cfg.Username,
            token:    cfg.APIToken,
        }
    }
    
    return &Client{
        config:     cfg,
        httpClient: httpClient,
        baseURL:    baseURL,
    }, nil
}

// WithOAuthToken creates a client with OAuth token (Jira Cloud)
func (c *Client) WithOAuthToken(token *oauth2.Token) *Client {
    newClient := *c
    newClient.httpClient = oauth2.NewClient(context.Background(), 
        oauth2.StaticTokenSource(token))
    return &newClient
}

type CreateIssueRequest struct {
    Summary       string
    Description   string
    Priority      string  // Highest, High, Medium, Low, Lowest
    Reporter      string  // Account email
    AccountAddr   string  // VirtEngine account address
    Category      string  // VEID, Billing, Provider, Technical
    Attachments   []Attachment
}

type Attachment struct {
    Filename    string
    ContentType string
    Data        []byte
}

type Issue struct {
    Key         string
    ID          string
    Summary     string
    Description string
    Status      string
    Priority    string
    Created     time.Time
    Updated     time.Time
    Reporter    string
    Assignee    string
    Comments    []Comment
}

type Comment struct {
    ID        string
    Author    string
    Body      string
    Created   time.Time
    Internal  bool  // Internal notes not visible to customer
}

func (c *Client) CreateIssue(ctx context.Context, req CreateIssueRequest) (*Issue, error) {
    // Build ADF (Atlassian Document Format) description
    description := buildADFDocument(req.Description)
    
    payload := map[string]interface{}{
        "fields": map[string]interface{}{
            "project": map[string]string{
                "key": c.config.ProjectKey,
            },
            "issuetype": map[string]string{
                "name": c.config.IssueType,
            },
            "summary":     req.Summary,
            "description": description,
            "priority": map[string]string{
                "name": req.Priority,
            },
        },
    }
    
    // Add custom fields
    if cf, ok := c.config.CustomFields["account_address"]; ok {
        payload["fields"].(map[string]interface{})[cf] = req.AccountAddr
    }
    if cf, ok := c.config.CustomFields["category"]; ok {
        payload["fields"].(map[string]interface{})[cf] = req.Category
    }
    
    resp, err := c.doRequest(ctx, "POST", "/issue", payload)
    if err != nil {
        return nil, fmt.Errorf("create issue: %w", err)
    }
    
    var createResp struct {
        ID   string `json:"id"`
        Key  string `json:"key"`
        Self string `json:"self"`
    }
    if err := json.Unmarshal(resp, &createResp); err != nil {
        return nil, err
    }
    
    // Upload attachments
    for _, att := range req.Attachments {
        if err := c.uploadAttachment(ctx, createResp.Key, att); err != nil {
            // Log but don't fail ticket creation
            fmt.Printf("Failed to upload attachment %s: %v\n", att.Filename, err)
        }
    }
    
    return c.GetIssue(ctx, createResp.Key)
}

func (c *Client) GetIssue(ctx context.Context, key string) (*Issue, error) {
    resp, err := c.doRequest(ctx, "GET", fmt.Sprintf("/issue/%s?expand=renderedFields", key), nil)
    if err != nil {
        return nil, err
    }
    
    var jiraIssue jiraIssueResponse
    if err := json.Unmarshal(resp, &jiraIssue); err != nil {
        return nil, err
    }
    
    return mapJiraIssue(jiraIssue), nil
}

func (c *Client) AddComment(ctx context.Context, issueKey string, body string, internal bool) (*Comment, error) {
    commentBody := buildADFDocument(body)
    
    payload := map[string]interface{}{
        "body": commentBody,
    }
    
    // Internal comments (Service Desk)
    if internal {
        payload["properties"] = []map[string]interface{}{
            {
                "key":   "sd.public.comment",
                "value": map[string]bool{"internal": true},
            },
        }
    }
    
    endpoint := fmt.Sprintf("/issue/%s/comment", issueKey)
    resp, err := c.doRequest(ctx, "POST", endpoint, payload)
    if err != nil {
        return nil, err
    }
    
    var commentResp jiraCommentResponse
    if err := json.Unmarshal(resp, &commentResp); err != nil {
        return nil, err
    }
    
    return &Comment{
        ID:       commentResp.ID,
        Author:   commentResp.Author.DisplayName,
        Body:     commentResp.Body.Content[0].Content[0].Text,
        Created:  commentResp.Created,
        Internal: internal,
    }, nil
}

func (c *Client) TransitionIssue(ctx context.Context, issueKey string, transitionName string) error {
    // Get available transitions
    resp, err := c.doRequest(ctx, "GET", fmt.Sprintf("/issue/%s/transitions", issueKey), nil)
    if err != nil {
        return err
    }
    
    var transitions struct {
        Transitions []struct {
            ID   string `json:"id"`
            Name string `json:"name"`
        } `json:"transitions"`
    }
    json.Unmarshal(resp, &transitions)
    
    // Find matching transition
    var transitionID string
    for _, t := range transitions.Transitions {
        if t.Name == transitionName {
            transitionID = t.ID
            break
        }
    }
    if transitionID == "" {
        return fmt.Errorf("transition %q not available", transitionName)
    }
    
    // Execute transition
    payload := map[string]interface{}{
        "transition": map[string]string{"id": transitionID},
    }
    _, err = c.doRequest(ctx, "POST", fmt.Sprintf("/issue/%s/transitions", issueKey), payload)
    return err
}
```

### Webhook Handler

```go
// pkg/support/jira/webhook.go

package jira

import (
    "context"
    "encoding/json"
    "net/http"
)

type WebhookHandler struct {
    processor WebhookProcessor
    secret    string
}

type WebhookProcessor interface {
    OnIssueCreated(ctx context.Context, issue Issue) error
    OnIssueUpdated(ctx context.Context, issue Issue, changelog Changelog) error
    OnCommentAdded(ctx context.Context, issueKey string, comment Comment) error
}

type WebhookEvent struct {
    WebhookEvent string          `json:"webhookEvent"`
    Issue        jiraIssueResponse `json:"issue"`
    Comment      *jiraCommentResponse `json:"comment"`
    Changelog    *Changelog      `json:"changelog"`
    User         struct {
        DisplayName string `json:"displayName"`
        EmailAddress string `json:"emailAddress"`
    } `json:"user"`
    Timestamp int64 `json:"timestamp"`
}

type Changelog struct {
    Items []ChangelogItem `json:"items"`
}

type ChangelogItem struct {
    Field      string `json:"field"`
    FromString string `json:"fromString"`
    ToString   string `json:"toString"`
}

func (h *WebhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // Verify webhook authenticity (check shared secret header)
    if h.secret != "" {
        secret := r.Header.Get("X-Atlassian-Webhook-Secret")
        if secret != h.secret {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
    }
    
    var event WebhookEvent
    if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
        http.Error(w, "Invalid payload", http.StatusBadRequest)
        return
    }
    
    ctx := r.Context()
    issue := mapJiraIssue(event.Issue)
    
    switch event.WebhookEvent {
    case "jira:issue_created":
        h.processor.OnIssueCreated(ctx, *issue)
        
    case "jira:issue_updated":
        var changelog Changelog
        if event.Changelog != nil {
            changelog = *event.Changelog
        }
        h.processor.OnIssueUpdated(ctx, *issue, changelog)
        
    case "comment_created":
        if event.Comment != nil {
            comment := &Comment{
                ID:      event.Comment.ID,
                Author:  event.Comment.Author.DisplayName,
                Body:    extractTextFromADF(event.Comment.Body),
                Created: event.Comment.Created,
            }
            h.processor.OnCommentAdded(ctx, issue.Key, *comment)
        }
    }
    
    w.WriteHeader(http.StatusOK)
}
```

### Support Service

```go
// pkg/support/service.go

package support

import (
    "context"
    "time"
    
    "github.com/virtengine/virtengine/pkg/support/jira"
)

type Service struct {
    jiraClient *jira.Client
    store      TicketStore
    notifier   Notifier
}

type Ticket struct {
    ID              string
    ExternalID      string  // Jira issue key
    AccountAddress  string
    Subject         string
    Description     string
    Status          TicketStatus
    Priority        TicketPriority
    Category        string
    AssignedTo      string
    CreatedAt       time.Time
    UpdatedAt       time.Time
    ResolvedAt      *time.Time
    
    // SLA tracking
    SLAResponseDue  time.Time
    SLAResolveDue   time.Time
    FirstResponseAt *time.Time
    SLABreached     bool
}

type TicketStatus string

const (
    StatusOpen       TicketStatus = "open"
    StatusInProgress TicketStatus = "in_progress"
    StatusWaiting    TicketStatus = "waiting_on_customer"
    StatusResolved   TicketStatus = "resolved"
    StatusClosed     TicketStatus = "closed"
)

type TicketPriority string

const (
    PriorityCritical TicketPriority = "critical"
    PriorityHigh     TicketPriority = "high"
    PriorityMedium   TicketPriority = "medium"
    PriorityLow      TicketPriority = "low"
)

func (s *Service) CreateTicket(ctx context.Context, req CreateTicketRequest) (*Ticket, error) {
    // Calculate SLA times based on priority
    slaTimes := s.getSLATimes(req.Priority)
    
    ticket := &Ticket{
        ID:             generateID(),
        AccountAddress: req.AccountAddress,
        Subject:        req.Subject,
        Description:    req.Description,
        Status:         StatusOpen,
        Priority:       req.Priority,
        Category:       req.Category,
        CreatedAt:      time.Now(),
        UpdatedAt:      time.Now(),
        SLAResponseDue: time.Now().Add(slaTimes.ResponseTime),
        SLAResolveDue:  time.Now().Add(slaTimes.ResolveTime),
    }
    
    // Create in Jira
    jiraIssue, err := s.jiraClient.CreateIssue(ctx, jira.CreateIssueRequest{
        Summary:     req.Subject,
        Description: req.Description,
        Priority:    mapPriorityToJira(req.Priority),
        Reporter:    req.Email,
        AccountAddr: req.AccountAddress,
        Category:    req.Category,
        Attachments: req.Attachments,
    })
    if err != nil {
        return nil, fmt.Errorf("create jira issue: %w", err)
    }
    
    ticket.ExternalID = jiraIssue.Key
    
    // Store locally
    if err := s.store.CreateTicket(ctx, ticket); err != nil {
        return nil, err
    }
    
    // Notify user
    s.notifier.NotifyTicketCreated(ctx, ticket)
    
    return ticket, nil
}

func (s *Service) AddReply(ctx context.Context, ticketID, accountAddr, message string, attachments []Attachment) error {
    ticket, err := s.store.GetTicket(ctx, ticketID)
    if err != nil {
        return err
    }
    
    if ticket.AccountAddress != accountAddr {
        return errors.New("unauthorized")
    }
    
    // Add comment in Jira
    _, err = s.jiraClient.AddComment(ctx, ticket.ExternalID, message, false)
    if err != nil {
        return fmt.Errorf("add jira comment: %w", err)
    }
    
    // Update local status if needed
    if ticket.Status == StatusWaiting {
        ticket.Status = StatusOpen
        ticket.UpdatedAt = time.Now()
        s.store.UpdateTicket(ctx, ticket)
    }
    
    return nil
}

// SLA Definitions
type SLATimes struct {
    ResponseTime time.Duration
    ResolveTime  time.Duration
}

func (s *Service) getSLATimes(priority TicketPriority) SLATimes {
    switch priority {
    case PriorityCritical:
        return SLATimes{
            ResponseTime: 1 * time.Hour,
            ResolveTime:  4 * time.Hour,
        }
    case PriorityHigh:
        return SLATimes{
            ResponseTime: 4 * time.Hour,
            ResolveTime:  24 * time.Hour,
        }
    case PriorityMedium:
        return SLATimes{
            ResponseTime: 8 * time.Hour,
            ResolveTime:  72 * time.Hour,
        }
    default:
        return SLATimes{
            ResponseTime: 24 * time.Hour,
            ResolveTime:  7 * 24 * time.Hour,
        }
    }
}
```

### Portal Support Component

```tsx
// portal/src/components/support/CreateTicketForm.tsx

'use client';

import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useVirtEngine } from '@virtengine/portal';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { FileUpload } from '@/components/ui/file-upload';

const ticketSchema = z.object({
  subject: z.string().min(5).max(200),
  description: z.string().min(20).max(10000),
  category: z.enum(['veid', 'billing', 'provider', 'technical', 'other']),
  priority: z.enum(['low', 'medium', 'high']),
  attachments: z.array(z.any()).optional(),
});

type TicketFormData = z.infer<typeof ticketSchema>;

export function CreateTicketForm({ onSuccess }: { onSuccess: (ticket: any) => void }) {
  const [isSubmitting, setIsSubmitting] = useState(false);
  const { address } = useVirtEngine();

  const {
    register,
    handleSubmit,
    setValue,
    watch,
    formState: { errors },
  } = useForm<TicketFormData>({
    resolver: zodResolver(ticketSchema),
    defaultValues: {
      priority: 'medium',
      category: 'technical',
    },
  });

  const onSubmit = async (data: TicketFormData) => {
    setIsSubmitting(true);
    try {
      const formData = new FormData();
      formData.append('subject', data.subject);
      formData.append('description', data.description);
      formData.append('category', data.category);
      formData.append('priority', data.priority);
      formData.append('accountAddress', address);
      
      if (data.attachments) {
        data.attachments.forEach((file, i) => {
          formData.append(`attachment_${i}`, file);
        });
      }

      const res = await fetch('/api/support/tickets', {
        method: 'POST',
        body: formData,
      });

      if (!res.ok) throw new Error('Failed to create ticket');

      const ticket = await res.json();
      onSuccess(ticket);
    } catch (error) {
      console.error('Create ticket error:', error);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
      <div>
        <label className="block text-sm font-medium mb-2">Subject</label>
        <Input
          {...register('subject')}
          placeholder="Brief description of your issue"
        />
        {errors.subject && (
          <p className="text-red-500 text-sm mt-1">{errors.subject.message}</p>
        )}
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div>
          <label className="block text-sm font-medium mb-2">Category</label>
          <Select onValueChange={(v) => setValue('category', v as any)}>
            <SelectTrigger>
              <SelectValue placeholder="Select category" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="veid">VEID Identity</SelectItem>
              <SelectItem value="billing">Billing</SelectItem>
              <SelectItem value="provider">Provider Issues</SelectItem>
              <SelectItem value="technical">Technical Support</SelectItem>
              <SelectItem value="other">Other</SelectItem>
            </SelectContent>
          </Select>
        </div>

        <div>
          <label className="block text-sm font-medium mb-2">Priority</label>
          <Select onValueChange={(v) => setValue('priority', v as any)}>
            <SelectTrigger>
              <SelectValue placeholder="Select priority" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="low">Low</SelectItem>
              <SelectItem value="medium">Medium</SelectItem>
              <SelectItem value="high">High</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </div>

      <div>
        <label className="block text-sm font-medium mb-2">Description</label>
        <Textarea
          {...register('description')}
          rows={6}
          placeholder="Please describe your issue in detail..."
        />
        {errors.description && (
          <p className="text-red-500 text-sm mt-1">{errors.description.message}</p>
        )}
      </div>

      <div>
        <label className="block text-sm font-medium mb-2">Attachments</label>
        <FileUpload
          accept="image/*,.pdf,.log,.txt"
          maxFiles={5}
          maxSize={10 * 1024 * 1024}
          onChange={(files) => setValue('attachments', files)}
        />
      </div>

      <Button type="submit" disabled={isSubmitting} className="w-full">
        {isSubmitting ? 'Creating...' : 'Create Support Ticket'}
      </Button>
    </form>
  );
}
```

---

## Directory Structure

```
pkg/support/
├── service.go            # Support service
├── store.go              # Ticket storage
├── notifier.go           # Notification sending
├── sla.go                # SLA tracking
├── jira/
│   ├── client.go         # Jira API client
│   ├── webhook.go        # Webhook handler
│   ├── adf.go            # Atlassian Document Format
│   └── types.go          # Jira types
└── zendesk/              # Future: Zendesk support
    └── client.go

portal/src/
├── app/support/
│   ├── page.tsx          # Support home
│   ├── tickets/
│   │   ├── page.tsx      # My tickets
│   │   └── [id]/
│   │       └── page.tsx  # Ticket detail
│   └── new/
│       └── page.tsx      # Create ticket
└── components/support/
    ├── CreateTicketForm.tsx
    ├── TicketList.tsx
    ├── TicketDetail.tsx
    └── TicketReply.tsx
```

---

## Testing Requirements

### Unit Tests
- Jira API request building
- ADF document generation
- SLA calculation

### Integration Tests
- Jira Cloud sandbox connection
- Webhook processing
- Bi-directional sync

### E2E Tests
- Full ticket lifecycle
- Comment sync
- Status updates

---

## Security Considerations

1. **OAuth Tokens**: Store refresh tokens encrypted
2. **API Keys**: Never log or expose
3. **Attachments**: Scan for malware before upload
4. **PII**: Don't expose customer data in internal notes
5. **Webhook Auth**: Validate shared secret
