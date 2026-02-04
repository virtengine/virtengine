@echo off
REM Commit script for MARKET-VEID-002 implementation

echo Staging files for commit...

git add x\provider\handler\server.go
git add x\provider\handler\handler.go
git add x\provider\module.go
git add app\modules.go
git add x\veid\keeper\keeper.go
git add x\provider\keeper\domain_verification.go
git add x\provider\keeper\key.go
git add x\provider\keeper\keeper.go
git add sdk\go\node\provider\v1beta4\key.go
git add sdk\go\node\provider\v1beta4\errors.go
git add sdk\go\node\provider\v1beta4\domain_events.go

echo.
echo Files staged. Committing...

git commit -m "feat(provider): enforce VEID score >=70, MFA, and DNS domain verification

- Add VEID score check requiring minimum score of 70 for provider registration
- Add MFA authorization session validation for provider registration
- Implement DNS TXT record-based domain verification system
  - Generate cryptographic verification tokens
  - Verify domains via _virtengine-verification.<domain> TXT records
  - Token expiration after 7 days
- Update provider module to accept VEID and MFA keepers
- Add GetVEIDScore to VEID keeper interface for cross-module access
- Add domain verification methods to provider keeper interface
- Add domain verification error codes and events

Implements MARKET-VEID-002 specification requirements.
Addresses domain verification similar to email service verification (Gmail, Zoho)."

echo.
echo Commit completed successfully!
echo.
pause
