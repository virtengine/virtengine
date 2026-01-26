# VirtEngine User Guide

**Version:** 1.0.0  
**Date:** 2026-01-24  
**Task Reference:** VE-803

---

## Table of Contents

1. [Getting Started](#getting-started)
2. [Account Setup](#account-setup)
3. [Identity Verification (VEID)](#identity-verification-veid)
4. [Multi-Factor Authentication](#multi-factor-authentication)
5. [Marketplace](#marketplace)
6. [HPC Jobs](#hpc-jobs)
7. [Account Security](#account-security)
8. [FAQ](#faq)

---

## Getting Started

Welcome to VirtEngine! This guide will help you:

1. Create and secure your account
2. Complete identity verification
3. Set up multi-factor authentication
4. Purchase cloud resources
5. Submit HPC jobs

### What is VirtEngine?

VirtEngine is a decentralized cloud marketplace where you can:

- **Purchase compute resources** from verified providers
- **Run HPC workloads** on distributed supercomputer clusters
- **Maintain privacy** with end-to-end encryption
- **Build trust** through identity verification

## Account Setup

### Creating Your Account

1. **Visit the Portal**: Navigate to [portal.virtengine.com](https://portal.virtengine.com)

2. **Create Wallet**: Choose one of:
   - Generate new wallet (recommended for new users)
   - Import existing wallet via mnemonic
   - Connect hardware wallet (Ledger)

3. **Secure Your Mnemonic**:
   > ⚠️ **CRITICAL**: Your 24-word mnemonic is the only way to recover your account. Write it down and store it securely offline. Never share it with anyone.

4. **Fund Your Account**: Obtain VE tokens via:
   - Exchange purchase
   - Faucet (testnet only)
   - Transfer from another account

### Account Dashboard

Your dashboard shows:

- **Balance**: Available VE tokens
- **Identity Score**: Your verified identity level (0-100)
- **MFA Status**: Enrolled authentication factors
- **Active Resources**: Running workloads and orders
- **Recent Activity**: Transaction history

## Identity Verification (VEID)

### Why Verify Your Identity?

Identity verification enables:
- Access to premium marketplace offerings
- Higher transaction limits
- Trust with providers
- Reduced fraud risk

### Identity Score Levels

| Score Range | Level | Access |
|-------------|-------|--------|
| 0-29 | Basic | Limited marketplace |
| 30-49 | Verified | Standard marketplace |
| 50-74 | Trusted | Full marketplace + HPC |
| 75-100 | Premium | Priority access + reduced fees |

### Verification Process

1. **Open VEID Section**: From dashboard, click "Verify Identity"

2. **Select Scopes**: Choose which identity elements to verify:
   - 📄 **Government ID**: Passport, driver's license, national ID
   - 📸 **Selfie**: Live photo with liveness detection
   - 📧 **Email**: Email ownership verification
   - 🔗 **Social**: OAuth from major providers
   - 🌐 **Domain**: DNS/HTTP verification (for businesses)

3. **Capture Documents**:
   - Use the VirtEngine mobile app for best results
   - Follow on-screen guidance for quality capture
   - Ensure good lighting and clear images

4. **Submit for Verification**:
   - Your data is encrypted before submission
   - Only validators can decrypt for scoring
   - No raw identity data is stored on-chain

5. **Receive Your Score**:
   - Processing typically takes 1-5 minutes
   - Score is computed by validator consensus
   - Score expires after 1 year

### Privacy Protections

Your identity data is protected by:

- **End-to-end encryption**: Data encrypted to validators only
- **Minimal storage**: Only hashes and score stored on-chain
- **No raw data**: Validators delete raw data after scoring
- **User control**: You control which scopes to share

## Multi-Factor Authentication

### Why Enable MFA?

MFA protects your account by requiring additional verification for:
- Account recovery
- High-value transactions
- Sensitive settings changes
- Key rotation

### Available MFA Factors

| Factor | Security Level | Convenience |
|--------|---------------|-------------|
| 📱 TOTP App | High | Medium |
| 🔑 FIDO2/WebAuthn | Very High | High |
| 📧 Email | Medium | High |
| 📲 SMS | Medium | High |
| 🔢 Backup Codes | High | Low (one-time use) |

### Setting Up MFA

1. **Navigate to Security**: Dashboard → Settings → Security

2. **Add MFA Factor**:
   - Click "Add Authentication Factor"
   - Select factor type
   - Follow setup wizard

3. **TOTP Setup**:
   - Scan QR code with authenticator app (Google Authenticator, Authy, etc.)
   - Enter 6-digit code to confirm

4. **FIDO2/WebAuthn Setup**:
   - Insert security key or enable biometric
   - Follow browser prompts
   - Name your device

5. **Backup Codes**:
   - Generate 10 single-use backup codes
   - Store securely offline
   - Each code can only be used once

### MFA for Transactions

Sensitive transactions require MFA verification:

| Transaction Type | MFA Required |
|-----------------|--------------|
| Normal transfer | No |
| High-value transfer (>1000 VE) | Yes |
| Key rotation | Yes |
| Account recovery | Yes (2+ factors) |
| Provider registration | Yes |

## Marketplace

### Browsing Offerings

1. **Open Marketplace**: Click "Marketplace" in navigation

2. **Filter Offerings**:
   - **Region**: Geographic location
   - **Resource Type**: Compute, Storage, GPU
   - **Price Range**: Min/max hourly rate
   - **Provider Score**: Minimum trust score
   - **Identity Requirement**: Required identity level

3. **Compare Providers**:
   - View benchmark scores
   - Check availability
   - Read reviews
   - Compare pricing

### Purchasing Resources

1. **Select Offering**: Click on desired offering

2. **Configure Order**:
   - Specify quantity
   - Set duration
   - Add configuration (if applicable)

3. **Review & Confirm**:
   - Check total cost
   - Verify identity meets requirements
   - Confirm order

4. **Order Lifecycle**:
   - **Pending**: Waiting for provider bids
   - **Allocated**: Provider assigned
   - **Provisioning**: Resources being created
   - **Active**: Resources ready to use
   - **Terminated**: Order complete

### Managing Orders

From "My Orders":

- View order status
- Access provisioned resources
- Monitor usage
- Terminate early (if needed)
- View invoices

### Encrypted Order Details

Your order configuration is encrypted:

- Only you and the provider can read details
- Stored encrypted on-chain
- Decrypted locally in your browser

## HPC Jobs

### What is HPC?

High-Performance Computing (HPC) allows you to:

- Run large-scale computations
- Access distributed supercomputer resources
- Pay only for compute time used
- Process ML training, simulations, analytics

### Submitting a Job

1. **Open HPC Portal**: Click "HPC" in navigation

2. **Choose Template or Custom**:
   - **Templates**: Pre-configured job types
     - ML Training (TensorFlow, PyTorch)
     - Scientific Computing (GROMACS, NAMD)
     - Data Processing (Spark, Dask)
   - **Custom**: Upload your own job manifest

3. **Configure Resources**:
   - CPUs: Number of cores
   - Memory: RAM in GB
   - GPUs: GPU units (if needed)
   - Duration: Maximum run time

4. **Upload Inputs** (optional):
   - Data files are encrypted
   - Only allocated provider can decrypt

5. **Submit Job**:
   - Job enters queue
   - Provider schedules execution
   - Monitor progress in real-time

### Job States

| State | Description |
|-------|-------------|
| Queued | Waiting for scheduling |
| Scheduled | Assigned to provider |
| Running | Actively executing |
| Completed | Finished successfully |
| Failed | Execution error |
| Cancelled | Cancelled by user |

### Retrieving Results

1. **Check Job Status**: HPC → My Jobs → [Job ID]

2. **Download Outputs**:
   - Outputs are encrypted to your key
   - Decrypt locally in browser
   - Download individual files or archive

3. **View Logs**:
   - Standard output
   - Error logs
   - Resource usage metrics

### Cost Estimation

Before submitting, the portal shows estimated costs:

```
Estimated Cost Breakdown:
├── CPUs (4 cores × 2 hours)    = 80 VE
├── Memory (16 GB × 2 hours)    = 160 VE
├── GPU (1 unit × 2 hours)      = 1000 VE
├── Platform Fee (2.5%)         = 31 VE
└── Total                       = 1,271 VE
```

## Account Security

### Security Best Practices

1. **Protect Your Mnemonic**:
   - Never share with anyone
   - Store offline in secure location
   - Consider metal backup for disaster recovery

2. **Enable MFA**:
   - Add at least 2 factors
   - Include hardware key if possible
   - Store backup codes securely

3. **Use Strong Passwords**:
   - For keyring encryption
   - Unique to VirtEngine

4. **Verify URLs**:
   - Only use official portal.virtengine.com
   - Check for HTTPS lock icon
   - Beware of phishing

5. **Keep Software Updated**:
   - Update browser regularly
   - Update mobile app when available

### Account Recovery

If you lose access:

1. **With Mnemonic**:
   - Import wallet using 24-word phrase
   - Full account access restored

2. **With MFA Backup Codes**:
   - Use backup code as MFA
   - Regain access to settings
   - Add new MFA factors

3. **Without Mnemonic or Backup**:
   - ⚠️ Account cannot be recovered
   - This is by design for security

### Reporting Issues

If you suspect unauthorized access:

1. Immediately transfer funds to new wallet
2. Revoke all sessions
3. Contact support@virtengine.com
4. File incident report in portal

## FAQ

### General

**Q: Is VirtEngine custodial?**  
A: No. You control your private keys. VirtEngine never has access to your funds.

**Q: What happens if I lose my mnemonic?**  
A: Your account cannot be recovered. This is a security feature, not a bug.

**Q: Are my identity documents stored on-chain?**  
A: No. Only encrypted hashes and your score are stored on-chain.

### Identity

**Q: How is my identity score calculated?**  
A: Multiple validators run ML models on your encrypted data and reach consensus on the score.

**Q: Can I improve my identity score?**  
A: Yes, by verifying additional scopes (document + selfie + email, etc.).

**Q: How long is my score valid?**  
A: Identity scores expire after 1 year and require re-verification.

### Marketplace

**Q: What if a provider doesn't deliver?**  
A: Funds are held in escrow. If SLAs aren't met, you can file a dispute.

**Q: Can I cancel an order early?**  
A: Yes, you can terminate early. You'll be charged for time used plus any penalties specified in the offering.

### HPC

**Q: How do I know my job data is secure?**  
A: Your data is encrypted end-to-end. Only you and the allocated provider can decrypt it, and the provider deletes data after job completion.

**Q: What happens if my job fails?**  
A: You're only charged for resources used up to the failure point. Check logs to diagnose the issue.

---

## Support

- Help Center: [help.virtengine.com](https://help.virtengine.com)
- Discord: [discord.gg/virtengine](https://discord.gg/virtengine)
- Email: support@virtengine.com

---

## Glossary

| Term | Definition |
|------|------------|
| **VE** | VirtEngine token, the native currency |
| **VEID** | VirtEngine Identity - the identity verification system |
| **MFA** | Multi-Factor Authentication |
| **HPC** | High-Performance Computing |
| **Mnemonic** | 24-word recovery phrase for your wallet |
| **Validator** | Network participant that processes transactions |
| **Provider** | Entity offering compute resources |
| **Offering** | A resource listing in the marketplace |
| **Order** | Your purchase of an offering |
| **Escrow** | Funds held securely during order fulfillment |
