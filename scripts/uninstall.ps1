# Remove envy binary and User PATH entry added by install.ps1.
# Usage:
#   irm https://raw.githubusercontent.com/RovierrHQ/envy/main/scripts/uninstall.ps1 | iex
#   .\uninstall.ps1 -Yes

[CmdletBinding()]
param(
    [string]$InstallDir = $(if ($env:ENVY_INSTALL_DIR) { $env:ENVY_INSTALL_DIR } else { Join-Path $env:LOCALAPPDATA 'envy\bin' }),
    [switch]$Yes
)

$ErrorActionPreference = 'Stop'
$bin = Join-Path $InstallDir 'envy.exe'

if (-not $Yes) {
    $ans = Read-Host "Remove envy from $InstallDir and strip User PATH? [y/N]"
    if ($ans -notmatch '^(y|yes)$') {
        Write-Host 'Aborted.'
        return
    }
}

if (Test-Path $bin) {
    Remove-Item -Force $bin
    Write-Host "Removed $bin"
} else {
    Write-Host "No binary at $bin"
}

$userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
if ($userPath) {
    $parts = $userPath -split ';' | Where-Object { $_ -ne '' -and $_ -ne $InstallDir }
    $newPath = ($parts -join ';')
    if ($newPath -ne $userPath) {
        [Environment]::SetEnvironmentVariable('Path', $newPath, 'User')
        Write-Host "Removed from User PATH: $InstallDir"
    }
}

$parent = Split-Path -Parent $InstallDir
if ((Test-Path $InstallDir) -and -not (Get-ChildItem -Force $InstallDir | Select-Object -First 1)) {
    Remove-Item -Force $InstallDir
    Write-Host "Removed empty $InstallDir"
}
$defaultParent = Join-Path $env:LOCALAPPDATA 'envy'
if ($parent -eq $defaultParent -and (Test-Path $parent) -and -not (Get-ChildItem -Force $parent | Select-Object -First 1)) {
    Remove-Item -Force $parent
}

Write-Host 'Done. Open a new terminal so PATH updates apply.'
