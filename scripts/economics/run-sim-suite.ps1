$ErrorActionPreference = "Stop"

$outputDir = if ($args.Count -gt 0) { $args[0] } else { ".\\sim-output" }
$scenario = if ($env:SCENARIO) { $env:SCENARIO } else { "baseline" }
$runs = if ($env:RUNS) { $env:RUNS } else { "50" }

Write-Host "Running economics simulation suite ($scenario)..."
go run ./cmd/ve-sim suite --scenario $scenario --output-dir $outputDir --runs $runs

Write-Host "Validating economics metrics..."
go run ./cmd/ve-sim check --metrics (Join-Path $outputDir "metrics.json")

Write-Host "Suite completed. Outputs in $outputDir"

