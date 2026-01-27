# Fix all error codes to use 100+ range
$files = Get-ChildItem -Recurse -Path x/, pkg/, sdk/ -Filter 'errors.go'
$count = 0

foreach ($file in $files) {
    $content = Get-Content $file.FullName -Raw
    $modified = $false
    
    # Find all Register calls with error codes < 100
    $updated = [regex]::Replace($content, 'Register\(([^,]+),\s*(\d{1,2})\s*,', {
            param($match)
            $code = [int]$match.Groups[2].Value
            if ($code -lt 100) {
                $script:modified = $true
                "Register($($match.Groups[1].Value), $($code + 99), "
            }
            else {
                $match.Value
            }
        })
    
    if ($modified) {
        Set-Content $file.FullName $updated -NoNewline -Encoding UTF8
        Write-Host "✓ Updated $($file.Name)" -ForegroundColor Green
        $count++
    }
}

Write-Host "`n✅ Updated $count files. All error codes now start at 100+" -ForegroundColor Cyan
