package main

import (
    "os"
    "fmt"
    "log"
    "flag"
    "bytes"
    "bufio"
    "encoding/base64"

    "golang.org/x/net/websocket"
)

// metadata header delimiter
var DELIMITER byte = '|'

var origin          = flag.String("origin", "http://localhost/", "origin")
var url             = flag.String("url", "", "url of websocket")
var metadata        = flag.String("metadata", "", "metadata")
var videoMagicBytes = flag.Bool("video-magic-bytes", false, "video magic bytes")

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

    header := "|{"

    if *metadata != "" {
        header += "\"metadata\":\"" + *metadata + "\""
    }

    if *videoMagicBytes {
        comma := ","
        if *metadata == "" {
            comma = ""
        }
        magicBytes := base64.StdEncoding.EncodeToString([]byte{0x6a, 0x73, 0x6d, 0x70, 0x01, 0x40, 0x00, 0xf0})
        header += comma + "\"magicBytes\": \"" + magicBytes + "\""
    }

    header += "}|"

    if header != "|{}|" {
        // write header
        if _, err = ws.Write([]byte(header)); err != nil {
            log.Fatal(err)
        }
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
