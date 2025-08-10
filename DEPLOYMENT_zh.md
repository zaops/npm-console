# npm-console 部署指南

本文档提供了在各种环境中部署 npm-console 的详细说明。

## 目录

- [快速安装](#快速安装)
- [手动安装](#手动安装)
- [Docker 部署](#docker-部署)
- [从源码构建](#从源码构建)
- [配置](#配置)
- [系统服务](#系统服务)
- [Windows 服务](#windows-服务)
- [故障排除](#故障排除)

## 快速安装

### Linux/macOS

```bash
# 下载并运行安装脚本
curl -fsSL https://raw.githubusercontent.com/npm-console/npm-console/main/scripts/install.sh | bash

# 或使用 wget
wget -qO- https://raw.githubusercontent.com/npm-console/npm-console/main/scripts/install.sh | bash
```

### Windows

```powershell
# 下载并运行 PowerShell 安装脚本
iwr -useb https://raw.githubusercontent.com/npm-console/npm-console/main/scripts/install.ps1 | iex

# 或手动下载运行
Invoke-WebRequest -Uri "https://raw.githubusercontent.com/npm-console/npm-console/main/scripts/install.ps1" -OutFile "install.ps1"
.\install.ps1
```

## 手动安装

### 下载预构建二进制文件

1. 访问 [Releases 页面](https://github.com/npm-console/npm-console/releases)
2. 下载适合您平台的二进制文件：
   - `npm-console-linux-amd64` - Linux 64位
   - `npm-console-linux-arm64` - Linux ARM64
   - `npm-console-darwin-amd64` - macOS Intel
   - `npm-console-darwin-arm64` - macOS Apple Silicon
   - `npm-console-windows-amd64.exe` - Windows 64位
   - `npm-console-windows-386.exe` - Windows 32位

3. 使二进制文件可执行 (Linux/macOS)：
   ```bash
   chmod +x npm-console-*
   ```

4. 移动到 PATH 目录中：
   ```bash
   # Linux/macOS
   sudo mv npm-console-* /usr/local/bin/npm-console
   
   # Windows (管理员权限)
   move npm-console-*.exe C:\Windows\System32\npm-console.exe
   ```

### 验证安装

```bash
npm-console version
npm-console --help
```

## Docker 部署

### 使用 Docker Compose (推荐)

1. 创建 `docker-compose.yml` 文件：
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

2. 启动服务：
   ```bash
   docker-compose up -d
   ```

3. 在 `http://localhost:8080` 访问 Web 界面

### 使用 Docker Run

```bash
docker run -d \
  --name npm-console \
  -p 8080:8080 \
  -v $(pwd)/projects:/app/projects:ro \
  -v npm-console-cache:/home/appuser/.cache \
  -v npm-console-config:/home/appuser/.config \
  npm-console/npm-console:latest
```

### 构建 Docker 镜像

```bash
# 克隆仓库
git clone https://github.com/npm-console/npm-console.git
cd npm-console

# 构建镜像
docker build -t npm-console:local .

# 运行容器
docker run -p 8080:8080 npm-console:local
```

## 从源码构建

### 环境要求

- Go 1.21 或更高版本
- Git
- Node.js 18+ (用于 Web 资源)

### 构建步骤

1. 克隆仓库：
   ```bash
   git clone https://github.com/npm-console/npm-console.git
   cd npm-console
   ```

2. 使用构建脚本：
   ```bash
   # Linux/macOS
   ./scripts/build.sh
   
   # Windows
   .\scripts\build.bat
   ```

3. 或手动构建：
   ```bash
   go build -ldflags="-s -w" -o npm-console .
   ```

### 跨平台构建

```bash
# 为所有平台构建
GOOS=linux GOARCH=amd64 go build -o npm-console-linux-amd64 .
GOOS=windows GOARCH=amd64 go build -o npm-console-windows-amd64.exe .
GOOS=darwin GOARCH=amd64 go build -o npm-console-darwin-amd64 .
```

## 配置

### 环境变量

- `NPM_CONSOLE_HOST` - Web 服务器主机 (默认: localhost)
- `NPM_CONSOLE_PORT` - Web 服务器端口 (默认: 8080)
- `NPM_CONSOLE_LOG_LEVEL` - 日志级别 (debug, info, warn, error)
- `NPM_CONSOLE_CONFIG_DIR` - 配置目录
- `NPM_CONSOLE_CACHE_DIR` - 缓存目录

### 配置文件

创建 `~/.config/npm-console/config.yaml`：

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

## 系统服务

### Linux (systemd)

创建 `/etc/systemd/system/npm-console.service`：

```ini
[Unit]
Description=npm-console - 统一包管理器控制台
After=network.target

[Service]
Type=simple
User=npm-console
Group=npm-console
ExecStart=/usr/local/bin/npm-console web --host 0.0.0.0 --port 8080
Restart=always
RestartSec=5
Environment=NPM_CONSOLE_LOG_LEVEL=info

# 安全设置
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/npm-console

[Install]
WantedBy=multi-user.target
```

启用并启动服务：

```bash
sudo systemctl daemon-reload
sudo systemctl enable npm-console
sudo systemctl start npm-console
sudo systemctl status npm-console
```

## Windows 服务

### 使用 NSSM (Non-Sucking Service Manager)

1. 从 https://nssm.cc/download 下载 NSSM
2. 安装服务：
   ```cmd
   nssm install npm-console "C:\Program Files\npm-console\npm-console.exe"
   nssm set npm-console Arguments "web --host 0.0.0.0 --port 8080"
   nssm set npm-console DisplayName "npm-console"
   nssm set npm-console Description "统一包管理器控制台"
   nssm start npm-console
   ```

### 使用 PowerShell (Windows 10+)

```powershell
# 创建启动时运行的计划任务
$Action = New-ScheduledTaskAction -Execute "C:\Program Files\npm-console\npm-console.exe" -Argument "web"
$Trigger = New-ScheduledTaskTrigger -AtStartup
$Settings = New-ScheduledTaskSettingsSet -AllowStartIfOnBatteries -DontStopIfGoingOnBatteries -StartWhenAvailable
$Principal = New-ScheduledTaskPrincipal -UserId "SYSTEM" -LogonType ServiceAccount

Register-ScheduledTask -TaskName "npm-console" -Action $Action -Trigger $Trigger -Settings $Settings -Principal $Principal
```

## 反向代理设置

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

## 故障排除

### 常见问题

1. **访问包管理器时权限被拒绝**
   - 确保 npm-console 以适当的用户权限运行
   - 检查包管理器是否已安装且可访问

2. **Web 界面无法访问**
   - 检查端口是否开放：`netstat -tlnp | grep 8080`
   - 验证防火墙设置
   - 查看日志：`npm-console web --verbose`

3. **未检测到包管理器**
   - 确保包管理器已安装且在 PATH 中
   - 检查：`npm-console packages list --verbose`

4. **缓存操作失败**
   - 验证对缓存目录的写权限
   - 检查磁盘空间可用性

### 日志

- **Linux/macOS**: `~/.local/share/npm-console/logs/`
- **Windows**: `%LOCALAPPDATA%\npm-console\logs\`

### 调试模式

使用详细日志运行：
```bash
npm-console --verbose [command]
```

### 健康检查

```bash
# 检查所有组件是否正常工作
npm-console version
npm-console cache list
npm-console packages list --global
```

## 安全考虑

1. **网络安全**
   - 在生产环境中使用 HTTPS
   - 使用防火墙规则限制访问
   - 考虑使用 VPN 进行远程访问

2. **文件权限**
   - 以最小必需权限运行
   - 保护配置文件 (600 权限)
   - 定期安全更新

3. **容器安全**
   - 在容器中使用非 root 用户
   - 扫描镜像漏洞
   - 保持基础镜像更新

## 性能调优

1. **资源限制**
   - 设置适当的内存限制
   - 为容器配置 CPU 限制
   - 监控资源使用情况

2. **缓存**
   - 启用缓存以获得更好性能
   - 配置适当的 TTL 值
   - 监控缓存命中率

3. **并发操作**
   - 调整工作池大小
   - 配置超时值
   - 监控响应时间

更多信息和更新，请访问 [GitHub 仓库](https://github.com/npm-console/npm-console)。
