param(
    [string]$Version = "v1.19.28"
)

$ErrorActionPreference = "Stop"
$projectRoot = Split-Path -Parent $PSScriptRoot
$binDir = Join-Path $projectRoot "bin"
$archive = Join-Path $env:TEMP "mihomo-$Version.zip"
$asset = "mihomo-windows-amd64-compatible-$Version.zip"
$url = "https://github.com/MetaCubeX/mihomo/releases/download/$Version/$asset"

if (-not (Test-Path -LiteralPath $binDir)) {
    New-Item -ItemType Directory -Path $binDir | Out-Null
}

Invoke-WebRequest -Uri $url -OutFile $archive
if ($Version -eq "v1.19.28") {
    $expected = "6d8a079d01b3631e73e56b7b42a067afc14f9e3ad99f2880d38bb141cf8fcbe7"
    $actual = (Get-FileHash -LiteralPath $archive -Algorithm SHA256).Hash.ToLowerInvariant()
    if ($actual -ne $expected) { throw "mihomo checksum mismatch: $actual" }
} else {
    Write-Warning "No pinned checksum is available for $Version"
}

Expand-Archive -LiteralPath $archive -DestinationPath $binDir -Force
$downloaded = Get-ChildItem -LiteralPath $binDir -Filter "mihomo*.exe" | Select-Object -First 1
if (-not $downloaded) { throw "mihomo.exe was not found in the archive" }
$target = Join-Path $binDir "mihomo.exe"
if ($downloaded.FullName -ne $target) { Move-Item -LiteralPath $downloaded.FullName -Destination $target -Force }
Remove-Item -LiteralPath $archive -Force
& $target -v
