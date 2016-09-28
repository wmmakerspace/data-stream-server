package main

import (
    "fmt"
    "log"
    "flag"
    "net/http"

    "github.com/wmmakerspace/data-stream-server"
)

var port = flag.String("port", "8080", "port")
var endpoint = flag.String("endpoint", "data", "endpoint")
var videoHeader = flag.Bool("video-header", false, "header")

func main() {
    flag.Parse()

    var header []byte = nil

    if *videoHeader {
        header = []byte{0x6a, 0x73, 0x6d, 0x70, 0x01, 0x40, 0x00, 0xf0}
    }

    streamserver.Start("/" + *endpoint, header)

    fmt.Println("Server listening on port :" + *port)
    fmt.Println("------------------------------")
    log.Fatal(http.ListenAndServe(":" + *port, nil))
}