# npm-console Installation Script for Windows
# Downloads and installs the latest version of npm-console

param(
    [string]$InstallDir = "$env:LOCALAPPDATA\npm-console",
    [switch]$AddToPath = $true,
    [switch]$Force = $false
)

# Configuration
$Repo = "npm-console/npm-console"
$BinaryName = "npm-console.exe"

Write-Host "🚀 npm-console Installation Script" -ForegroundColor Blue
Write-Host "===================================" -ForegroundColor Blue
Write-Host ""

# Detect architecture
$Arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }
Write-Host "Detected platform: windows/$Arch" -ForegroundColor Yellow

# Create install directory
if (-not (Test-Path $InstallDir)) {
    Write-Host "📁 Creating install directory: $InstallDir" -ForegroundColor Yellow
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
}

# Check if already installed
$ExistingPath = Join-Path $InstallDir $BinaryName
if ((Test-Path $ExistingPath) -and -not $Force) {
    $ExistingVersion = & $ExistingPath version --short 2>$null
    if ($ExistingVersion) {
        Write-Host "⚠️  npm-console $ExistingVersion is already installed" -ForegroundColor Yellow
        Write-Host "   Use -Force to reinstall" -ForegroundColor Yellow
        exit 0
    }
}

# Get latest release information
Write-Host "🔍 Fetching latest release information..." -ForegroundColor Yellow

try {
    $LatestRelease = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest"
    $Version = $LatestRelease.tag_name
    $Asset = $LatestRelease.assets | Where-Object { $_.name -like "*windows-$Arch.exe" }
    
    if (-not $Asset) {
        Write-Host "❌ Could not find release for windows/$Arch" -ForegroundColor Red
        exit 1
    }
    
    $DownloadUrl = $Asset.browser_download_url
    Write-Host "✅ Found version: $Version" -ForegroundColor Green
    Write-Host "📥 Download URL: $DownloadUrl" -ForegroundColor Yellow
}
catch {
    Write-Host "❌ Failed to fetch release information: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# Download binary
$TempFile = Join-Path $env:TEMP "npm-console-temp.exe"
Write-Host "⬇️  Downloading npm-console..." -ForegroundColor Yellow

try {
    Invoke-WebRequest -Uri $DownloadUrl -OutFile $TempFile -UseBasicParsing
    Write-Host "✅ Download completed" -ForegroundColor Green
}
catch {
    Write-Host "❌ Failed to download npm-console: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# Verify binary
Write-Host "🔍 Verifying binary..." -ForegroundColor Yellow
try {
    $TestResult = & $TempFile version 2>$null
    if (-not $TestResult) {
        throw "Binary verification failed"
    }
    Write-Host "✅ Binary verification passed" -ForegroundColor Green
}
catch {
    Write-Host "❌ Downloaded binary is not working: $($_.Exception.Message)" -ForegroundColor Red
    Remove-Item $TempFile -ErrorAction SilentlyContinue
    exit 1
}

# Install binary
Write-Host "📦 Installing to $InstallDir..." -ForegroundColor Yellow
try {
    Move-Item $TempFile $ExistingPath -Force
    Write-Host "✅ Binary installed successfully" -ForegroundColor Green
}
catch {
    Write-Host "❌ Failed to install binary: $($_.Exception.Message)" -ForegroundColor Red
    Remove-Item $TempFile -ErrorAction SilentlyContinue
    exit 1
}

# Add to PATH
if ($AddToPath) {
    $CurrentPath = [Environment]::GetEnvironmentVariable("PATH", "User")
    if ($CurrentPath -notlike "*$InstallDir*") {
        Write-Host "📝 Adding to PATH..." -ForegroundColor Yellow
        $NewPath = "$InstallDir;$CurrentPath"
        [Environment]::SetEnvironmentVariable("PATH", $NewPath, "User")
        Write-Host "✅ Added to PATH (restart shell to take effect)" -ForegroundColor Green
        
        # Update current session PATH
        $env:PATH = "$InstallDir;$env:PATH"
    } else {
        Write-Host "✅ Already in PATH" -ForegroundColor Green
    }
}

# Verify installation
try {
    $InstalledVersion = & $BinaryName version --short 2>$null
    if ($InstalledVersion) {
        Write-Host "✅ npm-console $InstalledVersion installed successfully!" -ForegroundColor Green
    } else {
        Write-Host "⚠️  npm-console installed but not accessible via PATH" -ForegroundColor Yellow
        Write-Host "   Binary location: $ExistingPath" -ForegroundColor Yellow
    }
}
catch {
    Write-Host "⚠️  npm-console installed but not accessible via PATH" -ForegroundColor Yellow
    Write-Host "   Binary location: $ExistingPath" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "🎉 Installation completed!" -ForegroundColor Blue
Write-Host ""
Write-Host "Quick start:" -ForegroundColor Yellow
Write-Host "  npm-console --help          # Show help"
Write-Host "  npm-console cache list      # List cache information"
Write-Host "  npm-console packages list   # List installed packages"
Write-Host "  npm-console web             # Start web interface"
Write-Host ""
Write-Host "For more information, visit: https://github.com/$Repo" -ForegroundColor Blue

# Create desktop shortcut (optional)
$CreateShortcut = Read-Host "Create desktop shortcut? (y/N)"
if ($CreateShortcut -eq "y" -or $CreateShortcut -eq "Y") {
    try {
        $WshShell = New-Object -comObject WScript.Shell
        $Shortcut = $WshShell.CreateShortcut("$env:USERPROFILE\Desktop\npm-console.lnk")
        $Shortcut.TargetPath = $ExistingPath
        $Shortcut.Arguments = "web"
        $Shortcut.Description = "npm-console - Unified Package Manager Console"
        $Shortcut.WorkingDirectory = $InstallDir
        $Shortcut.Save()
        Write-Host "✅ Desktop shortcut created" -ForegroundColor Green
    }
    catch {
        Write-Host "⚠️  Failed to create desktop shortcut: $($_.Exception.Message)" -ForegroundColor Yellow
    }
}
