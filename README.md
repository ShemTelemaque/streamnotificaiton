# Twitch Stream Notification Bot

A Go application that monitors Twitch streamers and sends notifications to Discord and Twitter when they go live.

## Features

- Monitor multiple Twitch streamers simultaneously
- Send notifications to Discord servers
- Post updates to Twitter
- PostgreSQL database for storing streamer and configuration data
- Web interface for managing monitored streamers and notification settings
- Live logging display on the web interface

## Project Structure

```
├── cmd/                  # Application entry points
│   └── server/           # Main server application
├── internal/             # Private application code
│   ├── api/              # API handlers
│   ├── config/           # Configuration management
│   ├── db/               # Database operations
│   ├── discord/          # Discord integration
│   ├── logger/           # Logging functionality
│   ├── models/           # Data models
│   ├── server/           # HTTP server implementation
│   ├── twitch/           # Twitch API integration
│   └── twitter/          # Twitter API integration
├── migrations/           # Database migrations
├── web/                  # Web interface assets
│   ├── static/           # Static files (CSS, JS, images)
│   └── templates/        # HTML templates
├── .env.example          # Example environment variables
├── go.mod                # Go module definition
├── go.sum                # Go module checksums
└── README.md             # Project documentation
```

## Setup

### Prerequisites

- Go 1.21 or higher
- PostgreSQL database
- Twitch Developer Account and API credentials
- Discord Bot Token
- Twitter Developer Account and API credentials

### Configuration

Copy the `.env.example` file to `.env` and fill in your configuration details:

```
# Server configuration
PORT=8080
ENVIRONMENT=development

# Database configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=streamnotification

# Twitch API configuration
TWITCH_CLIENT_ID=your_client_id
TWITCH_CLIENT_SECRET=your_client_secret

# Discord configuration
DISCORD_BOT_TOKEN=your_discord_bot_token

# Twitter configuration
TWITTER_API_KEY=your_twitter_api_key
TWITTER_API_SECRET=your_twitter_api_secret
TWITTER_ACCESS_TOKEN=your_twitter_access_token
TWITTER_ACCESS_SECRET=your_twitter_access_secret
```

### Running the Application

```bash
go run cmd/server/main.go
```

The web interface will be available at http://localhost:8080