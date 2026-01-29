param(
  [string]$Node = "tcp://localhost:26657",
  [string]$ChainID = "virtengine-1",
  [string]$GRPC = "localhost:9090",
  [string]$WaldurBaseUrl = "",
  [string]$WaldurToken = "",
  [string]$WaldurProjectUUID = "",
  [string]$ProviderKey = "provider",
  [string]$KeyringBackend = "test",
  [string]$KeyringDir = ""
)

if ([string]::IsNullOrWhiteSpace($WaldurBaseUrl)) { throw "WaldurBaseUrl is required" }
if ([string]::IsNullOrWhiteSpace($WaldurToken)) { throw "WaldurToken is required" }
if ([string]::IsNullOrWhiteSpace($WaldurProjectUUID)) { throw "WaldurProjectUUID is required" }

$repo = (Resolve-Path "$PSScriptRoot\..\").Path
$binary = Join-Path $repo "provider-daemon.exe"
if (-not (Test-Path $binary)) { $binary = "provider-daemon" }

$offeringMap = Join-Path $repo "config\waldur-offering-map.json"

& $binary start `
  --chain-id $ChainID `
  --node $Node `
  --provider-key $ProviderKey `
  --waldur-enabled `
  --waldur-base-url $WaldurBaseUrl `
  --waldur-token $WaldurToken `
  --waldur-project-uuid $WaldurProjectUUID `
  --waldur-offering-map $offeringMap `
  --waldur-chain-submit `
  --waldur-chain-key $ProviderKey `
  --waldur-chain-grpc $GRPC `
  --waldur-chain-keyring-backend $KeyringBackend `
  --waldur-chain-keyring-dir $KeyringDir
