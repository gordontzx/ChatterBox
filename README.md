# ChatterBox
Chatting app.

## Features

- **Accounts** - email/password registration and login, JWT access tokens
  (15 min) backed by DB-tracked refresh tokens (so a session survives a
  reload, and logout actually revokes it server-side).
- **Direct and group chats** - start a 1:1 chat with another user by email,
  create named group chats, add members, leave a chat.
- **Real-time messaging** - messages, typing indicators, read receipts, and
  online presence all pushed live over a WebSocket connection.
- **Read receipts** - per-message "seen by everyone" tracking in group
  chats.
- **Presence** - online/offline status with a "last seen" timestamp,
  updated live for anyone you share a chat with.
- **Live chat updates** - being added to a chat, or a group's membership
  changing, shows up immediately for everyone affected, no refresh needed.

## Quick start

Prerequisites: Docker and Docker Compose.

1. Clone the repo and set up your environment file:
   ```
   git clone git@github.com:gordontzx/ChatterBox.git
   cd ChatterBox
   cp .env.example .env
   ```
2. Generate a real `JWT_SECRET` and replace the placeholder line in `.env`
   with it.
   ```
   JWT_SECRET=$(openssl rand -base64 32) &&
   sed -i.bak "s#^JWT_SECRET=.*#JWT_SECRET=$JWT_SECRET#" .env &&
   rm -f .env.bak &&
   unset JWT_SECRET
   ```
3. Build and start everything:
   ```
   docker compose up --build -d
   ```

### Where to find it

- Frontend: http://localhost:5173
- API (direct, e.g. for curl): http://localhost:8080
- Postgres (e.g. for psql): localhost:5433

If any of these ports are already in use on your machine, that service will
fail to start - check `docker compose ps`/logs if something looks wrong.

### How to test
To use multiple accounts, open the frontend website on multiple browsers/incognito windows and login on different accounts.

### Checking it's up

```
docker compose ps       # all four services should show Up
docker compose logs -f  # follow logs if something looks wrong
```

### Shutting down

```
docker compose down     # stops everything, keeps your Postgres data
docker compose down -v  # also deletes the Postgres volume (wipes the DB)
```
