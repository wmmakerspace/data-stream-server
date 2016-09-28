package main

import (
    "fmt"
    "log"
    "net/http"

    "github.com/wmmakerspace/data-stream-server"
)

var port = ":8080"

func main() {
    streamserver.Start("/data", nil)

    fmt.Println("Server listening on port " + port)
    fmt.Println("------------------------------")
    log.Fatal(http.ListenAndServe(port, nil))
}