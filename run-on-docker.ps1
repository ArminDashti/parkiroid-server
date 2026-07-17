<#
.SYNOPSIS
    Build and run the dogan-api stack with Docker Compose locally or over SSH.

.DESCRIPTION
    Uses the repo-root Dockerfile, docker-compose.yml, and livekit.yaml.
    Builds the API image, then starts PostgreSQL, LiveKit, and dogan-api.
    When --ssh-string is omitted, the local Docker daemon is used. When --ssh-string
    is set, the API image is built locally, exported, transferred to the remote host,
    loaded there, and compose is started without a remote build. When --delete-volume=yes,
    existing compose volumes are removed before the stack is recreated.

.PARAMETER SshString
    SSH config alias for remote Docker (e.g. example). The script prepends "ssh"
    when connecting; do not include "ssh" in the value. When omitted, localhost Docker is used.

.PARAMETER DeleteImage
    Whether to remove built images during stack teardown. Default: no.

.PARAMETER DeleteVolume
    Whether to remove data volumes before starting. Default: no.

.PARAMETER ReverseProxy
    Reverse-proxy mode. Default: sslh (no host port publishing on remote deploy).

.PARAMETER DomainName
    Public hostname to route to the API container through nginx on the remote host.

.PARAMETER InternalPort
    Host port mapped to the API (default: 8080 from stack manifest).
    Remote with --domain: container port for domain routing (default: 8080).

.PARAMETER VolumeDir
    Deploy/data directory. Default: <USER-PROFILE-NAME>/docker/dogan-api.

.PARAMETER VolumeName
    Named volume label (default: dogan-api-vol). Compose still uses dogan-data / dogan-postgres-data.

.PARAMETER NetworkName
    Docker network name. Default: t3-net from stack manifest.

.PARAMETER PublicPort
    Public HTTPS port for sslh/nginx. Default: 443.

.PARAMETER Mode
    local or server (set by run-on-docker-local.ps1 / run-on-docker-server.ps1).

.EXAMPLE
    .\run-on-docker.ps1

.EXAMPLE
    .\run-on-docker.ps1 --delete-volume=yes

.EXAMPLE
    .\run-on-docker.ps1 --ssh-string=example --domain=dogan.example.com --internal-port=8080

.EXAMPLE
    .\run-on-docker.ps1 --ssh-string=example --delete-image=no --delete-volume=no
#>
[CmdletBinding()]
param(
    [Alias('ssh-string')]
    [string]$SshString,
    [Alias('delete-image')]
    [string]$DeleteImage,
    [Alias('delete-volume')]
    [string]$DeleteVolume,
    [Alias('reverse-proxy')]
    [string]$ReverseProxy = 'sslh',
    [Alias('domain')]
    [string]$DomainName,
    [Alias('internal-port')]
    [string]$InternalPort,
    [Alias('volume-dir')]
    [string]$VolumeDir,
    [Alias('volume-name')]
    [string]$VolumeName,
    [Alias('network-name')]
    [string]$NetworkName,
    [Alias('public-port')]
    [string]$PublicPort = '443',
    [string]$DeployMode,
    [switch]$Help,
    [Parameter(ValueFromRemainingArguments = $true)]
    [string[]]$RemainingArguments
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$Script:ComposeFile = 'docker-compose.yml'
$Script:LocalDeployDir = Join-Path $PSScriptRoot '.deploy'
$Script:TlsImageTag = 'nginx:alpine'
$Script:DeploySyncFiles = @(
    'docker-compose.yml',
    'Dockerfile',
    'livekit.yaml',
    '.dockerignore',
    '.docker/stack.manifest.json'
)

function Show-RunOnDockerHelp {
    Write-Host @'
run-on-docker.ps1 - build and deploy dogan-api in Docker (engine)

USAGE:
  .\run-on-docker.ps1 [flags]

Prefer entry scripts:
  .\run-on-docker-local.ps1
  .\run-on-docker-server.ps1 --ssh-string=<alias>

FLAGS:
  --ssh-string=<alias>       SSH alias; null -> local daemon (default: null)
  --delete-image=<no|yes>    Remove built images during teardown (default: null -> no)
  --delete-volume=<no|yes>   Remove volumes before recreate (default: null -> no)
  --internal-port=<port>     Host/API port (default: null -> 8080 from manifest)
  --volume-dir=<path>        Deploy data directory (default: null -> <USER>/docker/dogan-api)
  --volume-name=<name>       Volume label (default: null -> dogan-api-vol)
  --network-name=<name>      Docker network (default: null -> t3-net from manifest)
  --reverse-proxy=<sslh|none> Remote reverse-proxy mode (default: sslh)
  --domain=<hostname>        Remote hostname mapping (requires --ssh-string); also adds https://<hostname> to CORS
  --public-port=<port>       Public HTTPS port (default: 443)
  --mode=<local|server>      Deploy mode hint from wrapper scripts (also -DeployMode)
  --help                     Show this help

EXAMPLES:
  .\run-on-docker.ps1
  .\run-on-docker.ps1 --delete-volume=yes
  .\run-on-docker.ps1 --internal-port=8080
  .\run-on-docker.ps1 --ssh-string=myserver --domain=dogan.xaigrok.ir

NOTES:
  - Use SSH config alias only; do not include "ssh" in --ssh-string.
  - Null defaults resolve as described in FLAGS.
  - Truthy values for yes/no flags: yes, true, 1, y, on.
  - Default API host port is 8080 (from .docker/stack.manifest.json) if not specified.
  - CORS defaults include local web (:30808) and https://dogan.xaigrok.ir; --domain appends https://<domain>.
  - Image build uses create-image.ps1 when present; otherwise docker compose build.
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

function Normalize-CliParameterValue {
    param(
        [string]$Name,
        [string]$Value
    )

    if ([string]::IsNullOrWhiteSpace($Value)) { return $Value }

    $Value = Remove-SurroundingQuotes -Value $Value.Trim()
    if ($Value -match '^--?(?<param>[\w-]+)(?:=(?<rest>.*))?$') {
        $paramKey = ($Matches['param'] -replace '-', '_').ToLowerInvariant()
        $nameKey = ($Name -replace '-', '_').ToLowerInvariant()
        if ($paramKey -eq $nameKey) {
            if ($null -ne $Matches['rest'] -and $Matches['rest'] -ne '') {
                return Remove-SurroundingQuotes -Value $Matches['rest']
            }
            return $null
        }
    }
    return $Value
}

function Merge-CliArguments {
    param([hashtable]$BoundParameters, [string[]]$RemainingArguments)

    if ($null -eq $RemainingArguments) {
        $RemainingArguments = @()
    }
    else {
        $RemainingArguments = @($RemainingArguments | Where-Object { -not [string]::IsNullOrWhiteSpace($_) })
    }

    $merged = @{}
    foreach ($key in $BoundParameters.Keys) {
        $normalizedKey = ([regex]::Replace($key, '([a-z0-9])([A-Z])', '$1_$2')).ToLowerInvariant()
        if ($normalizedKey -in @('remainingarguments', 'help')) { continue }
        if ($null -eq $BoundParameters[$key] -or $BoundParameters[$key] -eq '') { continue }

        $normalizedValue = Normalize-CliParameterValue -Name $normalizedKey -Value ([string]$BoundParameters[$key])
        if ($null -ne $normalizedValue -and $normalizedValue -ne '') {
            $merged[$normalizedKey] = $normalizedValue
        }
    }

    $index = 0
    while ($index -lt $RemainingArguments.Count) {
        $argument = $RemainingArguments[$index]
        if ($argument -match '^--?(?<name>[\w-]+)(?:=(?<value>.*))?$') {
            $normalizedKey = ($Matches['name'] -replace '-', '_').ToLowerInvariant()
            if ($normalizedKey -in @('help', 'h')) {
                $merged['help'] = $true
                $index++
                continue
            }
            if ($null -ne $Matches['value'] -and $Matches['value'] -ne '') {
                $merged[$normalizedKey] = Remove-SurroundingQuotes -Value $Matches['value']
                $index++
            }
            elseif (($index + 1) -lt $RemainingArguments.Count -and $RemainingArguments[$index + 1] -notmatch '^-') {
                $merged[$normalizedKey] = Remove-SurroundingQuotes -Value $RemainingArguments[$index + 1]
                $index += 2
            }
            else {
                $merged[$normalizedKey] = $true
                $index++
            }
        }
        elseif ($argument -match '^(-h|-help|--help|-\?|/\?)$') {
            $merged['help'] = $true
            $index++
        }
        else {
            throw "Unknown argument: '$argument'. Run with --help."
        }
    }
    return $merged
}

function Test-Truthy {
    param([string]$Value)

    switch ($Value.ToLowerInvariant()) {
        { $_ -in @('yes', 'true', '1', 'y', 'on') } { return $true }
        default { return $false }
    }
}

function Write-RunStep {
    param(
        [int]$Step,
        [int]$Total,
        [string]$Message
    )

    $percent = [math]::Round(($Step / $Total) * 100)
    Write-Progress -Activity 'dogan-api Docker run' -Status $Message -PercentComplete $percent
    Write-Host ("[{0}/{1}] {2}" -f $Step, $Total, $Message) -ForegroundColor Yellow
}

function Resolve-SshTarget {
    param([string]$SshString)

    if ([string]::IsNullOrWhiteSpace($SshString)) {
        return [pscustomobject]@{
            IsLocal  = $true
            SshAlias = $null
        }
    }

    $alias = $SshString.Trim()

    if ($alias -match '^(?i)ssh(\s|$)') {
        throw 'Invalid --ssh-string value. Pass only the SSH config alias (e.g. --ssh-string=example). Do not include "ssh".'
    }

    if ([string]::IsNullOrWhiteSpace($alias)) {
        throw 'Invalid --ssh-string value. Example: --ssh-string=example'
    }

    return [pscustomobject]@{
        IsLocal  = $false
        SshAlias = $alias
    }
}

function Invoke-RemoteCommand {
    param(
        [pscustomobject]$Target,
        [string]$Command,
        [string]$WorkingDirectory = $null
    )

    $remoteCommand = if ($WorkingDirectory) { "cd '$WorkingDirectory' && $Command" } else { $Command }

    if ($Target.IsLocal) {
        if ($WorkingDirectory) {
            Push-Location $WorkingDirectory
            try { return (Invoke-Expression $Command 2>&1 | Out-String).Trim() }
            finally { Pop-Location }
        }
        return (Invoke-Expression $Command 2>&1 | Out-String).Trim()
    }

    $escapedCommand = $remoteCommand -replace "'", "'\''"
    $output = & ssh $Target.SshAlias "bash -lc '$escapedCommand'" 2>&1 | Out-String
    if ($LASTEXITCODE -ne 0) { throw "Remote command failed (exit $LASTEXITCODE): $remoteCommand" }
    return $output.Trim()
}

function Invoke-RemoteShell {
    param(
        [pscustomobject]$Target,
        [string]$Command,
        [string]$WorkingDirectory = $null
    )

    $remoteCommand = if ($WorkingDirectory) { "cd '$WorkingDirectory' && $Command" } else { $Command }

    if ($Target.IsLocal) {
        if ($WorkingDirectory) {
            Push-Location $WorkingDirectory
            try { Invoke-Expression $Command | Out-Null }
            finally { Pop-Location }
        }
        else {
            Invoke-Expression $Command | Out-Null
        }
        if ($LASTEXITCODE -ne 0) { throw "Command failed (exit $LASTEXITCODE): $Command" }
        return
    }

    $escapedCommand = $remoteCommand -replace "'", "'\''"
    & ssh $Target.SshAlias "bash -lc '$escapedCommand'"
    if ($LASTEXITCODE -ne 0) { throw "Remote command failed (exit $LASTEXITCODE): $remoteCommand" }
}

function Test-DockerCliAvailable {
    param([pscustomobject]$Target = $null)

    if ($null -eq $Target -or $Target.IsLocal) {
        & docker version | Out-Null
        if ($LASTEXITCODE -ne 0) { throw 'Docker CLI is not available or not running.' }
        return
    }

    Invoke-RemoteShell -Target $Target -Command 'docker version'
}

function Copy-FileToRemote {
    param(
        [pscustomobject]$Target,
        [string]$LocalPath,
        [string]$RemotePath
    )

    & scp -o StrictHostKeyChecking=accept-new $LocalPath "$($Target.SshAlias):$RemotePath"
    if ($LASTEXITCODE -ne 0) { throw "Failed to copy '$LocalPath' to remote." }
}

function Sync-DeployFilesToRemote {
    param(
        [pscustomobject]$Target,
        [string]$LocalRoot,
        [string]$RemotePath
    )

    Invoke-RemoteShell -Target $Target -Command "mkdir -p '$RemotePath' '$RemotePath/.docker'"

    foreach ($relativePath in $Script:DeploySyncFiles) {
        $localPath = Join-Path $LocalRoot $relativePath
        if (-not (Test-Path $localPath)) {
            throw "Missing deploy file: $relativePath"
        }

        $remoteTarget = if ($relativePath -like '.docker/*') {
            "$RemotePath/$relativePath"
        }
        else {
            "$RemotePath/"
        }

        Copy-FileToRemote -Target $Target -LocalPath $localPath -RemotePath $remoteTarget
    }
}

function Get-StackManifest {
    param([string]$ProjectRoot)

    $manifestPath = Join-Path $ProjectRoot '.docker/stack.manifest.json'
    if (-not (Test-Path $manifestPath)) { return $null }
    return Get-Content -Path $manifestPath -Raw | ConvertFrom-Json
}

function Get-StackImageTag {
    param([string]$ProjectRoot)

    $imageTag = 'dogan-api:latest'
    $manifest = Get-StackManifest -ProjectRoot $ProjectRoot
    if ($manifest) {
        if ($manifest.apiImageTag) { $imageTag = [string]$manifest.apiImageTag }
        elseif ($manifest.imageTag) { $imageTag = [string]$manifest.imageTag }
    }
    return $imageTag
}

function Get-ImageArchiveName {
    param([string]$StackName)

    return ($StackName -replace '[^a-zA-Z0-9._-]', '-') + '-images.tar'
}

function Build-LocalDockerImages {
    param(
        [string]$ProjectRoot,
        [string]$ImageTag = $null
    )

    $createImageScript = Join-Path $ProjectRoot 'create-image.ps1'
    if (Test-Path $createImageScript) {
        $buildArgs = @()
        if (-not [string]::IsNullOrWhiteSpace($ImageTag) -and $ImageTag -match '^(?<name>[^:]+):(?<tag>.+)$') {
            $buildArgs += "--image-name=$($Matches['name'])"
            $buildArgs += "--tag=$($Matches['tag'])"
        }
        & $createImageScript @buildArgs
        if ($LASTEXITCODE -ne 0) { throw "create-image.ps1 failed (exit $LASTEXITCODE)" }
        return
    }

    Push-Location $ProjectRoot
    try {
        & docker compose -p dogan -f $Script:ComposeFile build dogan-api
        if ($LASTEXITCODE -ne 0) { throw "docker compose build failed (exit $LASTEXITCODE)" }
    }
    finally {
        Pop-Location
    }
}

function Export-LocalDockerImages {
    param(
        [string[]]$ImageTags,
        [string]$ArchivePath
    )

    $parentDirectory = Split-Path -Parent $ArchivePath
    if (-not (Test-Path -LiteralPath $parentDirectory)) {
        New-Item -ItemType Directory -Path $parentDirectory -Force | Out-Null
    }
    if (Test-Path -LiteralPath $ArchivePath) {
        Remove-Item -LiteralPath $ArchivePath -Force
    }

    & docker save -o $ArchivePath @ImageTags
    if ($LASTEXITCODE -ne 0) { throw "docker save failed (exit $LASTEXITCODE)" }
}

function Transfer-DockerImagesToRemote {
    param(
        [pscustomobject]$Target,
        [string[]]$ImageTags,
        [string]$RemotePath,
        [string]$StackName
    )

    $archiveName = Get-ImageArchiveName -StackName $StackName
    $localArchive = Join-Path $Script:LocalDeployDir $archiveName
    $remoteArchive = "$RemotePath/$archiveName"

    try {
        Export-LocalDockerImages -ImageTags $ImageTags -ArchivePath $localArchive

        $tarSizeMb = [math]::Round((Get-Item $localArchive).Length / 1MB, 1)
        Write-Host "Transferring images ($tarSizeMb MB) to remote host..." -ForegroundColor Cyan
        Copy-FileToRemote -Target $Target -LocalPath $localArchive -RemotePath $remoteArchive

        Write-Host 'Loading images on remote host...' -ForegroundColor Cyan
        Invoke-RemoteShell -Target $Target -Command "docker load -i '$remoteArchive' && rm -f '$remoteArchive'"
        Write-Host 'Images loaded on remote host.' -ForegroundColor Green
    }
    finally {
        if (Test-Path -LiteralPath $localArchive) {
            Remove-Item -LiteralPath $localArchive -Force -ErrorAction SilentlyContinue
        }
    }
}

function Test-PortNumber {
    param(
        [string]$Value,
        [string]$ParameterName
    )

    if ($Value -notmatch '^\d+$') {
        throw "Invalid $ParameterName value '$Value'. Use a numeric port between 1 and 65535."
    }

    $port = [int]$Value
    if ($port -lt 1 -or $port -gt 65535) {
        throw "Invalid $ParameterName value '$Value'. Use a port between 1 and 65535."
    }
}

function Test-ReverseProxyMode {
    param([string]$Value)

    $normalized = $Value.Trim().ToLowerInvariant()
    if ($normalized -in @('sslh', 'none', 'direct', 'off')) { return $normalized }
    throw "Invalid --reverse-proxy value '$Value'. Allowed: sslh, none."
}

function Get-DefaultVolumeDir {
    param([string]$ContainerName)

    $userName = if ($env:USERNAME) { $env:USERNAME } elseif ($env:USER) { $env:USER } else { 'user' }
    $isWindowsHost = ($env:OS -match 'Windows') -or (-not [string]::IsNullOrWhiteSpace($env:USERPROFILE) -and $env:USERPROFILE -match '^[A-Za-z]:\\')
    if ($isWindowsHost) {
        return (Join-Path $env:USERPROFILE "docker\$ContainerName")
    }
    return "/$userName/docker/$ContainerName"
}

function Get-UnusedSafePort {
    param([int[]]$ExcludePorts = @())

    $min = 30000
    $max = 32767
    $used = @{}
    foreach ($excluded in @($ExcludePorts)) {
        if ($excluded -gt 0) { $used[$excluded] = $true }
    }

    try {
        Get-NetTCPConnection -ErrorAction SilentlyContinue |
            ForEach-Object { $used[[int]$_.LocalPort] = $true }
    }
    catch {
        $netstat = & netstat -ano 2>$null
        foreach ($line in @($netstat)) {
            if ($line -match ':(\d+)\s') {
                $used[[int]$Matches[1]] = $true
            }
        }
    }

    for ($attempt = 0; $attempt -lt 200; $attempt++) {
        $candidate = Get-Random -Minimum $min -Maximum ($max + 1)
        if (-not $used.ContainsKey($candidate)) {
            return [string]$candidate
        }
    }

    throw 'Could not find a free port in range 30000-32767.'
}

function Resolve-RemoteWorkDir {
    param(
        [string]$ProjectRoot,
        [pscustomobject]$Target,
        [string]$ContainerName,
        [string]$VolumeDir = $null
    )

    if ($Target.IsLocal) {
        return $ProjectRoot
    }

    $manifest = Get-StackManifest -ProjectRoot $ProjectRoot
    if ($manifest -and $manifest.remoteWorkDir) {
        return ([string]$manifest.remoteWorkDir).TrimEnd('/')
    }

    if (-not [string]::IsNullOrWhiteSpace($VolumeDir)) {
        return ($VolumeDir -replace '\\', '/').TrimEnd('/')
    }

    $stackName = if ($manifest -and $manifest.stackName) { [string]$manifest.stackName } else { 'dogan' }
    return "/cloud-admin/docker/$stackName"
}

function Get-SslhRuntimeInfo {
    param([pscustomobject]$Target)

    $inspectJson = Invoke-RemoteCommand -Target $Target -Command "docker inspect sslh --format '{{json .}}' 2>/dev/null || true"
    if ([string]::IsNullOrWhiteSpace($inspectJson)) {
        throw 'sslh container not found on remote host. Start sslh before using --reverse-proxy=sslh with --domain.'
    }

    $inspect = $inspectJson | ConvertFrom-Json
    $configMount = $inspect.Mounts | Where-Object { $_.Destination -eq '/etc/sslh' } | Select-Object -First 1
    if (-not $configMount) {
        throw 'Could not locate sslh config mount at /etc/sslh.'
    }

    $networkName = ($inspect.NetworkSettings.Networks.PSObject.Properties | Select-Object -First 1).Name
    if ([string]::IsNullOrWhiteSpace($networkName)) {
        throw 'Could not determine sslh Docker network.'
    }

    return [pscustomobject]@{
        ContainerName = [string]$inspect.Name.TrimStart('/')
        ConfigPath    = Join-Path $configMount.Source 'sslh.cfg'
        NetworkName   = [string]$networkName
    }
}

function Set-SslhDomainMapping {
    param(
        [pscustomobject]$Target,
        [string]$DomainName,
        [string]$TlsContainerName,
        [string]$ApiContainerPort,
        [string]$ApiContainerName,
        [string]$StackNetworkName,
        [pscustomobject]$SslhInfo,
        [string]$TlsImageTag = 'nginx:alpine'
    )

    $safeDomain = $DomainName.Trim().ToLowerInvariant()
    $configPath = ($SslhInfo.ConfigPath -replace '\\', '/')
    $sslhEntry = "  { name: `"tls`"; host: `"$TlsContainerName`"; port: `"443`"; sni_hostnames: [ `"$safeDomain`" ]; },"

    $remoteScriptPath = '/tmp/dogan-sslh-domain-map.sh'
    $localScriptPath = Join-Path $Script:LocalDeployDir 'sslh-domain-map.sh'
    $parentDirectory = Split-Path -Parent $localScriptPath
    if (-not (Test-Path -LiteralPath $parentDirectory)) {
        New-Item -ItemType Directory -Path $parentDirectory -Force | Out-Null
    }

    $bashScript = @"
#!/bin/bash
set -euo pipefail

TLS_HOST='$TlsContainerName'
API='$ApiContainerName'
API_PORT='$ApiContainerPort'
STACK_NET='$StackNetworkName'
SSLH_NET='$($SslhInfo.NetworkName)'
TLS_DIR='/cloud-admin/docker-volumes/dogan/tls'
CFG='$configPath'
DOMAIN='$safeDomain'
SSLH_ENTRY='$sslhEntry'

docker network connect "`$SSLH_NET" "`$API" 2>/dev/null || true

sudo mkdir -p "`$TLS_DIR"
if [ ! -f "`$TLS_DIR/cert.pem" ]; then
  sudo openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
    -keyout "`$TLS_DIR/key.pem" -out "`$TLS_DIR/cert.pem" \
    -subj "/CN=`$DOMAIN"
fi

sudo tee "`$TLS_DIR/default.conf" >/dev/null <<'NGINXEOF'
server {
    listen 443 ssl;
    server_name $safeDomain;
    ssl_certificate /etc/nginx/certs/cert.pem;
    ssl_certificate_key /etc/nginx/certs/key.pem;
    location / {
        proxy_pass http://${ApiContainerName}:${ApiContainerPort};
        proxy_http_version 1.1;
        proxy_set_header Host `$host;
        proxy_set_header X-Real-IP `$remote_addr;
        proxy_set_header X-Forwarded-For `$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto https;
    }
}
NGINXEOF

docker rm -f "`$TLS_HOST" 2>/dev/null || true
docker run -d --name "`$TLS_HOST" --network "`$SSLH_NET" \
  -v "`$TLS_DIR/cert.pem:/etc/nginx/certs/cert.pem:ro" \
  -v "`$TLS_DIR/key.pem:/etc/nginx/certs/key.pem:ro" \
  -v "`$TLS_DIR/default.conf:/etc/nginx/conf.d/default.conf:ro" \
  $TlsImageTag
docker network connect "`$STACK_NET" "`$TLS_HOST" 2>/dev/null || true

sudo python3 - <<'PY'
from pathlib import Path
import re

cfg_path = Path('$configPath')
domain = '$safeDomain'
entry = '$sslhEntry'
text = cfg_path.read_text()
if domain in text:
    pattern = r'  \{ name: "tls"; host: "[^"]+"; port: "[^"]+"; sni_hostnames: \[ "' + re.escape(domain) + r'" \]; \},?\n?'
    text = re.sub(pattern, entry + '\n', text)
else:
    text = text.replace('protocols:\n(\n', 'protocols:\n(\n' + entry + '\n', 1)
cfg_path.write_text(text)
PY

docker restart '$($SslhInfo.ContainerName)'
"@

    [System.IO.File]::WriteAllText($localScriptPath, ($bashScript -replace "`r`n", "`n"))

    Write-Host "Mapping domain '$safeDomain' via sslh -> ${TlsContainerName}:443 -> ${ApiContainerName}..." -ForegroundColor Cyan
    Copy-FileToRemote -Target $Target -LocalPath $localScriptPath -RemotePath $remoteScriptPath
    Invoke-RemoteShell -Target $Target -Command "chmod +x '$remoteScriptPath' && bash '$remoteScriptPath' && rm -f '$remoteScriptPath'"
    Write-Host "sslh domain mapping configured for https://${safeDomain}/" -ForegroundColor Green
}

function Set-DomainReverseProxyMapping {
    param(
        [pscustomobject]$Target,
        [string]$DomainName,
        [string]$ApiContainerName,
        [string]$ContainerPort,
        [string]$HttpsPort,
        [string]$ReverseProxyMode,
        [string]$StackNetworkName,
        [string]$TlsImageTag = 'nginx:alpine'
    )

    if ($Target.IsLocal) {
        throw '--domain requires --ssh-string for remote reverse-proxy configuration.'
    }

    $safeDomain = $DomainName.Trim().ToLowerInvariant()
    if ($safeDomain -notmatch '^[a-z0-9]([a-z0-9.-]*[a-z0-9])?$') {
        throw "Invalid --domain value '$DomainName'."
    }

    Test-PortNumber -Value $ContainerPort -ParameterName '--internal-port'
    Test-PortNumber -Value $HttpsPort -ParameterName '--public-port'

    if ($ReverseProxyMode -eq 'sslh') {
        $sslhInfo = Get-SslhRuntimeInfo -Target $Target
        $tlsContainerName = "${ApiContainerName}-tls"
        Set-SslhDomainMapping -Target $Target -DomainName $safeDomain -TlsContainerName $tlsContainerName -ApiContainerPort $ContainerPort -ApiContainerName $ApiContainerName -StackNetworkName $StackNetworkName -SslhInfo $sslhInfo -TlsImageTag $TlsImageTag
        return
    }

    $configFileName = "$safeDomain.conf"
    $availablePath = "/etc/nginx/sites-available/$configFileName"
    $enabledPath = "/etc/nginx/sites-enabled/$configFileName"
    $listenDirective = 'listen 80;'

    $nginxConfig = @"
server {
    $listenDirective
    server_name $safeDomain;
    location / {
        resolver 127.0.0.11 valid=30s ipv6=off;
        set `$upstream $ApiContainerName`:$ContainerPort;
        proxy_pass http://`$upstream;
        proxy_http_version 1.1;
        proxy_set_header Host `$host;
        proxy_set_header X-Real-IP `$remote_addr;
        proxy_set_header X-Forwarded-For `$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto `$scheme;
        proxy_set_header Upgrade `$http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
"@

    $encodedConfig = [Convert]::ToBase64String([System.Text.Encoding]::UTF8.GetBytes($nginxConfig))
    $installCommand = @"
set -e
tmp='$(Split-Path -Leaf $availablePath).tmp'
echo '$encodedConfig' | base64 -d | sudo tee "/tmp/`$tmp" >/dev/null
sudo mv "/tmp/`$tmp" '$availablePath'
sudo ln -sf '$availablePath' '$enabledPath'
if command -v nginx >/dev/null 2>&1; then
  sudo nginx -t
  sudo systemctl reload nginx || sudo service nginx reload
fi
"@

    Write-Host "Mapping domain '$safeDomain' -> ${ApiContainerName}:${ContainerPort} (public port $HttpsPort)..." -ForegroundColor Cyan
    Invoke-RemoteShell -Target $Target -Command $installCommand
    Write-Host "Domain mapping configured for https://${safeDomain}/" -ForegroundColor Green
}

function Get-DockerManifestDefaults {
    param([string]$ProjectRoot)

    $defaults = @{
        ApiHost           = 'dogan-api'
        ApiPort           = '8080'
        ApiContainerName  = 'dogan-api'
        ContainerName     = 'dogan-api'
        DockerNetwork     = 't3-net'
    }

    $manifest = Get-StackManifest -ProjectRoot $ProjectRoot
    if (-not $manifest) { return $defaults }

    if ($manifest.PSObject.Properties.Match('containerName').Count -gt 0 -and $manifest.containerName) {
        $defaults.ApiHost = [string]$manifest.containerName
        $defaults.ContainerName = [string]$manifest.containerName
    }
    if ($manifest.PSObject.Properties.Match('apiContainerName').Count -gt 0 -and $manifest.apiContainerName) {
        $defaults.ApiContainerName = [string]$manifest.apiContainerName
    }
    if ($manifest.PSObject.Properties.Match('internalPort').Count -gt 0 -and $manifest.internalPort) {
        $defaults.ApiPort = [string]$manifest.internalPort
    }
    elseif ($manifest.PSObject.Properties.Match('apiPort').Count -gt 0 -and $manifest.apiPort) {
        $defaults.ApiPort = [string]$manifest.apiPort
    }
    if ($manifest.PSObject.Properties.Match('dockerNetwork').Count -gt 0 -and $manifest.dockerNetwork) {
        $defaults.DockerNetwork = [string]$manifest.dockerNetwork
    }

    return $defaults
}

function Get-CorsAllowedOrigins {
    param([string]$DomainName = $null)

    $origins = [System.Collections.Generic.List[string]]::new()
    foreach ($origin in @(
            'http://localhost:30808',
            'http://127.0.0.1:30808',
            'https://dogan.xaigrok.ir'
        )) {
        $origins.Add($origin)
    }

    if (-not [string]::IsNullOrWhiteSpace($DomainName)) {
        $domainOrigin = "https://$($DomainName.Trim().TrimEnd('/'))"
        if (-not ($origins -contains $domainOrigin)) {
            $origins.Add($domainOrigin)
        }
    }

    return ($origins -join ',')
}

function Set-ComposeEnvironment {
    param(
        [string]$NetworkName,
        [string]$ApiImageTag,
        [string]$ApiContainerName,
        [bool]$PublishApiHostPort = $true,
        [string]$HostApiPort = $null,
        [string]$HostPostgresPort = $null,
        [string]$CorsAllowedOrigins = $null
    )

    $env:DOCKER_NETWORK = $NetworkName
    $env:API_IMAGE_TAG = $ApiImageTag
    $env:API_CONTAINER_NAME = $ApiContainerName

    if (-not [string]::IsNullOrWhiteSpace($CorsAllowedOrigins)) {
        $env:DOGAN_CORS_ALLOWED_ORIGINS = $CorsAllowedOrigins
    }
    else {
        Remove-Item Env:DOGAN_CORS_ALLOWED_ORIGINS -ErrorAction SilentlyContinue
    }

    if ($PublishApiHostPort) {
        if (-not [string]::IsNullOrWhiteSpace($HostApiPort)) {
            $env:DOGAN_PUBLISH_PORT = $HostApiPort
        }
        else {
            Remove-Item Env:DOGAN_PUBLISH_PORT -ErrorAction SilentlyContinue
        }
        if (-not [string]::IsNullOrWhiteSpace($HostPostgresPort)) {
            $env:POSTGRES_PUBLISH_PORT = $HostPostgresPort
        }
        else {
            Remove-Item Env:POSTGRES_PUBLISH_PORT -ErrorAction SilentlyContinue
        }
    }
    else {
        $env:DOGAN_PUBLISH_PORT = ''
        $env:POSTGRES_PUBLISH_PORT = ''
    }
}

function Get-RemoteComposeEnvironmentPrefix {
    param(
        [string]$NetworkName,
        [string]$ApiImageTag,
        [string]$ApiContainerName,
        [bool]$PublishApiHostPort = $false,
        [string]$HostApiPort = $null,
        [string]$HostPostgresPort = $null,
        [string]$CorsAllowedOrigins = $null
    )

    $prefix = "DOCKER_NETWORK='$NetworkName' API_IMAGE_TAG='$ApiImageTag' API_CONTAINER_NAME='$ApiContainerName' "
    if (-not [string]::IsNullOrWhiteSpace($CorsAllowedOrigins)) {
        $safeCors = $CorsAllowedOrigins.Replace("'", "'\''")
        $prefix += "DOGAN_CORS_ALLOWED_ORIGINS='$safeCors' "
    }
    if (-not $PublishApiHostPort) {
        $prefix += "DOGAN_PUBLISH_PORT='' POSTGRES_PUBLISH_PORT='' "
    }
    else {
        if (-not [string]::IsNullOrWhiteSpace($HostApiPort)) {
            $prefix += "DOGAN_PUBLISH_PORT='$HostApiPort' "
        }
        if (-not [string]::IsNullOrWhiteSpace($HostPostgresPort)) {
            $prefix += "POSTGRES_PUBLISH_PORT='$HostPostgresPort' "
        }
    }
    return $prefix
}

function Test-DockerComposeFile {
    param([string]$ProjectRoot)

    $composePath = Join-Path $ProjectRoot $Script:ComposeFile
    if (-not (Test-Path $composePath)) {
        throw "Missing $Script:ComposeFile in the repo root."
    }

    $dockerfilePath = Join-Path $ProjectRoot 'Dockerfile'
    if (-not (Test-Path $dockerfilePath)) {
        throw 'Missing Dockerfile in the repo root.'
    }

    $livekitPath = Join-Path $ProjectRoot 'livekit.yaml'
    if (-not (Test-Path $livekitPath)) {
        throw 'Missing livekit.yaml in the repo root.'
    }
}

function Invoke-ComposeStack {
    param(
        [pscustomobject]$Target,
        [string]$WorkingDirectory,
        [bool]$RemoveVolumes,
        [bool]$RemoveImages,
        [bool]$Build,
        [string]$NetworkName,
        [string]$ApiImageTag,
        [string]$ApiContainerName,
        [bool]$PublishApiHostPort,
        [string]$HostApiPort = $null,
        [string]$HostPostgresPort = $null,
        [string]$CorsAllowedOrigins = $null
    )

    $downFlag = if ($RemoveVolumes) { ' -v' } else { '' }
    $rmiFlag = if ($RemoveImages) { ' --rmi local' } else { '' }
    $composeDown = "docker compose -p dogan -f $Script:ComposeFile down$rmiFlag$downFlag"
    $buildFlag = if ($Build) { ' --build' } else { '' }
    $composeUp = "docker compose -p dogan -f $Script:ComposeFile up -d$buildFlag"

    if ($Target.IsLocal) {
        Push-Location $WorkingDirectory
        try {
            Set-ComposeEnvironment -NetworkName $NetworkName -ApiImageTag $ApiImageTag -ApiContainerName $ApiContainerName -PublishApiHostPort:$PublishApiHostPort -HostApiPort $HostApiPort -HostPostgresPort $HostPostgresPort -CorsAllowedOrigins $CorsAllowedOrigins
            Invoke-Expression $composeDown | Out-Null
            if ($LASTEXITCODE -ne 0) {
                Write-Host 'Compose down skipped or partial (stack may not exist yet).' -ForegroundColor DarkYellow
            }
            Invoke-Expression $composeUp | Out-Null
            if ($LASTEXITCODE -ne 0) { throw 'docker compose up failed.' }
        }
        finally {
            Pop-Location
        }
        return
    }

    try {
        Invoke-RemoteShell -Target $Target -Command $composeDown -WorkingDirectory $WorkingDirectory
    }
    catch {
        Write-Host "Compose down skipped: $($_.Exception.Message)" -ForegroundColor DarkYellow
    }

    $envPrefix = Get-RemoteComposeEnvironmentPrefix -NetworkName $NetworkName -ApiImageTag $ApiImageTag -ApiContainerName $ApiContainerName -PublishApiHostPort:$PublishApiHostPort -HostApiPort $HostApiPort -HostPostgresPort $HostPostgresPort -CorsAllowedOrigins $CorsAllowedOrigins
    Invoke-RemoteShell -Target $Target -Command "${envPrefix}$composeUp" -WorkingDirectory $WorkingDirectory
}

if ($Help) {
    Show-RunOnDockerHelp
    Get-Help $PSCommandPath -Full
    exit 0
}

$cliArgs = Merge-CliArguments -BoundParameters $PSBoundParameters -RemainingArguments $RemainingArguments
if ($cliArgs['help']) {
    Show-RunOnDockerHelp
    Get-Help $PSCommandPath -Full
    exit 0
}

$sshStringValue = if ($cliArgs['ssh_string']) { [string]$cliArgs['ssh_string'] } else { [string]$SshString }
$sshStringValue = Normalize-CliParameterValue -Name 'ssh_string' -Value $sshStringValue
$deleteImageValue = if ($cliArgs['delete_image']) { [string]$cliArgs['delete_image'] } elseif (-not [string]::IsNullOrWhiteSpace($DeleteImage)) { [string]$DeleteImage } else { 'no' }
$deleteImageValue = Normalize-CliParameterValue -Name 'delete_image' -Value $deleteImageValue
$deleteVolumeValue = if ($cliArgs['delete_volume']) { [string]$cliArgs['delete_volume'] } elseif (-not [string]::IsNullOrWhiteSpace($DeleteVolume)) { [string]$DeleteVolume } else { 'no' }
$deleteVolumeValue = Normalize-CliParameterValue -Name 'delete_volume' -Value $deleteVolumeValue
$reverseProxyValue = if ($cliArgs['reverse_proxy']) { [string]$cliArgs['reverse_proxy'] } else { [string]$ReverseProxy }
$reverseProxyValue = Normalize-CliParameterValue -Name 'reverse_proxy' -Value $reverseProxyValue
$domainValue = if ($cliArgs['domain']) { [string]$cliArgs['domain'] } else { [string]$DomainName }
$domainValue = Normalize-CliParameterValue -Name 'domain' -Value $domainValue
$publicPortValue = if ($cliArgs['public_port']) { [string]$cliArgs['public_port'] } else { [string]$PublicPort }
$publicPortValue = Normalize-CliParameterValue -Name 'public_port' -Value $publicPortValue
$volumeDirValue = if ($cliArgs['volume_dir']) { [string]$cliArgs['volume_dir'] } elseif (-not [string]::IsNullOrWhiteSpace($VolumeDir)) { [string]$VolumeDir } else { $null }
$volumeDirValue = Normalize-CliParameterValue -Name 'volume_dir' -Value $volumeDirValue
$volumeNameValue = if ($cliArgs['volume_name']) { [string]$cliArgs['volume_name'] } elseif (-not [string]::IsNullOrWhiteSpace($VolumeName)) { [string]$VolumeName } else { $null }
$volumeNameValue = Normalize-CliParameterValue -Name 'volume_name' -Value $volumeNameValue
$networkNameArg = if ($cliArgs['network_name']) { [string]$cliArgs['network_name'] } elseif (-not [string]::IsNullOrWhiteSpace($NetworkName)) { [string]$NetworkName } else { $null }
$networkNameArg = Normalize-CliParameterValue -Name 'network_name' -Value $networkNameArg
$modeValue = if ($cliArgs['mode']) { [string]$cliArgs['mode'] } elseif ($cliArgs['deploy_mode']) { [string]$cliArgs['deploy_mode'] } elseif (-not [string]::IsNullOrWhiteSpace($DeployMode)) { [string]$DeployMode } else { $null }
$modeValue = Normalize-CliParameterValue -Name 'mode' -Value $modeValue
$removeVolumes = Test-Truthy -Value $deleteVolumeValue
$removeImages = Test-Truthy -Value $deleteImageValue
$reverseProxyMode = Test-ReverseProxyMode -Value $reverseProxyValue

$ProjectRoot = $PSScriptRoot
$manifestDefaults = Get-DockerManifestDefaults -ProjectRoot $ProjectRoot
$containerName = $manifestDefaults.ContainerName
if ([string]::IsNullOrWhiteSpace($volumeNameValue)) {
    $volumeNameValue = "$containerName-volume"
}
if ([string]::IsNullOrWhiteSpace($volumeDirValue)) {
    $volumeDirValue = Get-DefaultVolumeDir -ContainerName $containerName
}

$target = Resolve-SshTarget -SshString $sshStringValue
if ($modeValue -eq 'server' -and $target.IsLocal) {
    throw 'Server mode requires --ssh-string=<alias>.'
}
if (-not [string]::IsNullOrWhiteSpace($domainValue) -and $target.IsLocal) {
    throw '--domain requires --ssh-string for remote nginx configuration.'
}

$internalPortSpecified = $false
if ($cliArgs['internal_port']) {
    $internalPortValue = [string]$cliArgs['internal_port']
    $internalPortSpecified = $true
}
elseif (-not [string]::IsNullOrWhiteSpace($InternalPort)) {
    $internalPortValue = [string]$InternalPort
    $internalPortSpecified = $true
}
else {
    $internalPortValue = $null
}
$internalPortValue = Normalize-CliParameterValue -Name 'internal_port' -Value $internalPortValue

$hostApiPort = $null
$hostPostgresPort = $null
if ($target.IsLocal) {
    if (-not $internalPortSpecified -or [string]::IsNullOrWhiteSpace($internalPortValue)) {
        $hostApiPort = $manifestDefaults.ApiPort
        $internalPortValue = $hostApiPort
    }
    else {
        $hostApiPort = $internalPortValue
    }
    $hostPostgresPort = Get-UnusedSafePort -ExcludePorts @([int]$hostApiPort)
}
else {
    if (-not $internalPortSpecified -or [string]::IsNullOrWhiteSpace($internalPortValue)) {
        $internalPortValue = $manifestDefaults.ApiPort
    }
    if ($reverseProxyMode -in @('none', 'direct', 'off')) {
        $hostApiPort = $internalPortValue
    }
}

Test-PortNumber -Value $internalPortValue -ParameterName '--internal-port'
Test-PortNumber -Value $publicPortValue -ParameterName '--public-port'

$networkValue = if (-not [string]::IsNullOrWhiteSpace($networkNameArg)) { $networkNameArg } else { $manifestDefaults.DockerNetwork }
$apiHostValue = $manifestDefaults.ApiHost
$apiPortValue = $manifestDefaults.ApiPort
$apiContainerName = $manifestDefaults.ApiContainerName
$publishApiHostPort = $target.IsLocal -or ($reverseProxyMode -in @('none', 'direct', 'off'))
$corsAllowedOrigins = Get-CorsAllowedOrigins -DomainName $domainValue

$workDir = Resolve-RemoteWorkDir -ProjectRoot $ProjectRoot -Target $target -ContainerName $containerName -VolumeDir $(if ($target.IsLocal) { $null } else { $volumeDirValue })
$imageTag = Get-StackImageTag -ProjectRoot $ProjectRoot
$stackManifest = Get-StackManifest -ProjectRoot $ProjectRoot
$stackName = if ($stackManifest -and $stackManifest.stackName) { [string]$stackManifest.stackName } else { 'dogan' }

$targetLabel = if ($target.IsLocal) { 'localhost' } else { "ssh $($target.SshAlias)" }
$volumeAction = if ($removeVolumes) { 'removing volumes' } else { 'keeping volumes' }
$imageAction = if ($removeImages) { 'removing images' } else { 'keeping images' }
$proxyLabel = if ($publishApiHostPort) { "API host port $hostApiPort" } else { 'sslh (docker network only)' }
$domainLabel = if ([string]::IsNullOrWhiteSpace($domainValue)) { 'none' } else { $domainValue }
$totalSteps = if ($target.IsLocal) {
    3
}
else {
    if ([string]::IsNullOrWhiteSpace($domainValue)) { 6 } else { 7 }
}

try {
    $deployMode = if ($target.IsLocal) { 'local Docker' } else { 'local build + image transfer' }
    Write-Host ("Target: {0} ({1}) | network: {2} | api: {3}:{4} | host-port: {5} | volume-dir: {6} | volume-name: {7} | domain: {8} | proxy: {9} | {10} | {11}" -f `
        $targetLabel, $deployMode, $networkValue, $apiHostValue, $apiPortValue, `
        $(if ($hostApiPort) { $hostApiPort } else { 'n/a' }), $volumeDirValue, $volumeNameValue, `
        $domainLabel, $proxyLabel, $volumeAction, $imageAction) -ForegroundColor Cyan

    if ($target.IsLocal -and -not (Test-Path -LiteralPath $volumeDirValue)) {
        New-Item -ItemType Directory -Path $volumeDirValue -Force | Out-Null
        Write-Host "Created volume-dir: $volumeDirValue" -ForegroundColor DarkYellow
    }

    Write-RunStep -Step 1 -Total $totalSteps -Message 'Checking Docker files'
    Test-DockerComposeFile -ProjectRoot $ProjectRoot
    Test-DockerCliAvailable -Target $target

    Write-RunStep -Step 2 -Total $totalSteps -Message 'Building dogan-api image'
    Build-LocalDockerImages -ProjectRoot $ProjectRoot -ImageTag $imageTag

    if ($target.IsLocal) {
        Write-RunStep -Step 3 -Total $totalSteps -Message $(if ($removeVolumes) { 'Recreating stack (volumes removed)' } else { 'Recreating stack (keeping volumes)' })
        Invoke-ComposeStack -Target $target -WorkingDirectory $workDir -RemoveVolumes:$removeVolumes -RemoveImages:$removeImages -Build:$false -NetworkName $networkValue -ApiImageTag $imageTag -ApiContainerName $apiContainerName -PublishApiHostPort:$publishApiHostPort -HostApiPort $hostApiPort -HostPostgresPort $hostPostgresPort -CorsAllowedOrigins $corsAllowedOrigins
    }
    else {
        Write-RunStep -Step 3 -Total $totalSteps -Message "Syncing compose files to $targetLabel"
        Sync-DeployFilesToRemote -Target $target -LocalRoot $ProjectRoot -RemotePath $workDir

        Write-RunStep -Step 4 -Total $totalSteps -Message 'Transferring images to remote host'
        Transfer-DockerImagesToRemote -Target $target -ImageTags @($imageTag) -RemotePath $workDir -StackName $stackName

        Write-RunStep -Step 5 -Total $totalSteps -Message 'Checking remote Docker'
        Test-DockerCliAvailable -Target $target

        $stackStep = if ([string]::IsNullOrWhiteSpace($domainValue)) { 6 } else { 6 }
        Write-RunStep -Step $stackStep -Total $totalSteps -Message $(if ($removeVolumes) { 'Recreating stack (volumes removed)' } else { 'Recreating stack (keeping volumes)' })
        Invoke-ComposeStack -Target $target -WorkingDirectory $workDir -RemoveVolumes:$removeVolumes -RemoveImages:$removeImages -Build:$false -NetworkName $networkValue -ApiImageTag $imageTag -ApiContainerName $apiContainerName -PublishApiHostPort:$publishApiHostPort -HostApiPort $hostApiPort -HostPostgresPort $hostPostgresPort -CorsAllowedOrigins $corsAllowedOrigins

        if (-not [string]::IsNullOrWhiteSpace($domainValue)) {
            Write-RunStep -Step 7 -Total $totalSteps -Message "Mapping domain '$domainValue' to $apiContainerName"
            Set-DomainReverseProxyMapping -Target $target -DomainName $domainValue -ApiContainerName $apiContainerName -ContainerPort $internalPortValue -HttpsPort $publicPortValue -ReverseProxyMode $reverseProxyMode -StackNetworkName $networkValue -TlsImageTag $Script:TlsImageTag
        }
    }

    Write-Progress -Activity 'dogan-api Docker run' -Completed -Status 'Done'
    Write-Host ''

    if ($target.IsLocal) {
        Write-Host 'Stack is running on localhost.' -ForegroundColor Green
        Write-Host ("  API:      http://localhost:{0}/dogan/api/v1/health" -f $hostApiPort) -ForegroundColor Green
        Write-Host ("  Postgres: localhost:{0}" -f $hostPostgresPort) -ForegroundColor Green
        Write-Host '  LiveKit:  ws://localhost:7880' -ForegroundColor Green
        Write-Host ("  Network:  {0}" -f $networkValue) -ForegroundColor Green
        Write-Host ("  Volume:   {0} ({1})" -f $volumeNameValue, $volumeDirValue) -ForegroundColor Green
    }
    else {
        Write-Host "Stack is running on remote host at $workDir (network: $networkValue, api: ${apiHostValue}:${apiPortValue})." -ForegroundColor Green
        Write-Host ("Images were built locally and deployed to {0} without a remote build." -f $target.SshAlias) -ForegroundColor Green
        Write-Host ("  Volume:   {0} ({1})" -f $volumeNameValue, $volumeDirValue) -ForegroundColor Green
        if (-not [string]::IsNullOrWhiteSpace($domainValue)) {
            Write-Host "  URL:  https://${domainValue}/" -ForegroundColor Green
        }
    }
}
catch {
    Write-Progress -Activity 'dogan-api Docker run' -Completed -Status 'Failed'
    Write-Host ''
    Write-Host $_.Exception.Message -ForegroundColor Red
    Write-Host ''
    Show-RunOnDockerHelp
    exit 1
}
