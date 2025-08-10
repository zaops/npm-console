@echo off
setlocal enabledelayedexpansion

REM npm-console Build Script for Windows
REM Builds binaries for multiple platforms

echo.
echo 🚀 Building npm-console
echo ========================

REM Configuration
set APP_NAME=npm-console
set OUTPUT_DIR=dist
set BIN_DIR=%OUTPUT_DIR%\bin

REM Get version information
for /f "tokens=*" %%i in ('git describe --tags --always --dirty 2^>nul') do set VERSION=%%i
if "%VERSION%"=="" set VERSION=dev

for /f "tokens=*" %%i in ('git rev-parse --short HEAD 2^>nul') do set GIT_COMMIT=%%i
if "%GIT_COMMIT%"=="" set GIT_COMMIT=unknown

REM Get current timestamp
for /f "tokens=2 delims==" %%a in ('wmic OS Get localdatetime /value') do set "dt=%%a"
set "YY=%dt:~2,2%" & set "YYYY=%dt:~0,4%" & set "MM=%dt:~4,2%" & set "DD=%dt:~6,2%"
set "HH=%dt:~8,2%" & set "Min=%dt:~10,2%" & set "Sec=%dt:~12,2%"
set BUILD_TIME=%YYYY%-%MM%-%DD%T%HH%:%Min%:%Sec%Z

REM Build flags
set LDFLAGS=-s -w -X main.Version=%VERSION% -X main.BuildTime=%BUILD_TIME% -X main.GitCommit=%GIT_COMMIT%

echo Version: %VERSION%
echo Build Time: %BUILD_TIME%
echo Git Commit: %GIT_COMMIT%
echo.

REM Clean previous builds
echo 🧹 Cleaning previous builds...
if exist "%OUTPUT_DIR%" rmdir /s /q "%OUTPUT_DIR%"
mkdir "%BIN_DIR%"

REM Build for Windows platforms
echo.
echo 📦 Building for Windows AMD64...
set GOOS=windows
set GOARCH=amd64
go build -ldflags="%LDFLAGS%" -o "%BIN_DIR%\%APP_NAME%-windows-amd64.exe" .
if errorlevel 1 (
    echo ❌ Failed to build for Windows AMD64
    exit /b 1
)
echo ✅ Built %APP_NAME%-windows-amd64.exe

echo.
echo 📦 Building for Windows 386...
set GOOS=windows
set GOARCH=386
go build -ldflags="%LDFLAGS%" -o "%BIN_DIR%\%APP_NAME%-windows-386.exe" .
if errorlevel 1 (
    echo ❌ Failed to build for Windows 386
    exit /b 1
)
echo ✅ Built %APP_NAME%-windows-386.exe

REM Build for Linux platforms
echo.
echo 📦 Building for Linux AMD64...
set GOOS=linux
set GOARCH=amd64
go build -ldflags="%LDFLAGS%" -o "%BIN_DIR%\%APP_NAME%-linux-amd64" .
if errorlevel 1 (
    echo ❌ Failed to build for Linux AMD64
    exit /b 1
)
echo ✅ Built %APP_NAME%-linux-amd64

echo.
echo 📦 Building for Linux 386...
set GOOS=linux
set GOARCH=386
go build -ldflags="%LDFLAGS%" -o "%BIN_DIR%\%APP_NAME%-linux-386" .
if errorlevel 1 (
    echo ❌ Failed to build for Linux 386
    exit /b 1
)
echo ✅ Built %APP_NAME%-linux-386

echo.
echo 📦 Building for Linux ARM64...
set GOOS=linux
set GOARCH=arm64
go build -ldflags="%LDFLAGS%" -o "%BIN_DIR%\%APP_NAME%-linux-arm64" .
if errorlevel 1 (
    echo ❌ Failed to build for Linux ARM64
    exit /b 1
)
echo ✅ Built %APP_NAME%-linux-arm64

REM Build for macOS platforms
echo.
echo 📦 Building for macOS AMD64...
set GOOS=darwin
set GOARCH=amd64
go build -ldflags="%LDFLAGS%" -o "%BIN_DIR%\%APP_NAME%-darwin-amd64" .
if errorlevel 1 (
    echo ❌ Failed to build for macOS AMD64
    exit /b 1
)
echo ✅ Built %APP_NAME%-darwin-amd64

echo.
echo 📦 Building for macOS ARM64...
set GOOS=darwin
set GOARCH=arm64
go build -ldflags="%LDFLAGS%" -o "%BIN_DIR%\%APP_NAME%-darwin-arm64" .
if errorlevel 1 (
    echo ❌ Failed to build for macOS ARM64
    exit /b 1
)
echo ✅ Built %APP_NAME%-darwin-arm64

echo.
echo 🎉 Build completed successfully!
echo 📁 Binaries are available in: %BIN_DIR%

REM List built binaries
echo.
echo 📋 Built binaries:
dir /b "%BIN_DIR%"

REM Generate checksums (if certutil is available)
echo.
echo 🔐 Generating checksums...
cd /d "%BIN_DIR%"
if exist checksums.txt del checksums.txt
for %%f in (*) do (
    if not "%%f"=="checksums.txt" (
        for /f "tokens=*" %%h in ('certutil -hashfile "%%f" SHA256 ^| find /v ":" ^| find /v "CertUtil"') do (
            echo %%h *%%f >> checksums.txt
        )
    )
)
echo ✅ Checksums saved to checksums.txt

echo.
echo 🚀 Build process completed!
pause
