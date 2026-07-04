$ErrorActionPreference = "Stop"

$repo = "ipriverdev/cli"
$binary = "ipriver"

$arch = if ([Environment]::Is64BitOperatingSystem) {
    switch ($env:PROCESSOR_ARCHITECTURE) {
        "ARM64" { "arm64" }
        default { "amd64" }
    }
} else {
    Write-Error "Unsupported: 32-bit Windows is not supported"
    exit 1
}

if ($arch -eq "arm64") {
    Write-Error "Windows ARM64 builds are not yet available. Use Windows x64 emulation or build from source."
    exit 1
}

$release = (Invoke-RestMethod "https://api.github.com/repos/$repo/releases/latest").tag_name
if (-not $release) {
    Write-Error "Could not determine latest release"
    exit 1
}

$version = $release.TrimStart("v")
$archive = "${binary}_${version}_windows_${arch}.zip"
$url = "https://github.com/$repo/releases/download/$release/$archive"
$checksumsUrl = "https://github.com/$repo/releases/download/$release/checksums.txt"

Write-Host "Installing $binary $release (windows/$arch)..."

$tmp = Join-Path ([System.IO.Path]::GetTempPath()) "ipriver-install"
if (Test-Path $tmp) { Remove-Item $tmp -Recurse -Force }
New-Item -ItemType Directory -Path $tmp | Out-Null

$zipPath = Join-Path $tmp $archive
Invoke-WebRequest -Uri $url -OutFile $zipPath

$checksumsPath = Join-Path $tmp "checksums.txt"
Invoke-WebRequest -Uri $checksumsUrl -OutFile $checksumsPath

$expected = (Get-Content $checksumsPath | Where-Object { $_ -match $archive }) -replace '\s+.*$', ''
if (-not $expected) {
    Write-Error "No checksum found for $archive"
    exit 1
}

$actual = (Get-FileHash -Path $zipPath -Algorithm SHA256).Hash.ToLower()
if ($actual -ne $expected) {
    Write-Error "Checksum mismatch for ${archive}: expected $expected, got $actual"
    exit 1
}

Expand-Archive -Path $zipPath -DestinationPath $tmp -Force

$installDir = Join-Path $env:LOCALAPPDATA "ipriver"
if (-not (Test-Path $installDir)) {
    New-Item -ItemType Directory -Path $installDir | Out-Null
}

Copy-Item (Join-Path $tmp "$binary.exe") (Join-Path $installDir "$binary.exe") -Force
Copy-Item (Join-Path $installDir "$binary.exe") (Join-Path $installDir "ipr.exe") -Force
Remove-Item $tmp -Recurse -Force

$currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($currentPath -notlike "*$installDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$currentPath;$installDir", "User")
    Write-Host "Added $installDir to your PATH (restart your terminal to use it)"
}

Write-Host "Installed $binary to $installDir\$binary.exe (also available as 'ipr')"
