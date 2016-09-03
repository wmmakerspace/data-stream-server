package main

import (
    "os"
    "log"
    "bufio"

    "golang.org/x/net/websocket"
)

var origin = "http://localhost/"
var url = "ws://localhost:8080/video/in"

func main() {
    ws, err := websocket.Dial(url, "", origin)
    if err != nil {
        log.Fatal(err)
    }

    s := bufio.NewScanner(os.Stdin)

    for s.Scan() {
        _, err = ws.Write([]byte(s.Bytes()))
        if err != nil {
            log.Fatal(err)
        }
    }
}
