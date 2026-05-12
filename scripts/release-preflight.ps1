[CmdletBinding()]
param(
    [string]$Tag = 'v0.4.3',
    [string]$Remote = 'origin',
    [string]$ExpectedPushUrl = 'https://github.com/loopgap/HAL-Proxy-build-by-go.git',
    [string]$ExpectedCommitPrefix = ''
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$RepoRoot = (Resolve-Path -LiteralPath (Join-Path $PSScriptRoot '..')).Path
Set-Location -LiteralPath $RepoRoot

function Invoke-Git {
    param([Parameter(Mandatory = $true)][string[]]$Arguments)
    & git @Arguments
    if ($LASTEXITCODE -ne 0) {
        throw "git $($Arguments -join ' ') failed with exit code $LASTEXITCODE"
    }
}

function Get-GitOutput {
    param([Parameter(Mandatory = $true)][string[]]$Arguments)
    $output = & git @Arguments
    if ($LASTEXITCODE -ne 0) {
        throw "git $($Arguments -join ' ') failed with exit code $LASTEXITCODE"
    }
    return ($output -join "`n").Trim()
}

function Get-GitHubRepo {
    param([Parameter(Mandatory = $true)][string]$RemoteUrl)

    if ($RemoteUrl -match '^https://github\.com/([^/]+)/(.+?)(\.git)?$') {
        return @{ Owner = $Matches[1]; Repo = $Matches[2] }
    }
    if ($RemoteUrl -match '^git@github\.com:([^/]+)/(.+?)(\.git)?$') {
        return @{ Owner = $Matches[1]; Repo = $Matches[2] }
    }
    throw "Unsupported GitHub remote URL: $RemoteUrl"
}

function Assert-Step {
    param(
        [Parameter(Mandatory = $true)][string]$Name,
        [Parameter(Mandatory = $true)][scriptblock]$Action
    )

    Write-Host "==> $Name"
    & $Action
}

Assert-Step 'Current branch is main' {
    $branch = Get-GitOutput @('branch', '--show-current')
    if ($branch -ne 'main') {
        throw "Expected current branch main, got $branch"
    }
}

Assert-Step 'Working tree is clean' {
    $status = Get-GitOutput @('status', '--porcelain')
    if ($status) {
        throw "Working tree must be clean before release push:`n$status"
    }
}

Assert-Step 'No staged changes remain' {
    & git diff --cached --quiet
    if ($LASTEXITCODE -ne 0) {
        throw 'Staged changes remain.'
    }
}

Assert-Step 'Remote push URL is expected HTTPS origin' {
    $pushUrl = Get-GitOutput @('remote', 'get-url', '--push', $Remote)
    if ($pushUrl -ne $ExpectedPushUrl) {
        throw "Expected push URL $ExpectedPushUrl, got $pushUrl"
    }
}

Assert-Step 'Git remote is reachable with local credentials' {
    Invoke-Git @('ls-remote', '--exit-code', $Remote, 'HEAD')
}

Assert-Step "Local tag $Tag is annotated" {
    $type = Get-GitOutput @('cat-file', '-t', $Tag)
    if ($type -ne 'tag') {
        throw "Expected $Tag to be an annotated tag object, got $type"
    }
    $commit = Get-GitOutput @('rev-parse', "$Tag^{}")
    if ($ExpectedCommitPrefix -and -not $commit.StartsWith($ExpectedCommitPrefix, [System.StringComparison]::OrdinalIgnoreCase)) {
        throw "Expected $Tag to resolve to $ExpectedCommitPrefix..., got $commit"
    }
    if (-not $ExpectedCommitPrefix) {
        $head = Get-GitOutput @('rev-parse', 'HEAD')
        if ($commit -ne $head) {
            throw "Expected $Tag to resolve to current HEAD $head, got $commit"
        }
    }
}

Assert-Step "Remote tag $Tag does not exist" {
    $remoteTag = Get-GitOutput @('ls-remote', '--tags', $Remote, "refs/tags/$Tag")
    if ($remoteTag) {
        throw "Remote tag already exists:`n$remoteTag"
    }
}

Assert-Step "GitHub Release $Tag does not exist" {
    $pushUrl = Get-GitOutput @('remote', 'get-url', '--push', $Remote)
    $repo = Get-GitHubRepo $pushUrl
    $uri = "https://api.github.com/repos/$($repo.Owner)/$($repo.Repo)/releases/tags/$Tag"
    $headers = @{
        Accept = 'application/vnd.github+json'
        'User-Agent' = 'BridgeOS-release-preflight'
    }

    try {
        Invoke-WebRequest -Uri $uri -Headers $headers -Method Get -UseBasicParsing | Out-Null
        throw "GitHub Release already exists for $Tag"
    }
    catch {
        $response = $_.Exception.Response
        $statusCode = $null
        if ($response -and $response.StatusCode) {
            $statusCode = [int]$response.StatusCode
        }
        if ($statusCode -ne 404) {
            throw
        }
    }
}

Assert-Step 'Atomic push dry-run succeeds' {
    Invoke-Git @('push', '--atomic', '--dry-run', $Remote, 'HEAD:refs/heads/main', "refs/tags/$Tag")
}

Write-Host ''
Write-Host "Release preflight passed for $Tag."
