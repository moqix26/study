param(
    [switch]$Run
)

$ErrorActionPreference = 'Stop'
$root = Split-Path -Parent $PSScriptRoot
$strictUtf8 = [System.Text.UTF8Encoding]::new($false, $true)
$errors = [System.Collections.Generic.List[string]]::new()
$markdownFiles = Get-ChildItem -LiteralPath $root -Recurse -File -Filter '*.md'
$localLinkCount = 0
$goFenceCount = 0
$tickFence = ([string][char]96) * 3

foreach ($file in $markdownFiles) {
    try {
        $text = $strictUtf8.GetString([System.IO.File]::ReadAllBytes($file.FullName))
    }
    catch {
        $errors.Add("UTF-8 decode failed: $($file.FullName)")
        continue
    }

    if ($text.Contains([char]0xFFFD)) {
        $errors.Add("Replacement character found: $($file.FullName)")
    }

    foreach ($forbidden in @('EXPANSION-STANDARD', 'notehub')) {
        if ($text.Contains($forbidden)) {
            $errors.Add("Forbidden legacy marker '$forbidden': $($file.FullName)")
        }
    }

    $lines = $text -split "\r?\n"
    $inFence = $false
    $fenceMarker = ''
    $fenceLanguage = ''
    $goFenceLines = [System.Collections.Generic.List[string]]::new()
    $h1Count = 0

    for ($i = 0; $i -lt $lines.Length; $i++) {
        $line = $lines[$i]
        $marker = ''
        if ($line.StartsWith($tickFence)) {
            $marker = $tickFence
        }
        elseif ($line.StartsWith('~~~')) {
            $marker = '~~~'
        }

        if ($marker.Length -gt 0) {
            if (-not $inFence) {
                $inFence = $true
                $fenceMarker = $marker
                $fenceLanguage = $line.Substring($marker.Length).Trim().ToLowerInvariant()
                $goFenceLines.Clear()
            }
            elseif ($marker -eq $fenceMarker) {
                if ($fenceLanguage -eq 'go') {
                    $goFenceCount++
                    $code = $goFenceLines -join [Environment]::NewLine
                    $source = if ($code -match '^\s*package\s+\w+') {
                        $code
                    }
                    else {
                        'package p' + [Environment]::NewLine + [Environment]::NewLine + $code
                    }
                    $gofmtOutput = $source | & gofmt 2>&1
                    if ($LASTEXITCODE -ne 0) {
                        $errors.Add("Invalid Go fence in $($file.FullName): $($gofmtOutput -join ' ')")
                    }
                }
                $inFence = $false
                $fenceMarker = ''
                $fenceLanguage = ''
            }
            continue
        }

        if ($inFence -and $fenceLanguage -eq 'go') {
            $goFenceLines.Add($line)
        }

        if (-not $inFence -and $line -match '^# ') {
            $h1Count++
        }
    }

    if ($inFence) {
        $errors.Add("Unclosed code fence: $($file.FullName)")
    }
    if ($h1Count -ne 1) {
        $errors.Add("Expected exactly one H1, got ${h1Count}: $($file.FullName)")
    }

    foreach ($match in [regex]::Matches($text, '\[[^\]]+\]\(([^)]+)\)')) {
        $target = $match.Groups[1].Value.Trim()
        if ($target -match '^(https?://|mailto:|#)') {
            continue
        }

        $pathPart = ($target -split '#', 2)[0]
        if ([string]::IsNullOrWhiteSpace($pathPart)) {
            continue
        }

        $localLinkCount++
        try {
            $decoded = [System.Uri]::UnescapeDataString($pathPart)
            $resolved = [System.IO.Path]::GetFullPath(
                (Join-Path $file.DirectoryName $decoded)
            )
            if (-not (Test-Path -LiteralPath $resolved)) {
                $errors.Add("Broken local link in $($file.FullName): $target")
            }
        }
        catch {
            $errors.Add("Invalid local link in $($file.FullName): $target")
        }
    }
}

$goMod = Get-ChildItem -LiteralPath $root -Recurse -File -Filter 'go.mod' |
    Select-Object -First 1
$goRoot = if ($null -eq $goMod) { $null } else { Split-Path -Parent $goMod.FullName }
if ($null -eq $goRoot -or -not (Test-Path -LiteralPath $goRoot -PathType Container)) {
    $errors.Add("Missing Go lab directory under: $root")
}
else {
    Push-Location $goRoot
    try {
        & go test ./...
        if ($LASTEXITCODE -ne 0) {
            $errors.Add('go test ./... failed')
        }

        if ($Run) {
            $labDirs = Get-ChildItem -LiteralPath $goRoot -Directory |
                Where-Object { $_.Name -match '^\d{2}-' } |
                Sort-Object Name

            foreach ($lab in $labDirs) {
                Write-Host "[RUN] $($lab.Name)"
                & go run $lab.FullName
                if ($LASTEXITCODE -ne 0) {
                    $errors.Add("go run failed: $($lab.FullName)")
                }
            }
        }
    }
    finally {
        Pop-Location
    }
}

Write-Host "Markdown files: $($markdownFiles.Count)"
Write-Host "Local links checked: $localLinkCount"
Write-Host "Go fences parsed: $goFenceCount"

if ($errors.Count -gt 0) {
    Write-Host ''
    Write-Host 'Validation errors:'
    foreach ($errorMessage in $errors) {
        Write-Host " - $errorMessage"
    }
    throw "System design validation failed with $($errors.Count) error(s)."
}

Write-Host 'System design validation passed.'
