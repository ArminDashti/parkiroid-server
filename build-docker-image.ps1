<#
.SYNOPSIS
    Build the dogan-server Docker image.

.DESCRIPTION
    Builds the dogan-server:latest image from docker-compose.yml in the project root.
    Optionally exports the image to .deploy/dogan-server.tar for remote deployment.

.EXAMPLE
    .\build-docker-image.ps1

.EXAMPLE
    .\build-docker-image.ps1 --export=yes

.EXAMPLE
    .\build-docker-image.ps1 --no-cache=yes
#>
Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$Script:ProjectRoot = $PSScriptRoot
$Script:ComposeFile = 'docker-compose.yml'
$Script:ImageName = 'dogan-server:latest'
$Script:ImageArchiveName = 'dogan-server.tar'
$Script:LocalDeployDir = Join-Path $Script:ProjectRoot '.deploy'

function Show-BuildDockerImageHelp {
    Write-Host @'
dogan-server Docker build - build the API container image

Usage:
  .\build-docker-image.ps1 [--export=<no|yes>] [--no-cache=<no|yes>]

Arguments:
  --export=<no|yes>     Export image to .deploy/dogan-server.tar after build (default: no)
  --no-cache=<no|yes>   Build without using Docker cache (default: no)

Examples:
  .\build-docker-image.ps1
  .\build-docker-image.ps1 --export=yes
  .\build-docker-image.ps1 --no-cache=yes
  .\build-docker-image.ps1 --export=yes --no-cache=yes

Image:  dogan-server:latest
Output: .deploy/dogan-server.tar (when --export=yes)
'@ -ForegroundColor Cyan
}

function ConvertTo-BuildArguments {
    param([string[]]$RawArguments)

    $parsed = @{
        export   = 'no'
        no_cache = 'no'
        help     = $false
    }

    foreach ($argument in $RawArguments) {
        if ($argument -match '^--(?<name>[\w-]+)(?:=(?<value>.*))?$') {
            $key = ($Matches['name'] -replace '-', '_').ToLowerInvariant()
            $value = if ($Matches.ContainsKey('value')) { $Matches['value'] } else { 'true' }

            switch ($key) {
                'help' { $parsed['help'] = $true }
                'export' { $parsed['export'] = $value.Trim().ToLowerInvariant() }
                'no_cache' { $parsed['no_cache'] = $value.Trim().ToLowerInvariant() }
                default { throw "Unknown argument: --$($Matches['name']). Run with --help." }
            }
        }
        elseif ($argument -match '^-(?<flag>[h?])$') {
            $parsed['help'] = $true
        }
        else {
            throw "Unknown argument: $argument. Run with --help."
        }
    }

    foreach ($flag in @('export', 'no_cache')) {
        if ($parsed[$flag] -notin @('no', 'yes', 'false', 'true', '0', '1')) {
            throw "Invalid --$($flag -replace '_', '-') value '$($parsed[$flag])'. Allowed: no, yes."
        }
    }

    return $parsed
}

function Test-FlagEnabled {
    param([string]$Value)

    return $Value -in @('yes', 'true', '1')
}

function Write-BuildStep {
    param(
        [int]$Step,
        [int]$Total,
        [string]$Message
    )

    $percent = [math]::Round(($Step / $Total) * 100)
    Write-Progress -Activity 'dogan-server Docker build' -Status $Message -PercentComplete $percent
    Write-Host ("[{0}/{1}] {2}" -f $Step, $Total, $Message) -ForegroundColor Yellow
}

function Test-DockerCliAvailable {
    docker version | Out-Null

    if ($LASTEXITCODE -ne 0) {
        throw 'Docker CLI is not available or not running.'
    }
}

function Invoke-DockerImageBuild {
    param([bool]$NoCache)

    Push-Location $Script:ProjectRoot
    try {
        $buildCommand = "docker compose -f $Script:ComposeFile build"
        if ($NoCache) {
            $buildCommand += ' --no-cache'
        }

        Invoke-Expression $buildCommand
        if ($LASTEXITCODE -ne 0) {
            throw 'Docker image build failed.'
        }
    }
    finally {
        Pop-Location
    }
}

function Export-DockerImage {
    param([string]$ArchivePath)

    $parentDirectory = Split-Path -Parent $ArchivePath
    if (-not (Test-Path -LiteralPath $parentDirectory)) {
        New-Item -ItemType Directory -Path $parentDirectory -Force | Out-Null
    }

    if (Test-Path -LiteralPath $ArchivePath) {
        Remove-Item -LiteralPath $ArchivePath -Force
    }

    docker save $Script:ImageName -o $ArchivePath
    if ($LASTEXITCODE -ne 0) {
        throw "Failed to export image '$($Script:ImageName)'."
    }
}

function Invoke-BuildDockerImage {
    param(
        [bool]$Export,
        [bool]$NoCache
    )

    $totalSteps = if ($Export) { 3 } else { 2 }

    Write-BuildStep -Step 1 -Total $totalSteps -Message 'Checking Docker'
    Test-DockerCliAvailable

    $buildMessage = if ($NoCache) {
        "Building image ($($Script:ImageName)) without cache"
    }
    else {
        "Building image ($($Script:ImageName))"
    }
    Write-BuildStep -Step 2 -Total $totalSteps -Message $buildMessage
    Invoke-DockerImageBuild -NoCache:$NoCache

    if ($Export) {
        $archivePath = Join-Path $Script:LocalDeployDir $Script:ImageArchiveName
        Write-BuildStep -Step 3 -Total $totalSteps -Message "Exporting image to $archivePath"
        Export-DockerImage -ArchivePath $archivePath
    }

    Write-Progress -Activity 'dogan-server Docker build' -Completed -Status 'Done'
    Write-Host ''
    Write-Host "Build complete. Image: $($Script:ImageName)" -ForegroundColor Green

    if ($Export) {
        Write-Host "Archive: $(Join-Path $Script:LocalDeployDir $Script:ImageArchiveName)" -ForegroundColor Green
    }
}

$argumentMap = ConvertTo-BuildArguments -RawArguments $args
if ($argumentMap['help']) {
    Show-BuildDockerImageHelp
    exit 0
}

$export = Test-FlagEnabled -Value $argumentMap['export']
$noCache = Test-FlagEnabled -Value $argumentMap['no_cache']

try {
    Invoke-BuildDockerImage -Export:$export -NoCache:$noCache
}
catch {
    Write-Progress -Activity 'dogan-server Docker build' -Completed -Status 'Failed'
    Write-Host ''
    Write-Host $_.Exception.Message -ForegroundColor Red
    Write-Host ''
    Show-BuildDockerImageHelp
    exit 1
}
