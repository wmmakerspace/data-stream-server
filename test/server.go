package main

import (
    "fmt"
    "log"
    "flag"
    "net/http"

    "github.com/wmmakerspace/data-stream-server"
)

var port        = flag.String("port", "8080", "port")
var endpoint    = flag.String("endpoint", "data", "endpoint")

func main() {
    flag.Parse()

    streamserver.Start("/" + *endpoint)

    fmt.Println("Server listening on port :" + *port)
    fmt.Println("------------------------------")
    log.Fatal(http.ListenAndServe(":" + *port, nil))
}