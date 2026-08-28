# Builds castctl and installs it to a per-user dir on PATH (Windows).
# Usage:  powershell -ExecutionPolicy Bypass -File install.ps1
param(
    [string]$Version = "0.1.0"
)

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $root

# Ensure go is reachable even if the shell hasn't picked it up post-install.
if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    $goBin = "C:\Program Files\Go\bin"
    if (Test-Path (Join-Path $goBin "go.exe")) { $env:Path = "$goBin;$env:Path" }
    else { throw "Go not found. Install from https://go.dev/dl/ or 'winget install GoLang.Go'." }
}

$destDir = Join-Path $env:LOCALAPPDATA "Programs\castctl"
New-Item -ItemType Directory -Force -Path $destDir | Out-Null
$dest = Join-Path $destDir "castctl.exe"

Write-Host "Building castctl $Version ..."
$env:CGO_ENABLED = "0"
go build -ldflags "-s -w -X main.version=$Version" -o $dest ./
Write-Host "Installed: $dest"

# Add to the user PATH if it isn't already there.
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*$destDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$userPath;$destDir", "User")
    Write-Host "Added $destDir to your user PATH. Open a new terminal to use 'castctl'."
} else {
    Write-Host "$destDir already on PATH."
}

& $dest --version
