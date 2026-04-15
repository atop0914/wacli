# wacli — WhatsApp CLI

A command-line interface for WhatsApp, built with Go and [whatsmeow](https://github.com/htrc/WhatsApp).

## Features

- **Authentication**: QR code based login
- **Message Sync**: Local SQLite storage with continuous capture
- **Search**: Full-text search powered by SQLite FTS5
- **Send**: Send text messages and files
- **Group Management**: List and rename groups
- **History**: Backfill older messages from your primary device

## Installation

### From Source

```bash
git clone https://github.com/atop0914/wacli.git
cd wacli
make build
```

### From Homebrew

```bash
brew install steipete/tap/wacli
```

## Quick Start

### 1. Authenticate

```bash
./dist/wacli auth
```

Scan the QR code with your WhatsApp app.

### 2. Sync Messages

```bash
# Continuous sync
./dist/wacli sync --follow

# One-shot sync
./dist/wacli sync
```

### 3. Search Messages

```bash
./dist/wacli messages search "meeting"

# JSON output
./dist/wacli messages search "meeting" --json
```

### 4. Send Messages

```bash
# Send text
./dist/wacli send text --to 1234567890 --message "Hello!"

# Send file
./dist/wacli send file --to 1234567890 --file ./photo.jpg --caption "Check this out"

# Override display filename
./dist/wacli send file --to 1234567890 --file /tmp/abc123 --filename report.pdf
```

### 5. Group Management

```bash
# List groups
./dist/wacli groups list

# Rename group
./dist/wacli groups rename --jid 123456789@g.us --name "New Group Name"
```

### 6. History Backfill

```bash
./dist/wacli history backfill --chat 1234567890@s.whatsapp.net --requests 10 --count 50
```

### 7. Diagnostics

```bash
./dist/wacli doctor
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `WACLI_STORE` | Storage directory | `~/.wacli` |
| `WACLI_DEVICE_LABEL` | Linked device label | - |
| `WACLI_DEVICE_PLATFORM` | Device platform | `CHROME` |

### CLI Flags

| Flag | Description |
|------|-------------|
| `--store DIR` | Custom storage directory |
| `--json` | JSON output format |

## Development

```bash
# Build
make build

# Run tests
make test

# Clean
make clean
```

## Architecture

```
wacli/
├── cmd/wacli/          # Entry point
├── internal/
│   ├── auth/           # Authentication
│   ├── client/         # WhatsApp client
│   ├── commands/       # CLI commands
│   ├── db/             # SQLite layer
│   └── store/          # Storage management
└── docs/
    └── spec.md         # Design specification
```

## License

MIT
