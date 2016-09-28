package main

import (
    "log"
    "golang.org/x/net/websocket"
)

// ingest data from the server
func main() {
    ws, err := websocket.Dial("ws://localhost:8080/data/out/1", "", "http://localhost")
    if err != nil {
        log.Fatal(err)
    }
    data := make([]byte, 1024)
    for {
        if _, err = ws.Read(data); err != nil {
            log.Fatal(err)
        }
    }
}


