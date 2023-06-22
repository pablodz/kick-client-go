package main

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type Message struct {
	Event string          `json:"event"`
	Data  json.RawMessage `json:"data"`
}

func main() {
	url := "wss://ws-us2.pusher.com/app/eb1d5f283081a78b932c?protocol=7&client=js&version=7.6.0&flash=false"

	// Getting channel id from name in https://kick.com/api/v2/channels/%s/chatroom require javascript
	// Instead, we can get it from the website

	channelID := "75062"
	// Create a WebSocket connection
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatal("WebSocket connection error:", err)
	}

	// Function to handle received messages
	go func() {
		for {
			_, msgOrg, err := conn.ReadMessage()
			if err != nil {
				log.Println("WebSocket read error:", err)
				return
			}

			// Parse the received message
			msg := &Message{}
			if err := json.Unmarshal(msgOrg, &msg); err != nil {
				log.Println("Error parsing message:", err)
				continue
			}

			// fmt.Println("Received message:", string(msgOrg))

			// Handle the message based on the event
			switch msg.Event {
			case "pusher:connection_established":
				handleConnectionEstablished(conn, channelID)
				fmt.Println(">>> Subscription message sent")
			case "App\\Events\\ChatMessageEvent":
				handleChatMessageEvent(string(msg.Data))
			case "pusher_internal:subscription_succeeded":
				fmt.Println("<<< Subscription successful")
				continue
			default:
				fmt.Println("[EVENT] Received message:", string(msgOrg))
			}
		}
	}()

	// Keep the program running until interrupted
	<-make(chan struct{})
}

func handleConnectionEstablished(conn *websocket.Conn, channelID string) {
	event := &Message{
		Event: "pusher:subscribe",
		Data:  []byte(`{"auth":"","channel":"chatrooms.` + channelID + `.v2"}`),
	}

	b, err := json.Marshal(event)
	if err != nil {
		log.Println("Error marshalling subscription message:", err)
		conn.Close()
		return
	}

	if err := conn.WriteMessage(websocket.BinaryMessage, b); err != nil {
		log.Println("Error sending subscription message:", err)
		conn.Close()
		return
	}
}

type MessageData struct {
	ID         string    `json:"id"`
	ChatroomID int       `json:"chatroom_id"`
	Content    string    `json:"content"`
	Type       string    `json:"type"`
	CreatedAt  time.Time `json:"created_at"`
	Sender     struct {
		ID       int    `json:"id"`
		Username string `json:"username"`
		Slug     string `json:"slug"`
		Identity struct {
			Color  string `json:"color"`
			Badges []struct {
				Type  string `json:"type"`
				Text  string `json:"text"`
				Count int    `json:"count"`
			} `json:"badges"`
		} `json:"identity"`
	} `json:"sender"`
}

func handleChatMessageEvent(data string) (*MessageData, error) {

	original := data

	// First try to unmarshal the message
	data = strings.TrimSpace(original)
	data = data[1 : len(data)-1]
	data = fixString(data)
	dataAttempt1 := data
	dataAttempt2 := ""

	messageEvent := &MessageData{}
	err := json.Unmarshal([]byte(data), &messageEvent)
	if err != nil {
		data = original
		data = strings.TrimSpace(data)
		data = data[1 : len(data)-1]
		data = strings.ReplaceAll(data, `\\`, `\`)
		data = fixString(data)
		dataAttempt2 = data
	}

	err = json.Unmarshal([]byte(data), &messageEvent)
	if err != nil {
		fmt.Println("[Original]", original)
		fmt.Println("[Attempt1][Fixer]", dataAttempt1)
		fmt.Println("[Attempt2][Fixer]", dataAttempt2)
		return nil, err
	}

	// fmt.Println("Sender:", messageEvent.Sender.Username)
	messageEvent.Content = fixString(messageEvent.Content)
	fmt.Printf("%s: %s\n", messageEvent.Sender.Username, messageEvent.Content)
	// fmt.Println("Type:", messageEvent.Type)
	// fmt.Println("CreatedAt:", messageEvent.CreatedAt)
	// fmt.Println("Identity:", messageEvent.Sender.Identity)

	return messageEvent, nil

}

func fixString(str string) string {
	// Replace Unicode escape sequences with corresponding characters
	str = fixUnicodeEscapes(str)

	// Replace escaped double quotes with double quotes
	str = strings.ReplaceAll(str, "\\\"", "\"")

	return str
}

func fixUnicodeEscapes(str string) string {
	// Regular expression pattern to match Unicode escape sequences
	unicodePattern := regexp.MustCompile(`\\u([0-9a-fA-F]{4})`)

	// Replace Unicode escape sequences with corresponding characters
	fixedStr := unicodePattern.ReplaceAllStringFunc(str, func(unicodeMatch string) string {
		unicodeValue := unicodePattern.FindStringSubmatch(unicodeMatch)[1]
		unicodeInt, _ := strconv.ParseInt(unicodeValue, 16, 32)
		return string(rune(unicodeInt))
	})

	// Replace double backslashes with a single backslash if encountered as an error
	if strings.Contains(fixedStr, `\\`) {
		fixedStr = strings.ReplaceAll(fixedStr, `\\`, `\`)
	}

	return fixedStr
}
