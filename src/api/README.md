# REST API Server

This REST API exposes the PostgreSQL database for reading player data. It's designed to be consumed by a Netlify Node.js TypeScript frontend.

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

## Environment Variables

The API requires the `POSTGRES_CONNECTION_STRING` environment variable to be set. This can be loaded from `dev.env` (in the `src/` directory) or set directly in your environment.

### Option 1: Using dev.env file
The API will automatically look for `dev.env` in several locations:
- `../../dev.env` (when running from `src/cmd/api-server/`)
- `../dev.env` (when running from `src/`)
- `dev.env` (current directory)

### Option 2: Setting environment variable directly
If the `dev.env` file cannot be found, you can set the environment variable directly:

```bash
export POSTGRES_CONNECTION_STRING="host=localhost port=5432 user=vb_su password=YOUR_PASSWORD dbname=rl_db sslmode=disable"
go run cmd/api-server/main.go
```

Or in a single command:
```bash
POSTGRES_CONNECTION_STRING="host=localhost port=5432 user=vb_su password=YOUR_PASSWORD dbname=rl_db sslmode=disable" go run cmd/api-server/main.go
```

**Note:** The connection string should NOT include quotes when set as an environment variable. The code will automatically handle quotes if present in the env file.

Example connection string format:
```
host=localhost port=5432 user=vb_su password=YOUR_PASSWORD dbname=rl_db sslmode=disable
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
- Returns: Array of all players sorted by MMR (descending)
- Same format as `/api/players`

## CORS

The API includes CORS middleware that allows requests from any origin (`Access-Control-Allow-Origin: *`). This is configured for development use with your Netlify frontend.

## Example Usage

```bash
# Health check
curl http://localhost:8080/health

# Get all players
curl http://localhost:8080/api/players

# Get specific player
curl http://localhost:8080/api/players/123456789

# Get leaderboard
curl http://localhost:8080/api/leaderboard
```

## Frontend Integration

From your Netlify TypeScript app, you can fetch data like this:

```typescript
const API_BASE_URL = 'http://localhost:8080'; // or your deployed API URL

// Fetch all players
const players = await fetch(`${API_BASE_URL}/api/players`)
  .then(res => res.json());

// Fetch leaderboard
const leaderboard = await fetch(`${API_BASE_URL}/api/leaderboard`)
  .then(res => res.json());

// Fetch specific player
const player = await fetch(`${API_BASE_URL}/api/players/123456789`)
  .then(res => res.json());
```

