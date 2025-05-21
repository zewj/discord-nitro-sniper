# Discord Nitro Sniper

A high-performance Discord Nitro gift sniper written in Go. This tool monitors Discord messages for Nitro gift links and automatically claims them using your tokens.

## Features

- Written in Go for maximum performance
- Supports multiple tokens for monitoring
- Uses a main token for claiming gifts
- Discord webhook notifications
- Connection status monitoring
- Automatic reconnection handling

## Setup

1. Install Go (1.16 or higher)
2. Install required dependencies:
   ```bash
   go mod init discord-nitro-sniper
   go get github.com/gorilla/websocket
   go get github.com/joho/godotenv
   ```

3. Create a `.env` file with your main token:
   ```
   MAIN_TOKEN=your_main_token_here
   ```

4. Create a `tokens.txt` file with your additional tokens and webhook URL:
   ```
   WEBHOOK_URL=your_webhook_url_here
   token1_here
   token2_here
   token3_here
   ```

## Usage

Run the sniper:
```bash
go run main.go
```

## Security

- Never share your tokens or webhook URLs
- Keep your `.env` and `tokens.txt` files secure
- These files are automatically ignored by Git

## License

MIT License 