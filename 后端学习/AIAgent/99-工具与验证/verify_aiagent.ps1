param(
    [switch]$Run
)

$ErrorActionPreference = 'Stop'
$root = Split-Path -Parent $PSScriptRoot
$workspace = [System.IO.Path]::GetFullPath((Join-Path $root '..\..'))
$strictUtf8 = [System.Text.UTF8Encoding]::new($false, $true)
$errors = [System.Collections.Generic.List[string]]::new()
$tickFence = ([string][char]96) * 3
$markdownFiles = Get-ChildItem -LiteralPath $root -Recurse -File -Filter '*.md'
$localLinks = 0
$workspaceAIAgentLinks = 0
$goFences = 0
$goFenceSnippetsSkipped = 0

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
    foreach ($forbidden in @('EXPANSION-STANDARD', 'verified-facts')) {
        if ($text.Contains($forbidden)) {
            $errors.Add("Forbidden legacy marker '$forbidden': $($file.FullName)")
        }
    }

    $lines = $text -split "\r?\n"
    $inFence = $false
    $fenceMarker = ''
    $fenceLanguage = ''
    $goLines = [System.Collections.Generic.List[string]]::new()
    $h1Count = 0

    foreach ($line in $lines) {
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
                $goLines.Clear()
            }
            elseif ($marker -eq $fenceMarker) {
                if ($fenceLanguage -eq 'go') {
                    $goFences++
                    $code = $goLines -join [Environment]::NewLine
                    if ($code -match '\.\.\.' -or $code -match '省略') {
                        $goFenceSnippetsSkipped++
                    }
                    elseif ($code -match '^\s*package\s+\w+' -or $code -match '^\s*(func|type|var|const|import)\b') {
                        $source = if ($code -match '^\s*package\s+\w+') {
                            $code
                        }
                        else {
                            'package p' + [Environment]::NewLine + [Environment]::NewLine + $code
                        }
                        $previousPreference = $ErrorActionPreference
                        $ErrorActionPreference = 'Continue'
                        $gofmtOutput = $source | & gofmt 2>&1
                        $gofmtExitCode = $LASTEXITCODE
                        $ErrorActionPreference = $previousPreference
                        if ($gofmtExitCode -ne 0) {
                            $errors.Add("Invalid Go fence in $($file.FullName): $($gofmtOutput -join ' ')")
                        }
                    }
                    else {
                        $goFenceSnippetsSkipped++
                    }
                }
                $inFence = $false
                $fenceMarker = ''
                $fenceLanguage = ''
            }
            continue
        }

        if ($inFence -and $fenceLanguage -eq 'go') {
            $goLines.Add($line)
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
        $localLinks++
        try {
            $decoded = [System.Uri]::UnescapeDataString($pathPart)
            $resolved = [System.IO.Path]::GetFullPath((Join-Path $file.DirectoryName $decoded))
            if (-not (Test-Path -LiteralPath $resolved)) {
                $errors.Add("Broken local link in $($file.FullName): $target")
            }
        }
        catch {
            $errors.Add("Invalid local link in $($file.FullName): $target")
        }
    }
}

$workspaceMarkdown = Get-ChildItem -LiteralPath $workspace -Recurse -File -Filter '*.md'
foreach ($file in $workspaceMarkdown) {
    $text = $strictUtf8.GetString([System.IO.File]::ReadAllBytes($file.FullName))
    foreach ($match in [regex]::Matches($text, '\[[^\]]+\]\(([^)]*AIAgent/[^)]+)\)')) {
        $target = $match.Groups[1].Value.Trim()
        if ($target -match '^https?://') {
            continue
        }
        $pathPart = ($target -split '#', 2)[0]
        if ([string]::IsNullOrWhiteSpace($pathPart)) {
            continue
        }
        $workspaceAIAgentLinks++
        try {
            $decoded = [System.Uri]::UnescapeDataString($pathPart)
            $resolved = [System.IO.Path]::GetFullPath((Join-Path $file.DirectoryName $decoded))
            if (-not (Test-Path -LiteralPath $resolved)) {
                $errors.Add("Broken workspace AIAgent link in $($file.FullName): $target")
            }
        }
        catch {
            $errors.Add("Invalid workspace AIAgent link in $($file.FullName): $target")
        }
    }
}

$goModule = Get-ChildItem -LiteralPath $root -Recurse -File -Filter 'go.mod' |
    Where-Object { $_.Directory.Name -eq 'agentgo' } |
    Select-Object -First 1
$goRoot = if ($null -eq $goModule) { $null } else { $goModule.Directory.FullName }
if ($null -eq $goRoot) {
    $errors.Add("Missing Go project named agentgo under: $root")
}
else {
    Push-Location $goRoot
    try {
        $goFiles = Get-ChildItem -LiteralPath $goRoot -Recurse -File -Filter '*.go' |
            Select-Object -ExpandProperty FullName
        $unformatted = if ($goFiles.Count -eq 0) { @() } else { & gofmt -l $goFiles }
        if ($LASTEXITCODE -ne 0) {
            $errors.Add('gofmt check failed')
        }
        if ($unformatted.Count -gt 0) {
            $errors.Add("Unformatted Go files: $($unformatted -join ', ')")
        }

        & go test ./...
        if ($LASTEXITCODE -ne 0) {
            $errors.Add('go test ./... failed')
        }

        if ($Run) {
            $exe = Join-Path $env:TEMP 'agentgo-verify.exe'
            & go build -o $exe ./cmd/server
            if ($LASTEXITCODE -ne 0) {
                $errors.Add('go build ./cmd/server failed')
            }
            else {
                $env:SERVER_ADDR = '127.0.0.1:18081'
                $env:AI_PROVIDER = 'mock'
                $env:RAG_STORE = 'memory'
                $process = Start-Process -FilePath $exe -PassThru -WindowStyle Hidden
                try {
                    $ready = $false
                    for ($i = 0; $i -lt 30; $i++) {
                        try {
                            $health = Invoke-RestMethod -Uri 'http://127.0.0.1:18081/health' -TimeoutSec 1
                            if ($health.status -eq 'ok') {
                                $ready = $true
                                break
                            }
                        }
                        catch {
                            Start-Sleep -Milliseconds 200
                        }
                    }
                    if (-not $ready) {
                        $errors.Add('agentgo smoke server did not become ready')
                    }
                    else {
                        $headers = @{ 'X-User-ID' = 'verify-user' }
                        $body = @{ message = 'hello' } | ConvertTo-Json
                        $chat = Invoke-RestMethod -Uri 'http://127.0.0.1:18081/api/chat' -Method POST -Headers $headers -ContentType 'application/json; charset=utf-8' -Body ([Text.Encoding]::UTF8.GetBytes($body))
                        if ([string]::IsNullOrWhiteSpace($chat.answer)) {
                            $errors.Add('agentgo smoke chat returned empty answer')
                        }

                        $agentBody = @{ message = 'calculate 12+8' } | ConvertTo-Json
                        $agentResult = Invoke-RestMethod -Uri 'http://127.0.0.1:18081/api/agent/run' -Method POST -Headers $headers -ContentType 'application/json; charset=utf-8' -Body ([Text.Encoding]::UTF8.GetBytes($agentBody))
                        if ($agentResult.stop_reason -ne 'COMPLETED' -or @($agentResult.steps).Count -lt 1) {
                            $errors.Add('agentgo smoke agent did not complete a Tool call')
                        }

                        $ingestBody = @{
                            id = 'verify-document'
                            title = 'Verification document'
                            content = 'AgentGo offline V1 uses a Mock Provider and therefore needs no API key.'
                            metadata = @{ source = 'verify_aiagent.ps1' }
                        } | ConvertTo-Json -Depth 5
                        $ingest = Invoke-RestMethod -Uri 'http://127.0.0.1:18081/api/rag/ingest' -Method POST -Headers $headers -ContentType 'application/json; charset=utf-8' -Body ([Text.Encoding]::UTF8.GetBytes($ingestBody))
                        if ($ingest.chunk_count -lt 1) {
                            $errors.Add('agentgo smoke RAG ingest created no chunks')
                        }

                        $askBody = @{ question = 'Why does offline V1 need no API key?'; top_k = 3 } | ConvertTo-Json
                        $ragAnswer = Invoke-RestMethod -Uri 'http://127.0.0.1:18081/api/rag/ask' -Method POST -Headers $headers -ContentType 'application/json; charset=utf-8' -Body ([Text.Encoding]::UTF8.GetBytes($askBody))
                        if ([string]::IsNullOrWhiteSpace($ragAnswer.answer) -or @($ragAnswer.citations).Count -lt 1) {
                            $errors.Add('agentgo smoke RAG ask returned no answer or citation')
                        }

                        $streamBody = @{ message = 'stream smoke test' } | ConvertTo-Json -Compress
                        $stream = Invoke-WebRequest -UseBasicParsing -Uri 'http://127.0.0.1:18081/api/chat/stream' -Method POST -Headers $headers -ContentType 'application/json; charset=utf-8' -Body ([Text.Encoding]::UTF8.GetBytes($streamBody)) -TimeoutSec 10
                        if ($stream.Content -notmatch 'event: meta' -or $stream.Content -notmatch 'event: delta' -or $stream.Content -notmatch 'event: done') {
                            $errors.Add('agentgo smoke SSE response missed meta, delta, or done')
                        }
                    }
                }
                finally {
                    if (-not $process.HasExited) {
                        Stop-Process -Id $process.Id -Force
                    }
                    Remove-Item -LiteralPath $exe -Force -ErrorAction SilentlyContinue
                }
            }
        }
    }
    finally {
        Pop-Location
    }
}

Write-Host "Markdown files: $($markdownFiles.Count)"
Write-Host "AIAgent local links checked: $localLinks"
Write-Host "Workspace links into AIAgent checked: $workspaceAIAgentLinks"
Write-Host "Go fences parsed: $goFences"
Write-Host "Go statement/pseudocode fences skipped: $goFenceSnippetsSkipped"

if ($errors.Count -gt 0) {
    Write-Host ''
    Write-Host 'Validation errors:'
    foreach ($message in $errors) {
        Write-Host " - $message"
    }
    throw "AIAgent validation failed with $($errors.Count) error(s)."
}

Write-Host 'AIAgent validation passed.'
