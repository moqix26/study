param(
    [string]$Root = (Split-Path -Parent $PSScriptRoot),
    [switch]$Run
)

$ErrorActionPreference = 'Stop'
$utf8NoBom = New-Object System.Text.UTF8Encoding($false)
$tempRoot = [System.IO.Path]::GetFullPath([System.IO.Path]::GetTempPath())
$work = Join-Path $tempRoot ("data-structure-verify-" + [guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $work -Force | Out-Null

$counts = @{ cpp = 0; java = 0; go = 0 }
$passed = 0

function Assert-Command([string]$Name) {
    if (-not (Get-Command $Name -ErrorAction SilentlyContinue)) {
        throw "Required command not found: $Name"
    }
}

function Invoke-Checked(
    [string]$Program,
    [string[]]$Arguments,
    [string]$WorkingDirectory,
    [string]$Label
) {
    Push-Location $WorkingDirectory
    try {
        $previousPreference = $ErrorActionPreference
        $ErrorActionPreference = 'Continue'
        try {
            $output = & $Program @Arguments 2>&1
            $exitCode = $LASTEXITCODE
        }
        finally {
            $ErrorActionPreference = $previousPreference
        }
        if ($exitCode -ne 0) {
            $details = ($output | Out-String).Trim()
            throw "$Label failed with exit code $exitCode`n$details"
        }
    }
    finally {
        Pop-Location
    }
}

try {
    Assert-Command 'g++'
    Assert-Command 'javac'
    Assert-Command 'java'
    Assert-Command 'go'

    $markdownFiles = Get-ChildItem -LiteralPath $Root -Recurse -File -Filter '*.md' |
        Sort-Object FullName

    $pattern = '(?ms)^```(?<lang>cpp|c\+\+|java|go)\s*\r?\n(?<code>.*?)^```\s*$'
    $caseNumber = 0

    foreach ($file in $markdownFiles) {
        $text = [System.IO.File]::ReadAllText($file.FullName, [System.Text.Encoding]::UTF8)
        $matches = [regex]::Matches($text, $pattern)
        foreach ($match in $matches) {
            $caseNumber++
            $lang = $match.Groups['lang'].Value.ToLowerInvariant()
            if ($lang -eq 'c++') { $lang = 'cpp' }
            $counts[$lang]++

            $caseDir = Join-Path $work ("case-{0:D3}-{1}" -f $caseNumber, $lang)
            New-Item -ItemType Directory -Path $caseDir -Force | Out-Null
            $code = $match.Groups['code'].Value.TrimEnd() + [Environment]::NewLine
            $label = "$($file.FullName) example $($counts[$lang]) [$lang]"

            switch ($lang) {
                'cpp' {
                    $source = Join-Path $caseDir 'main.cpp'
                    $binary = Join-Path $caseDir 'main.exe'
                    [System.IO.File]::WriteAllText($source, $code, $utf8NoBom)
                    Invoke-Checked 'g++' @('-std=c++17', '-Wall', '-Wextra', '-pedantic', '-Werror', $source, '-o', $binary) $caseDir $label
                    if ($Run) { Invoke-Checked $binary @() $caseDir "$label run" }
                }
                'java' {
                    $source = Join-Path $caseDir 'Main.java'
                    [System.IO.File]::WriteAllText($source, $code, $utf8NoBom)
                    Invoke-Checked 'javac' @('-encoding', 'UTF-8', $source) $caseDir $label
                    if ($Run) { Invoke-Checked 'java' @('-cp', $caseDir, 'Main') $caseDir "$label run" }
                }
                'go' {
                    $source = Join-Path $caseDir 'main.go'
                    $binary = Join-Path $caseDir 'main.exe'
                    [System.IO.File]::WriteAllText($source, $code, $utf8NoBom)
                    Invoke-Checked 'gofmt' @('-w', $source) $caseDir "$label gofmt"
                    Invoke-Checked 'go' @('build', '-o', $binary, $source) $caseDir $label
                    if ($Run) { Invoke-Checked $binary @() $caseDir "$label run" }
                }
            }
            $passed++
        }
    }

    Write-Host "Verified programs: $passed"
    Write-Host "C++: $($counts.cpp)  Java: $($counts.java)  Go: $($counts.go)"
}
finally {
    if (Test-Path -LiteralPath $work) {
        $resolvedWork = [System.IO.Path]::GetFullPath($work)
        if (-not $resolvedWork.StartsWith($tempRoot, [System.StringComparison]::OrdinalIgnoreCase)) {
            throw "Refusing to clean a non-temp directory: $resolvedWork"
        }
        Remove-Item -LiteralPath $resolvedWork -Recurse -Force
    }
}
