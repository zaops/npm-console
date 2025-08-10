# npm-console

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Build Status](https://github.com/npm-console/npm-console/workflows/CI/badge.svg)](https://github.com/npm-console/npm-console/actions)

**npm-console** 是一个统一的包管理器控制台，为 npm、pnpm、yarn 和 bun 提供一致的管理界面。

[English](README_en.md) | 简体中文

## ✨ 功能特性

- 🔧 **统一管理**: 支持 npm、pnpm、yarn、bun 四种包管理器
- 🗄️ **缓存管理**: 查看、清理、统计缓存信息
- 📦 **包管理**: 全局包查看、搜索、统计
- ⚙️ **配置管理**: 镜像源和代理设置
- 📁 **项目管理**: 项目扫描、分析、依赖树
- 💻 **CLI界面**: 完整的命令行工具
- 🌐 **Web界面**: 现代化的Web管理界面
- 📱 **响应式设计**: 支持桌面和移动设备
- 🌍 **跨平台**: Windows、Linux、macOS 支持

## 🚀 快速开始

### 安装

#### 自动安装脚本

**Linux/macOS:**
```bash
curl -fsSL https://raw.githubusercontent.com/npm-console/npm-console/main/scripts/install.sh | bash
```

**Windows (PowerShell):**
```powershell
iwr -useb https://raw.githubusercontent.com/npm-console/npm-console/main/scripts/install.ps1 | iex
```

#### 手动安装

1. 从 [Releases 页面](https://github.com/npm-console/npm-console/releases) 下载适合您平台的二进制文件
2. 解压并将二进制文件移动到 PATH 目录中
3. 验证安装：`npm-console version`

#### Docker 运行

```bash
docker run -p 8080:8080 npm-console/npm-console:latest
```

### 基本使用

```bash
# 显示帮助信息
npm-console --help

# 查看版本信息
npm-console version

# 列出缓存信息
npm-console cache list

# 清理所有缓存
npm-console cache clean

# 列出全局包
npm-console packages list --global

# 搜索包
npm-console packages search react

# 扫描项目
npm-console projects scan

# 启动 Web 界面
npm-console web
```

## 📖 详细文档

### CLI 命令

#### 缓存管理
```bash
npm-console cache list              # 列出所有缓存信息
npm-console cache clean             # 清理所有缓存
npm-console cache clean --manager npm  # 清理指定管理器缓存
npm-console cache info              # 显示缓存详细信息
npm-console cache size              # 显示总缓存大小
```

#### 包管理
```bash
npm-console packages list           # 列出项目包
npm-console packages list --global  # 列出全局包
npm-console packages search <query> # 搜索包
npm-console packages info <name>    # 显示包信息
npm-console packages stats          # 显示包统计
```

#### 配置管理
```bash
npm-console registry list           # 列出镜像源配置
npm-console registry set <url>      # 设置镜像源
npm-console registry test           # 测试镜像源连接
npm-console proxy set <url>         # 设置代理
npm-console proxy unset             # 移除代理
```

#### 项目管理
```bash
npm-console projects scan           # 扫描项目
npm-console projects analyze        # 分析项目
npm-console projects stats          # 项目统计
npm-console projects deps           # 显示依赖树
```

#### Web 界面
```bash
npm-console web                     # 启动 Web 服务器
npm-console web --port 3000         # 指定端口
npm-console web --host 0.0.0.0      # 指定主机
```

### Web 界面

启动 Web 服务器后，在浏览器中访问 `http://localhost:8080` 即可使用图形化界面：

- 📊 **仪表板**: 系统概览和状态监控
- 🗄️ **缓存管理**: 可视化缓存管理
- 📦 **包管理**: 包浏览和搜索
- ⚙️ **配置管理**: 镜像源和代理设置
- 📁 **项目管理**: 项目扫描和分析

## 🏗️ 架构设计

```
npm-console/
├── cmd/                    # CLI 命令
├── internal/
│   ├── core/              # 核心数据结构
│   ├── managers/          # 包管理器实现
│   ├── services/          # 业务服务层
│   └── web/               # Web 服务器
├── pkg/
│   ├── config/            # 配置管理
│   ├── logger/            # 日志系统
│   └── utils/             # 工具函数
├── web/dist/              # Web 前端资源
└── scripts/               # 构建和安装脚本
```

## 🔧 开发

### 环境要求

- Go 1.21+
- Node.js 18+ (用于 Web 资源)
- Git

### 构建

```bash
# 克隆仓库
git clone https://github.com/npm-console/npm-console.git
cd npm-console

# 构建
go build -o npm-console .

# 或使用构建脚本
./scripts/build.sh
```

### 测试

```bash
# 运行所有测试
go test ./...

# 运行基准测试
go test -bench=. ./...

# 生成覆盖率报告
go test -cover ./...
```

## 📦 部署

### Docker 部署

```bash
# 使用 Docker Compose
docker-compose up -d

# 或直接运行
docker run -d \
  --name npm-console \
  -p 8080:8080 \
  -v $(pwd)/projects:/app/projects:ro \
  npm-console/npm-console:latest
```

### 系统服务

#### Linux (systemd)

创建 `/etc/systemd/system/npm-console.service`:

```ini
[Unit]
Description=npm-console
After=network.target

[Service]
Type=simple
User=npm-console
ExecStart=/usr/local/bin/npm-console web
Restart=always

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl enable npm-console
sudo systemctl start npm-console
```

#### Windows 服务

使用 NSSM 或 PowerShell 计划任务创建 Windows 服务。

详细部署说明请参考 [部署指南](DEPLOYMENT.md)。

## 🤝 贡献

我们欢迎各种形式的贡献！

1. Fork 本仓库
2. 创建功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add some amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

### 开发指南

- 遵循 Go 代码规范
- 添加适当的测试
- 更新相关文档
- 确保 CI 通过

## 📄 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件。

## 🙏 致谢

感谢所有为这个项目做出贡献的开发者！

## 📞 支持

- 🐛 [报告问题](https://github.com/npm-console/npm-console/issues)
- 💡 [功能请求](https://github.com/npm-console/npm-console/issues)
- 📖 [文档](https://github.com/npm-console/npm-console/wiki)
- 💬 [讨论](https://github.com/npm-console/npm-console/discussions)

---

<div align="center">
  <p>如果这个项目对您有帮助，请给我们一个 ⭐️</p>
  <p>Made with ❤️ by the npm-console team</p>
</div>
