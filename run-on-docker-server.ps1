<#
.SYNOPSIS
    Deploy dogan-api on a remote Docker host over SSH.

.DESCRIPTION
    Builds the API image locally, transfers it, and starts Compose on the remote host.
    Requires --ssh-string=<SSH config alias>. Optional --domain / --reverse-proxy for sslh/nginx.
#>
Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

function Show-Help {
    Write-Host @'
run-on-docker-server.ps1 - deploy dogan-api on remote Docker over SSH

USAGE:
  .\run-on-docker-server.ps1 --ssh-string=<alias> [flags]

FLAGS:
  --ssh-string=<alias>       SSH config alias (required; default: null -> error)
  --delete-image=<no|yes>    Remove built images during teardown (default: null -> no)
  --delete-volume=<no|yes>   Remove volumes before recreate (default: null -> no)
  --internal-port=<port>     Container port for domain routing / host publish (default: null -> 8080 from manifest)
  --volume-dir=<path>        Remote deploy directory (default: null -> <USER-PROFILE-NAME>/docker/dogan-api)
  --volume-name=<name>       Named Docker volume label (default: null -> dogan-api-vol)
  --network-name=<name>      Docker network (default: null -> t3-net from manifest)
  --reverse-proxy=<sslh|none> Reverse-proxy mode (default: sslh)
  --domain=<hostname>        Map hostname via remote sslh/nginx to dogan-api
  --public-port=<port>       Public HTTPS port for sslh/nginx (default: 443)
  --help                     Show this help

EXAMPLES:
  .\run-on-docker-server.ps1 --ssh-string=myserver
  .\run-on-docker-server.ps1 --ssh-string=myserver --delete-volume=yes
  .\run-on-docker-server.ps1 --ssh-string=myserver --domain=dogan.xaigrok.ir --internal-port=8080
  .\run-on-docker-server.ps1 --ssh-string=myserver --reverse-proxy=none --internal-port=30042

NOTES:
  --ssh-string is required. Use SSH config alias only; do not include "ssh".
  - Null defaults resolve as described in FLAGS.
  - Truthy values for yes/no flags: yes, true, 1, y, on.
  - Image is built locally, saved, SCP'd, and loaded on the remote host (no remote build).
  - With --reverse-proxy=sslh (default), API host ports are not published; use --domain for HTTPS routing.
  - --domain also sets DOGAN_CORS_ALLOWED_ORIGINS to include https://<domain> (plus local web + https://dogan.xaigrok.ir).
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
        'volume_dir', 'volume_name', 'network_name', 'reverse_proxy',
        'domain', 'public_port', 'help'
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

    if ([string]::IsNullOrWhiteSpace($flags['ssh_string'])) {
        throw 'Missing required --ssh-string=<alias>. Example: --ssh-string=myserver'
    }

    if ($flags['ssh_string'] -match '^(?i)ssh(\s|$)') {
        throw 'Invalid --ssh-string. Pass only the SSH config alias (e.g. myserver). Do not include "ssh".'
    }

    $forward = [System.Collections.Generic.List[string]]::new()
    $forward.Add('--mode=server')
    $forward.Add("--ssh-string=$($flags['ssh_string'])")
    if ($flags['delete_image']) { $forward.Add("--delete-image=$($flags['delete_image'])") }
    if ($flags['delete_volume']) { $forward.Add("--delete-volume=$($flags['delete_volume'])") }
    if ($flags['internal_port']) { $forward.Add("--internal-port=$($flags['internal_port'])") }
    if ($flags['volume_dir']) { $forward.Add("--volume-dir=$($flags['volume_dir'])") }
    if ($flags['volume_name']) { $forward.Add("--volume-name=$($flags['volume_name'])") }
    if ($flags['network_name']) { $forward.Add("--network-name=$($flags['network_name'])") }
    if ($flags['reverse_proxy']) { $forward.Add("--reverse-proxy=$($flags['reverse_proxy'])") }
    if ($flags['domain']) { $forward.Add("--domain=$($flags['domain'])") }
    if ($flags['public_port']) { $forward.Add("--public-port=$($flags['public_port'])") }

    $engine = Join-Path $PSScriptRoot 'run-on-docker.ps1'
    if (-not (Test-Path $engine)) {
        throw 'Missing run-on-docker.ps1 (deploy engine) in the project root.'
    }

    Write-Host ("Starting remote Docker deploy via ssh {0}..." -f $flags['ssh_string']) -ForegroundColor Cyan
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
