package main

import (
    "log"
    "golang.org/x/net/websocket"
)

// stream data into the server
func main() {
    ws, err := websocket.Dial("ws://localhost:8080/data/in", "", "http://localhost")
    if err != nil {
        log.Fatal(err)
    }
    for {
        if _, err = ws.Write([]byte("1")); err != nil {
            log.Fatal(err)
        }
    }
}
