# Quick protobuf generation script for PowerShell
# Bypasses buf's slow vendor scanning on Windows

$ErrorActionPreference = "Stop"
$sdkRoot = "C:\Users\jON\Documents\source\repos\virtengine-gh\virtengine\sdk"
Set-Location $sdkRoot

# Ensure tools are available
$protoc = "$sdkRoot\.cache\bin\protoc.exe"
$protocGenGo = "$sdkRoot\.cache\bin\protoc-gen-gocosmos.exe"
$protocGenGrpcGateway = "$sdkRoot\.cache\bin\protoc-gen-grpc-gateway.exe"

if (-not (Test-Path $protoc)) {
    Write-Error "protoc not found. Run 'make cache' first."
}

# Set up proto include paths
$includes = @(
    "-I$sdkRoot\go\vendor"
    "-I$sdkRoot\go\vendor\github.com\cosmos\gogoproto"
    "-I$sdkRoot\go\vendor\github.com\cosmos\cosmos-sdk\proto"
    "-I$sdkRoot\go\vendor\github.com\cosmos\cosmos-proto\proto"
    "-I$sdkRoot\go\vendor\github.com\cosmos\ibc-go\v10\proto"
    "-I$sdkRoot\proto\node"
    "-I$sdkRoot\proto\provider"
)

# Find all proto files
$protoFiles = Get-ChildItem -Path "$sdkRoot\proto" -Recurse -Filter "*.proto" |
Where-Object { $_.FullName -notmatch 'v1beta3|v1beta5|v2beta1|v1beta4|market' } |
Select-Object -ExpandProperty FullName

Write-Host "Found $($protoFiles.Count) proto files to process"

# Generate for each proto file
$goOut = "$sdkRoot\go"
foreach ($protoFile in $protoFiles) {
    $relPath = $protoFile.Replace("$sdkRoot\", "")
    Write-Host "Generating: $relPath"
    
    & $protoc $includes `
        --gocosmos_out=plugins=grpc:$goOut `
        --grpc-gateway_out=logtostderr=true, allow_colon_final_segments=true:$goOut `
        $protoFile
    
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Failed to generate $relPath"
    }
}

Write-Host "Proto generation complete!"
