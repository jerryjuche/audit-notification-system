package websocket

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "sync"
    "time"

    "github.com/gorilla/websocket"
)

// Upgrader configures the WebSocket upgrade process.
var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool {
        return true // Allow all origins for dev; restrict in prod.
    },
}

// Connections map: username -> WebSocket conn.
var (
    connections = make(map[string]*websocket.Conn)
    mu          sync.Mutex // Lock for thread-safe map access.
)

// Notification struct for JSON parsing.
type Notification struct {
    TargetUser string `json:"targetUser"`
    Message    string `json:"message"`
}

// EchoHandler upgrades HTTP to WS, stores connection, and handles messages (echo for testing).
func EchoHandler(w http.ResponseWriter, r *http.Request) {
    // For now, fake username from query param (e.g., ?user=jerry). Later, from OAuth.
    username := r.URL.Query().Get("user")
    if username == "" {
        http.Error(w, "Missing user", http.StatusBadRequest)
        return
    }

    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Printf("Upgrade error: %v", err)
        return
    }

    // Store connection.
    mu.Lock()
    connections[username] = conn
    mu.Unlock()

    defer func() {
        mu.Lock()
        delete(connections, username)
        mu.Unlock()
        conn.Close()
    }()

    log.Printf("User %s connected", username)

    // Set pong handler.
    conn.SetPongHandler(func(string) error {
        log.Printf("Pong from %s", username)
        return nil
    })

    // Ping every 30s to keep alive.
    go func() {
        ticker := time.NewTicker(30 * time.Second)
        defer ticker.Stop()
        for {
            <-ticker.C
            mu.Lock()
            if _, ok := connections[username]; !ok {
                mu.Unlock()
                return
            }
            mu.Unlock()
            if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
                return
            }
        }
    }()

    for {
        msgType, msg, err := conn.ReadMessage()
        if err != nil {
            log.Printf("Read error for %s: %v", username, err)
            return
        }
        // Echo back for testing.
        if err := conn.WriteMessage(msgType, msg); err != nil {
            log.Printf("Write error: %v", err)
            return
        }
    }
}

// NotifyHandler sends a notification to a specific user via POST.
func NotifyHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var notif Notification
    if err := json.NewDecoder(r.Body).Decode(&notif); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

    mu.Lock()
    conn, exists := connections[notif.TargetUser]
    mu.Unlock()

    if !exists {
        http.Error(w, "User not connected", http.StatusNotFound)
        return
    }

    // Send as text message.
    if err := conn.WriteMessage(websocket.TextMessage, []byte(notif.Message)); err != nil {
        log.Printf("Send error to %s: %v", notif.TargetUser, err)
        http.Error(w, "Send failed", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    fmt.Fprintln(w, "Notification sent")
}