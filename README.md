# Versus Bot

<div align="center">
  <img src="src/logo.png" alt="RL-Discord-Matchmaking Logo" width="200"/>
</div>

<h3 align="center">Professional Matchmaking Bot</h3>

<p align="center">
  <a href="#features">Features</a> •
  <a href="#installation">Installation</a> •
  <a href="#setup">Setup</a> •
  <a href="#api">API</a> •
  <a href="#contributing">Contributing</a>
</p>

---

## Overview

A sophisticated Discord matchmaking bot written in Go, designed for Rocket League but easily adaptable for any competitive game requiring matchmaking ladders. This bot facilitates competitive 1v1 and 2v2 matches with automated matchmaking, MMR tracking, leaderboards, and real-time Discord integration.

## Features

### Bot Features
- **Dual Queue System**: Separate 1v1 and 2v2 queues
- **Automated Matchmaking**: Smart queue popping and team balancing
- **MMR System**: Elo-based rating system for fair matchmaking
- **Role Management**: Automatic king role assignment for leaders
- **Timeout System**: 20-minute queue timeouts with automatic removal
- **Real-time Updates**: Live queue status and match notifications

### Game Modes
- **1v1 Matches**: Direct player vs player competitions
- **2v2 Matches**: Team-based 4-player matches
- **Independent Queues**: Players can queue for multiple modes simultaneously

### Technical Features
- **REST API**: Secure API with authentication for frontend integration
- **PostgreSQL Database**: Separate tables for 1v1 and 2v2 MMR tracking
- **CORS Security**: Restricted to your Netlify frontend
- **Docker Support**: Containerized deployment
- **Comprehensive Logging**: Detailed bot activity tracking

## Installation

### Prerequisites
- Go 1.18 or higher
- PostgreSQL database
- Discord Bot Token

### Quick Start

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd RL-Discord-Matchmaking
   ```

2. **Install dependencies**
   ```bash
   cd src
   go mod tidy
   ```

3. **Set up the database**
   ```sql
   -- Run the migration script
   psql -h localhost -U your_user -d your_db -f ../migrations/create_tables.sql
   ```

4. **Configure environment variables**
   ```bash
   # Copy and edit the dev.env file
   cp dev.env dev.env.local
   # Edit dev.env.local with your values
   ```

5. **Run the bot**
   ```bash
   go run main.go
   ```

## Setup

### Environment Variables

Create a `dev.env` file in the `src/` directory:

```env
# Discord Bot Configuration
TOKEN=your-discord-bot-token

# Database Configuration
POSTGRES_CONNECTION_STRING=host=localhost port=5432 user=your_user password=your_password dbname=rl_db sslmode=disable

# API Configuration
API_KEY=your-secret-api-key-here

# Optional: Supabase (if using hosted database)
PUBLIC_SUPABASE_URL=https://your-project.supabase.co
PUBLIC_SUPABASE_ANON_KEY=your-supabase-anon-key
SUPABASE_CONNECTION_STRING=user=postgres.password=your_password host=aws-0-us-east-1.pooler.supabase.com port=5432 dbname=postgres
```

### Discord Bot Setup

1. Go to [Discord Developer Portal](https://discord.com/developers/applications)
2. Create a new application
3. Go to the "Bot" section
4. Copy the token and add it to your `dev.env`
5. Invite the bot to your server with appropriate permissions

### Database Schema

The bot uses separate tables for different game modes:

- `rocketleague_1v1`: 1v1 player statistics and MMR
- `rocketleague_2v2`: 2v2 player statistics and MMR

## Usage

### Bot Commands

| Command | Description |
|---------|-------------|
| `!q` | Join the queue for the current channel |
| `!leave` | Leave the queue |
| `!status` | View current queue status |
| `!report win` | Report a match victory |
| `!leaderboard` | View leaderboard link |
| `!help` | Display help menu |
| `!activematches` | View active matches |
| `!clear` | Clear queue (admin only) |

### Queue System

- **Channel-based**: Use `!q` in the appropriate channel for 1v1 or 2v2
- **Independent**: Players can be in both 1v1 and 2v2 queues simultaneously
- **Timeout**: 20-minute timeout with automatic removal
- **Match Creation**: Automatic team balancing and match announcements

## API

The bot includes a comprehensive REST API for frontend integration. See [API Documentation](src/api/README.md) for detailed information.

### Quick API Setup

```bash
cd src
go run cmd/api-server/main.go
```

### API Endpoints

- `GET /health` - Health check
- `GET /api/players` - Get all players
- `GET /api/players/{id}` - Get specific player
- `GET /api/leaderboard` - Get leaderboard
- `GET /api/leaderboard/1v1` - Get 1v1 leaderboard
- `GET /api/leaderboard/2v2` - Get 2v2 leaderboard

## Docker Deployment

### Build the image
```bash
docker build -t rl-matchmaking-bot .
```

### Run with environment variables
```bash
docker run -e TOKEN=your-token -e POSTGRES_CONNECTION_STRING=your-connection rl-matchmaking-bot
```

## Architecture

```
├── application/          # Business logic layer
│   ├── delegator.go     # Main bot logic and commands
│   ├── command.go       # Command constants
│   └── match.go         # Match structures
├── domain/              # Domain entities
│   ├── player.go        # Player model
│   └── queue.go         # Queue logic
├── infrastructure/      # Data access layer
│   ├── playerHandler.go # Database operations
│   └── matchHandler.go  # Match storage
├── interfaces/          # Repository interfaces
├── api/                 # REST API server
└── cmd/                 # Application entry points
```

## Security Features

- **API Key Authentication**: All API endpoints protected with X-API-Key
- **CORS Protection**: Restricted to your Netlify frontend
- **Input Validation**: Sanitized user inputs
- **SQL Injection Protection**: Parameterized queries
- **Rate Limiting**: Built-in request throttling

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines

- Follow Go conventions and best practices
- Add tests for new features
- Update documentation
- Ensure all tests pass

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built with [DiscordGo](https://github.com/bwmarrin/discordgo)
- Database powered by PostgreSQL
- Inspired by 6 Mans matchmaking systems
- Created in partnership with Ritter Gustave

---

<div align="center">
  <p><strong>Built with ❤️ for the Rocket League community</strong></p>
  <p>
    <a href="#rl-discord-matchmaking">Back to top</a>
  </p>
</div>
