Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$BaseUrl = if ($env:DOGAN_BASE_URL) { $env:DOGAN_BASE_URL } else { 'http://localhost:8080/dogan/api/v1' }
$ModelsDir = if ($env:DOGAN_MODELS_DIR) { $env:DOGAN_MODELS_DIR } else { Join-Path (Join-Path $PSScriptRoot '..') 'models' }
$EmbeddedToken = if ($env:DOGAN_EMBEDDED_API_TOKEN) {
    $env:DOGAN_EMBEDDED_API_TOKEN
} else {
    'dg_dev_a8f3c2e1b9d74f6a0e5c3b9d2f7a1e4c8b6d0f3a7e2c9b5d1f8a4e6c0b3d7f9'
}

$defaultLabels = @{
    yolov8_nano    = @('person', 'car', 'motorcycle', 'truck', 'speed_camera', 'speed_limit_sign')
    yolov8_small   = @('person', 'car', 'motorcycle', 'truck', 'speed_camera', 'speed_limit_sign')
    mobilenet_ssd  = @('person', 'car', 'motorcycle', 'truck')
}

function Get-FileSha256Hex {
    param([string]$Path)
    $hash = Get-FileHash -LiteralPath $Path -Algorithm SHA256
    return $hash.Hash.ToLowerInvariant()
}

function Invoke-DoganRequest {
    param(
        [string]$Method,
        [string]$Path,
        [hashtable]$Headers = @{},
        [object]$Body = $null
    )

    $uri = "$BaseUrl$Path"
    $params = @{
        Method          = $Method
        Uri             = $uri
        Headers         = $Headers
        UseBasicParsing = $true
        ErrorAction     = 'Stop'
    }

    if ($null -ne $Body) {
        $params['Body'] = ($Body | ConvertTo-Json -Depth 10)
        $params['ContentType'] = 'application/json'
    }

    return Invoke-RestMethod @params
}

if (-not (Test-Path -LiteralPath $ModelsDir)) {
    Write-Error "Models directory not found: $ModelsDir"
}

$authHeaders = @{ Authorization = "Bearer $EmbeddedToken" }
$registered = 0

Write-Host "Scanning models in $ModelsDir" -ForegroundColor Cyan

Get-ChildItem -LiteralPath $ModelsDir -Directory | ForEach-Object {
    $modelId = $_.Name
    $paramPath = Join-Path $_.FullName 'model.param'
    $binPath = Join-Path $_.FullName 'model.bin'

    if (-not (Test-Path -LiteralPath $paramPath) -or -not (Test-Path -LiteralPath $binPath)) {
        Write-Host "Skipping $modelId (missing model.param or model.bin)" -ForegroundColor Yellow
        return
    }

    $labels = if ($defaultLabels.ContainsKey($modelId)) { $defaultLabels[$modelId] } else { @() }
    $body = @{
        model_name    = $modelId
        param_sha256  = (Get-FileSha256Hex -Path $paramPath)
        bin_sha256    = (Get-FileSha256Hex -Path $binPath)
        labels        = $labels
        format        = 'ncnn'
        version       = '1.0.0'
    }

    Write-Host "Registering $modelId"
    Invoke-DoganRequest -Method POST -Path '/ai-models' -Headers $authHeaders -Body $body | Out-Null
    $registered++
}

Write-Host "Registered $registered model(s)." -ForegroundColor Green

if ($registered -gt 0) {
    $manifest = Invoke-DoganRequest -Method GET -Path '/models' -Headers $authHeaders
    Write-Host "Manifest models: $($manifest.models.Count)"
}
