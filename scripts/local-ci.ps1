[CmdletBinding()]
param()

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$RepoRoot = (Resolve-Path -LiteralPath (Join-Path $PSScriptRoot '..')).Path
$UiRoot = Join-Path $RepoRoot 'ui'
$CoverageDir = Join-Path $RepoRoot '.tmp\coverage\go'

function Invoke-Step {
    param(
        [Parameter(Mandatory = $true)][string]$Name,
        [Parameter(Mandatory = $true)][scriptblock]$Action
    )

    Write-Host ""
    Write-Host "==> $Name"
    & $Action
}

function Invoke-Native {
    param(
        [Parameter(Mandatory = $true)][string]$FilePath,
        [Parameter(Mandatory = $true)][string[]]$Arguments,
        [string]$WorkingDirectory = $RepoRoot
    )

    Push-Location -LiteralPath $WorkingDirectory
    try {
        & $FilePath @Arguments
        if ($LASTEXITCODE -ne 0) {
            throw "Command failed with exit code ${LASTEXITCODE}: $FilePath $($Arguments -join ' ')"
        }
    }
    finally {
        Pop-Location
    }
}

Invoke-Step 'Go unit tests' {
    Invoke-Native go @('test', './cmd/...', './internal/...')
}

Invoke-Step 'Go race tests' {
    Invoke-Native go @('test', '-race', './cmd/...', './internal/...')
}

Invoke-Step 'Go integration tests' {
    Invoke-Native go @('test', '-tags=integration', './cmd/...', './internal/...')
}

Invoke-Step 'Go coverage' {
    New-Item -ItemType Directory -Force -Path $CoverageDir | Out-Null
    Invoke-Native go @('test', '-coverpkg=./cmd/...,./internal/...', '-coverprofile=.tmp/coverage/go/coverage.out', './cmd/...', './internal/...')
    Invoke-Native go @('tool', 'cover', '-func=.tmp/coverage/go/coverage.out')
}

Invoke-Step 'Go vet' {
    Invoke-Native go @('vet', './cmd/...', './internal/...')
}

Invoke-Step 'Go build' {
    Invoke-Native go @('build', './cmd/...')
}

Invoke-Step 'UI install' {
    Invoke-Native npm @('ci') $UiRoot
}

Invoke-Step 'UI unit tests' {
    Invoke-Native npm @('test', '--', '--run') $UiRoot
}

Invoke-Step 'UI coverage' {
    Invoke-Native npm @('run', 'test:coverage') $UiRoot
}

Invoke-Step 'UI typecheck' {
    Invoke-Native npx @('tsc', '--noEmit') $UiRoot
}

Invoke-Step 'UI lint' {
    Invoke-Native npm @('run', 'lint') $UiRoot
}

Invoke-Step 'UI build' {
    Invoke-Native npm @('run', 'build') $UiRoot
}

Invoke-Step 'UI Playwright' {
    Invoke-Native npx @('playwright', 'test') $UiRoot
}

Invoke-Step 'UI audit' {
    Invoke-Native npm @('audit', '--audit-level=moderate') $UiRoot
}

Write-Host ""
Write-Host "Local CI completed successfully."
