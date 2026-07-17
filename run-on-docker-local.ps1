<#
.SYNOPSIS
    Deploy dogan-api on the local Docker daemon.

.DESCRIPTION
    Builds/runs the Compose stack (postgres, LiveKit, dogan-api) locally.
    Forwards to run-on-docker.ps1 after resolving skill defaults.
#>
Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

function Show-Help {
    Write-Host @'
run-on-docker-local.ps1 - deploy dogan-api on local Docker

USAGE:
  .\run-on-docker-local.ps1 [flags]

FLAGS:
  --ssh-string=<alias>       SSH alias; null -> local daemon (default: null)
  --delete-image=<no|yes>    Remove built images during teardown (default: null -> no)
  --delete-volume=<no|yes>   Remove volumes before recreate (default: null -> no)
  --internal-port=<port>     Host port mapped to the API container (default: null -> 8080)
  --volume-dir=<path>        Bind-mount / deploy data directory (default: null -> <USER-PROFILE-NAME>/docker/dogan-api)
  --volume-name=<name>       Named Docker volume label (default: null -> dogan-api-vol)
  --network-name=<name>      Docker network (default: null -> t3-net from manifest)
  --help                     Show this help

EXAMPLES:
  .\run-on-docker-local.ps1
  .\run-on-docker-local.ps1 --delete-volume=yes
  .\run-on-docker-local.ps1 --internal-port=8080
  .\run-on-docker-local.ps1 --network-name=t3-net --delete-image=no

NOTES:
  - Use SSH config alias only; do not include "ssh" in --ssh-string.
  - For local deploy, omit --ssh-string (or leave null). Non-null values are ignored with a warning.
  - Null defaults resolve as described in FLAGS.
  - Truthy values for yes/no flags: yes, true, 1, y, on.
  - Default API host port is 8080 if not specified.
  - Stack: dogan-api + postgres + LiveKit (see docker-compose.yml).
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

    $allowed = @(
        'ssh_string', 'delete_image', 'delete_volume', 'internal_port',
        'volume_dir', 'volume_name', 'network_name', 'help'
    )
    $parsed = @{}
    foreach ($key in $allowed) { $parsed[$key] = $null }
    $parsed['help'] = $false

    foreach ($argument in @($RawArguments)) {
        if ($argument -match '^(--help|-h|/\?)$') {
            $parsed['help'] = $true
            continue
        }
        if ($argument -match '^--(?<name>[\w-]+)(?:=(?<value>.*))?$') {
            $key = ($Matches['name'] -replace '-', '_').ToLowerInvariant()
            if ($key -eq 'help') {
                $parsed['help'] = $true
                continue
            }
            if ($key -notin $allowed) {
                throw "Unknown argument: --$($Matches['name']). Run with --help."
            }
            if ($null -eq $Matches['value'] -or $Matches['value'] -eq '') {
                throw "Missing value for --$($Matches['name']). Use --flag=value. Run with --help."
            }
            $parsed[$key] = Remove-SurroundingQuotes -Value $Matches['value']
        }
        else {
            throw "Unknown argument: $argument. Run with --help."
        }
    }

    return $parsed
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

    if (-not [string]::IsNullOrWhiteSpace($flags['ssh_string'])) {
        Write-Host 'Warning: --ssh-string is ignored by run-on-docker-local.ps1. Use run-on-docker-server.ps1 for remote deploy.' -ForegroundColor DarkYellow
    }

    $forward = [System.Collections.Generic.List[string]]::new()
    $forward.Add('--mode=local')
    if ($flags['delete_image']) { $forward.Add("--delete-image=$($flags['delete_image'])") }
    if ($flags['delete_volume']) { $forward.Add("--delete-volume=$($flags['delete_volume'])") }
    if ($flags['internal_port']) { $forward.Add("--internal-port=$($flags['internal_port'])") }
    if ($flags['volume_dir']) { $forward.Add("--volume-dir=$($flags['volume_dir'])") }
    if ($flags['volume_name']) { $forward.Add("--volume-name=$($flags['volume_name'])") }
    if ($flags['network_name']) { $forward.Add("--network-name=$($flags['network_name'])") }

    $engine = Join-Path $PSScriptRoot 'run-on-docker.ps1'
    if (-not (Test-Path $engine)) {
        throw 'Missing run-on-docker.ps1 (deploy engine) in the project root.'
    }

    Write-Host 'Starting local Docker deploy...' -ForegroundColor Cyan
    & $engine -RemainingArguments ($forward.ToArray())
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
}
catch {
    Write-Host ''
    Write-Host $_.Exception.Message -ForegroundColor Red
    Write-Host ''
    Show-Help
    exit 1
}
