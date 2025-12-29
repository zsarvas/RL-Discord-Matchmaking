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

The API requires two environment variables:
- `POSTGRES_CONNECTION_STRING` - Database connection string
- `API_KEY` - Secret key for API authentication

These can be loaded from `dev.env` (in the `src/` directory) or set directly in your environment.

### Option 1: Using dev.env file
Add both variables to your `dev.env` file:
```
POSTGRES_CONNECTION_STRING="host=localhost port=5432 user=vb_su password=YOUR_PASSWORD dbname=rl_db sslmode=disable"
API_KEY="your-secret-api-key-here"
```

The API will automatically look for `dev.env` in several locations:
- `../../dev.env` (when running from `src/cmd/api-server/`)
- `../dev.env` (when running from `src/`)
- `dev.env` (current directory)

### Option 2: Setting environment variables directly
If the `dev.env` file cannot be found, you can set the environment variables directly:

```bash
export POSTGRES_CONNECTION_STRING="host=localhost port=5432 user=vb_su password=YOUR_PASSWORD dbname=rl_db sslmode=disable"
export API_KEY="your-secret-api-key-here"
go run cmd/api-server/main.go
```

Or in a single command:
```bash
POSTGRES_CONNECTION_STRING="host=localhost port=5432 user=vb_su password=YOUR_PASSWORD dbname=rl_db sslmode=disable" API_KEY="your-secret-api-key-here" go run cmd/api-server/main.go
```

**Note:** The connection string should NOT include quotes when set as an environment variable. The code will automatically handle quotes if present in the env file.

**Security:** Choose a strong, random API key. You can generate one using:
```bash
openssl rand -hex 32
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

// Fetch leaderboard
const leaderboard = await fetch(`${API_BASE_URL}/api/leaderboard`, {
  headers: {
    'X-API-Key': API_KEY
  }
})
  .then(res => {
    if (!res.ok) throw new Error('Failed to fetch leaderboard');
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

