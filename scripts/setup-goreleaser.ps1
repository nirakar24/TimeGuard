<#!
.SYNOPSIS
Installs goreleaser (portable) into local .bin folder and updates PATH for current session.
#>
param(
  [string]$Version = "latest",
  [switch]$Prerelease
)

$ErrorActionPreference = 'Stop'
$bin = Join-Path $PSScriptRoot '..' | Resolve-Path | Join-Path -ChildPath '.bin'
if (!(Test-Path $bin)) { New-Item -ItemType Directory -Path $bin | Out-Null }

function Get-LatestTag {
  param([switch]$Prerelease)
  $base = 'https://api.github.com/repos/goreleaser/goreleaser/releases'
  $url  = if ($Prerelease) { "$base" } else { "$base/latest" }
  $headers = @{ 'User-Agent' = 'timeguard-setup-script' }
  try {
    $resp = Invoke-RestMethod -Headers $headers -UseBasicParsing -Uri $url
    if ($Prerelease) {
      $rel = $resp | Where-Object { -not $_.draft } | Select-Object -First 1
      return $rel.tag_name.TrimStart('v')
    } else {
      return $resp.tag_name.TrimStart('v')
    }
  }
  catch {
    Write-Warning "Failed to query GitHub API for latest tag: $_"
    return $null
  }
}

if ($Version -eq 'latest') {
  $lookup = Get-LatestTag -Prerelease:$Prerelease
  if ($lookup) { $Version = $lookup } else { Write-Warning 'Falling back to pinned version 2.1.0'; $Version = '2.1.0' }
}

$exe = Join-Path $bin 'goreleaser.exe'
if (Test-Path $exe) {
  Write-Host "goreleaser already installed at $exe" -ForegroundColor Green
} else {
  $arch = if ($env:PROCESSOR_ARCHITECTURE -match 'ARM64') { 'arm64' } else { 'x86_64' }
  $zipName = "goreleaser_Windows_${arch}.zip"
  $downloadUrl = "https://github.com/goreleaser/goreleaser/releases/download/v$Version/$zipName"
  $tmpDir = New-Item -ItemType Directory -Path ([System.IO.Path]::Combine([System.IO.Path]::GetTempPath(), [System.Guid]::NewGuid().ToString()))
  $tmpZip = Join-Path $tmpDir 'goreleaser.zip'
  Write-Host "Downloading $downloadUrl" -ForegroundColor Cyan
  try {
    Invoke-WebRequest -Uri $downloadUrl -OutFile $tmpZip -UseBasicParsing -Headers @{ 'User-Agent'='timeguard-setup-script' }
  } catch {
    throw "Download failed: $_"
  }
  try {
    Expand-Archive -Path $tmpZip -DestinationPath $bin -Force
  } catch {
    throw "Extraction failed: $_"
  } finally {
    if (Test-Path $tmpDir) { Remove-Item $tmpDir -Recurse -Force }
  }
  Write-Host "Installed goreleaser v$Version" -ForegroundColor Green
}

if (-not ($env:PATH.Split(';') -contains $bin)) {
  $env:PATH = "$bin;" + $env:PATH
  Write-Host "Updated PATH for current session." -ForegroundColor Yellow
}

try {
  & $exe --version
} catch {
  Write-Warning "goreleaser installed but failed to execute: $_"
  Write-Host "Check Antivirus quarantine or execution policy." -ForegroundColor Yellow
}

Write-Host "Run 'goreleaser release --clean' from the repo root after tagging (git tag -a vX.Y.Z -m vX.Y.Z)." -ForegroundColor Cyan
