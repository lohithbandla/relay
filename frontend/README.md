# Relay — Frontend

A React + Vite client for [Relay](https://github.com/lohithbandla/relay), the Go/Fiber/Redis realtime
chat backend. Covers the full feature set exposed by the API: auth, servers,
channels, message history + live delivery, typing indicators, and presence.

## Stack

- React 18 + Vite
- No UI framework — hand-written CSS design system in `src/index.css`
- Talks to the backend over REST (`/api/v1/...`) for auth/history/servers/channels,
  and over a raw WebSocket (`/ws/:channelID`) for live messages, typing, and join/leave events

## Setup

```bash
cd relay-frontend
npm install
cp .env.example .env   # point VITE_API_URL at your running backend
npm run dev
```

Requires the Relay backend running (see its own `docker-compose up --build -d`,
default `http://localhost:7777`).

## Notes on wiring to the backend

- **Sending messages goes over the WebSocket, not REST.** The backend's REST
  `POST /channels/:id/messages` persists a message but does *not* broadcast it
  to connected clients (only the Hub's WS path does) — so the composer sends
  a `{"type":"message","payload":{"content":...}}` frame instead. REST
  `GET /channels/:id/messages` is used to preload history when a channel opens.
- **Presence** comes from `GET /servers/:id/presence`, which only returns member
  IDs + online booleans (no usernames), so the member panel labels everyone but
  yourself by a shortened ID. Wire up a "list server members" endpoint on the
  backend if you want real names there.
- **Auth** is a JWT stored in `localStorage`, sent as `Authorization: Bearer <token>`
  on REST calls and as a `?token=` query param on the WebSocket upgrade, matching
  the backend's `middleware.Protected` and `ws.UpgradeMiddleware`.

## Structure

```
src/
├── lib/api.js              # REST client + session storage
├── hooks/useChannelSocket.js  # WebSocket lifecycle per active channel
├── context/AuthContext.jsx
└── components/
    ├── AuthScreen.jsx
    ├── StatusBar.jsx        # connection/presence readout
    ├── ServerRail.jsx
    ├── ChannelSidebar.jsx
    ├── ChatView.jsx         # history + live feed + composer
    ├── MemberPanel.jsx
    └── *Modal.jsx           # create/join server, create channel
```
