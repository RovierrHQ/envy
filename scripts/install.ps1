# Install envy from GitHub Releases into %LOCALAPPDATA%\envy\bin and optionally update User PATH.
# Usage:
#   irm https://raw.githubusercontent.com/RovierrHQ/envy/main/scripts/install.ps1 | iex
#   $env:ENVY_VERSION = 'v0.2.0'; irm ... | iex
#   .\install.ps1 -NoPath
#   .\install.ps1 -Version v0.2.0

[CmdletBinding()]
param(
    [string]$Version = $(if ($env:ENVY_VERSION) { $env:ENVY_VERSION } else { 'latest' }),
    [string]$InstallDir = $(if ($env:ENVY_INSTALL_DIR) { $env:ENVY_INSTALL_DIR } else { Join-Path $env:LOCALAPPDATA 'envy\bin' }),
    [switch]$NoPath
)

$ErrorActionPreference = 'Stop'
$Repo = 'RovierrHQ/envy'

function Get-ReleaseTag {
    param([string]$Version)
    if ($Version -eq 'latest') {
        $rel = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest"
        return $rel.tag_name
    }
    $tag = if ($Version.StartsWith('v')) { $Version } else { "v$Version" }
    $rel = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/tags/$tag"
    return $rel.tag_name
}

$tag = Get-ReleaseTag -Version $Version
$ver = $tag.TrimStart('v')
$asset = "envy_${ver}_windows_amd64.zip"
$base = "https://github.com/$Repo/releases/download/$tag"

$dest = Join-Path $InstallDir 'envy.exe'
$oldVer = $null
if (Test-Path $dest) {
    try { $oldVer = (& $dest version 2>$null).Trim() } catch { $oldVer = $null }
}

if ($oldVer) {
    Write-Host "Upgrading envy $oldVer -> $tag (windows/amd64) -> $InstallDir"
} else {
    Write-Host "Installing envy $tag (windows/amd64) -> $InstallDir"
}

$tmp = Join-Path ([System.IO.Path]::GetTempPath()) ("envy-install-" + [guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $tmp | Out-Null
try {
    $zipPath = Join-Path $tmp $asset
    $sumPath = Join-Path $tmp 'checksums.txt'
    Invoke-WebRequest -Uri "$base/$asset" -OutFile $zipPath -UseBasicParsing
    Invoke-WebRequest -Uri "$base/checksums.txt" -OutFile $sumPath -UseBasicParsing

    $expected = $null
    Get-Content $sumPath | ForEach-Object {
        $parts = $_ -split '\s+', 2
        if ($parts.Count -ge 2 -and $parts[1].Trim() -eq $asset) {
            $expected = $parts[0].Trim().ToLowerInvariant()
        }
    }
    if (-not $expected) {
        throw "checksum not found for $asset"
    }

    $actual = (Get-FileHash -Path $zipPath -Algorithm SHA256).Hash.ToLowerInvariant()
    if ($actual -ne $expected) {
        throw "checksum mismatch for $asset`n  expected: $expected`n  actual:   $actual"
    }

    Expand-Archive -Path $zipPath -DestinationPath $tmp -Force
    $bin = Join-Path $tmp 'envy.exe'
    if (-not (Test-Path $bin)) {
        throw 'archive missing envy.exe'
    }

    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    Copy-Item -Path $bin -Destination $dest -Force

    if (-not $NoPath) {
        $userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
        if (-not $userPath) { $userPath = '' }
        $parts = $userPath -split ';' | Where-Object { $_ -ne '' }
        if ($parts -notcontains $InstallDir) {
            $newPath = if ($userPath.Trim() -eq '') { $InstallDir } else { "$userPath;$InstallDir" }
            [Environment]::SetEnvironmentVariable('Path', $newPath, 'User')
            Write-Host "Added to User PATH: $InstallDir"
        }
        if ($env:Path -notlike "*$InstallDir*") {
            $env:Path = "$InstallDir;$env:Path"
        }
    }

    $newVer = $ver
    try { $newVer = (& $dest version 2>$null).Trim() } catch { }
    if ($oldVer) {
        Write-Host "Upgraded $dest ($oldVer -> $newVer)"
    } else {
        Write-Host "Installed $dest ($newVer)"
    }
    Write-Host 'Done. Open a new terminal if envy is not found on PATH.'
    Write-Host 'Upgrade anytime by re-running this script. Uninstall: scripts/uninstall.ps1'
}
finally {
    Remove-Item -Recurse -Force $tmp -ErrorAction SilentlyContinue
}
