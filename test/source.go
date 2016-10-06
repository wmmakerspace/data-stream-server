package main

import (
    "log"
    "flag"
    "golang.org/x/net/websocket"
)

var HEADER_DELIMITER = "|"

var host     = flag.String("host", "localhost", "host")
var port     = flag.String("port", "8080", "port")
var endpoint = flag.String("endpoint", "data", "endpoint")
var header   = flag.String("header", "{\"magicBytes\": \"hello\", \"metadata\": \"{}\"}", "header")

// stream data into the server
func main() {
    flag.Parse()

    if (*port != "") {
        *port = ":" + *port
    }

    url := "ws://" + *host + *port + "/" + *endpoint + "/in"

    ws, err := websocket.Dial(url, "", "http://localhost")
    if err != nil {
        log.Fatal(err)
    }

    if (*header != "") {
        // write header
        if _, err = ws.Write([]byte(HEADER_DELIMITER + *header + HEADER_DELIMITER)); err != nil {
            log.Fatal(err)
        }
    }

    log.Println("connected to server!")

    for {
        if _, err = ws.Write([]byte("1")); err != nil {
            log.Fatal(err)
        }
    }
}
