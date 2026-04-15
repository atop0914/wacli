# wacli Design Specification

## Overview

wacli is a WhatsApp CLI client that interfaces with WhatsApp Web protocol via the whatsmeow library. It provides a terminal-based interface for messaging, searching, and group management.

## Architecture

### Components

1. **CLI Layer** (`cmd/wacli/`) - Cobra-based command framework
2. **Client Layer** (`internal/client/`) - WhatsApp network communication
3. **Auth Layer** (`internal/auth/`) - QR authentication and session management
4. **DB Layer** (`internal/db/`) - SQLite with FTS5 for message storage
5. **Store Layer** (`internal/store/`) - File system storage management

### Data Flow

```
User Input → CLI → Client → WhatsApp Web
                ↓
            Database ← Client Events
```

## Database Schema

### messages
```sql
CREATE TABLE messages (
    id TEXT PRIMARY KEY,
    chat_id TEXT NOT NULL,
    sender_id TEXT NOT NULL,
    content TEXT,
    timestamp INTEGER NOT NULL,
    is_group INTEGER DEFAULT 0,
    has_media INTEGER DEFAULT 0,
    media_type TEXT,
    reply_to TEXT
);
```

### chats
```sql
CREATE TABLE chats (
    jid TEXT PRIMARY KEY,
    name TEXT,
    is_group INTEGER DEFAULT 0,
    last_message INTEGER
);
```

### FTS5 Virtual Table
```sql
CREATE VIRTUAL TABLE messages_fts USING fts5(
    content,
    content='messages',
    content_rowid='rowid'
);
```

## Command Structure

```
wacli
├── auth          # QR login
├── sync          # Message sync
│   └── --follow  # Continuous mode
├── messages
│   └── search    # Full-text search
├── send
│   ├── text      # Send text message
│   └── file      # Send file/media
├── groups
│   ├── list      # List groups
│   └── rename    # Rename group
├── history
│   └── backfill  # History sync
└── doctor        # Diagnostics
```

## Authentication Flow

1. Check for existing session in store
2. If no session, generate QR code
3. User scans QR with WhatsApp app
4. Store session credentials locally
5. On restart, use stored session (no QR needed)

## Message Sync Strategy

- **Continuous mode**: Keep connection alive, store all incoming messages
- **Backfill**: Request older messages from primary device on demand
- **Best-effort**: WhatsApp may not return full history

## Error Handling

- Network errors: Retry with exponential backoff
- Auth errors: Prompt for re-authentication
- DB errors: Log and continue (non-fatal)

## Security Considerations

- Session tokens stored with 0600 permissions
- No plaintext password storage
- QR codes are single-use
