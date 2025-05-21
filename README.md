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

1. Install Go (1.21 or higher)
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

## Running on a VPS

To run the Discord Nitro Sniper on a VPS, follow these steps:

1. **Set Up Your VPS**:
   - Choose a VPS provider (e.g., DigitalOcean, AWS, Linode) and create a new instance.
   - Ensure your VPS has a Linux-based OS (e.g., Ubuntu).

2. **Install Go**:
   - SSH into your VPS and install Go 1.21 or higher:
     ```bash
     wget https://golang.org/dl/go1.21.0.linux-amd64.tar.gz
     sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
     export PATH=$PATH:/usr/local/go/bin
     ```

3. **Clone the Repository**:
   - Clone the repository to your VPS:
     ```bash
     git clone https://github.com/zewj/discord-nitro-sniper.git
     cd discord-nitro-sniper
     ```

4. **Configure Environment**:
   - Create a `.env` file with your main token:
     ```
     MAIN_TOKEN=your_main_token_here
     ```
   - Create a `tokens.txt` file with your additional tokens and webhook URL:
     ```
     WEBHOOK_URL=your_webhook_url_here
     token1_here
     token2_here
     token3_here
     ```

5. **Run the Sniper**:
   - Execute the sniper:
     ```bash
     go run main.go
     ```

6. **(Recommended) Install screen or tmux**:
   - To keep the sniper running after you disconnect from SSH, you should install `screen` or `tmux`.

   **How to Install and Verify `screen` and `tmux`:**

   1. Update your package list:
      ```bash
      sudo apt update
      ```
   2. Install both screen and tmux:
      ```bash
      sudo apt install screen tmux
      ```
   3. Verify installation:
      ```bash
      screen --version
      tmux -V
      ```
      If you see version numbers, the installation was successful.
   4. Basic usage:
      - To start a new screen session:
        `screen -S nitro-sniper`
      - To start a new tmux session:
        `tmux new -s nitro-sniper`
      - To detach:
        - In screen: `Ctrl+A` then `D`
        - In tmux: `Ctrl+B` then `D`
      - To reattach:
        - In screen: `screen -r nitro-sniper`
        - In tmux: `tmux attach -t nitro-sniper`

7. **Keep the Sniper Running**:
   - Use a process manager like `screen` or `tmux` to keep the sniper running in the background:
     ```bash
     screen -S nitro-sniper
     go run main.go
     ```
   - Detach from the screen session with `Ctrl+A+D` and reattach with `screen -r nitro-sniper`.

## Security

- Never share your tokens or webhook URLs
- Keep your `.env` and `tokens.txt` files secure
- These files are automatically ignored by Git

## License

MIT License 