package main

import (
    "fmt"
    "log"
    "net/http"

    ".."
)

var port = ":8080"

func main() {
    streamserver.Start("/video")

    fmt.Println("Server listening on port " + port)
    fmt.Println("------------------------------")
    log.Fatal(http.ListenAndServe(port, nil))
}