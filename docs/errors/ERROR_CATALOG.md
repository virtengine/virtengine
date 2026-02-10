# VirtEngine Error Catalog

Complete catalog of all error codes in VirtEngine.

## Format

Each error entry includes:
- **Code**: Module and numeric code
- **Category**: Error category (validation, not_found, etc.)
- **Message**: Default error message
- **Retryable**: Whether the error can be retried
- **Action**: Recommended client action

## Blockchain Modules (x/)

### veid (1000-1099) - Identity Verification

| Code | Category | Message | Retryable | Action |
|------|----------|---------|-----------|--------|
| 1000 | validation | invalid address | No | Fix address format |
| 1001 | validation | invalid scope | No | Fix scope format |
| 1002 | validation | invalid scope type | No | Use valid scope type |
| 1003 | validation | invalid scope version | No | Update scope version |
| 1004 | validation | invalid payload | No | Fix encrypted payload |
| 1005 | validation | invalid salt | No | Generate new salt |
| 1006 | conflict | salt already used | No | Generate new salt |
| 1007 | validation | invalid device info | No | Fix device information |
| 1008 | validation | invalid client ID | No | Use approved client ID |
| 1009 | unauthorized | client not approved | No | Register and approve client |
| 1010 | validation | invalid client signature | No | Re-sign with correct key |
| 1011 | validation | invalid user signature | No | Re-sign with correct key |
| 1012 | validation | invalid payload hash | No | Recalculate hash |
| 1013 | validation | invalid verification status | No | Use valid status |
| 1014 | validation | invalid verification event | No | Fix event data |
| 1015 | validation | invalid score | No | Score must be 0-100 |
| 1016 | validation | invalid tier | No | Use valid tier |
| 1017 | validation | invalid identity record | No | Fix record format |
| 1018 | validation | invalid wallet | No | Fix wallet format |
| 1019 | not_found | scope not found | No | Create scope first |
| 1020 | not_found | identity record not found | No | Create identity first |
| 1021 | conflict | scope already exists | No | Use existing scope |
| 1022 | state | scope revoked | No | Cannot use revoked scope |
| 1023 | state | scope expired | No | Renew scope |
| 1024 | unauthorized | unauthorized | No | Check permissions |
| 1025 | state | invalid status transition | No | Check state machine |
| 1026 | state | identity locked | No | Wait for unlock |
| 1027 | validation | max scopes exceeded | No | Delete unused scopes |
| 1028 | state | verification in progress | No | Wait for completion |
| 1029 | unauthorized | validator only | No | Must be validator |
| 1030 | validation | signature mismatch | No | Verify signatures |
| 1031 | validation | invalid parameters | No | Fix parameters |
| 1032 | validation | invalid verification request | No | Fix request format |
| 1033 | validation | invalid verification result | No | Fix result format |
| 1034 | not_found | verification request not found | No | Create request first |
| 1035 | internal | decryption failed | No | Check encryption keys |
| 1036 | internal | ML inference failed | Yes | Retry or report |
| 1037 | timeout | verification timeout | Yes | Retry with longer timeout |
| 1038 | validation | max retries exceeded | No | Contact support |
| 1039 | unauthorized | not block proposer | No | Only proposer can execute |
| 1040 | not_found | validator key not found | No | Register validator key |

### mfa (1200-1299) - Multi-Factor Authentication

| Code | Category | Message | Retryable | Action |
|------|----------|---------|-----------|--------|
| 1200 | validation | invalid address | No | Fix address format |
| 1201 | validation | invalid factor type | No | Use valid factor type |
| 1202 | validation | invalid MFA policy | No | Fix policy format |
| 1203 | not_found | MFA policy not found | No | Create policy first |
| 1204 | validation | invalid enrollment | No | Fix enrollment data |
| 1205 | not_found | factor enrollment not found | No | Enroll factor first |
| 1206 | conflict | factor enrollment already exists | No | Use existing enrollment |
| 1207 | validation | invalid challenge | No | Generate new challenge |
| 1208 | not_found | challenge not found | No | Create challenge first |
| 1209 | state | challenge expired | No | Generate new challenge |
| 1210 | state | challenge already used | No | Generate new challenge |
| 1211 | rate_limit | max attempts exceeded | No | Wait before retrying |
| 1212 | validation | invalid challenge response | No | Fix response data |
| 1213 | unauthorized | verification failed | No | Check credentials |
| 1214 | unauthorized | MFA required | No | Complete MFA verification |
| 1215 | unauthorized | insufficient factors verified | No | Verify more factors |
| 1216 | not_found | authorization session not found | No | Create session first |
| 1217 | state | session expired | No | Create new session |
| 1218 | state | session already used | No | Create new session |
| 1219 | unauthorized | unauthorized | No | Check permissions |
| 1220 | validation | invalid sensitive tx type | No | Use valid tx type |
| 1221 | validation | invalid sensitive tx config | No | Fix config format |
| 1222 | not_found | trusted device not found | No | Register device |
| 1223 | state | trusted device expired | No | Re-register device |
| 1224 | validation | max trusted devices limit reached | No | Remove old devices |
| 1225 | internal | challenge creation failed | Yes | Retry |
| 1226 | state | MFA disabled | No | Enable MFA first |
| 1227 | state | factor revoked | No | Enroll new factor |
| 1228 | state | factor expired | No | Renew factor |
| 1229 | unauthorized | VEID score insufficient | No | Improve identity score |
| 1230 | validation | device mismatch | No | Use correct device |
| 1231 | rate_limit | cooldown active | No | Wait for cooldown period |
| 1232 | validation | invalid MFA proof | No | Fix proof data |
| 1233 | state | no active factors enrolled | No | Enroll MFA factor |
| 1234 | validation | invalid certificate | No | Fix certificate |
| 1235 | state | certificate expired | No | Renew certificate |
| 1236 | state | certificate not yet valid | No | Check certificate dates |
| 1237 | state | certificate revoked | No | Use valid certificate |
| 1238 | validation | certificate chain invalid | No | Fix certificate chain |
| 1239 | external | revocation check failed | Yes | Retry |
| 1240 | unauthorized | smart card auth failed | No | Check smart card |
| 1241 | validation | invalid signature | No | Re-sign |
| 1242 | unauthorized | key usage not allowed | No | Use correct key |
| 1243 | unauthorized | PIN required | No | Provide PIN |
| 1244 | state | smart card expired | No | Renew smart card |

### encryption (1300-1399) - Encryption Services

| Code | Category | Message | Retryable | Action |
|------|----------|---------|-----------|--------|
| 1300 | validation | invalid address | No | Fix address format |
| 1301 | validation | invalid public key | No | Fix key format |
| 1302 | not_found | recipient key not found | No | Register key first |
| 1303 | conflict | key already exists | No | Use existing key |
| 1304 | state | key revoked | No | Use valid key |
| 1305 | validation | invalid envelope | No | Fix envelope format |
| 1306 | validation | unsupported algorithm | No | Use supported algorithm |
| 1307 | validation | unsupported envelope version | No | Use supported version |
| 1308 | validation | invalid signature | No | Re-sign |
| 1309 | internal | encryption failed | Yes | Retry |
| 1310 | internal | decryption failed | No | Check keys |
| 1311 | validation | invalid nonce | No | Generate new nonce |
| 1312 | unauthorized | unauthorized | No | Check permissions |
| 1313 | unauthorized | not a recipient | No | Must be envelope recipient |
| 1314 | validation | invalid key fingerprint | No | Recalculate fingerprint |
| 1315 | validation | max recipients exceeded | No | Reduce recipient count |
| 1316 | internal | crypto agility error | Yes | Retry |
| 1317 | not_found | algorithm not found | No | Register algorithm |
| 1318 | state | algorithm deprecated | No | Migrate to new algorithm |
| 1319 | state | algorithm disabled | No | Use enabled algorithm |
| 1320 | state | key rotation in progress | No | Wait for completion |
| 1321 | not_found | key rotation not found | No | Start key rotation |
| 1322 | internal | migration failed | Yes | Retry or contact support |

### market (1400-1499) - Marketplace

| Code | Category | Message | Retryable | Action |
|------|----------|---------|-----------|--------|
| 1400 | validation | invalid order | No | Fix order parameters |
| 1401 | validation | invalid bid | No | Fix bid parameters |
| 1402 | validation | invalid lease | No | Fix lease parameters |
| 1410 | not_found | order not found | No | Verify order ID |
| 1411 | not_found | bid not found | No | Verify bid ID |
| 1412 | not_found | lease not found | No | Verify lease ID |
| 1420 | conflict | order already exists | No | Use existing order |
| 1421 | conflict | bid already exists | No | Use existing bid |
| 1422 | conflict | lease already exists | No | Use existing lease |
| 1430 | unauthorized | unauthorized | No | Check permissions |
| 1431 | unauthorized | identity verification required | No | Complete VEID verification |
| 1432 | unauthorized | MFA verification required | No | Complete MFA verification |
| 1440 | state | order closed | No | Cannot modify closed order |
| 1441 | state | bid rejected | No | Submit new bid |
| 1442 | state | lease terminated | No | Create new lease |

### Common Error Code Patterns

Within each 100-code range, modules follow these patterns:

- **00-09**: Validation errors
- **10-19**: Not found errors
- **20-29**: Conflict errors
- **30-39**: Authorization errors
- **40-49**: State errors
- **50-59**: External service errors
- **60-69**: Internal errors
- **70-79**: Verification errors
- **80-89**: Rate limit errors
- **90-99**: Reserved

## Off-Chain Services (pkg/)

### provider_daemon (100-199)

| Code | Category | Message | Retryable | Action |
|------|----------|---------|-----------|--------|
| 100 | validation | invalid configuration | No | Fix daemon config |
| 101 | validation | invalid manifest | No | Fix manifest format |
| 110 | not_found | deployment not found | No | Verify deployment ID |
| 111 | not_found | provider not found | No | Register provider |
| 120 | conflict | deployment already exists | No | Use existing deployment |
| 130 | unauthorized | unauthorized | No | Check provider keys |
| 140 | state | deployment terminated | No | Create new deployment |
| 150 | external | external service error | Yes | Retry |
| 151 | timeout | operation timeout | Yes | Retry with longer timeout |
| 160 | internal | internal error | No | Contact support |

### inference (200-299) - ML Inference

| Code | Category | Message | Retryable | Action |
|------|----------|---------|-----------|--------|
| 200 | validation | invalid input | No | Fix input format |
| 210 | not_found | model not found | No | Load model first |
| 211 | not_found | feature not found | No | Extract features first |
| 250 | external | model loading failed | Yes | Retry |
| 251 | timeout | inference timeout | Yes | Retry with longer timeout |
| 260 | internal | inference error | Yes | Retry or contact support |
| 261 | internal | non-deterministic result | No | Check determinism config |

### workflow (300-399) - Workflow Engine

| Code | Category | Message | Retryable | Action |
|------|----------|---------|-----------|--------|
| 300 | validation | invalid workflow | No | Fix workflow definition |
| 310 | not_found | workflow not found | No | Create workflow first |
| 311 | not_found | execution not found | No | Start execution first |
| 320 | conflict | workflow already exists | No | Use existing workflow |
| 340 | state | workflow completed | No | Cannot modify completed workflow |
| 341 | state | workflow failed | No | Retry or fix workflow |
| 350 | external | step execution failed | Yes | Retry |
| 360 | internal | workflow engine error | Yes | Retry or contact support |

## Severity Guidelines

### Info
- Not found errors
- Validation errors

### Warning
- Deprecated features
- State transition warnings

### Error (Default)
- Most application errors
- External service errors
- Timeout errors

### Critical
- Internal system errors
- Data corruption
- Security violations

## Retryability Guidelines

### Retryable
- Timeout errors (CategoryTimeout)
- External service errors (CategoryExternal)
- Rate limit errors (CategoryRateLimit) - after reset time
- Some internal errors (infrastructure issues)

### Not Retryable
- Validation errors (CategoryValidation)
- Not found errors (CategoryNotFound)
- Conflict errors (CategoryConflict)
- Authorization errors (CategoryUnauthorized)
- State errors that require intervention

## Adding New Error Codes

When adding a new error code:

1. Check the module's allocated range in `pkg/errors/codes.go`
2. Choose a code within the module's range
3. Follow the code pattern for the error category (00-09 for validation, etc.)
4. Validate the code: `errors.ValidateCode(module, code)`
5. Register in the module's `types/errors.go`
6. Add to this catalog
7. Document in module README

## References

- Error Code Registry: `pkg/errors/codes.go`
- Error Types: `pkg/errors/types.go`
- Error Handling Guide: `_docs/ERROR_HANDLING.md`
- Client Error Guide: `docs/api/ERROR_HANDLING.md`
