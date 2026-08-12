# Relay Chat Backend — REST API Reference

Complete REST API reference for the Relay chat backend. All endpoints return JSON responses and require JWT authentication (except auth endpoints).

**Base URL:** `http://localhost:7777/api/v1`

---

## 🔐 Authentication Endpoints

### 1. Register New User

Create a new user account.

| Setting | Value |
|---|---|
| **Method** | POST |
| **URL** | `/auth/register` |
| **Headers** | `Content-Type: application/json` |

**Request Body:**

```json
{
  "username": "user1",
  "email": "user1@gmail.com",
  "password": "123456"
}
Field	Type	Required	Constraints
username	string	Yes	3-30 characters, unique
email	string	Yes	Valid email format, unique
password	string	Yes	Min 6 characters

Response (201 Created):

{
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": "5ae8517e-349f-44a7-8f0f-a0211f703f7c",
      "username": "user1",
      "email": "user1@gmail.com",
      "created_at": "2026-08-10T13:56:51.194531383Z"
    }
  },
  "success": true
}

Error Responses:

// 400 Bad Request
{
  "success": false,
  "error": "Email already registered"
}
2. Login

Authenticate and receive JWT token.

Setting	Value
Method	POST
URL	/auth/login
Headers	Content-Type: application/json

Request Body:

{
  "email": "user1@gmail.com",
  "password": "123456"
}
Field	Type	Required	Description
email	string	Yes	Registered email
password	string	Yes	Account password

Response (200 OK):

{
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": "5ae8517e-349f-44a7-8f0f-a0211f703f7c",
      "username": "user1",
      "email": "user1@gmail.com"
    }
  },
  "success": true
}

Error Response (401):

{
  "success": false,
  "error": "Invalid email or password"
}
🏠 Server Endpoints
3. Create Server

Create a new server (like a Discord server).

Setting	Value
Method	POST
URL	/servers
Headers	Authorization: Bearer <JWT_TOKEN>
Content-Type: application/json

Request Body:

{
  "name": "Demo Server"
}
Field	Type	Required	Constraints
name	string	Yes	1-50 characters

Response (201 Created):

{
  "data": {
    "id": "096bfbd9-1f2d-4e88-bbea-b0ef18d03b18",
    "name": "Demo Server",
    "owner_id": "5ae8517e-349f-44a7-8f0f-a0211f703f7c",
    "invite_code": "HHMdCcjT",
    "created_at": "2026-08-10T13:57:55.846130219Z"
  },
  "success": true
}
4. List My Servers

Get all servers you're a member of.

Setting	Value
Method	GET
URL	/servers
Headers	Authorization: Bearer <JWT_TOKEN>

Response (200 OK):

{
  "data": {
    "servers": [
      {
        "id": "096bfbd9-1f2d-4e88-bbea-b0ef18d03b18",
        "name": "Demo Server",
        "invite_code": "HHMdCcjT",
        "owner_id": "5ae8517e-349f-44a7-8f0f-a0211f703f7c",
        "member_count": 2,
        "created_at": "2026-08-10T13:57:55.84613Z"
      }
    ]
  },
  "success": true
}
5. Join Server via Invite Code

Join an existing server using an invite code.

Setting	Value
Method	POST
URL	/servers/join
Headers	Authorization: Bearer <JWT_TOKEN>
Content-Type: application/json

Request Body:

{
  "invite_code": "HHMdCcjT"
}
Field	Type	Required	Description
invite_code	string	Yes	8-character invite code

Response (200 OK):

{
  "data": {
    "id": "096bfbd9-1f2d-4e88-bbea-b0ef18d03b18",
    "name": "Demo Server"
  },
  "success": true
}

Error Response (404):

{
  "success": false,
  "error": "Invalid invite code"
}
6. Get Server Presence (Online Users)

Check who's online in a server.

Setting	Value
Method	GET
URL	/servers/:server_id/presence
Headers	Authorization: Bearer <JWT_TOKEN>

Path Parameters:

Parameter	Type	Description
server_id	UUID	Server ID

Response (200 OK):

{
  "data": {
    "online_count": 2,
    "presence": {
      "5ae8517e-349f-44a7-8f0f-a0211f703f7c": true,
      "94a74e0e-81d0-4e74-b961-e0ab71b556c3": true
    }
  },
  "success": true
}
📢 Channel Endpoints
7. Create Channel

Create a channel inside a server.

Setting	Value
Method	POST
URL	/servers/:server_id/channels
Headers	Authorization: Bearer <JWT_TOKEN>
Content-Type: application/json

Path Parameters:

Parameter	Type	Description
server_id	UUID	Server ID

Request Body:

{
  "name": "general"
}
Field	Type	Required	Constraints
name	string	Yes	1-30 characters

Response (201 Created):

{
  "data": {
    "id": "3f316751-c5d0-4400-a0ba-5cd21a4e7ec1",
    "name": "general",
    "server_id": "096bfbd9-1f2d-4e88-bbea-b0ef18d03b18",
    "position": 0,
    "is_private": false,
    "created_at": "2026-08-10T14:00:46.353084972Z"
  },
  "success": true
}
8. List Channels in Server

Get all channels in a server.

Setting	Value
Method	GET
URL	/servers/:server_id/channels
Headers	Authorization: Bearer <JWT_TOKEN>

Path Parameters:

Parameter	Type	Description
server_id	UUID	Server ID

Response (200 OK):

{
  "data": {
    "channels": [
      {
        "id": "3f316751-c5d0-4400-a0ba-5cd21a4e7ec1",
        "name": "general",
        "server_id": "096bfbd9-1f2d-4e88-bbea-b0ef18d03b18",
        "position": 0,
        "is_private": false,
        "created_at": "2026-08-10T14:00:46.353084972Z"
      }
    ]
  },
  "success": true
}
💬 Message Endpoints
9. Send Message (REST API)

Send a message to a channel via REST.

Setting	Value
Method	POST
URL	/channels/:channel_id/messages
Headers	Authorization: Bearer <JWT_TOKEN>
Content-Type: application/json

Path Parameters:

Parameter	Type	Description
channel_id	UUID	Channel ID

Request Body:

{
  "content": "Hello from Postman!"
}
Field	Type	Required	Constraints
content	string	Yes	1-2000 characters

Response (201 Created):

{
  "data": {
    "id": "678ac8f9-1970-4f88-a25a-3b10b9df1d56",
    "content": "Hello from Postman!",
    "channel_id": "3f316751-c5d0-4400-a0ba-5cd21a4e7ec1",
    "sender_id": "5ae8517e-349f-44a7-8f0f-a0211f703f7c",
    "created_at": "2026-08-10T14:01:22.177928458Z"
  },
  "success": true
}
10. Get Messages (Paginated)

Retrieve messages from a channel with pagination.

Setting	Value
Method	GET
URL	/channels/:channel_id/messages?limit=50&offset=0
Headers	Authorization: Bearer <JWT_TOKEN>

Path Parameters:

Parameter	Type	Description
channel_id	UUID	Channel ID

Query Parameters:

Parameter	Type	Default	Description
limit	integer	50	Max messages to return (1-100)
offset	integer	0	Number of messages to skip

Response (200 OK):

{
  "data": {
    "messages": [
      {
        "id": "678ac8f9-1970-4f88-a25a-3b10b9df1d56",
        "content": "Hello from Postman!",
        "channel_id": "3f316751-c5d0-4400-a0ba-5cd21a4e7ec1",
        "sender": {
          "id": "5ae8517e-349f-44a7-8f0f-a0211f703f7c",
          "username": "user1"
        },
        "created_at": "2026-08-10T14:01:22.177928458Z"
      }
    ],
    "pagination": {
      "limit": 50,
      "offset": 0,
      "total": 1
    }
  },
  "success": true
}
🔌 WebSocket Protocol
Connect to Channel

URL:

ws://localhost:7777/ws/:channel_id?token=<JWT_TOKEN>

Query Parameters:

Parameter	Type	Description
token	string	JWT token from login

Example:

ws://localhost:7777/ws/3f316751-c5d0-4400-a0ba-5cd21a4e7ec1?token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
Client → Server Events
Send Message
{
  "type": "message",
  "payload": {
    "content": "Hello from WebSocket!"
  }
}
Typing Start
{
  "type": "typing_start",
  "payload": {}
}
Typing Stop
{
  "type": "typing_stop",
  "payload": {}
}
Server → Client Events
New Message
{
  "type": "new_message",
  "payload": {
    "id": "b36b340a-ccad-47df-8cf4-2778e12de484",
    "content": "Hello from WebSocket!",
    "channel_id": "3f316751-c5d0-4400-a0ba-5cd21a4e7ec1",
    "sender": {
      "id": "5ae8517e-349f-44a7-8f0f-a0211f703f7c",
      "username": "user1"
    },
    "created_at": "2026-08-10T14:01:57.716818063Z"
  }
}
User Typing
{
  "type": "user_typing",
  "payload": {
    "user_id": "5ae8517e-349f-44a7-8f0f-a0211f703f7c",
    "username": "user1",
    "channel_id": "3f316751-c5d0-4400-a0ba-5cd21a4e7ec1"
  }
}
User Stopped Typing
{
  "type": "user_stopped_typing",
  "payload": {
    "user_id": "5ae8517e-349f-44a7-8f0f-a0211f703f7c",
    "username": "user1",
    "channel_id": "3f316751-c5d0-4400-a0ba-5cd21a4e7ec1"
  }
}
User Joined Channel
{
  "type": "user_joined",
  "payload": {
    "user_id": "5ae8517e-349f-44a7-8f0f-a0211f703f7c",
    "username": "user1",
    "channel_id": "3f316751-c5d0-4400-a0ba-5cd21a4e7ec1"
  }
}
User Left Channel
{
  "type": "user_left",
  "payload": {
    "user_id": "5ae8517e-349f-44a7-8f0f-a0211f703f7c",
    "username": "user1",
    "channel_id": "3f316751-c5d0-4400-a0ba-5cd21a4e7ec1"
  }
}
📊 Status Codes
Code	Description
200	Success
201	Created
400	Bad Request
401	Unauthorized
403	Forbidden
404	Not Found
429	Too Many Requests
500	Internal Server Error
🛡️ Rate Limits
Endpoint	Limit
Auth Routes (register/login)	10 requests/min
API Routes	60 requests/min
✅ Quick Reference (Sample Data)
USERNAME: user1
EMAIL: user1@gmail.com
PASSWORD: 123456

USERNAME: user2
EMAIL: user2@gmail.com
PASSWORD: 123456
🎯 Interview Demo Flow
1. Health Check → "Server is running"
2. Register user1 → "user1 created"
3. Login user1 → "Got JWT token"
4. Create Server → "Got server_id + invite_code"
5. Create Channel → "Got channel_id"
6. Send Message → "Hello from REST API!"
7. Get Messages → "Messages with pagination"

8. WebSocket (user1) → "Real-time connection"
9. Send WebSocket message → "Hi from user1!"
10. Typing indicators → "user1 is typing..."

11. Register user2 → "user2 created"
12. Login user2 → "Got user2 token"
13. user2 joins server → "Using invite code"
14. WebSocket (user2) → "Second user connected"
15. user1 sends "Hello!" → user2 sees it
16. user2 replies "Hi!" → user1 sees it
17. Check Presence → "2 users online"
📝 Notes for Interview
JWT tokens expire after 72 hours
Passwords are hashed with bcrypt (cost factor 12)
Messages use soft deletes (never permanently lost)
Presence tracked via Redis TTL keys (90-second expiry)
WebSocket uses Hub pattern for concurrent connections

API Documentation Complete! 🚀