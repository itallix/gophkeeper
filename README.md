# GophKeeper

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

A secure and reliable client-server system for storing sensitive information, written in Go. GophKeeper provides encrypted storage for various types of data including credentials, binary files, text notes, and more.

## Features

- 🔐 Secure user authentication
- 📁 Binary file storage and retrieval
- 🔄 Cross-platform support (Linux, macOS, Windows)
- 🔒 End-to-end encryption
- 🚀 High-performance gRPC communication
- 💾 Reliable PostgreSQL storage backend
- ☁️ S3-compatible object storage support

## Installation

### Prerequisites

- Go 1.23 or higher
- PostgreSQL
- Docker

### Building from Source

1. Clone the repository:
```bash
git clone https://github.com/itallix/gophkeeper.git
cd gophkeeper
```

2. Build the client and server:
```bash
make all       # Build CLI for all platforms
# or
make linux     # Build CLI for linux platform
```

### Docker Installation

```bash
make up        # Start Docker containers
make down      # Stop Docker containers
```

## Quick Start

1. Start the server:
```bash
make server
./bin/server
```

2. Setup the client:
```bash
# For Linux AMD64
make linux
mv ./bin/cli-linux-amd64 ./bin/cli
# For other platforms, use appropriate binary
```

3. Register and authenticate:
```bash
./bin/cli user register -l <username>
./bin/cli user auth -l <username>
```

## Usage Guide

### Command Structure

All commands follow the pattern:
```bash
./bin/cli <category> <action> [options]
```

Most commands require the `-p` (path) parameter as a reference to the secret. Binary retrieval operation use `-f` (filepath) for the source file path.

### User Management

```bash
# Register a new user
./bin/cli user register -l adam

# Authenticate user
./bin/cli user auth -l adam
```

### Binary Operations

```bash
# Upload a binary file
./bin/cli binary create -f path/to/file.mp4

# List all stored binaries
./bin/cli binary list

# Retrieve a binary file
./bin/cli binary get -p original_name.mp4 -o output_name.mp4

# Delete a binary file
./bin/cli binary delete -p filename.mp4
```

### Flags Reference

| Flag | Description | Used With |
|------|-------------|-----------|
| `-p` | Path/reference to the secret | Most commands |
| `-f` | Source file path | Binary creation |
| `-o` | Output file path | Binary retrieval |
| `-l` | Username | User operations |

## Project Structure

```
.
├── cmd/
│   ├── client/        # Client application
│   └── server/        # Server application
├── internal/
│   ├── client/        # Client logic
│   ├── server/        # Server logic
│   └── common/        # Shared code
├── pkg/
│   └── generated/     # Generated protobuf code
├── api/
│   └── proto/         # Protocol buffer definitions
└── db/
    └── migrations/    # Database migrations
```

## Development

### Running Tests

```bash
make test
```

### Database Migrations

```bash
make migrate-up     # Apply migrations
make migrate-down   # Rollback migrations
```

### Generate Protocol Buffers

```bash
make proto
```

## Security Considerations

- All data is encrypted before storage
- Communication is secured via gRPC with TLS
- Passwords are hashed using modern algorithms

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [MinIO](https://min.io/) for object storage
- [gRPC](https://grpc.io/) for the communication framework
- [PostgreSQL](https://www.postgresql.org/) for reliable data storage
- The [Yandex.Practicum](https://practicum.yandex.ru/go-advanced/) and it's rock-star team for inspiration and support
