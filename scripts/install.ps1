# install.ps1 - PowerShell installation script for SpecLedger CLI on Windows

$ErrorActionPreference = "Stop"

# Default variables
$Version = if ($env:VERSION) { $env:VERSION } else { "latest" }
$DownloadUrl = if ($env:DOWNLOAD_URL) { $env:DOWNLOAD_URL } else { "" }
$InstallDir = if ($env:INSTALL_DIR) { $env:INSTALL_DIR } else { "$env:USERPROFILE\.local\bin" }
$UseAdmin = if ($env:USE_SUDO) { $true } else { $false }

# Create install directory if it doesn't exist
if (-not (Test-Path -Path $InstallDir)) {
    Write-Host "Creating install directory: $InstallDir" -ForegroundColor Yellow
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
}

# Add to PATH if not already there
$PathEnv = [Environment]::GetEnvironmentVariable("Path", "User")
if ($PathEnv -notlike "*$InstallDir*") {
    Write-Host "Adding $InstallDir to PATH" -ForegroundColor Yellow
    [Environment]::SetEnvironmentVariable("Path", "$PathEnv;$InstallDir", "User")
}

# Get latest version from GitHub API
function Get-LatestVersion {
    try {
        $response = Invoke-RestMethod -Uri "https://api.github.com/repos/specledger/specledger/releases/latest" -ErrorAction Stop
        return $response.tag_name -replace '^v', ''
    } catch {
        return "1.0.12"  # Fallback version
    }
}

# Get version if "latest"
if ($Version -eq "latest") {
    $Version = Get-LatestVersion
    Write-Host "Detected latest version: $Version" -ForegroundColor Cyan
}

# Strip 'v' prefix for filename
$FileVersion = $Version -replace '^v', ''
$Arch = if ($env:ARCH) { $env:ARCH } else { "amd64" }

# Get download URL
if ([string]::IsNullOrWhiteSpace($DownloadUrl)) {
    # Add 'v' prefix for URL path
    $UrlVersion = if ($Version -match '^v') { $Version } else { "v$Version" }

    $DownloadUrl = "https://github.com/specledger/specledger/releases/download/$UrlVersion/specledger_${FileVersion}_windows_$Arch.zip"
}

Write-Host "Installing SpecLedger $Version" -ForegroundColor Cyan
Write-Host "Platform: Windows" -ForegroundColor Cyan
Write-Host "Architecture: $Arch" -ForegroundColor Cyan
Write-Host "Install Directory: $InstallDir" -ForegroundColor Cyan
Write-Host "Download URL: $DownloadUrl" -ForegroundColor Cyan
Write-Host ""

# Download file
$tempFile = "$env:TEMP\specledger-download.zip"
Write-Host "Downloading SpecLedger..." -ForegroundColor Yellow

try {
    Invoke-WebRequest -Uri $DownloadUrl -OutFile $tempFile -ErrorAction Stop
} catch {
    Write-Host "Error: Failed to download from $DownloadUrl" -ForegroundColor Red
    Write-Host $_.Exception.Message -ForegroundColor Red
    exit 1
}

# Extract to temp directory
$tempExtract = "$env:TEMP\specledger-extract"
Remove-Item -Path $tempExtract -Recurse -Force -ErrorAction SilentlyContinue
New-Item -ItemType Directory -Path $tempExtract -Force | Out-Null

try {
    Add-Type -AssemblyName System.IO.Compression.FileSystem
    [System.IO.Compression.ZipFile]::ExtractToDirectory($tempFile, $tempExtract)
} catch {
    Write-Host "Error: Failed to extract archive" -ForegroundColor Red
    Write-Host $_.Exception.Message -ForegroundColor Red
    Remove-Item -Path $tempFile -Force
    exit 1
}

# Find the binary - GoReleaser puts sl.exe at the root of the archive
$binaryPath = Join-Path $tempExtract "sl.exe"
if (-not (Test-Path -Path $binaryPath)) {
    # Try alternative paths
    $altPaths = @(
        (Join-Path $tempExtract "specledger_${FileVersion}_windows_$Arch\sl.exe"),
        (Join-Path $tempExtract "specledger_${FileVersion}_windows_amd64\sl.exe")
    )
    foreach ($altPath in $altPaths) {
        if (Test-Path -Path $altPath) {
            $binaryPath = $altPath
            break
        }
    }
}

if (-not (Test-Path -Path $binaryPath)) {
    Write-Host "Error: Binary not found" -ForegroundColor Red
    Write-Host "Contents of extract directory:" -ForegroundColor Yellow
    Get-ChildItem -Path $tempExtract -Recurse | Select-Object FullName
    Remove-Item -Path $tempFile, $tempExtract -Recurse -Force
    exit 1
}

# Copy binary to install directory
$targetBinary = Join-Path $InstallDir "sl.exe"

if (-not $UseAdmin) {
    # Check if we can write to install directory
    try {
        $null = [System.IO.File]::OpenWrite($targetBinary)
        [System.IO.File]::Close($openStream)
    } catch {
        $UseAdmin = $true
        Write-Host "Warning: Admin privileges required for system-wide install" -ForegroundColor Yellow
        Write-Host "This script will attempt to run with elevated privileges..." -ForegroundColor Yellow
    }
}

if ($UseAdmin) {
    try {
        Start-Process powershell -Verb RunAs -ArgumentList "-Command", "Copy-Item -Path `"$binaryPath`" -Destination `"$targetBinary`"; Set-Content -Path `"$targetBinary`" -Value (Get-Content `"$targetBinary`" -Raw) -Encoding UTF8; Write-Host 'Installation complete' -ForegroundColor Green"
        Write-Host "Please close and reopen your terminal" -ForegroundColor Green
    } catch {
        Write-Host "Error: Failed to install with elevated privileges" -ForegroundColor Red
        Write-Host $_.Exception.Message -ForegroundColor Red
        Remove-Item -Path $tempFile, $tempExtract -Recurse -Force
        exit 1
    }
} else {
    Copy-Item -Path $binaryPath -Destination $targetBinary -Force
    Write-Host "✓ Installed SpecLedger $Version to $InstallDir/sl.exe" -ForegroundColor Green
}

# Cleanup
Remove-Item -Path $tempFile, $tempExtract -Recurse -Force

Write-Host ""
Write-Host "Installation complete!" -ForegroundColor Green
Write-Host ""
Write-Host "To verify the installation, run:" -ForegroundColor Cyan
Write-Host "  sl version" -ForegroundColor Yellow

# Verify installation
Start-Sleep -Seconds 1
$env:Path = [System.Environment]::GetEnvironmentVariable("Path","User") + ";" + [System.Environment]::GetEnvironmentVariable("Path","Machine")

if (Get-Command sl -ErrorAction SilentlyContinue) {
    Write-Host ""
    Write-Host "SpecLedger version:" -ForegroundColor Green
    sl version
}
