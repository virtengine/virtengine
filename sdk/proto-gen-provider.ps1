#!/usr/bin/env pwsh
# Generate provider protos using protoc directly to handle k8s.io vendor imports

$ErrorActionPreference = "Stop"

Write-Host "Generating provider Go protos with protoc..." -ForegroundColor Cyan

# Ensure protoc plugins are in PATH
$env:PATH = "$env:GOPATH\bin;$PWD\.cache\bin;$env:PATH"

# Output directory
$outDir = "go/inventory"
Remove-Item -Recurse -Force $outDir -ErrorAction SilentlyContinue
New-Item -ItemType Directory -Path $outDir -Force | Out-Null

# Find all provider proto files
$protoFiles = Get-ChildItem -Path provider -Filter *.proto -Recurse | Select-Object -ExpandProperty FullName

# Include paths for protoc
$includes = @(
    "--proto_path=provider",                    # Provider protos
    "--proto_path=proto/node",                  # Node protos (for virtengine.base.attributes)
    "--proto_path=go/vendor",                   # k8s.io vendor (resolves k8s.io/apimachinery/...)
    "--proto_path=$PWD\.cache\include"          # Well-known types (google/protobuf/...)
)

# Check if cache include exists, if not add buf registry cache
if (-not (Test-Path "$PWD\.cache\include")) {
    $bufCache = "$env:LOCALAPPDATA\buf"
    if (Test-Path $bufCache) {
        $includes += "--proto_path=$bufCache"
    }
}

# Generator options
$goOpts = @(
    "--gocosmos_out=plugins=grpc:.",
    "--gocosmos_opt=Mgoogle/protobuf/any.proto=github.com/cosmos/cosmos-sdk/codec/types",
    "--gocosmos_opt=Mgoogle/protobuf/duration.proto=github.com/cosmos/gogoproto/types",
    "--gocosmos_opt=Mgoogle/protobuf/timestamp.proto=github.com/cosmos/gogoproto/types",
    "--gocosmos_opt=Mgogoproto/gogo.proto=github.com/cosmos/gogoproto/gogoproto",
    "--gocosmos_opt=Mcosmos/base/v1beta1/coin.proto=github.com/cosmos/cosmos-sdk/types",
    "--gocosmos_opt=Mcosmos/base/query/v1beta1/pagination.proto=github.com/cosmos/cosmos-sdk/types/query"
)

# Convert Windows paths to relative for protoc
$protoFilesRelative = $protoFiles | ForEach-Object { 
    $rel = Resolve-Path -Relative -Path $_
    $rel.Replace('\', '/').TrimStart('./')
}

Write-Host "Found $($protoFiles.Count) provider proto files"

# Run protoc
$protocCmd = @(".cache\bin\protoc") + $includes + $goOpts + $protoFilesRelative

Write-Host "Running: protoc $($includes -join ' ') $($goOpts -join ' ') [provider protos]"

try {
    & $protocCmd[0] $protocCmd[1..($protocCmd.Length - 1)]
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ Provider Go protos generated successfully" -ForegroundColor Green
        
        # Move generated files to correct location
        if (Test-Path "provider/virtengine") {
            Write-Host "Moving generated files to go/inventory/"
            Move-Item -Path "provider/virtengine/*" -Destination "go/inventory/" -Force
            Remove-Item -Recurse -Force "provider/virtengine" -ErrorAction SilentlyContinue
        }
        
        # Count generated files
        $generatedFiles = Get-ChildItem -Path $outDir -Filter *.pb.go -Recurse
        Write-Host "✓ Generated $($generatedFiles.Count) Go files in $outDir" -ForegroundColor Green
    }
    else {
        Write-Host "✗ protoc failed with exit code $LASTEXITCODE" -ForegroundColor Red
        exit 1
    }
}
catch {
    Write-Host "✗ Error running protoc: $_" -ForegroundColor Red
    exit 1
}
