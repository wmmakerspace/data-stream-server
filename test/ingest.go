package main

import (
    "log"
    "golang.org/x/net/websocket"
    "flag"
)

var host     = flag.String("host", "localhost", "host")
var endpoint = flag.String("endpoint", "data", "endpoint")
var port     = flag.String("port", "8080", "port")
var sourceId = flag.String("source-id", "1", "source id")

// ingest data from the server
func main() {
    flag.Parse()

    if (*port != "") {
        *port = ":" + *port
    }

    url := "ws://" + *host + *port + "/" + *endpoint + "/out/" + *sourceId

    ws, err := websocket.Dial(url, "", "http://localhost")
    if err != nil {
        log.Fatal(err)
    }

    // buffer for reading data
    data := make([]byte, 1024)

    for {
        if _, err = ws.Read(data); err != nil {
            log.Fatal(err)
        }
    }
}


