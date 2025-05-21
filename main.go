package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

type GatewayPayload struct {
	Op int         `json:"op"`
	D  interface{} `json:"d"`
}

type IdentifyData struct {
	Token      string                 `json:"token"`
	Intents    int                    `json:"intents"`
	Properties map[string]interface{} `json:"properties"`
}

type MessageCreate struct {
	Content string `json:"content"`
}

type DiscordEvent struct {
	T  string          `json:"t"`
	Op int             `json:"op"`
	D  json.RawMessage `json:"d"`
}

type Heartbeat struct {
	Op int         `json:"op"`
	D  interface{} `json:"d"`
}

type TokenStatus struct {
	Token     string
	Connected bool
	Error     string
}

var lastConnectionWebhook time.Time

func loadMainToken() string {
	_ = godotenv.Load()
	return os.Getenv("MAIN_TOKEN")
}

func loadAdditionalTokensAndWebhook() ([]string, string) {
	file, err := os.Open("tokens.txt")
	if err != nil {
		log.Fatalf("Failed to open tokens.txt: %v", err)
	}
	defer file.Close()

	var tokens []string
	var webhook string
	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			log.Fatalf("Error reading tokens.txt: %v", err)
		}
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			if err == io.EOF {
				break
			}
			continue
		}
		if strings.HasPrefix(line, "WEBHOOK_URL=") {
			webhook = strings.TrimPrefix(line, "WEBHOOK_URL=")
		} else {
			tokens = append(tokens, line)
		}
		if err == io.EOF {
			break
		}
	}
	return tokens, webhook
}

func sendWebhook(webhookURL, content string) {
	if webhookURL == "" {
		return
	}
	payload := map[string]interface{}{
		"content":    content,
		"username":   "Nitro Sniper",
		"avatar_url": "https://i.imgur.com/4M34hi2.png",
	}
	b, _ := json.Marshal(payload)
	_, err := http.Post(webhookURL, "application/json", strings.NewReader(string(b)))
	if err != nil {
		log.Printf("Failed to send webhook: %v", err)
	}
}

func extractNitroCode(content string) string {
	re := regexp.MustCompile(`discord\.gift\/(\w+)`)
	match := re.FindStringSubmatch(content)
	if len(match) > 1 {
		return match[1]
	}
	return ""
}

func claimNitro(token, code string, index int, isMain bool, wg *sync.WaitGroup, results chan<- map[string]interface{}) {
	defer wg.Done()
	url := fmt.Sprintf("https://discordapp.com/api/v9/entitlements/gift-codes/%s/redeem", code)

	// Create request body
	requestBody := map[string]interface{}{
		"channel_id": nil,
	}
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		results <- map[string]interface{}{"index": index, "isMain": isMain, "success": false, "message": err.Error()}
		return
	}

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		results <- map[string]interface{}{"index": index, "isMain": isMain, "success": false, "message": err.Error()}
		return
	}
	defer resp.Body.Close()
	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	if resp.StatusCode == 200 && body["message"] == nil {
		results <- map[string]interface{}{"index": index, "isMain": isMain, "success": true, "message": ""}
	} else {
		msg := "Unknown error"
		if m, ok := body["message"].(string); ok {
			msg = m
		}
		results <- map[string]interface{}{"index": index, "isMain": isMain, "success": false, "message": msg}
	}
}

func sendConnectionWebhook(webhookURL string, statuses []TokenStatus) {
	if webhookURL == "" {
		return
	}

	// Check if 24 hours have passed since last webhook
	if time.Since(lastConnectionWebhook) < 24*time.Hour {
		return
	}

	// Build status message
	var message strings.Builder
	message.WriteString("🔌 Connection Status Update\n\n")

	// Add main token status
	message.WriteString("**Main Token:**\n")
	if statuses[0].Connected {
		message.WriteString("✅ Connected\n\n")
	} else {
		message.WriteString(fmt.Sprintf("❌ Failed: %s\n\n", statuses[0].Error))
	}

	// Add additional tokens status
	if len(statuses) > 1 {
		message.WriteString("**Additional Tokens:**\n")
		for i, status := range statuses[1:] {
			if status.Connected {
				message.WriteString(fmt.Sprintf("✅ Token %d: Connected\n", i+1))
			} else {
				message.WriteString(fmt.Sprintf("❌ Token %d: %s\n", i+1, status.Error))
			}
		}
	}

	// Send webhook
	payload := map[string]interface{}{
		"content":    message.String(),
		"username":   "Nitro Sniper",
		"avatar_url": "https://i.imgur.com/4M34hi2.png",
	}
	b, _ := json.Marshal(payload)
	_, err := http.Post(webhookURL, "application/json", strings.NewReader(string(b)))
	if err != nil {
		log.Printf("Failed to send connection webhook: %v", err)
		return
	}

	// Update last webhook time
	lastConnectionWebhook = time.Now()
}

func main() {
	mainToken := loadMainToken()
	if mainToken == "" {
		log.Fatal("MAIN_TOKEN is not set in .env file")
	}
	additionalTokens, webhook := loadAdditionalTokensAndWebhook()
	allTokens := append([]string{mainToken}, additionalTokens...)
	log.Printf("Loaded %d tokens (1 main + %d additional)", len(allTokens), len(additionalTokens))
	if webhook != "" {
		log.Println("Webhook notifications enabled")
	} else {
		log.Println("Webhook notifications disabled (no webhook URL found)")
	}

	// Initialize connection statuses
	statuses := make([]TokenStatus, len(allTokens))
	for i, token := range allTokens {
		statuses[i] = TokenStatus{
			Token:     token,
			Connected: false,
			Error:     "",
		}
	}

	// Create a WaitGroup to wait for all connections
	var wg sync.WaitGroup
	connections := make([]*websocket.Conn, len(allTokens))

	// Connect all tokens
	for i, token := range allTokens {
		wg.Add(1)
		go func(index int, token string) {
			defer wg.Done()
			c, _, err := websocket.DefaultDialer.Dial("wss://gateway.discord.gg/?v=9&encoding=json", nil)
			if err != nil {
				statuses[index].Error = err.Error()
				log.Printf("Token %d failed to connect: %v", index+1, err)
				return
			}
			connections[index] = c
			statuses[index].Connected = true
			log.Printf("Token %d connected successfully", index+1)

			// Identify payload
			identify := GatewayPayload{
				Op: 2,
				D: IdentifyData{
					Token:   token,
					Intents: 513,
					Properties: map[string]interface{}{
						"os":      "windows",
						"browser": "chrome",
						"device":  "chrome",
					},
				},
			}
			c.WriteJSON(identify)

			// Start heartbeat for this connection
			go func(conn *websocket.Conn) {
				for {
					_, msg, err := conn.ReadMessage()
					if err != nil {
						log.Printf("Token %d WebSocket error: %v", index+1, err)
						statuses[index].Connected = false
						statuses[index].Error = err.Error()
						sendConnectionWebhook(webhook, statuses)
						return
					}
					var event DiscordEvent
					if err := json.Unmarshal(msg, &event); err != nil {
						continue
					}
					if event.Op == 10 {
						var d struct {
							HeartbeatInterval float64 `json:"heartbeat_interval"`
						}
						json.Unmarshal(event.D, &d)
						heartbeatInterval := time.Duration(d.HeartbeatInterval) * time.Millisecond
						ticker := time.NewTicker(heartbeatInterval)
						go func() {
							for range ticker.C {
								hb := Heartbeat{Op: 1, D: nil}
								conn.WriteJSON(hb)
							}
						}()
					}
					if event.T == "MESSAGE_CREATE" {
						var msgCreate MessageCreate
						json.Unmarshal(event.D, &msgCreate)
						code := extractNitroCode(msgCreate.Content)
						if code != "" {
							startTime := time.Now()
							log.Printf("Found Nitro gift: %s", code)
							var claimWg sync.WaitGroup
							results := make(chan map[string]interface{}, 1) // Only need space for one result

							// Only use main token to claim
							claimWg.Add(1)
							go claimNitro(mainToken, code, 1, true, &claimWg, results)

							go func() {
								claimWg.Wait()
								close(results)
							}()

							var successful []string
							var failMsg string
							for res := range results {
								if res["success"].(bool) {
									successful = append(successful, "Main Token")
									log.Printf("Main Token successfully claimed Nitro!")
								} else {
									log.Printf("Main Token failed to claim: %s", res["message"].(string))
									failMsg = res["message"].(string)
								}
							}
							snipeTime := time.Since(startTime)
							if len(successful) > 0 {
								msg := fmt.Sprintf("🎉 Successfully claimed Nitro gift!\nCode: `%s`\nClaimed with: Main Token\n⏱️ Snipe time: %s",
									code,
									snipeTime.Round(time.Millisecond))
								sendWebhook(webhook, msg)
							} else {
								msg := fmt.Sprintf("❌ Failed to claim Nitro gift\nCode: `%s`\nReason: %s\n⏱️ Attempt time: %s",
									code,
									failMsg,
									snipeTime.Round(time.Millisecond))
								sendWebhook(webhook, msg)
							}
						}
					}
				}
			}(c)
		}(i, token)
	}

	// Wait for all connections to be established
	wg.Wait()

	// Send startup notification
	if webhook != "" {
		startupMsg := fmt.Sprintf("🚀 Nitro Sniper Started\n\n"+
			"✅ Successfully connected to Discord\n"+
			"📊 Loaded %d tokens (1 main + %d additional)\n"+
			"⏰ Started at: %s",
			len(allTokens),
			len(additionalTokens),
			time.Now().Format("2006-01-02 15:04:05"))
		sendWebhook(webhook, startupMsg)
	}

	// Send initial connection status
	sendConnectionWebhook(webhook, statuses)

	// Keep the main goroutine alive
	select {}
}
