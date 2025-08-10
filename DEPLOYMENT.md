# npm-console Deployment Guide

This document provides comprehensive instructions for deploying npm-console in various environments.

## Table of Contents

- [Quick Installation](#quick-installation)
- [Manual Installation](#manual-installation)
- [Docker Deployment](#docker-deployment)
- [Building from Source](#building-from-source)
- [Configuration](#configuration)
- [Systemd Service](#systemd-service)
- [Windows Service](#windows-service)
- [Troubleshooting](#troubleshooting)

## Quick Installation

### Linux/macOS

```bash
# Download and run the installation script
curl -fsSL https://raw.githubusercontent.com/npm-console/npm-console/main/scripts/install.sh | bash

# Or with wget
wget -qO- https://raw.githubusercontent.com/npm-console/npm-console/main/scripts/install.sh | bash
```

### Windows

```powershell
# Download and run the PowerShell installation script
iwr -useb https://raw.githubusercontent.com/npm-console/npm-console/main/scripts/install.ps1 | iex

# Or download and run manually
Invoke-WebRequest -Uri "https://raw.githubusercontent.com/npm-console/npm-console/main/scripts/install.ps1" -OutFile "install.ps1"
.\install.ps1
```

## Manual Installation

### Download Pre-built Binaries

1. Go to the [Releases page](https://github.com/npm-console/npm-console/releases)
2. Download the appropriate binary for your platform:
   - `npm-console-linux-amd64` - Linux 64-bit
   - `npm-console-linux-arm64` - Linux ARM64
   - `npm-console-darwin-amd64` - macOS Intel
   - `npm-console-darwin-arm64` - macOS Apple Silicon
   - `npm-console-windows-amd64.exe` - Windows 64-bit
   - `npm-console-windows-386.exe` - Windows 32-bit

3. Make the binary executable (Linux/macOS):
   ```bash
   chmod +x npm-console-*
   ```

4. Move to a directory in your PATH:
   ```bash
   # Linux/macOS
   sudo mv npm-console-* /usr/local/bin/npm-console
   
   # Windows (as Administrator)
   move npm-console-*.exe C:\Windows\System32\npm-console.exe
   ```

### Verify Installation

```bash
npm-console version
npm-console --help
```

## Docker Deployment

### Using Docker Compose (Recommended)

1. Create a `docker-compose.yml` file:
   ```yaml
   version: '3.8'
   services:
     npm-console:
       image: npm-console/npm-console:latest
       ports:
         - "8080:8080"
       volumes:
         - ./projects:/app/projects:ro
         - npm-console-cache:/home/appuser/.cache
         - npm-console-config:/home/appuser/.config
       environment:
         - NPM_CONSOLE_HOST=0.0.0.0
         - NPM_CONSOLE_PORT=8080
       restart: unless-stopped
   
   volumes:
     npm-console-cache:
     npm-console-config:
   ```

2. Start the service:
   ```bash
   docker-compose up -d
   ```

3. Access the web interface at `http://localhost:8080`

### Using Docker Run

```bash
docker run -d \
  --name npm-console \
  -p 8080:8080 \
  -v $(pwd)/projects:/app/projects:ro \
  -v npm-console-cache:/home/appuser/.cache \
  -v npm-console-config:/home/appuser/.config \
  npm-console/npm-console:latest
```

### Building Docker Image

```bash
# Clone the repository
git clone https://github.com/npm-console/npm-console.git
cd npm-console

# Build the image
docker build -t npm-console:local .

# Run the container
docker run -p 8080:8080 npm-console:local
```

## Building from Source

### Prerequisites

- Go 1.21 or later
- Git
- Node.js 18+ (for web assets)

### Build Steps

1. Clone the repository:
   ```bash
   git clone https://github.com/npm-console/npm-console.git
   cd npm-console
   ```

2. Build using the build script:
   ```bash
   # Linux/macOS
   ./scripts/build.sh
   
   # Windows
   .\scripts\build.bat
   ```

3. Or build manually:
   ```bash
   go build -ldflags="-s -w" -o npm-console .
   ```

### Cross-platform Building

```bash
# Build for all platforms
GOOS=linux GOARCH=amd64 go build -o npm-console-linux-amd64 .
GOOS=windows GOARCH=amd64 go build -o npm-console-windows-amd64.exe .
GOOS=darwin GOARCH=amd64 go build -o npm-console-darwin-amd64 .
```

## Configuration

### Environment Variables

- `NPM_CONSOLE_HOST` - Web server host (default: localhost)
- `NPM_CONSOLE_PORT` - Web server port (default: 8080)
- `NPM_CONSOLE_LOG_LEVEL` - Log level (debug, info, warn, error)
- `NPM_CONSOLE_CONFIG_DIR` - Configuration directory
- `NPM_CONSOLE_CACHE_DIR` - Cache directory

### Configuration File

Create `~/.config/npm-console/config.yaml`:

```yaml
web:
  host: "0.0.0.0"
  port: 8080
  enabled: true
  cors:
    enabled: true
    allowed_origins: ["*"]
    allowed_methods: ["GET", "POST", "PUT", "DELETE"]
    allowed_headers: ["*"]

logger:
  level: "info"
  format: "json"
  output: "stdout"

cache:
  enabled: true
  ttl: "1h"
```

## Systemd Service

Create `/etc/systemd/system/npm-console.service`:

```ini
[Unit]
Description=npm-console - Unified Package Manager Console
After=network.target

[Service]
Type=simple
User=npm-console
Group=npm-console
ExecStart=/usr/local/bin/npm-console web --host 0.0.0.0 --port 8080
Restart=always
RestartSec=5
Environment=NPM_CONSOLE_LOG_LEVEL=info

# Security settings
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/npm-console

[Install]
WantedBy=multi-user.target
```

Enable and start the service:

```bash
sudo systemctl daemon-reload
sudo systemctl enable npm-console
sudo systemctl start npm-console
sudo systemctl status npm-console
```

## Windows Service

### Using NSSM (Non-Sucking Service Manager)

1. Download NSSM from https://nssm.cc/download
2. Install the service:
   ```cmd
   nssm install npm-console "C:\Program Files\npm-console\npm-console.exe"
   nssm set npm-console Arguments "web --host 0.0.0.0 --port 8080"
   nssm set npm-console DisplayName "npm-console"
   nssm set npm-console Description "Unified Package Manager Console"
   nssm start npm-console
   ```

### Using PowerShell (Windows 10+)

```powershell
# Create a scheduled task that runs at startup
$Action = New-ScheduledTaskAction -Execute "C:\Program Files\npm-console\npm-console.exe" -Argument "web"
$Trigger = New-ScheduledTaskTrigger -AtStartup
$Settings = New-ScheduledTaskSettingsSet -AllowStartIfOnBatteries -DontStopIfGoingOnBatteries -StartWhenAvailable
$Principal = New-ScheduledTaskPrincipal -UserId "SYSTEM" -LogonType ServiceAccount

Register-ScheduledTask -TaskName "npm-console" -Action $Action -Trigger $Trigger -Settings $Settings -Principal $Principal
```

## Reverse Proxy Setup

### Nginx

```nginx
server {
    listen 80;
    server_name npm-console.example.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### Apache

```apache
<VirtualHost *:80>
    ServerName npm-console.example.com
    
    ProxyPreserveHost On
    ProxyPass / http://localhost:8080/
    ProxyPassReverse / http://localhost:8080/
</VirtualHost>
```

## Troubleshooting

### Common Issues

1. **Permission denied when accessing package managers**
   - Ensure npm-console runs with appropriate user permissions
   - Check that package managers are installed and accessible

2. **Web interface not accessible**
   - Check if the port is open: `netstat -tlnp | grep 8080`
   - Verify firewall settings
   - Check logs: `npm-console web --verbose`

3. **Package managers not detected**
   - Ensure package managers are installed and in PATH
   - Check with: `npm-console packages list --verbose`

4. **Cache operations fail**
   - Verify write permissions to cache directories
   - Check disk space availability

### Logs

- **Linux/macOS**: `~/.local/share/npm-console/logs/`
- **Windows**: `%LOCALAPPDATA%\npm-console\logs\`

### Debug Mode

Run with verbose logging:
```bash
npm-console --verbose [command]
```

### Health Check

```bash
# Check if all components are working
npm-console version
npm-console cache list
npm-console packages list --global
```

## Security Considerations

1. **Network Security**
   - Use HTTPS in production
   - Restrict access with firewall rules
   - Consider VPN for remote access

2. **File Permissions**
   - Run with minimal required permissions
   - Secure configuration files (600 permissions)
   - Regular security updates

3. **Container Security**
   - Use non-root user in containers
   - Scan images for vulnerabilities
   - Keep base images updated

## Performance Tuning

1. **Resource Limits**
   - Set appropriate memory limits
   - Configure CPU limits for containers
   - Monitor resource usage

2. **Caching**
   - Enable caching for better performance
   - Configure appropriate TTL values
   - Monitor cache hit rates

3. **Concurrent Operations**
   - Adjust worker pool sizes
   - Configure timeout values
   - Monitor response times

For more information and updates, visit the [GitHub repository](https://github.com/npm-console/npm-console).
