package main

import (
    "log"
    "time"

    "golang.org/x/net/websocket"
)

var origin = "http://localhost/"
var url = "ws://localhost:8080/video/in"

func main() {
    ws, err := websocket.Dial(url, "", origin)
    if err != nil {
        log.Fatal(err)
    }

    message := []byte("hello")

    for {
        _, err = ws.Write(message)
        if err != nil {
            log.Fatal(err)
        }
        time.Sleep(time.Millisecond * 3000)
    }
}
