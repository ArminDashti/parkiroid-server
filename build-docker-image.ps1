<#
.SYNOPSIS
    Compatibility wrapper for create-image.ps1.
#>
Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$createImage = Join-Path $PSScriptRoot 'create-image.ps1'
if (-not (Test-Path $createImage)) {
    Write-Host 'Missing create-image.ps1. Use .\create-image.ps1 instead of build-docker-image.ps1.' -ForegroundColor Red
    exit 1
}

Write-Host 'build-docker-image.ps1 -> create-image.ps1' -ForegroundColor DarkYellow
& $createImage @args
exit $LASTEXITCODE
