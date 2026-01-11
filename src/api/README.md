# 🔌 REST API Server

<div align="center">
  <img src="../logo.png" alt="RL-Discord-Matchmaking Logo" width="150"/>
</div>

<h3 align="center">Rocket League Matchmaking API</h3>

<p align="center">
  A secure REST API for accessing player data and leaderboards from the Rocket League Discord matchmaking bot.
</p>

---

## 📋 Table of Contents

- [Overview](#-overview)
- [Quick Start](#-quick-start)
- [Environment Variables](#-environment-variables)
- [API Endpoints](#-api-endpoints)
- [Security](#-security)
- [Example Usage](#-example-usage)
- [Frontend Integration](#-frontend-integration)

## 📋 Overview

This REST API provides secure access to player statistics, MMR data, and leaderboards for both 1v1 and 2v2 game modes. Built with authentication and CORS protection for seamless frontend integration.

## 🚀 Quick Start

```bash
# Navigate to src directory
cd src

# Run the API server
go run cmd/api-server/main.go

# Or with custom port
go run cmd/api-server/main.go -port 3000
```

The API will be available at `http://localhost:8080` (or your custom port).

## 🔧 Environment Variables

The API requires two environment variables:
- `POSTGRES_CONNECTION_STRING` - Database connection string
- `API_KEY` - Secret key for API authentication

### Using dev.env file
Add both variables to your `dev.env` file:
```env
POSTGRES_CONNECTION_STRING="host=localhost port=5432 user=vb_su password=YOUR_PASSWORD dbname=rl_db sslmode=disable"
API_KEY="your-secret-api-key-here"
```

### Setting environment variables directly
```bash
export POSTGRES_CONNECTION_STRING="host=localhost port=5432 user=vb_su password=YOUR_PASSWORD dbname=rl_db sslmode=disable"
export API_KEY="your-secret-api-key-here"
go run cmd/api-server/main.go
```

## Running the API

From the `src/` directory:

```bash
go run cmd/api-server/main.go
```

Or with a custom port:

```bash
go run cmd/api-server/main.go -port 3000
```

Or build and run:

```bash
go build -o api-server cmd/api-server/main.go
./api-server
```


## API Endpoints

### Health Check
- **GET** `/health`
- Returns: `{"status": "ok"}`

### Get All Players
- **GET** `/api/players`
- Returns: Array of all players sorted by MMR (descending)
- Response format:
```json
[
  {
    "id": "player_name",
    "name": "player_name",
    "matchId": "uuid",
    "mentionName": "player_name",
    "numWins": 10,
    "numLosses": 5,
    "mmr": 1200.5,
    "isInGame": false,
    "isAdmin": false,
    "discordId": 123456789
  }
]
```

### Get Player by Discord ID
- **GET** `/api/players/{discordId}`
- Returns: Single player object
- Example: `GET /api/players/123456789`

### Get Leaderboard
- **GET** `/api/leaderboard`
- Returns: Array of all players sorted by MMR (descending) - defaults to 2v2
- Same format as `/api/players`

### Get 1v1 Leaderboard
- **GET** `/api/leaderboard/1v1`
- Returns: Array of players sorted by 1v1 MMR (descending)

### Get 2v2 Leaderboard
- **GET** `/api/leaderboard/2v2`
- Returns: Array of players sorted by 2v2 MMR (descending)

## 📄 Response Formats

### Player Object
```json
{
  "id": "player_name",
  "name": "player_name",
  "matchId": "uuid-string",
  "mentionName": "player_name",
  "numWins": 15,
  "numLosses": 8,
  "mmr": 1250.5,
  "isInGame": false,
  "isAdmin": false,
  "discordId": 123456789
}
```

### Error Response
```json
{
  "error": "Missing X-API-Key header"
}
```

## 🚨 Error Handling

The API uses standard HTTP status codes:

- `200` - Success
- `400` - Bad Request (invalid parameters)
- `401` - Unauthorized (missing or invalid API key)
- `404` - Not Found (player not found)
- `500` - Internal Server Error

All error responses include a JSON object with an "error" field containing the error message.

## Security

### API Key Authentication
All API endpoints (except `/health`) require an `X-API-Key` header with a valid API key. Requests without a valid API key will receive a `401 Unauthorized` response.

### CORS
The API restricts CORS to only allow requests from `https://versusbot.netlify.app/`. Requests from other origins will be blocked by the browser's CORS policy.

## Example Usage

```bash
# Health check (no API key required)
curl http://localhost:8080/health

# Get all players (API key required)
curl -H "X-API-Key: your-secret-api-key-here" http://localhost:8080/api/players

# Get specific player (API key required)
curl -H "X-API-Key: your-secret-api-key-here" http://localhost:8080/api/players/123456789

# Get leaderboard (API key required)
curl -H "X-API-Key: your-secret-api-key-here" http://localhost:8080/api/leaderboard

# Get 1v1 leaderboard (API key required)
curl -H "X-API-Key: your-secret-api-key-here" http://localhost:8080/api/leaderboard/1v1

# Get 2v2 leaderboard (API key required)
curl -H "X-API-Key: your-secret-api-key-here" http://localhost:8080/api/leaderboard/2v2
```

## Frontend Integration

From your Netlify TypeScript app at `https://versusbot.netlify.app/`, you can fetch data like this:

**Important:** Store your API key as an environment variable in Netlify (Settings → Environment variables) and use it in your frontend code. Never commit the API key to your repository.

```typescript
const API_BASE_URL = 'http://localhost:8080'; // or your deployed API URL
const API_KEY = process.env.REACT_APP_API_KEY || process.env.NEXT_PUBLIC_API_KEY; // Adjust based on your framework

// Fetch all players
const players = await fetch(`${API_BASE_URL}/api/players`, {
  headers: {
    'X-API-Key': API_KEY
  }
})
  .then(res => {
    if (!res.ok) throw new Error('Failed to fetch players');
    return res.json();
  });

// Fetch leaderboard (defaults to 2v2)
const leaderboard = await fetch(`${API_BASE_URL}/api/leaderboard`, {
  headers: {
    'X-API-Key': API_KEY
  }
})
  .then(res => {
    if (!res.ok) throw new Error('Failed to fetch leaderboard');
    return res.json();
  });

// Fetch 1v1 leaderboard
const leaderboard1v1 = await fetch(`${API_BASE_URL}/api/leaderboard/1v1`, {
  headers: {
    'X-API-Key': API_KEY
  }
})
  .then(res => {
    if (!res.ok) throw new Error('Failed to fetch 1v1 leaderboard');
    return res.json();
  });

// Fetch 2v2 leaderboard
const leaderboard2v2 = await fetch(`${API_BASE_URL}/api/leaderboard/2v2`, {
  headers: {
    'X-API-Key': API_KEY
  }
})
  .then(res => {
    if (!res.ok) throw new Error('Failed to fetch 2v2 leaderboard');
    return res.json();
  });

// Fetch specific player
const player = await fetch(`${API_BASE_URL}/api/players/123456789`, {
  headers: {
    'X-API-Key': API_KEY
  }
})
  .then(res => {
    if (!res.ok) throw new Error('Failed to fetch player');
    return res.json();
  });
```

### Netlify Environment Variables Setup

1. Go to your Netlify site dashboard
2. Navigate to **Site settings** → **Environment variables**
3. Add a new variable:
   - **Key:** `REACT_APP_API_KEY` (or `NEXT_PUBLIC_API_KEY` for Next.js)
   - **Value:** Your API key (the same value you set in `API_KEY` on the server)
4. Redeploy your site for the changes to take effect

## 🚀 Deployment

### Building for Production

```bash
cd src
go build -o api-server cmd/api-server/main.go
```

### Running in Production

```bash
# Set environment variables
export POSTGRES_CONNECTION_STRING="your-production-connection-string"
export API_KEY="your-production-api-key"

# Run the server
./api-server -port 8080
```

### Docker Deployment

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN cd src && go build -o api-server cmd/api-server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/src/api-server .
EXPOSE 8080
CMD ["./api-server"]
```

## 🔧 Troubleshooting

### Common Issues

**API Key Authentication Failed**
- Ensure you're sending the `X-API-Key` header
- Check that the API key matches the `API_KEY` environment variable
- Verify the header is not being blocked by CORS

**CORS Errors**
- The API only accepts requests from `https://versusbot.netlify.app/`
- For development, you may need to temporarily modify the `allowedOrigin` in the code

**Database Connection Issues**
- Verify your `POSTGRES_CONNECTION_STRING` is correct
- Ensure the database tables exist (run the migration script)
- Check network connectivity to your PostgreSQL server

### Health Check

Always test with the health endpoint first:
```bash
curl http://localhost:8080/health
```

This endpoint doesn't require authentication and will confirm the API is running.

---

<div align="center">
  <p><strong>Part of the RL-Discord-Matchmaking system</strong></p>
  <p><a href="../README.md">← Back to main README</a></p>
</div>