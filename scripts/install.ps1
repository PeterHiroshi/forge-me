# cfmon installer for Windows
# Usage: iwr -useb https://raw.githubusercontent.com/PeterHiroshi/cfmon/main/scripts/install.ps1 | iex

$ErrorActionPreference = 'Stop'

$Repo = "PeterHiroshi/cfmon"
$BinaryName = "cfmon.exe"
$InstallDir = "$env:LOCALAPPDATA\Programs\cfmon"

# Detect architecture
$Arch = if ([Environment]::Is64BitOperatingSystem) { "x86_64" } else { "i386" }

# Get latest release version
Write-Host "Fetching latest release..."
$LatestRelease = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest"
$Version = $LatestRelease.tag_name

if (-not $Version) {
    Write-Error "Failed to get latest version"
    exit 1
}

Write-Host "Installing cfmon $Version for Windows $Arch..."

# Download URLs
$Archive = "${BinaryName}_${Version}_Windows_${Arch}.zip"
$DownloadUrl = "https://github.com/$Repo/releases/download/$Version/$Archive"
$ChecksumUrl = "https://github.com/$Repo/releases/download/$Version/checksums.txt"

# Create temp directory
$TmpDir = [System.IO.Path]::GetTempPath() + [System.Guid]::NewGuid().ToString()
New-Item -ItemType Directory -Path $TmpDir | Out-Null

try {
    # Download archive
    Write-Host "Downloading $Archive..."
    $ArchivePath = Join-Path $TmpDir $Archive
    Invoke-WebRequest -Uri $DownloadUrl -OutFile $ArchivePath

    # Download and verify checksum
    Write-Host "Verifying checksum..."
    $ChecksumPath = Join-Path $TmpDir "checksums.txt"
    Invoke-WebRequest -Uri $ChecksumUrl -OutFile $ChecksumPath

    $ExpectedChecksum = (Get-Content $ChecksumPath | Select-String $Archive).Line.Split()[0]
    $ActualChecksum = (Get-FileHash -Path $ArchivePath -Algorithm SHA256).Hash.ToLower()

    if ($ExpectedChecksum -ne $ActualChecksum) {
        Write-Error "Checksum verification failed!"
        exit 1
    }

    # Extract
    Write-Host "Extracting..."
    Expand-Archive -Path $ArchivePath -DestinationPath $TmpDir -Force

    # Create install directory
    if (-not (Test-Path $InstallDir)) {
        New-Item -ItemType Directory -Path $InstallDir | Out-Null
    }

    # Install
    Write-Host "Installing to $InstallDir..."
    $SourceBinary = Join-Path $TmpDir $BinaryName
    $DestBinary = Join-Path $InstallDir $BinaryName
    Copy-Item -Path $SourceBinary -Destination $DestBinary -Force

    # Add to PATH if not already there
    $UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($UserPath -notlike "*$InstallDir*") {
        Write-Host "Adding $InstallDir to PATH..."
        [Environment]::SetEnvironmentVariable("Path", "$UserPath;$InstallDir", "User")
        $env:Path += ";$InstallDir"
    }

    Write-Host "✓ cfmon installed successfully!" -ForegroundColor Green
    Write-Host "Run 'cfmon --help' to get started"
    Write-Host ""
    Write-Host "Note: You may need to restart your terminal for PATH changes to take effect"
    Write-Host ""
    Write-Host "Alternative: Install via Scoop (if available):"
    Write-Host "  scoop bucket add cfmon https://github.com/PeterHiroshi/scoop-cfmon"
    Write-Host "  scoop install cfmon"

} finally {
    # Cleanup
    Remove-Item -Path $TmpDir -Recurse -Force -ErrorAction SilentlyContinue
}
