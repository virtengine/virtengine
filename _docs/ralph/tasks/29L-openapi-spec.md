# Task 29L: OpenAPI Spec for Provider API

**ID:** 29L  
**Title:** docs(portal): OpenAPI spec for provider API  
**Priority:** P2 (Medium)  
**Wave:** 4 (After 29F)  
**Estimated LOC:** ~500  
**Dependencies:** 29F (Enhanced portal_api.go)  
**Blocking:** None  

---

## Problem Statement

The provider Portal API lacks formal documentation:

1. **No OpenAPI spec** - Third parties can't easily integrate
2. **Manual documentation** - Prone to getting out of sync
3. **No TypeScript generation** - Types maintained manually
4. **No client libraries** - Must write custom clients

---

## Acceptance Criteria

### AC-1: OpenAPI 3.0 Specification
- [ ] Complete OpenAPI 3.0.3 spec in YAML
- [ ] All endpoints documented
- [ ] Request/response schemas defined
- [ ] Authentication documented
- [ ] Error responses documented

### AC-2: Endpoint Documentation
- [ ] Health endpoints
- [ ] Deployment endpoints
- [ ] Organization endpoints
- [ ] Ticket endpoints
- [ ] Billing endpoints
- [ ] Provider info endpoints

### AC-3: Schema Definitions
- [ ] Deployment schema
- [ ] Lease schema
- [ ] Metrics schema
- [ ] Invoice schema
- [ ] Ticket schema
- [ ] Error schema

### AC-4: Authentication Documentation
- [ ] Wallet-signed auth flow
- [ ] HMAC auth (legacy)
- [ ] Public vs protected endpoints
- [ ] Token refresh flow

### AC-5: Code Generation
- [ ] Generate TypeScript types from spec
- [ ] Generate Go types from spec
- [ ] CI validation of spec
- [ ] Auto-generate on changes

### AC-6: Documentation Site
- [ ] Swagger UI deployment
- [ ] Redoc alternative
- [ ] Interactive API testing
- [ ] Code samples

---

## Technical Requirements

### OpenAPI Specification

```yaml
# api/openapi/portal_api.yaml

openapi: 3.0.3
info:
  title: VirtEngine Provider Portal API
  description: |
    REST API for interacting with VirtEngine provider portals.
    
    ## Authentication
    
    Most endpoints require wallet-signed authentication. Include the following headers:
    
    - `X-VE-Address`: Your VirtEngine wallet address
    - `X-VE-Timestamp`: Unix milliseconds timestamp
    - `X-VE-Nonce`: Random 32-character hex string
    - `X-VE-Signature`: Base64-encoded signature
    - `X-VE-PubKey`: Base64-encoded public key
    
    The signature is created by signing the canonical request data using ADR-036.
    
  version: 1.0.0
  contact:
    name: VirtEngine
    url: https://virtengine.io
    email: api@virtengine.io
  license:
    name: Apache 2.0
    url: https://www.apache.org/licenses/LICENSE-2.0.html

servers:
  - url: https://{provider-endpoint}
    description: Provider API endpoint
    variables:
      provider-endpoint:
        default: provider.example.com:8443
        description: Provider's public API endpoint

tags:
  - name: Health
    description: Provider health and status
  - name: Deployments
    description: Deployment management
  - name: Organizations
    description: Organization management
  - name: Tickets
    description: Support ticket management
  - name: Billing
    description: Invoices and usage
  - name: Provider
    description: Provider information

security:
  - walletAuth: []

paths:
  /api/v1/health:
    get:
      tags:
        - Health
      summary: Check provider health
      description: Returns the health status of the provider
      operationId: getHealth
      security: []  # Public endpoint
      responses:
        '200':
          description: Provider is healthy
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HealthResponse'
        '503':
          description: Provider is unhealthy
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'

  /api/v1/deployments:
    get:
      tags:
        - Deployments
      summary: List deployments
      description: Returns all deployments for the authenticated user
      operationId: listDeployments
      parameters:
        - $ref: '#/components/parameters/limitParam'
        - $ref: '#/components/parameters/cursorParam'
        - name: status
          in: query
          schema:
            type: string
            enum: [active, closed, pending]
      responses:
        '200':
          description: List of deployments
          content:
            application/json:
              schema:
                type: object
                properties:
                  deployments:
                    type: array
                    items:
                      $ref: '#/components/schemas/Deployment'
                  next_cursor:
                    type: string
        '401':
          $ref: '#/components/responses/Unauthorized'

  /api/v1/deployments/{deploymentId}:
    get:
      tags:
        - Deployments
      summary: Get deployment details
      operationId: getDeployment
      parameters:
        - $ref: '#/components/parameters/deploymentId'
      responses:
        '200':
          description: Deployment details
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Deployment'
        '404':
          $ref: '#/components/responses/NotFound'

  /api/v1/deployments/{deploymentId}/status:
    get:
      tags:
        - Deployments
      summary: Get deployment status
      operationId: getDeploymentStatus
      parameters:
        - $ref: '#/components/parameters/deploymentId'
      responses:
        '200':
          description: Deployment status
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/DeploymentStatus'

  /api/v1/deployments/{deploymentId}/logs:
    get:
      tags:
        - Deployments
      summary: Get deployment logs
      operationId: getDeploymentLogs
      parameters:
        - $ref: '#/components/parameters/deploymentId'
        - name: tail
          in: query
          schema:
            type: integer
            default: 100
        - name: since
          in: query
          schema:
            type: string
            format: date-time
        - name: timestamps
          in: query
          schema:
            type: boolean
            default: false
      responses:
        '200':
          description: Log lines
          content:
            application/json:
              schema:
                type: array
                items:
                  type: string

  /api/v1/deployments/{deploymentId}/metrics:
    get:
      tags:
        - Deployments
      summary: Get deployment metrics
      operationId: getDeploymentMetrics
      parameters:
        - $ref: '#/components/parameters/deploymentId'
      responses:
        '200':
          description: Current metrics
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ResourceMetrics'

  /api/v1/deployments/{deploymentId}/shell:
    get:
      tags:
        - Deployments
      summary: Connect to deployment shell
      description: WebSocket endpoint for interactive shell
      operationId: connectShell
      parameters:
        - $ref: '#/components/parameters/deploymentId'
        - name: service
          in: query
          schema:
            type: string
      responses:
        '101':
          description: WebSocket connection established
        '401':
          $ref: '#/components/responses/Unauthorized'

  /api/v1/deployments/{deploymentId}/actions:
    post:
      tags:
        - Deployments
      summary: Perform deployment action
      operationId: performAction
      parameters:
        - $ref: '#/components/parameters/deploymentId'
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required:
                - action
              properties:
                action:
                  type: string
                  enum: [start, stop, restart]
      responses:
        '200':
          description: Action performed
          content:
            application/json:
              schema:
                type: object
                properties:
                  success:
                    type: boolean
                  message:
                    type: string

  /api/v1/organizations:
    get:
      tags:
        - Organizations
      summary: List organizations
      operationId: listOrganizations
      responses:
        '200':
          description: List of organizations
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/Organization'

  /api/v1/organizations/{orgId}/members:
    get:
      tags:
        - Organizations
      summary: List organization members
      operationId: getOrganizationMembers
      parameters:
        - name: orgId
          in: path
          required: true
          schema:
            type: string
      responses:
        '200':
          description: List of members
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/Member'

  /api/v1/tickets:
    get:
      tags:
        - Tickets
      summary: List support tickets
      operationId: listTickets
      parameters:
        - name: status
          in: query
          schema:
            type: string
            enum: [open, in_progress, resolved, closed]
        - name: deployment_id
          in: query
          schema:
            type: string
      responses:
        '200':
          description: List of tickets
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/Ticket'
    post:
      tags:
        - Tickets
      summary: Create support ticket
      operationId: createTicket
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateTicketRequest'
      responses:
        '201':
          description: Ticket created
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Ticket'

  /api/v1/invoices:
    get:
      tags:
        - Billing
      summary: List invoices
      operationId: listInvoices
      parameters:
        - name: status
          in: query
          schema:
            type: string
            enum: [pending, paid, overdue]
        - $ref: '#/components/parameters/limitParam'
        - $ref: '#/components/parameters/cursorParam'
      responses:
        '200':
          description: List of invoices
          content:
            application/json:
              schema:
                type: object
                properties:
                  invoices:
                    type: array
                    items:
                      $ref: '#/components/schemas/Invoice'
                  next_cursor:
                    type: string

  /api/v1/usage:
    get:
      tags:
        - Billing
      summary: Get current usage
      operationId: getCurrentUsage
      responses:
        '200':
          description: Current usage summary
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/UsageSummary'

  /api/v1/provider/info:
    get:
      tags:
        - Provider
      summary: Get provider info
      operationId: getProviderInfo
      security: []
      responses:
        '200':
          description: Provider information
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ProviderInfo'

  /api/v1/provider/pricing:
    get:
      tags:
        - Provider
      summary: Get provider pricing
      operationId: getProviderPricing
      security: []
      responses:
        '200':
          description: Pricing information
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Pricing'

components:
  securitySchemes:
    walletAuth:
      type: apiKey
      in: header
      name: X-VE-Signature
      description: |
        Wallet-signed authentication. Required headers:
        - X-VE-Address: Wallet address
        - X-VE-Timestamp: Unix milliseconds
        - X-VE-Nonce: Random hex string
        - X-VE-Signature: Base64 signature
        - X-VE-PubKey: Base64 public key

  parameters:
    deploymentId:
      name: deploymentId
      in: path
      required: true
      schema:
        type: string
      description: Deployment/Lease ID
    limitParam:
      name: limit
      in: query
      schema:
        type: integer
        minimum: 1
        maximum: 100
        default: 20
    cursorParam:
      name: cursor
      in: query
      schema:
        type: string

  responses:
    Unauthorized:
      description: Authentication required
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'
    NotFound:
      description: Resource not found
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'

  schemas:
    Error:
      type: object
      required:
        - error
        - message
      properties:
        error:
          type: string
        message:
          type: string
        code:
          type: string

    HealthResponse:
      type: object
      properties:
        status:
          type: string
          enum: [ok, degraded, down]
        version:
          type: string
        uptime:
          type: integer

    Deployment:
      type: object
      properties:
        id:
          type: string
        owner:
          type: string
        provider:
          type: string
        state:
          type: string
          enum: [pending, active, closed]
        created_at:
          type: string
          format: date-time
        resources:
          $ref: '#/components/schemas/ResourceMetrics'

    DeploymentStatus:
      type: object
      properties:
        lease_id:
          type: string
        state:
          type: string
        replicas:
          type: object
          properties:
            ready:
              type: integer
            total:
              type: integer
        services:
          type: array
          items:
            $ref: '#/components/schemas/ServiceStatus'

    ServiceStatus:
      type: object
      properties:
        name:
          type: string
        state:
          type: string
        replicas:
          type: integer

    ResourceMetrics:
      type: object
      properties:
        cpu:
          $ref: '#/components/schemas/UsageMetric'
        memory:
          $ref: '#/components/schemas/UsageMetric'
        storage:
          $ref: '#/components/schemas/UsageMetric'

    UsageMetric:
      type: object
      properties:
        usage:
          type: number
        limit:
          type: number

    Organization:
      type: object
      properties:
        id:
          type: string
        name:
          type: string
        description:
          type: string
        created_at:
          type: string
          format: date-time

    Member:
      type: object
      properties:
        address:
          type: string
        role:
          type: string
          enum: [admin, member, viewer]
        joined_at:
          type: string
          format: date-time

    Ticket:
      type: object
      properties:
        id:
          type: string
        deployment_id:
          type: string
        subject:
          type: string
        description:
          type: string
        status:
          type: string
          enum: [open, in_progress, waiting_customer, resolved, closed]
        priority:
          type: string
          enum: [low, medium, high, critical]
        created_at:
          type: string
          format: date-time

    CreateTicketRequest:
      type: object
      required:
        - deployment_id
        - subject
        - description
      properties:
        deployment_id:
          type: string
        subject:
          type: string
        description:
          type: string
        category:
          type: string
          enum: [technical, billing, general, security]
        priority:
          type: string
          enum: [low, medium, high, critical]
          default: medium

    Invoice:
      type: object
      properties:
        id:
          type: string
        number:
          type: string
        status:
          type: string
          enum: [pending, paid, overdue]
        total:
          type: string
        currency:
          type: string
        due_date:
          type: string
          format: date-time

    UsageSummary:
      type: object
      properties:
        period:
          type: object
          properties:
            start:
              type: string
              format: date-time
            end:
              type: string
              format: date-time
        total_cost:
          type: string
        resources:
          $ref: '#/components/schemas/ResourceMetrics'

    ProviderInfo:
      type: object
      properties:
        address:
          type: string
        name:
          type: string
        version:
          type: string
        capabilities:
          type: array
          items:
            type: string
        region:
          type: string

    Pricing:
      type: object
      properties:
        cpu:
          $ref: '#/components/schemas/ResourcePrice'
        memory:
          $ref: '#/components/schemas/ResourcePrice'
        storage:
          $ref: '#/components/schemas/ResourcePrice'
        currency:
          type: string

    ResourcePrice:
      type: object
      properties:
        unit:
          type: string
        price:
          type: string
        interval:
          type: string
          enum: [hourly, daily, monthly]
```

### Code Generation Script

```bash
#!/bin/bash
# scripts/generate-api-types.sh

set -e

SPEC_FILE="api/openapi/portal_api.yaml"
TS_OUTPUT="lib/portal/src/provider-api/generated"
GO_OUTPUT="pkg/provider_daemon/api/generated"

# Validate spec
npx @redocly/cli lint "$SPEC_FILE"

# Generate TypeScript types
npx openapi-typescript "$SPEC_FILE" -o "$TS_OUTPUT/types.ts"

# Generate Go types
oapi-codegen -generate types -package generated "$SPEC_FILE" > "$GO_OUTPUT/types.go"

echo "API types generated successfully"
```

### CI Validation

```yaml
# .github/workflows/api-spec.yaml
name: API Spec Validation

on:
  push:
    paths:
      - 'api/openapi/**'
      - 'pkg/provider_daemon/**'
  pull_request:
    paths:
      - 'api/openapi/**'
      - 'pkg/provider_daemon/**'

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Install Redocly CLI
        run: npm install -g @redocly/cli
      
      - name: Lint OpenAPI spec
        run: redocly lint api/openapi/portal_api.yaml
      
      - name: Generate types
        run: ./scripts/generate-api-types.sh
      
      - name: Check for changes
        run: |
          if [[ -n $(git status --porcelain) ]]; then
            echo "Generated types are out of date. Run ./scripts/generate-api-types.sh"
            exit 1
          fi
```

---

## Files to Create

| Path | Description | Est. Lines |
|------|-------------|------------|
| `api/openapi/portal_api.yaml` | OpenAPI specification | 800 |
| `scripts/generate-api-types.sh` | Type generation script | 30 |
| `.github/workflows/api-spec.yaml` | CI validation | 40 |
| `lib/portal/src/provider-api/generated/types.ts` | Generated TS types | auto |
| `pkg/provider_daemon/api/generated/types.go` | Generated Go types | auto |
| `docs/api/README.md` | API documentation | 100 |

**Total: ~970 lines (manual) + auto-generated**

---

## Validation Checklist

- [ ] OpenAPI spec is valid (Redocly lint passes)
- [ ] All endpoints documented
- [ ] TypeScript types generated
- [ ] Go types generated
- [ ] CI validation passing
- [ ] Swagger UI accessible
- [ ] Code samples provided

---

## Vibe-Kanban Task ID

`f657eef7-b495-4369-92b6-2d530cb46f52`
