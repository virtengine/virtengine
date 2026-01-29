# CDN and WAF Configuration Guide for DDoS Protection

This guide provides recommendations for configuring Content Delivery Networks (CDN) and Web Application Firewalls (WAF) to complement VirtEngine's built-in rate limiting and provide comprehensive DDoS protection.

## Table of Contents

1. [Overview](#overview)
2. [Cloudflare Configuration](#cloudflare-configuration)
3. [AWS CloudFront + WAF](#aws-cloudfront--waf)
4. [Azure Front Door + WAF](#azure-front-door--waf)
5. [NGINX Rate Limiting](#nginx-rate-limiting)
6. [Best Practices](#best-practices)

---

## Overview

VirtEngine provides comprehensive application-level rate limiting, but adding a CDN/WAF layer provides:

- **Network-level DDoS protection** (Layer 3/4)
- **Edge caching** to reduce load on origin servers
- **Geographic distribution** for better performance
- **SSL/TLS termination** at the edge
- **Advanced bot detection** and challenge mechanisms

### Architecture

```
Client → CDN/WAF (Edge Protection) → VirtEngine (Application Rate Limiting) → Backend
```

---

## Cloudflare Configuration

### Basic Setup

1. **Add your domain to Cloudflare**
   - Point your domain's nameservers to Cloudflare

2. **Enable DDoS Protection**
   - Go to Security → DDoS
   - Set to "High" sensitivity
   - Enable "Advanced DDoS Protection"

3. **Configure Rate Limiting Rules**

```javascript
// Cloudflare Rate Limiting Rule
{
  "expression": "(http.request.uri.path contains \"/api/\")",
  "action": "challenge",
  "characteristics": [
    "ip.src"
  ],
  "period": 10,
  "requests_per_period": 100,
  "mitigation_timeout": 600
}
```

4. **Bot Fight Mode**
   - Go to Security → Bots
   - Enable "Bot Fight Mode" or "Super Bot Fight Mode"
   - Configure challenge rules for suspicious traffic

5. **WAF Rules**

```javascript
// Block requests with suspicious patterns
(http.request.uri.query contains "sql" or
 http.request.uri.query contains "union" or
 http.request.uri.query contains "select") and
not cf.client.bot
```

### Advanced Configuration

**Create custom rules for VirtEngine endpoints:**

```javascript
// Protect VEID verification endpoints
{
  "expression": "(http.request.uri.path eq \"/veid/verify\")",
  "action": "managed_challenge",
  "characteristics": ["ip.src"],
  "period": 60,
  "requests_per_period": 10
}

// Protect market endpoints
{
  "expression": "(http.request.uri.path contains \"/market/\")",
  "action": "js_challenge",
  "characteristics": ["ip.src"],
  "period": 10,
  "requests_per_period": 50
}
```

**Page Rules:**

```
URL: virtengine.example.com/api/*
Settings:
  - Security Level: High
  - Cache Level: Bypass
  - Rate Limit: 1000 req/min per IP
```

---

## AWS CloudFront + WAF

### CloudFront Setup

1. **Create CloudFront Distribution**

```bash
aws cloudfront create-distribution \
  --origin-domain-name virtengine.example.com \
  --default-root-object index.html \
  --comment "VirtEngine API Distribution"
```

2. **Configure Caching Behavior**

```json
{
  "PathPattern": "/api/*",
  "TargetOriginId": "virtengine-origin",
  "ViewerProtocolPolicy": "redirect-to-https",
  "AllowedMethods": ["GET", "HEAD", "OPTIONS", "PUT", "POST", "PATCH", "DELETE"],
  "CachedMethods": ["GET", "HEAD"],
  "Compress": true,
  "DefaultTTL": 0,
  "MinTTL": 0,
  "MaxTTL": 0
}
```

### AWS WAF Configuration

1. **Create Web ACL**

```bash
aws wafv2 create-web-acl \
  --name virtengine-waf \
  --scope CLOUDFRONT \
  --default-action Block={} \
  --region us-east-1
```

2. **Add Rate-Based Rules**

```json
{
  "Name": "RateLimitRule",
  "Priority": 1,
  "Statement": {
    "RateBasedStatement": {
      "Limit": 2000,
      "AggregateKeyType": "IP"
    }
  },
  "Action": {
    "Block": {}
  },
  "VisibilityConfig": {
    "SampledRequestsEnabled": true,
    "CloudWatchMetricsEnabled": true,
    "MetricName": "RateLimitRule"
  }
}
```

3. **Geo-Blocking (Optional)**

```json
{
  "Name": "GeoBlockRule",
  "Priority": 2,
  "Statement": {
    "GeoMatchStatement": {
      "CountryCodes": ["CN", "RU", "KP"]
    }
  },
  "Action": {
    "Block": {}
  }
}
```

4. **Managed Rule Groups**

```bash
# Enable AWS Managed Rules
aws wafv2 update-web-acl \
  --name virtengine-waf \
  --scope CLOUDFRONT \
  --add-managed-rule-group-statement \
    Name=AWSManagedRulesCommonRuleSet,VendorName=AWS
```

---

## Azure Front Door + WAF

### Front Door Setup

```bash
# Create Front Door profile
az afd profile create \
  --profile-name virtengine-fd \
  --resource-group virtengine-rg \
  --sku Premium_AzureFrontDoor

# Create endpoint
az afd endpoint create \
  --profile-name virtengine-fd \
  --endpoint-name virtengine-api \
  --resource-group virtengine-rg
```

### Azure WAF Policy

```json
{
  "name": "virtengineWAF",
  "properties": {
    "policySettings": {
      "mode": "Prevention",
      "requestBodyCheck": true,
      "maxRequestBodySizeInKb": 128
    },
    "customRules": {
      "rules": [
        {
          "name": "RateLimitRule",
          "priority": 1,
          "ruleType": "RateLimitRule",
          "rateLimitThreshold": 1000,
          "rateLimitDurationInMinutes": 1,
          "matchConditions": [
            {
              "matchVariable": "RequestUri",
              "operator": "Contains",
              "matchValue": ["/api/"]
            }
          ],
          "action": "Block"
        }
      ]
    },
    "managedRules": {
      "managedRuleSets": [
        {
          "ruleSetType": "OWASP",
          "ruleSetVersion": "3.2"
        },
        {
          "ruleSetType": "Microsoft_BotManagerRuleSet",
          "ruleSetVersion": "1.0"
        }
      ]
    }
  }
}
```

---

## NGINX Rate Limiting

If you're using NGINX as a reverse proxy:

### Configuration

```nginx
# Define rate limit zones
http {
    # IP-based rate limiting
    limit_req_zone $binary_remote_addr zone=ip_limit:10m rate=10r/s;

    # User-based rate limiting (requires authentication)
    limit_req_zone $http_x_user_id zone=user_limit:10m rate=50r/s;

    # Connection limiting
    limit_conn_zone $binary_remote_addr zone=conn_limit:10m;

    # Request body size limit
    client_max_body_size 1m;

    # Timeouts
    client_body_timeout 10s;
    client_header_timeout 10s;

    server {
        listen 443 ssl http2;
        server_name api.virtengine.example.com;

        # Apply rate limits
        limit_req zone=ip_limit burst=20 nodelay;
        limit_conn conn_limit 10;

        # Security headers
        add_header X-Frame-Options "DENY" always;
        add_header X-Content-Type-Options "nosniff" always;
        add_header X-XSS-Protection "1; mode=block" always;
        add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;

        # API endpoints with stricter limits
        location /api/veid/ {
            limit_req zone=ip_limit burst=5 nodelay;
            proxy_pass http://virtengine-backend;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        }

        location /api/market/ {
            limit_req zone=ip_limit burst=10 nodelay;
            proxy_pass http://virtengine-backend;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        }

        # Custom error page for rate limiting
        error_page 429 /rate_limit.json;
        location = /rate_limit.json {
            internal;
            default_type application/json;
            return 429 '{"error":"rate_limit_exceeded","message":"Too many requests. Please try again later."}';
        }
    }
}
```

### ModSecurity WAF (with NGINX)

```nginx
# Load ModSecurity module
load_module modules/ngx_http_modsecurity_module.so;

http {
    # Enable ModSecurity
    modsecurity on;
    modsecurity_rules_file /etc/nginx/modsec/main.conf;
}
```

**ModSecurity Rules:**

```
# Include OWASP Core Rule Set
Include /usr/share/modsecurity-crs/owasp-crs.conf

# Custom rules for VirtEngine
SecRule REQUEST_URI "@contains /api/" \
    "id:1001,phase:1,deny,status:403,msg:'Suspicious API request'"

# Rate limiting rule
SecRule IP:RATE_LIMIT "@gt 100" \
    "id:1002,phase:1,deny,status:429,msg:'Rate limit exceeded',\
    setvar:ip.rate_limit=+1,expirevar:ip.rate_limit=60"
```

---

## Best Practices

### 1. Defense in Depth

Implement multiple layers of protection:

```
Layer 1: CDN/WAF (Cloudflare/AWS/Azure)
  ↓
Layer 2: NGINX/Load Balancer
  ↓
Layer 3: VirtEngine Rate Limiting
  ↓
Layer 4: Application Logic
```

### 2. Monitoring and Alerts

Set up monitoring for:

- **CDN/WAF metrics**: Blocked requests, challenge completions
- **Application metrics**: Rate limit hits, bypass attempts
- **Infrastructure metrics**: CPU, memory, network

**Example Prometheus alerts:**

```yaml
groups:
  - name: rate_limiting
    rules:
      - alert: HighRateLimitBlocks
        expr: rate(virtengine_ratelimit_blocked_requests[5m]) > 100
        for: 5m
        annotations:
          summary: "High rate of blocked requests"

      - alert: PotentialDDoS
        expr: virtengine_ratelimit_bypass_attempts > 1000
        for: 1m
        annotations:
          summary: "Potential DDoS attack detected"
```

### 3. Gradual Tightening

Start with loose limits and gradually tighten based on observed traffic:

1. **Week 1**: Monitor only, no blocking
2. **Week 2**: Enable blocking with generous limits
3. **Week 3**: Adjust limits based on metrics
4. **Week 4**: Enable strict limits

### 4. Whitelist Critical Services

Always whitelist:
- Health check endpoints
- Monitoring systems
- Internal services
- Known good actors (validators, trusted partners)

### 5. Incident Response Plan

Create a runbook for DDoS incidents:

1. **Detection**: Automated alerts trigger
2. **Assessment**: Check metrics dashboards
3. **Mitigation**: Apply emergency rate limits
4. **Communication**: Notify stakeholders
5. **Post-mortem**: Analyze and improve

---

## Testing DDoS Protection

### Load Testing Tools

```bash
# Apache Bench
ab -n 10000 -c 100 https://api.virtengine.example.com/health

# wrk
wrk -t12 -c400 -d30s https://api.virtengine.example.com/api/market/orders

# Vegeta
echo "GET https://api.virtengine.example.com/api/veid/verify" | \
  vegeta attack -rate=100 -duration=30s | vegeta report
```

### Expected Behavior

- **Normal traffic**: All requests succeed
- **Burst traffic**: Some requests delayed (burst allowance)
- **Sustained high traffic**: Rate limiting kicks in, returns 429
- **DDoS attack**: CDN/WAF blocks malicious traffic, application rate limiting catches overflow

---

## Configuration Examples by Use Case

### Public Blockchain Node

```yaml
# High throughput, moderate protection
cloudflare:
  rate_limit: 5000/min per IP
  bot_fight: enabled
  ddos_sensitivity: medium

virtengine:
  ip_limits:
    requests_per_second: 50
    requests_per_minute: 2000
```

### Private Enterprise Deployment

```yaml
# Lower throughput, strict protection
aws_waf:
  rate_limit: 500/min per IP
  geo_blocking: enabled
  managed_rules: enabled

virtengine:
  ip_limits:
    requests_per_second: 10
    requests_per_minute: 300
  whitelist: [corporate_ips]
```

### Development/Testnet

```yaml
# Generous limits for testing
rate_limiting:
  enabled: true
  ip_limits:
    requests_per_second: 100
    requests_per_minute: 5000
  bypass_detection: disabled
```

---

## Support and Resources

- **VirtEngine Docs**: https://docs.virtengine.com/security/rate-limiting
- **Cloudflare Docs**: https://developers.cloudflare.com/waf/
- **AWS WAF Docs**: https://docs.aws.amazon.com/waf/
- **Azure WAF Docs**: https://docs.microsoft.com/azure/web-application-firewall/

For issues or questions, please contact security@virtengine.com or file an issue on GitHub.
