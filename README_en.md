# npm-console

A powerful npm package manager tool built with Go 1.24, supporting multiple package managers with both CLI and Web interfaces.

## Features

- 🚀 **Multi-Package Manager Support**: npm, pnpm, yarn, bun
- 💻 **Dual Interface**: Command-line interface and Web dashboard
- 🧹 **Cache Management**: View cache size, one-click cleanup
- 📦 **Package Management**: List installed packages, manage dependencies
- 🌐 **Registry Management**: Configure and switch between registries
- 🔗 **Proxy Management**: HTTP/HTTPS proxy configuration
- 📁 **Project Management**: Scan and analyze local projects
- 🔒 **Cross-Platform**: Windows, Linux, macOS support
- ⚡ **Single Binary**: One-click deployment

## Installation

### Download Binary

Download the latest release for your platform from the [releases page](https://github.com/your-org/npm-console/releases).

### Build from Source

```bash
# Clone the repository
git clone https://github.com/your-org/npm-console.git
cd npm-console

# Build the binary
make build

# Or build for all platforms
make build-all
```

## Usage

### Command Line Interface

```bash
# Show help
npm-console --help

# Cache management
npm-console cache list          # List all caches
npm-console cache clean         # Clean all caches
npm-console cache info          # Show cache information

# Package management
npm-console packages list       # List installed packages
npm-console packages search     # Search packages

# Registry management
npm-console registry list       # List configured registries
npm-console registry set        # Set registry URL
npm-console registry test       # Test registry connectivity

# Proxy management
npm-console proxy set           # Set proxy configuration
npm-console proxy unset         # Remove proxy
npm-console proxy test          # Test proxy connectivity

# Project management
npm-console projects scan       # Scan for projects
npm-console projects analyze    # Analyze project dependencies

# Web interface
npm-console web                 # Start web server (default: http://localhost:8080)
```

### Web Interface

Start the web server:

```bash
npm-console web
```

Then open your browser and navigate to `http://localhost:8080` to access the web dashboard.

## Configuration

npm-console uses a YAML configuration file. The default locations are:

- `$HOME/.npm-console.yaml`
- `$XDG_CONFIG_HOME/npm-console/config.yaml` (Linux)
- `~/Library/Application Support/npm-console/config.yaml` (macOS)
- `%APPDATA%/npm-console/config.yaml` (Windows)

### Example Configuration

```yaml
app:
  name: npm-console
  version: 1.0.0
  environment: production

logger:
  level: info
  format: text
  output: stdout

web:
  enabled: true
  host: localhost
  port: 8080
  cors:
    enabled: true
    allowed_origins: ["*"]

managers:
  npm:
    enabled: true
    registry: https://registry.npmjs.org/
  pnpm:
    enabled: true
    registry: https://registry.npmjs.org/
  yarn:
    enabled: true
    registry: https://registry.npmjs.org/
  bun:
    enabled: true
    registry: https://registry.npmjs.org/

cache:
  auto_clean: false
  max_size: 10GB
  max_age: 30d
```

## Development

### Prerequisites

- Go 1.24 or later
- Make (optional, for using Makefile)

### Setup

```bash
# Clone the repository
git clone https://github.com/your-org/npm-console.git
cd npm-console

# Install dependencies
make deps

# Run in development mode
make dev

# Run tests
make test

# Run with coverage
make coverage

# Format code
make fmt

# Run linter
make lint
```

### Project Structure

```
npm-console/
├── cmd/                 # CLI commands (Cobra)
├── internal/
│   ├── core/           # Core interfaces and models
│   ├── managers/       # Package manager implementations
│   ├── services/       # Business services
│   └── web/           # Web server (Fiber)
├── pkg/               # Public packages
│   ├── config/        # Configuration management
│   ├── logger/        # Logging utilities
│   └── utils/         # Utility functions
├── web/               # Frontend assets
├── docs/              # Documentation
└── configs/           # Configuration files
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built with [Go 1.24](https://golang.org/)
- CLI powered by [Cobra](https://github.com/spf13/cobra)
- Web server powered by [Fiber](https://github.com/gofiber/fiber)
- Configuration management with [Viper](https://github.com/spf13/viper)
