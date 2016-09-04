package main

import (
    "os"
    "log"
    "bytes"
    "bufio"

    "golang.org/x/net/websocket"
)

var origin = "http://localhost/"
var url = "ws://localhost:8080/video/in"

var BUFFER_LEN = 8

func main() {
    ws, err := websocket.Dial(url, "", origin)
    if err != nil {
        log.Fatal(err)
    }

    s := bufio.NewScanner(os.Stdin)
    s.Split(bufio.ScanBytes)

    i := 0
    var buffer bytes.Buffer

    for s.Scan() {
        i++
        buffer.WriteByte(s.Bytes()[0])

        if i == BUFFER_LEN {
            _, err = ws.Write(buffer.Bytes())
            if err != nil {
                log.Fatal(err)
            }
            i = 0
            buffer.Reset()
        }
    }
}
