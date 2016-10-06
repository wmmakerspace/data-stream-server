package main

import (
    "os"
    "fmt"
    "log"
    "flag"
    "bytes"
    "bufio"

    "golang.org/x/net/websocket"
)

// metadata header delimiter
var DELIMITER byte = '|'

var origin   = flag.String("origin", "http://localhost/", "origin")
var url      = flag.String("url", "", "url of websocket")
var metadata = flag.String("metadata", "", "metadata")

var BUFFER_LEN = 8

func main() {
    flag.Parse()

    if *url == "" {
        fmt.Println("no url provided")
        os.Exit(1)
    }

    ws, err := websocket.Dial(*url, "", *origin)
    if err != nil {
        log.Fatal(err)
    }

    // write header
    if _, err = ws.Write(append([]byte(*metadata), DELIMITER)); err != nil {
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
