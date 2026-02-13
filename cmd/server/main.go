package main

import (
    "log"
    "net/http"
    "os"

    "github.com/jerryjuche/audit-notification-system/pkg/websocket" // Update to your module path.
)

func main() {
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080" // Default.
    }

    http.HandleFunc("/ws", websocket.EchoHandler)

    log.Printf("Starting server on :%s", port)
    if err := http.ListenAndServe(":"+port, nil); err != nil {
        log.Fatalf("Server error: %v", err)
    }
}