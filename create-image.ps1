<#
.SYNOPSIS
    Build the project Docker image.

.DESCRIPTION
    Builds the API image from the project Dockerfile (or compose service).
    Name and tag resolve from .docker/stack.manifest.json when omitted.
#>
Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$Script:ProjectRoot = $PSScriptRoot
$Script:ComposeFile = 'docker-compose.yml'
$Script:LocalDeployDir = Join-Path $Script:ProjectRoot '.deploy'

function Show-Help {
    Write-Host @'
create-image.ps1 - build dogan-api Docker image

USAGE:
  .\create-image.ps1 [flags]

FLAGS:
  --image-name=<name>     Image repository name (default: null -> dogan-api from manifest/repo)
  --tag=<tag>             Image tag (default: null -> latest)
  --dockerfile=<path>     Dockerfile path (default: null -> Dockerfile)
  --context=<path>        Build context directory (default: null -> .)
  --export=<no|yes>       Export image to .deploy/<image>.tar after build (default: null -> no)
  --no-cache=<no|yes>     Build without Docker cache (default: null -> no)
  --help                  Show this help

EXAMPLES:
  .\create-image.ps1
  .\create-image.ps1 --tag=dev
  .\create-image.ps1 --image-name=dogan-api --tag=latest --export=yes
  .\create-image.ps1 --no-cache=yes --dockerfile=Dockerfile --context=.

NOTES:
  - Null defaults resolve from .docker/stack.manifest.json (apiImageTag) or repo folder name.
  - Truthy values for yes/no flags: yes, true, 1, y, on.
  - Prefer this script for image-only builds; run-on-docker-local.ps1 / run-on-docker-server.ps1 call it when a rebuild is needed.
'@ -ForegroundColor Cyan
}

function Remove-SurroundingQuotes {
    param([string]$Value)
    if ([string]::IsNullOrWhiteSpace($Value)) { return $Value }
    $Value = $Value.Trim()
    if (($Value.StartsWith('"') -and $Value.EndsWith('"')) -or ($Value.StartsWith("'") -and $Value.EndsWith("'"))) {
        return $Value.Substring(1, $Value.Length - 2).Trim()
    }
    return $Value
}

function ConvertTo-FlagMap {
    param([string[]]$RawArguments)

    $parsed = @{
        image_name = $null
        tag        = $null
        dockerfile = $null
        context    = $null
        export     = $null
        no_cache   = $null
        help       = $false
    }

    foreach ($argument in @($RawArguments)) {
        if ($argument -match '^(--help|-h|/\?)$') {
            $parsed['help'] = $true
            continue
        }
        if ($argument -match '^--(?<name>[\w-]+)(?:=(?<value>.*))?$') {
            $key = ($Matches['name'] -replace '-', '_').ToLowerInvariant()
            $value = if ($null -ne $Matches['value'] -and $Matches['value'] -ne '') {
                Remove-SurroundingQuotes -Value $Matches['value']
            }
            else {
                'true'
            }
            if (-not $parsed.ContainsKey($key) -and $key -ne 'help') {
                throw "Unknown argument: --$($Matches['name']). Run with --help."
            }
            if ($key -eq 'help') { $parsed['help'] = $true }
            else { $parsed[$key] = $value }
        }
        else {
            throw "Unknown argument: $argument. Run with --help."
        }
    }

    return $parsed
}

function Test-Truthy {
    param([string]$Value)
    if ([string]::IsNullOrWhiteSpace($Value)) { return $false }
    return $Value.ToLowerInvariant() -in @('yes', 'true', '1', 'y', 'on')
}

function Get-StackManifest {
    $manifestPath = Join-Path $Script:ProjectRoot '.docker/stack.manifest.json'
    if (-not (Test-Path $manifestPath)) { return $null }
    return Get-Content -Path $manifestPath -Raw | ConvertFrom-Json
}

function Resolve-ImageDefaults {
    param(
        [string]$ImageName,
        [string]$Tag
    )

    $resolvedName = $ImageName
    $resolvedTag = $Tag
    $manifest = Get-StackManifest

    if ($manifest -and $manifest.apiImageTag) {
        $parts = ([string]$manifest.apiImageTag).Split(':', 2)
        if ([string]::IsNullOrWhiteSpace($resolvedName)) { $resolvedName = $parts[0] }
        if ([string]::IsNullOrWhiteSpace($resolvedTag) -and $parts.Count -gt 1) { $resolvedTag = $parts[1] }
    }

    if ([string]::IsNullOrWhiteSpace($resolvedName)) {
        $resolvedName = Split-Path -Leaf $Script:ProjectRoot
        if ($resolvedName -eq 'parkiroid-server') { $resolvedName = 'dogan-api' }
    }
    if ([string]::IsNullOrWhiteSpace($resolvedTag)) { $resolvedTag = 'latest' }

    return [pscustomobject]@{
        Name     = $resolvedName
        Tag      = $resolvedTag
        FullTag  = "${resolvedName}:${resolvedTag}"
    }
}

function Write-BuildStep {
    param([int]$Step, [int]$Total, [string]$Message)
    $percent = [math]::Round(($Step / $Total) * 100)
    Write-Progress -Activity 'create-image' -Status $Message -PercentComplete $percent
    Write-Host ("[{0}/{1}] {2}" -f $Step, $Total, $Message) -ForegroundColor Yellow
}

function Test-DockerCliAvailable {
    & docker version | Out-Null
    if ($LASTEXITCODE -ne 0) { throw 'Docker CLI is not available or not running.' }
}

function Invoke-ImageBuild {
    param(
        [string]$FullTag,
        [string]$Dockerfile,
        [string]$Context,
        [bool]$NoCache
    )

    $composePath = Join-Path $Script:ProjectRoot $Script:ComposeFile
    $dockerfilePath = Join-Path $Script:ProjectRoot $Dockerfile

    if (-not (Test-Path $dockerfilePath)) {
        throw "Missing Dockerfile at '$Dockerfile'."
    }

    Push-Location $Script:ProjectRoot
    try {
        if ((Test-Path $composePath) -and $Dockerfile -eq 'Dockerfile' -and $Context -eq '.') {
            $env:API_IMAGE_TAG = $FullTag
            $buildCommand = "docker compose -p dogan -f $Script:ComposeFile build dogan-api"
            if ($NoCache) { $buildCommand += ' --no-cache' }
            Invoke-Expression $buildCommand
            if ($LASTEXITCODE -ne 0) { throw 'Docker compose build failed.' }
            return
        }

        $dockerArgs = @('build', '-t', $FullTag, '-f', $Dockerfile)
        if ($NoCache) { $dockerArgs += '--no-cache' }
        $dockerArgs += $Context
        & docker @dockerArgs
        if ($LASTEXITCODE -ne 0) { throw 'Docker build failed.' }
    }
    finally {
        Pop-Location
    }
}

function Export-DockerImage {
    param(
        [string]$FullTag,
        [string]$ArchivePath
    )

    $parentDirectory = Split-Path -Parent $ArchivePath
    if (-not (Test-Path -LiteralPath $parentDirectory)) {
        New-Item -ItemType Directory -Path $parentDirectory -Force | Out-Null
    }
    if (Test-Path -LiteralPath $ArchivePath) {
        Remove-Item -LiteralPath $ArchivePath -Force
    }

    & docker save $FullTag -o $ArchivePath
    if ($LASTEXITCODE -ne 0) { throw "Failed to export image '$FullTag'." }
}

if ($args -match '^(--help|-h|/\?)$') {
    Show-Help
    exit 0
}

try {
    $flags = ConvertTo-FlagMap -RawArguments $args
    if ($flags['help']) {
        Show-Help
        exit 0
    }

    $image = Resolve-ImageDefaults -ImageName $flags['image_name'] -Tag $flags['tag']
    $dockerfile = if ($flags['dockerfile']) { $flags['dockerfile'] } else { 'Dockerfile' }
    $context = if ($flags['context']) { $flags['context'] } else { '.' }
    $export = Test-Truthy -Value $flags['export']
    $noCache = Test-Truthy -Value $flags['no_cache']
    $totalSteps = if ($export) { 3 } else { 2 }

    Write-BuildStep -Step 1 -Total $totalSteps -Message 'Checking Docker'
    Test-DockerCliAvailable

    $buildMsg = if ($noCache) { "Building $($image.FullTag) without cache" } else { "Building $($image.FullTag)" }
    Write-BuildStep -Step 2 -Total $totalSteps -Message $buildMsg
    Invoke-ImageBuild -FullTag $image.FullTag -Dockerfile $dockerfile -Context $context -NoCache:$noCache

    if ($export) {
        $safeName = ($image.Name -replace '[^a-zA-Z0-9._-]', '-')
        $archivePath = Join-Path $Script:LocalDeployDir "$safeName.tar"
        Write-BuildStep -Step 3 -Total $totalSteps -Message "Exporting to $archivePath"
        Export-DockerImage -FullTag $image.FullTag -ArchivePath $archivePath
    }

    Write-Progress -Activity 'create-image' -Completed -Status 'Done'
    Write-Host ''
    Write-Host "Build complete. Image: $($image.FullTag)" -ForegroundColor Green
    if ($export) {
        Write-Host "Archive: $(Join-Path $Script:LocalDeployDir (($image.Name -replace '[^a-zA-Z0-9._-]', '-') + '.tar'))" -ForegroundColor Green
    }
}
catch {
    Write-Progress -Activity 'create-image' -Completed -Status 'Failed'
    Write-Host ''
    Write-Host $_.Exception.Message -ForegroundColor Red
    Write-Host ''
    Show-Help
    exit 1
}
