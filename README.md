<div align="center">

# ⚡ Relay

### Production-grade Discord-style realtime chat backend built with Go

[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?style=for-the-badge&logo=go)](https://golang.org)
[![Fiber](https://img.shields.io/badge/Fiber-v2-00ACD7?style=for-the-badge)](https://gofiber.io)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-336791?style=for-the-badge&logo=postgresql)](https://postgresql.org)
[![Redis](https://img.shields.io/badge/Redis-7-DC382D?style=for-the-badge&logo=redis)](https://redis.io)
[![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?style=for-the-badge&logo=docker)](https://docker.com)
[![WebSocket](https://img.shields.io/badge/WebSocket-Realtime-010101?style=for-the-badge)](https://developer.mozilla.org/en-US/docs/Web/API/WebSockets_API)

[Features](#-features) • [Architecture](#-architecture) • [Quick Start](#-quick-start) • [API Reference](#-api-reference) • [WebSocket Protocol](#-websocket-protocol)

</div>

---

## 🚀 What is Relay?

Relay is a **production-quality realtime chat backend** inspired by Discord. Built entirely in Go, it handles WebSocket connections, realtime message broadcasting, presence tracking, and Redis-powered horizontal scaling — all within a clean, layered architecture.

> Built to demonstrate real backend engineering: not just CRUD, but goroutines, channels, mutexes, pub/sub, connection pooling, and graceful shutdown.

---

## ✨ Features

| Feature | Technology |
|---|---|
| JWT Authentication | `golang-jwt/jwt` |
| Realtime Messaging | WebSockets + Goroutines |
| Multi-server Scaling | Redis Pub/Sub |
| Presence Tracking | Redis TTL keys |
| Typing Indicators | WebSocket events |
| Rate Limiting | Fiber middleware (10 req/min auth) |
| Structured Logging | Fiber logger middleware |
| Graceful Shutdown | OS signal handling |
| Dockerized | Multi-stage Docker build (~15MB image) |
| Password Security | bcrypt (cost factor 12) |
| Soft Deletes | GORM DeletedAt |
| Connection Pooling | PostgreSQL (25 max) + Redis (10 max) |

---

## 🏗 Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        CLIENT                               │
│              (WebSocket + REST API)                         │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│                    FIBER HTTP SERVER                        │
│         Rate Limiter │ Logger │ JWT Middleware              │
└──────┬───────────────┬────────────────────┬─────────────────┘
       │               │                    │
       ▼               ▼                    ▼
  REST APIs      WebSocket            Health Check
  /api/v1        /ws/:channelID       /health
       │               │
       ▼               ▼
┌─────────────────────────────────────────────────────────────┐
│                    HUB (single goroutine)                   │
│                                                             │
│   register ──→ map[channelID]map[*Client]bool               │
│   unregister ──→ delete + close(client.send)                │
│   broadcast ──→ handleInboundMessage()                      │
│                        │                                    │
│               ┌────────┴────────┐                           │
│               ▼                 ▼                           │
│         Save to DB      Publish to Redis                    │
└─────────────────────────────────────────────────────────────┘
       │               │
       ▼               ▼
┌────────────┐  ┌─────────────────────────────────────────────┐
│ PostgreSQL │  │              REDIS                          │
│            │  │                                             │
│ users      │  │  chat:<channelID>  ← Pub/Sub               │
│ servers    │  │  presence:<userID> ← TTL 90s               │
│ channels   │  │                                             │
│ messages   │  └─────────────────────────────────────────────┘
└────────────┘
```

### Concurrency Model

```
For each WebSocket connection:

  ReadPump goroutine          WritePump goroutine
  ──────────────────          ───────────────────
  reads from browser    →     writes to browser
  sends to hub.broadcast      reads from client.send channel
                              sends ping every 54s
                              refreshes Redis presence every 30s

Hub goroutine (single)
──────────────────────
processes register/unregister/broadcast sequentially
owns the clients map — NO mutex needed
```

---

## 📁 Project Structure

```
backend/
├── cmd/server/
├── internal/
│   ├── auth/          # JWT generation + validation
│   ├── users/         # User model, register, login, presence API
│   ├── servers/       # Server + member management
│   ├── channels/      # Channel management
│   ├── messages/      # Message persistence + pagination
│   ├── websocket/     # Hub, Client, ReadPump, WritePump, Pub/Sub
│   ├── middleware/    # JWT auth, rate limiter, logger, recover
│   ├── config/        # Environment config loader
│   ├── database/      # GORM connection + migrations
│   └── redis/         # Redis connection, presence, pub/sub
├── Dockerfile         # Multi-stage build (~15MB final image)
├── docker-compose.yml # PostgreSQL + Redis + App
├── .env.example       # Environment variable template
└── main.go            # DI wiring + graceful shutdown
```

---

## ⚡ Quick Start

### Prerequisites
- Docker + Docker Compose
- Go 1.25+

### 1. Clone the repo
```bash
git clone https://github.com/lohithbandla/relay.git
cd relay/backend
```

### 2. Set up environment
```bash
cp .env.example .env
# Edit .env with your values
```

### 3. Start everything
```bash
docker-compose up --build -d
```

### 4. Verify
```bash
curl http://localhost:7777/health
# {"success":true,"message":"Server is healthy","env":"production"}
```

---

## 📡 API Reference

### Authentication

```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "username": "lohith",
  "email": "lohith@example.com",
  "password": "secret123"
}
```

```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "lohith@example.com",
  "password": "secret123"
}
```

### Servers

```http
POST   /api/v1/servers              # Create server
GET    /api/v1/servers              # List my servers
POST   /api/v1/servers/join         # Join via invite code
POST   /api/v1/servers/:id/channels # Create channel
GET    /api/v1/servers/:id/channels # List channels
GET    /api/v1/servers/:id/presence # Get online members
```

### Messages

```http
POST /api/v1/channels/:id/messages  # Send message
GET  /api/v1/channels/:id/messages?limit=50&offset=0  # Get messages
```

---

## 🔌 WebSocket Protocol

### Connect
```
ws://localhost:7777/ws/:channelID?token=<JWT>
```

### Send a message
```json
{
  "type": "message",
  "payload": {
    "content": "Hello, World!"
  }
}
```

### Receive a message
```json
{
  "type": "new_message",
  "payload": {
    "id": "uuid",
    "content": "Hello, World!",
    "channel_id": "uuid",
    "sender": {
      "id": "uuid",
      "username": "lohith"
    },
    "created_at": "2026-05-10T07:00:00Z"
  }
}
```

### Event Types

| Event | Direction | Description |
|---|---|---|
| `message` | Client → Server | Send a chat message |
| `new_message` | Server → Client | Broadcast new message |
| `typing_start` | Client → Server | User started typing |
| `typing_stop` | Client → Server | User stopped typing |
| `user_joined` | Server → Client | User connected to channel |
| `user_left` | Server → Client | User disconnected |

---

## 🛡 Security

- Passwords hashed with **bcrypt** (cost factor 12)
- JWT tokens with configurable expiry (default 72h)
- Rate limiting: **10 req/min** on auth routes, **60 req/min** on API
- WebSocket auth via JWT query parameter
- UUIDs for all IDs (prevents enumeration attacks)
- Soft deletes (data never permanently lost)
- Environment variables for all secrets

---

## 🐳 Docker

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f app

# Stop everything
docker-compose down

# Wipe all data
docker-compose down -v
```

**Image size:** ~15MB (multi-stage build)
**Services:** Go app + PostgreSQL 16 + Redis 7

---

## 🧠 Key Engineering Decisions

**Why the Hub pattern?**
A single goroutine owns the clients map. All operations go through channels. Zero mutexes on the map — zero race conditions.

**Why Redis Pub/Sub?**
When running multiple server instances, WebSocket connections are split across them. Redis Pub/Sub ensures every message reaches every connected client regardless of which instance they're on.

**Why UUID over auto-increment?**
Auto-increment IDs allow enumeration attacks (`/users/1`, `/users/2`...). UUIDs are random and unpredictable.

**Why soft deletes?**
Chat apps need audit trails. Messages are never truly deleted — just marked with `deleted_at`. Recoverable, auditable, compliant.

---

## 📊 Performance

| Metric | Value |
|---|---|
| PostgreSQL max connections | 25 |
| Redis max connections | 10 |
| WebSocket message buffer | 256 messages/client |
| Auth rate limit | 10 req/min/IP |
| API rate limit | 60 req/min/IP |
| Graceful shutdown timeout | 10 seconds |
| Presence TTL | 90 seconds |
| Presence heartbeat | 30 seconds |

---

<div align="center">

Built with ❤️ by [Lohith Bandla](https://github.com/lohithbandla)

⭐ Star this repo if you found it useful!

</div>