param(
    [switch]$Run
)

$ErrorActionPreference = 'Stop'
$linuxRoot = (Resolve-Path -LiteralPath (Join-Path $PSScriptRoot '..')).Path
$strictUtf8 = New-Object System.Text.UTF8Encoding($false, $true)
$errors = New-Object System.Collections.Generic.List[string]
$warnings = New-Object System.Collections.Generic.List[string]

function Add-Error([string]$Message) {
    $errors.Add($Message)
}

function Add-Warning([string]$Message) {
    $warnings.Add($Message)
}

function Read-Utf8([string]$Path) {
    try {
        return [System.IO.File]::ReadAllText($Path, $strictUtf8)
    }
    catch {
        Add-Error "UTF-8 decode failed: $Path"
        return $null
    }
}

$expectedCoreFiles = @(
    'README.md',
    '00-学习路线与环境选择.md',
    '01-终端文件文本与权限.md',
    '02-进程信号systemd与日志.md',
    '03-网络端口DNS防火墙与curl.md',
    '04-APT与GoMySQLRedis环境.md',
    '05-Shell与可靠发布脚本.md',
    '06-SSH密钥与远程运维.md',
    '07-Docker与Compose.md',
    '08-NginxTLS与反向代理.md',
    '09-Go短链服务完整部署.md',
    '10-故障排查与面试速查.md'
)

foreach ($relative in $expectedCoreFiles) {
    $path = Join-Path $linuxRoot $relative
    if (-not (Test-Path -LiteralPath $path -PathType Leaf)) {
        Add-Error "missing core file: $relative"
    }
}

$legacyFiles = Get-ChildItem -LiteralPath $linuxRoot -File -Filter '*.md' |
    Where-Object { $_.Name -match '^(01-Linux入门|02-文件系统|03-文件与目录操作命令|04-文本查看|05-用户组|06-进程与服务管理|07-网络命令|08-软件包管理|09-Shell脚本入门|10-SSH远程登录与文件传输|11-日志分析|12-Docker容器基础|13-Nginx与Web服务部署|14-全栈项目|15-面试专题)' }
foreach ($legacy in $legacyFiles) {
    Add-Error "legacy flat document still exists: $($legacy.Name)"
}

$markdownFiles = Get-ChildItem -LiteralPath $linuxRoot -Recurse -File -Filter '*.md' |
    Sort-Object FullName
$relativeLinksChecked = 0
$codeFences = 0

foreach ($file in $markdownFiles) {
    $text = Read-Utf8 $file.FullName
    if ($null -eq $text) {
        continue
    }

    $lines = $text -split "`r?`n"
    $insideFence = $false
    $fenceMarker = $null
    $h1Count = 0
    $fenceCount = 0

    for ($i = 0; $i -lt $lines.Length; $i++) {
        $line = $lines[$i]
        if (-not $insideFence -and $line -match '^(```|~~~)') {
            $insideFence = $true
            $fenceMarker = $Matches[1]
            $fenceCount++
            continue
        }
        if ($insideFence -and $line.StartsWith($fenceMarker)) {
            $insideFence = $false
            $fenceMarker = $null
            $fenceCount++
            continue
        }
        if (-not $insideFence -and $line -match '^# ') {
            $h1Count++
        }
        if ($line -match '[ \t]+$') {
            Add-Warning "trailing whitespace: $($file.FullName):$($i + 1)"
        }
    }

    if ($insideFence -or $fenceCount % 2 -ne 0) {
        Add-Error "unclosed code fence: $($file.FullName)"
    }
    $codeFences += $fenceCount
    if ($h1Count -ne 1) {
        Add-Error "expected exactly one H1, got ${h1Count}: $($file.FullName)"
    }

    foreach ($match in [regex]::Matches($text, '\[[^\]]+\]\((?<target>[^)]+)\)')) {
        $target = $match.Groups['target'].Value.Trim()
        if ($target -match '^(https?://|mailto:|#)') {
            continue
        }

        $pathPart = ($target -split '#', 2)[0].Trim()
        if ([string]::IsNullOrWhiteSpace($pathPart)) {
            continue
        }
        if ($pathPart.StartsWith('<') -and $pathPart.EndsWith('>')) {
            $pathPart = $pathPart.Substring(1, $pathPart.Length - 2)
        }

        $relativeLinksChecked++
        try {
            $decoded = [uri]::UnescapeDataString($pathPart)
            $resolved = [System.IO.Path]::GetFullPath((Join-Path $file.DirectoryName $decoded))
            if (-not (Test-Path -LiteralPath $resolved)) {
                Add-Error "broken local link: $($file.FullName) -> $target"
            }
        }
        catch {
            Add-Error "invalid local link: $($file.FullName) -> $target"
        }
    }
}

$contentScanFiles = $markdownFiles | Where-Object {
    -not $_.FullName.StartsWith((Join-Path $linuxRoot '99-工具与验证'), [System.StringComparison]::OrdinalIgnoreCase)
}
$allMarkdown = ($contentScanFiles | ForEach-Object { Read-Utf8 $_.FullName }) -join "`n"
$forbiddenPatterns = [ordered]@{
    'internal expansion marker' = 'EXPANSION-STANDARD'
    'deleted todo link' = '\.\./\.\./todo\.md|F:\\study\\todo\.md'
    'deleted modification spec link' = '修改规范\.md'
    'legacy notehub project' = '(?i)notehub'
    'Spring Boot mainline' = '(?i)Spring Boot'
    'global MySQL grant' = '(?i)GRANT\s+ALL\s+PRIVILEGES\s+ON\s+\*\.\*'
    'hard-coded weak database password' = '(?i)(MYSQL_ROOT_PASSWORD|MYSQL_PASSWORD)\s*[:=]\s*[" ]?(123456|rootpass|notepass)'
}

foreach ($name in $forbiddenPatterns.Keys) {
    if ([regex]::IsMatch($allMarkdown, $forbiddenPatterns[$name])) {
        Add-Error "forbidden content found: $name"
    }
}

$javaCount = ([regex]::Matches($allMarkdown, '(?<![A-Za-z])Java(?![A-Za-z])', 'IgnoreCase')).Count
$goCount = ([regex]::Matches($allMarkdown, '(?<![A-Za-z])Go(?![A-Za-z])', 'IgnoreCase')).Count
$shortlinkCount = ([regex]::Matches($allMarkdown, '(?i)shortlink|短链')).Count
if ($goCount -lt 30) {
    Add-Warning "Go references are unexpectedly low: $goCount"
}
if ($shortlinkCount -lt 20) {
    Add-Warning "shortlink references are unexpectedly low: $shortlinkCount"
}
if ($javaCount -gt 20) {
    Add-Warning "Java references remain high for a Go-first library: $javaCount"
}

$goRoot = Join-Path $linuxRoot 'Go实验'
if (Test-Path -LiteralPath $goRoot -PathType Container) {
    $goFiles = Get-ChildItem -LiteralPath $goRoot -Recurse -File -Filter '*.go'
    $unformatted = @()
    foreach ($goFile in $goFiles) {
        $result = & gofmt -l $goFile.FullName
        if ($LASTEXITCODE -ne 0) {
            Add-Error "gofmt failed: $($goFile.FullName)"
        }
        elseif ($result) {
            $unformatted += $result
        }
    }
    foreach ($path in $unformatted) {
        Add-Error "Go file is not gofmt formatted: $path"
    }

    if ($Run) {
        Push-Location $goRoot
        try {
            & go test ./...
            if ($LASTEXITCODE -ne 0) {
                Add-Error 'go test ./... failed under Go实验'
            }
        }
        finally {
            Pop-Location
        }
    }
}
else {
    Add-Error 'missing Go实验 directory'
}

$shellFiles = Get-ChildItem -LiteralPath $linuxRoot -Recurse -File -Filter '*.sh'
if ($Run -and $shellFiles.Count -gt 0) {
    $wsl = Get-Command wsl.exe -ErrorAction SilentlyContinue
    $distros = @()
    if ($wsl) {
        $distros = @(& wsl.exe --list --quiet 2>$null | Where-Object { -not [string]::IsNullOrWhiteSpace($_) })
    }
    if ($wsl -and $distros.Count -gt 0) {
        foreach ($shellFile in $shellFiles) {
            $wslPath = (& wsl.exe wslpath -a $shellFile.FullName 2>$null | Select-Object -First 1).Trim()
            if ([string]::IsNullOrWhiteSpace($wslPath)) {
                Add-Warning "could not translate path for bash -n: $($shellFile.FullName)"
                continue
            }
            & wsl.exe bash -n -- $wslPath
            if ($LASTEXITCODE -ne 0) {
                Add-Error "bash -n failed: $($shellFile.FullName)"
            }
        }
    }
    else {
        Add-Warning 'WSL distro unavailable; skipped bash -n validation'
    }
}

$composePath = Join-Path $goRoot 'deploy\compose.yaml'
if ($Run -and (Test-Path -LiteralPath $composePath)) {
    $docker = Get-Command docker -ErrorAction SilentlyContinue
    if ($docker) {
        $envExample = Join-Path $goRoot 'deploy\.env.example'
        & docker compose --env-file $envExample -f $composePath config --quiet
        if ($LASTEXITCODE -ne 0) {
            Add-Error 'docker compose config failed'
        }
    }
    else {
        Add-Warning 'Docker CLI unavailable; skipped docker compose config'
    }
}

Write-Host "Linux root: $linuxRoot"
Write-Host "Markdown files: $($markdownFiles.Count)"
Write-Host "Code fence markers: $codeFences"
Write-Host "Relative links checked: $relativeLinksChecked"
Write-Host "Go / Java / shortlink references: $goCount / $javaCount / $shortlinkCount"
Write-Host "Go files: $((Get-ChildItem -LiteralPath $goRoot -Recurse -File -Filter '*.go').Count)"
Write-Host "Shell files: $($shellFiles.Count)"
Write-Host "Warnings: $($warnings.Count)"
foreach ($warning in $warnings) {
    Write-Warning $warning
}
Write-Host "Errors: $($errors.Count)"
foreach ($message in $errors) {
    Write-Error $message -ErrorAction Continue
}

if ($errors.Count -gt 0) {
    exit 1
}

Write-Host 'Linux learning library verification passed.' -ForegroundColor Green
