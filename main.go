package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gorilla/websocket"
)

type Message struct {
	Event string          `json:"event"`
	Data  json.RawMessage `json:"data"`
}

type ChannelResponse struct {
	ID       int `json:"id"`
	SlowMode struct {
		Enabled         bool `json:"enabled"`
		MessageInterval int  `json:"message_interval"`
	} `json:"slow_mode"`
	SubscribersMode struct {
		Enabled bool `json:"enabled"`
	} `json:"subscribers_mode"`
	FollowersMode struct {
		Enabled     bool `json:"enabled"`
		MinDuration int  `json:"min_duration"`
	} `json:"followers_mode"`
	EmotesMode struct {
		Enabled bool `json:"enabled"`
	} `json:"emotes_mode"`
	PinnedMessage any `json:"pinned_message"`
}

func GetChannelID(channelName string) (string, error) {

	url := fmt.Sprintf("https://kick.com/api/v2/channels/%s/chatroom", channelName)

	// Create a cookie jar to store cookies
	cookieJar, _ := cookiejar.New(nil)

	client := &http.Client{
		Jar: cookieJar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return "", err
	}

	req.Header.Set("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("accept-language", "en-US,en;q=0.9")
	req.Header.Set("sec-ch-ua", "\"Not.A/Brand\";v=\"8\", \"Chromium\";v=\"114\", \"Microsoft Edge\";v=\"114\"")
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", "\"Linux\"")
	req.Header.Set("sec-fetch-dest", "document")
	req.Header.Set("sec-fetch-mode", "navigate")
	req.Header.Set("sec-fetch-site", "none")
	req.Header.Set("sec-fetch-user", "?1")
	req.Header.Set("upgrade-insecure-requests", "1")
	req.Header.Set("referrerPolicy", "strict-origin-when-cross-origin")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return "", err
	}

	fmt.Println("Response body:", string(body))

	channelResponse := &ChannelResponse{}
	if err := json.Unmarshal(body, &channelResponse); err != nil {
		fmt.Println("Error parsing response body:", err)
		return "", err
	}
	return fmt.Sprintf("%d", channelResponse.ID), nil
}

func main() {
	url := "wss://ws-us2.pusher.com/app/eb1d5f283081a78b932c?protocol=7&client=js&version=7.6.0&flash=false"

	// Getting channel id from name in https://kick.com/api/v2/channels/%s/chatroom require javascript
	// Instead, we can get it from the website

	// channelID, err := GetChannelID("xposed")
	// if err != nil {
	// 	log.Fatal("Error getting channel ID:", err)
	// 	return
	// }
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

func handleChatMessageEvent(messageData string) (*MessageData, error) {

	original := messageData
	messageData = strings.TrimSpace(messageData)
	messageData = messageData[1 : len(messageData)-1]
	messageData = strings.ReplaceAll(messageData, "\\\"", "\"")
	messageData = strings.ReplaceAll(messageData, `\\"`, `'`)
	messageData = strings.ReplaceAll(messageData, `\",\"`, `","`)
	messageEvent := &MessageData{}
	err := json.Unmarshal([]byte(messageData), &messageEvent)
	if err != nil {
		fmt.Println("[Original]", original)
		fmt.Println("[Fixed]", messageData)
		fmt.Println("Error parsing message:", err)
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
	// Convert Unicode escape sequence to the actual character
	// by parsing the string as UTF-8
	fixed := make([]byte, 0, len(str))
	for len(str) > 0 {
		r, size := utf8.DecodeRuneInString(str)
		fixed = append(fixed, []byte(string(r))...)
		str = str[size:]
	}
	return string(fixed)
}
