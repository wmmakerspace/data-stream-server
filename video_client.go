package main

import (
    "log"

    "golang.org/x/net/websocket"
)

var origin = "http://localhost/"
var url = "ws://localhost:8080/video/out"

func main() {
    ws, err := websocket.Dial(url, "", origin)
    if err != nil {
        log.Fatal(err)
    }

    var data []byte
    for {
        websocket.Message.Receive(ws, &data)
        log.Println(data)
    }
}
